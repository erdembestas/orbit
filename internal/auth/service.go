package auth

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"orbit/internal/secrets"
	"orbit/internal/store"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUnauthorized       = errors.New("unauthorized")
)

var defaultRoles = []string{"admin", "viewer", "approver", "operator"}

type Logger interface {
	Info(msg string, attrs ...any)
	Error(msg string, attrs ...any)
}

type UserStore interface {
	EnsureRoles(ctx context.Context, roles []string) error
	GetUserByUsername(ctx context.Context, username string) (store.User, error)
	CreateUser(ctx context.Context, params store.CreateUserParams) (store.User, error)
	AssignRoleToUser(ctx context.Context, userID, roleName string) error
	GetRolesForUser(ctx context.Context, userID string) ([]string, error)
	CreateSession(ctx context.Context, session store.Session) error
	GetSessionByTokenID(ctx context.Context, tokenID string) (store.Session, error)
	UpdateLastLogin(ctx context.Context, userID string, at time.Time) error
	CreateAuditEvent(ctx context.Context, event store.AuditEvent) error
}

type Service struct {
	store     UserStore
	bootstrap secrets.BootstrapConfig
	tokenTTL  time.Duration
	logger    Logger
	now       func() time.Time
	newID     func() string
}

type LoginResponse struct {
	AccessToken string `json:"accessToken"`
	TokenType   string `json:"tokenType"`
	ExpiresIn   int    `json:"expiresIn"`
}

type Principal struct {
	UserID   string
	Username string
	Roles    []string
	Source   string
}

type Claims struct {
	Username string   `json:"username"`
	Roles    []string `json:"roles"`
	Source   string   `json:"source"`
	jwt.RegisteredClaims
}

func NewService(userStore UserStore, bootstrap secrets.BootstrapConfig, tokenTTL time.Duration, logger Logger) *Service {
	return &Service{
		store:     userStore,
		bootstrap: bootstrap,
		tokenTTL:  tokenTTL,
		logger:    logger,
		now: func() time.Time {
			return time.Now().UTC()
		},
		newID: func() string {
			return uuid.NewString()
		},
	}
}

func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	return string(hash), nil
}

func VerifyPassword(hash, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

func (s *Service) EnsureBootstrapAdmin(ctx context.Context) error {
	if err := s.store.EnsureRoles(ctx, defaultRoles); err != nil {
		return err
	}
	s.logger.Info("default roles ensured", "roles", defaultRoles)

	user, err := s.store.GetUserByUsername(ctx, s.bootstrap.AdminUsername)
	if err == nil {
		s.logger.Info("bootstrap admin already exists", "username", user.Username)
		return nil
	}
	if !errors.Is(err, store.ErrNotFound) {
		return err
	}

	now := s.now()
	passwordHash, err := HashPassword(s.bootstrap.AdminPassword)
	if err != nil {
		return err
	}

	user, err = s.store.CreateUser(ctx, store.CreateUserParams{
		ID:           s.newID(),
		Username:     s.bootstrap.AdminUsername,
		PasswordHash: passwordHash,
		Source:       "local",
		Status:       "active",
		CreatedAt:    now,
		UpdatedAt:    now,
	})
	if err != nil {
		return err
	}

	if err := s.store.AssignRoleToUser(ctx, user.ID, "admin"); err != nil {
		return err
	}

	if err := s.store.CreateAuditEvent(ctx, store.AuditEvent{
		ID:        s.newID(),
		Username:  user.Username,
		EventType: "bootstrap_admin_created",
		Success:   true,
		Reason:    "created local bootstrap admin",
		CreatedAt: now,
	}); err != nil {
		return err
	}

	s.logger.Info("bootstrap admin created", "username", user.Username)
	return nil
}

func (s *Service) Login(ctx context.Context, username, password, remoteAddr, userAgent string) (LoginResponse, error) {
	user, err := s.store.GetUserByUsername(ctx, username)
	if err != nil {
		s.audit(ctx, store.AuditEvent{
			ID:         s.newID(),
			Username:   username,
			EventType:  "login",
			Success:    false,
			Reason:     "invalid_credentials",
			RemoteAddr: remoteAddr,
			UserAgent:  userAgent,
			CreatedAt:  s.now(),
		})
		if errors.Is(err, store.ErrNotFound) {
			return LoginResponse{}, ErrInvalidCredentials
		}
		return LoginResponse{}, err
	}

	if user.Status != "active" || VerifyPassword(user.PasswordHash, password) != nil {
		s.audit(ctx, store.AuditEvent{
			ID:         s.newID(),
			Username:   username,
			EventType:  "login",
			Success:    false,
			Reason:     "invalid_credentials",
			RemoteAddr: remoteAddr,
			UserAgent:  userAgent,
			CreatedAt:  s.now(),
		})
		return LoginResponse{}, ErrInvalidCredentials
	}

	roles, err := s.store.GetRolesForUser(ctx, user.ID)
	if err != nil {
		return LoginResponse{}, err
	}

	now := s.now()
	expiresAt := now.Add(s.tokenTTL)
	tokenID := s.newID()
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		Username: user.Username,
		Roles:    roles,
		Source:   user.Source,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   user.ID,
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(now),
			ID:        tokenID,
		},
	}).SignedString([]byte(s.bootstrap.JWTSecret))
	if err != nil {
		return LoginResponse{}, err
	}

	if err := s.store.CreateSession(ctx, store.Session{
		ID:        s.newID(),
		UserID:    user.ID,
		TokenID:   tokenID,
		ExpiresAt: expiresAt,
		CreatedAt: now,
	}); err != nil {
		return LoginResponse{}, err
	}

	if err := s.store.UpdateLastLogin(ctx, user.ID, now); err != nil {
		return LoginResponse{}, err
	}

	s.audit(ctx, store.AuditEvent{
		ID:         s.newID(),
		Username:   username,
		EventType:  "login",
		Success:    true,
		RemoteAddr: remoteAddr,
		UserAgent:  userAgent,
		CreatedAt:  now,
	})

	return LoginResponse{
		AccessToken: token,
		TokenType:   "Bearer",
		ExpiresIn:   int(s.tokenTTL.Seconds()),
	}, nil
}

func (s *Service) Authenticate(ctx context.Context, token, remoteAddr, userAgent string) (Principal, error) {
	claims := &Claims{}
	parsedToken, err := jwt.ParseWithClaims(token, claims, func(candidate *jwt.Token) (any, error) {
		if candidate.Method != jwt.SigningMethodHS256 {
			return nil, ErrUnauthorized
		}
		return []byte(s.bootstrap.JWTSecret), nil
	})
	if err != nil || !parsedToken.Valid {
		s.audit(ctx, store.AuditEvent{
			ID:         s.newID(),
			Username:   claims.Username,
			EventType:  "token_invalid",
			Success:    false,
			Reason:     "invalid_signature_or_expiry",
			RemoteAddr: remoteAddr,
			UserAgent:  userAgent,
			CreatedAt:  s.now(),
		})
		return Principal{}, ErrUnauthorized
	}

	session, err := s.store.GetSessionByTokenID(ctx, claims.ID)
	if err != nil {
		s.audit(ctx, store.AuditEvent{
			ID:         s.newID(),
			Username:   claims.Username,
			EventType:  "token_invalid",
			Success:    false,
			Reason:     "session_not_found",
			RemoteAddr: remoteAddr,
			UserAgent:  userAgent,
			CreatedAt:  s.now(),
		})
		if errors.Is(err, store.ErrNotFound) {
			return Principal{}, ErrUnauthorized
		}
		return Principal{}, err
	}

	if session.UserID != claims.Subject || session.RevokedAt.Valid || session.ExpiresAt.Before(s.now()) {
		s.audit(ctx, store.AuditEvent{
			ID:         s.newID(),
			Username:   claims.Username,
			EventType:  "token_invalid",
			Success:    false,
			Reason:     "session_invalid",
			RemoteAddr: remoteAddr,
			UserAgent:  userAgent,
			CreatedAt:  s.now(),
		})
		return Principal{}, ErrUnauthorized
	}

	return Principal{
		UserID:   claims.Subject,
		Username: claims.Username,
		Roles:    claims.Roles,
		Source:   claims.Source,
	}, nil
}

func (s *Service) audit(ctx context.Context, event store.AuditEvent) {
	if err := s.store.CreateAuditEvent(ctx, event); err != nil {
		s.logger.Error("audit event failed", "event_type", event.EventType, "error", err.Error())
	}
}

func nullTime(timeValue time.Time) sql.NullTime {
	if timeValue.IsZero() {
		return sql.NullTime{}
	}
	return sql.NullTime{Time: timeValue, Valid: true}
}
