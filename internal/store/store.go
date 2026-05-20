package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

var ErrNotFound = errors.New("not found")

type User struct {
	ID           string
	Username     string
	PasswordHash string
	Source       string
	Status       string
	CreatedAt    time.Time
	UpdatedAt    time.Time
	LastLoginAt  sql.NullTime
}

type CreateUserParams struct {
	ID           string
	Username     string
	PasswordHash string
	Source       string
	Status       string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type Session struct {
	ID        string
	UserID    string
	TokenID   string
	ExpiresAt time.Time
	RevokedAt sql.NullTime
	CreatedAt time.Time
}

type AuditEvent struct {
	ID         string
	Username   string
	EventType  string
	Success    bool
	Reason     string
	RemoteAddr string
	UserAgent  string
	CreatedAt  time.Time
}

type Cluster struct {
	ID         string       `json:"id"`
	Name       string       `json:"name"`
	Type       string       `json:"type"`
	Mode       string       `json:"mode"`
	Status     string       `json:"status"`
	CreatedAt  time.Time    `json:"created_at"`
	UpdatedAt  time.Time    `json:"updated_at"`
	LastSeenAt sql.NullTime `json:"last_seen_at"`
}

type KubernetesResource struct {
	ID              string            `json:"id"`
	ClusterID       string            `json:"cluster_id"`
	APIVersion      string            `json:"api_version"`
	Kind            string            `json:"kind"`
	Namespace       string            `json:"namespace,omitempty"`
	Name            string            `json:"name"`
	UID             string            `json:"uid"`
	ResourceVersion string            `json:"resource_version,omitempty"`
	Labels          map[string]string `json:"labels"`
	Annotations     map[string]string `json:"annotations"`
	Status          string            `json:"status,omitempty"`
	RawJSON         json.RawMessage   `json:"raw_json,omitempty"`
	ObservedAt      time.Time         `json:"observed_at"`
}

type UpsertKubernetesResourceParams struct {
	ID              string
	ClusterID       string
	APIVersion      string
	Kind            string
	Namespace       string
	Name            string
	UID             string
	ResourceVersion string
	Labels          map[string]string
	Annotations     map[string]string
	Status          string
	RawJSON         json.RawMessage
	ObservedAt      time.Time
}

type KubernetesEvent struct {
	ID           string       `json:"id"`
	ClusterID    string       `json:"cluster_id"`
	InvolvedUID  string       `json:"involved_uid,omitempty"`
	Namespace    string       `json:"namespace,omitempty"`
	InvolvedKind string       `json:"involved_kind,omitempty"`
	InvolvedName string       `json:"involved_name,omitempty"`
	Type         string       `json:"type,omitempty"`
	Reason       string       `json:"reason,omitempty"`
	Message      string       `json:"message,omitempty"`
	Count        int          `json:"count"`
	FirstSeenAt  sql.NullTime `json:"first_seen_at"`
	LastSeenAt   sql.NullTime `json:"last_seen_at"`
	ObservedAt   time.Time    `json:"observed_at"`
}

type UpsertKubernetesEventParams struct {
	ID           string
	ClusterID    string
	InvolvedUID  string
	Namespace    string
	InvolvedKind string
	InvolvedName string
	Type         string
	Reason       string
	Message      string
	Count        int
	FirstSeenAt  sql.NullTime
	LastSeenAt   sql.NullTime
	ObservedAt   time.Time
}

type ResourceMetric struct {
	ID            string    `json:"id"`
	ClusterID     string    `json:"cluster_id"`
	Namespace     string    `json:"namespace,omitempty"`
	PodName       string    `json:"pod_name,omitempty"`
	ContainerName string    `json:"container_name,omitempty"`
	CPUMillicores string    `json:"cpu_millicores,omitempty"`
	MemoryBytes   string    `json:"memory_bytes,omitempty"`
	ObservedAt    time.Time `json:"observed_at"`
}

type InsertResourceMetricParams struct {
	ID            string
	ClusterID     string
	Namespace     string
	PodName       string
	ContainerName string
	CPUMillicores string
	MemoryBytes   string
	ObservedAt    time.Time
}

type Finding struct {
	ID                string          `json:"id"`
	ClusterID         string          `json:"cluster_id"`
	ResourceID        string          `json:"resource_id,omitempty"`
	Severity          string          `json:"severity"`
	Category          string          `json:"category"`
	Title             string          `json:"title"`
	Description       string          `json:"description"`
	Status            string          `json:"status"`
	EvidenceJSON      json.RawMessage `json:"evidence_json"`
	CreatedAt         time.Time       `json:"created_at"`
	UpdatedAt         time.Time       `json:"updated_at"`
	ResolvedAt        sql.NullTime    `json:"resolved_at"`
	ResourceKind      string          `json:"resource_kind,omitempty"`
	ResourceNamespace string          `json:"resource_namespace,omitempty"`
	ResourceName      string          `json:"resource_name,omitempty"`
}

type UpsertFindingParams struct {
	ID           string
	ClusterID    string
	ResourceID   string
	Severity     string
	Category     string
	Title        string
	Description  string
	Status       string
	EvidenceJSON json.RawMessage
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type EvidencePack struct {
	ID             string          `json:"id"`
	FindingID      string          `json:"finding_id,omitempty"`
	ClusterID      string          `json:"cluster_id"`
	ResourceID     string          `json:"resource_id,omitempty"`
	ScopeType      string          `json:"scope_type"`
	ScopeNamespace string          `json:"scope_namespace,omitempty"`
	ScopeName      string          `json:"scope_name,omitempty"`
	TokenEstimate  int             `json:"token_estimate"`
	PackJSON       json.RawMessage `json:"pack_json"`
	CreatedAt      time.Time       `json:"created_at"`
}

type CreateEvidencePackParams struct {
	ID             string
	FindingID      string
	ClusterID      string
	ResourceID     string
	ScopeType      string
	ScopeNamespace string
	ScopeName      string
	TokenEstimate  int
	PackJSON       json.RawMessage
	CreatedAt      time.Time
}

type LLMRun struct {
	ID                  string          `json:"id"`
	FindingID           string          `json:"finding_id"`
	EvidencePackID      string          `json:"evidence_pack_id,omitempty"`
	Provider            string          `json:"provider"`
	Model               string          `json:"model"`
	PromptVersion       string          `json:"prompt_version"`
	InputTokenEstimate  int             `json:"input_token_estimate"`
	OutputTokenEstimate int             `json:"output_token_estimate"`
	OutputJSON          json.RawMessage `json:"output_json"`
	CreatedAt           time.Time       `json:"created_at"`
}

type CreateLLMRunParams struct {
	ID                  string
	FindingID           string
	EvidencePackID      string
	Provider            string
	Model               string
	PromptVersion       string
	InputTokenEstimate  int
	OutputTokenEstimate int
	OutputJSON          json.RawMessage
	CreatedAt           time.Time
}

type ActionPlan struct {
	ID             string          `json:"id"`
	FindingID      string          `json:"finding_id"`
	EvidencePackID string          `json:"evidence_pack_id,omitempty"`
	Title          string          `json:"title"`
	Summary        string          `json:"summary"`
	RiskLevel      string          `json:"risk_level"`
	Status         string          `json:"status"`
	PlanJSON       json.RawMessage `json:"plan_json"`
	CreatedAt      time.Time       `json:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at"`
}

type CreateActionPlanDraftParams struct {
	ID             string
	FindingID      string
	EvidencePackID string
	Title          string
	Summary        string
	RiskLevel      string
	Status         string
	PlanJSON       json.RawMessage
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type ListResourceFilters struct {
	Kind      string
	Namespace string
	Name      string
}

type ListFindingFilters struct {
	Severity  string
	Status    string
	Namespace string
	Kind      string
}

type ListEvidencePackFilters struct {
	ScopeType string
	Namespace string
	Name      string
	FindingID string
}

type ControllerStatus struct {
	ClusterName             string         `json:"cluster_name"`
	Mode                    string         `json:"mode"`
	LastSeenAt              sql.NullTime   `json:"last_seen_at"`
	ResourceCountsByKind    map[string]int `json:"resource_counts_by_kind"`
	OpenFindingCount        int            `json:"open_finding_count"`
	LatestEvidencePackCount int            `json:"latest_evidence_pack_count"`
}

type Store struct {
	db *sql.DB
}

func New(db *sql.DB) *Store {
	return &Store{db: db}
}

func (s *Store) EnsureSchema(ctx context.Context) error {
	const schemaLockKey int64 = 420420420

	if _, err := s.db.ExecContext(ctx, `SELECT pg_advisory_lock($1)`, schemaLockKey); err != nil {
		return err
	}
	defer func() {
		_, _ = s.db.ExecContext(context.Background(), `SELECT pg_advisory_unlock($1)`, schemaLockKey)
	}()

	const schema = `
CREATE TABLE IF NOT EXISTS users (
	id UUID PRIMARY KEY,
	username TEXT UNIQUE NOT NULL,
	password_hash TEXT NOT NULL,
	source TEXT NOT NULL DEFAULT 'local',
	status TEXT NOT NULL DEFAULT 'active',
	created_at TIMESTAMP NOT NULL,
	updated_at TIMESTAMP NOT NULL,
	last_login_at TIMESTAMP NULL
);

CREATE TABLE IF NOT EXISTS roles (
	id UUID PRIMARY KEY,
	name TEXT UNIQUE NOT NULL,
	created_at TIMESTAMP NOT NULL
);

CREATE TABLE IF NOT EXISTS user_roles (
	user_id UUID REFERENCES users(id) ON DELETE CASCADE,
	role_id UUID REFERENCES roles(id) ON DELETE CASCADE,
	PRIMARY KEY (user_id, role_id)
);

CREATE TABLE IF NOT EXISTS auth_audit_events (
	id UUID PRIMARY KEY,
	username TEXT,
	event_type TEXT NOT NULL,
	success BOOLEAN NOT NULL,
	reason TEXT,
	remote_addr TEXT,
	user_agent TEXT,
	created_at TIMESTAMP NOT NULL
);

CREATE TABLE IF NOT EXISTS sessions (
	id UUID PRIMARY KEY,
	user_id UUID REFERENCES users(id) ON DELETE CASCADE,
	token_id TEXT UNIQUE NOT NULL,
	expires_at TIMESTAMP NOT NULL,
	revoked_at TIMESTAMP NULL,
	created_at TIMESTAMP NOT NULL
);

CREATE TABLE IF NOT EXISTS clusters (
	id UUID PRIMARY KEY,
	name TEXT UNIQUE NOT NULL,
	type TEXT NOT NULL,
	mode TEXT NOT NULL DEFAULT 'single-cluster',
	status TEXT NOT NULL DEFAULT 'active',
	created_at TIMESTAMP NOT NULL,
	updated_at TIMESTAMP NOT NULL,
	last_seen_at TIMESTAMP NULL
);

CREATE TABLE IF NOT EXISTS kubernetes_resources (
	id UUID PRIMARY KEY,
	cluster_id UUID REFERENCES clusters(id) ON DELETE CASCADE,
	api_version TEXT NOT NULL,
	kind TEXT NOT NULL,
	namespace TEXT NULL,
	name TEXT NOT NULL,
	uid TEXT NOT NULL,
	resource_version TEXT NULL,
	labels_json JSONB NOT NULL DEFAULT '{}'::jsonb,
	annotations_json JSONB NOT NULL DEFAULT '{}'::jsonb,
	status TEXT NULL,
	raw_json JSONB NOT NULL,
	observed_at TIMESTAMP NOT NULL,
	UNIQUE(cluster_id, uid)
);

CREATE TABLE IF NOT EXISTS kubernetes_events (
	id UUID PRIMARY KEY,
	cluster_id UUID REFERENCES clusters(id) ON DELETE CASCADE,
	involved_uid TEXT NULL,
	namespace TEXT NULL,
	involved_kind TEXT NULL,
	involved_name TEXT NULL,
	type TEXT NULL,
	reason TEXT NULL,
	message TEXT NULL,
	count INTEGER,
	first_seen_at TIMESTAMP NULL,
	last_seen_at TIMESTAMP NULL,
	observed_at TIMESTAMP NOT NULL
);

CREATE TABLE IF NOT EXISTS resource_metrics (
	id UUID PRIMARY KEY,
	cluster_id UUID REFERENCES clusters(id) ON DELETE CASCADE,
	namespace TEXT NULL,
	pod_name TEXT NULL,
	container_name TEXT NULL,
	cpu_millicores NUMERIC NULL,
	memory_bytes NUMERIC NULL,
	observed_at TIMESTAMP NOT NULL
);

CREATE TABLE IF NOT EXISTS findings (
	id UUID PRIMARY KEY,
	cluster_id UUID REFERENCES clusters(id) ON DELETE CASCADE,
	resource_id UUID REFERENCES kubernetes_resources(id) ON DELETE SET NULL,
	severity TEXT NOT NULL,
	category TEXT NOT NULL,
	title TEXT NOT NULL,
	description TEXT NOT NULL,
	status TEXT NOT NULL DEFAULT 'open',
	evidence_json JSONB NOT NULL DEFAULT '{}'::jsonb,
	created_at TIMESTAMP NOT NULL,
	updated_at TIMESTAMP NOT NULL,
	resolved_at TIMESTAMP NULL
);

CREATE TABLE IF NOT EXISTS evidence_packs (
	id UUID PRIMARY KEY,
	finding_id UUID REFERENCES findings(id) ON DELETE CASCADE,
	cluster_id UUID REFERENCES clusters(id) ON DELETE CASCADE,
	resource_id UUID REFERENCES kubernetes_resources(id) ON DELETE SET NULL,
	scope_type TEXT NOT NULL DEFAULT 'finding',
	scope_namespace TEXT NULL,
	scope_name TEXT NULL,
	token_estimate INTEGER NOT NULL,
	pack_json JSONB NOT NULL,
	created_at TIMESTAMP NOT NULL
);

CREATE TABLE IF NOT EXISTS llm_runs (
	id UUID PRIMARY KEY,
	finding_id UUID REFERENCES findings(id) ON DELETE CASCADE,
	evidence_pack_id UUID REFERENCES evidence_packs(id) ON DELETE SET NULL,
	provider TEXT NOT NULL,
	model TEXT NOT NULL,
	prompt_version TEXT NOT NULL,
	input_token_estimate INTEGER NULL,
	output_token_estimate INTEGER NULL,
	output_json JSONB NOT NULL,
	created_at TIMESTAMP NOT NULL
);

CREATE TABLE IF NOT EXISTS action_plans (
	id UUID PRIMARY KEY,
	finding_id UUID REFERENCES findings(id) ON DELETE CASCADE,
	evidence_pack_id UUID REFERENCES evidence_packs(id) ON DELETE SET NULL,
	title TEXT NOT NULL,
	summary TEXT NOT NULL,
	risk_level TEXT NOT NULL,
	status TEXT NOT NULL DEFAULT 'draft',
	plan_json JSONB NOT NULL,
	created_at TIMESTAMP NOT NULL,
	updated_at TIMESTAMP NOT NULL
);
`

	if _, err := s.db.ExecContext(ctx, schema); err != nil {
		return err
	}

	migrations := []string{
		`ALTER TABLE evidence_packs ADD COLUMN IF NOT EXISTS scope_type TEXT NOT NULL DEFAULT 'finding'`,
		`ALTER TABLE evidence_packs ADD COLUMN IF NOT EXISTS scope_namespace TEXT NULL`,
		`ALTER TABLE evidence_packs ADD COLUMN IF NOT EXISTS scope_name TEXT NULL`,
		`UPDATE evidence_packs SET scope_type = 'finding' WHERE scope_type IS NULL OR scope_type = ''`,
	}
	for _, statement := range migrations {
		if _, err := s.db.ExecContext(ctx, statement); err != nil {
			return err
		}
	}

	return nil
}

func (s *Store) EnsureRoles(ctx context.Context, roles []string) error {
	now := time.Now().UTC()
	for _, roleName := range roles {
		_, err := s.db.ExecContext(
			ctx,
			`INSERT INTO roles (id, name, created_at) VALUES ($1, $2, $3) ON CONFLICT (name) DO NOTHING`,
			uuid.NewString(),
			roleName,
			now,
		)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *Store) GetUserByUsername(ctx context.Context, username string) (User, error) {
	var user User
	err := s.db.QueryRowContext(
		ctx,
		`SELECT id, username, password_hash, source, status, created_at, updated_at, last_login_at FROM users WHERE username = $1`,
		username,
	).Scan(
		&user.ID,
		&user.Username,
		&user.PasswordHash,
		&user.Source,
		&user.Status,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.LastLoginAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return User{}, ErrNotFound
	}

	return user, err
}

func (s *Store) CreateUser(ctx context.Context, params CreateUserParams) (User, error) {
	_, err := s.db.ExecContext(
		ctx,
		`INSERT INTO users (id, username, password_hash, source, status, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		params.ID,
		params.Username,
		params.PasswordHash,
		params.Source,
		params.Status,
		params.CreatedAt,
		params.UpdatedAt,
	)
	if err != nil {
		return User{}, err
	}

	return User{
		ID:           params.ID,
		Username:     params.Username,
		PasswordHash: params.PasswordHash,
		Source:       params.Source,
		Status:       params.Status,
		CreatedAt:    params.CreatedAt,
		UpdatedAt:    params.UpdatedAt,
	}, nil
}

func (s *Store) AssignRoleToUser(ctx context.Context, userID, roleName string) error {
	result, err := s.db.ExecContext(
		ctx,
		`INSERT INTO user_roles (user_id, role_id)
		 SELECT $1, id FROM roles WHERE name = $2
		 ON CONFLICT (user_id, role_id) DO NOTHING`,
		userID,
		roleName,
	)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		var exists bool
		err = s.db.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM roles WHERE name = $1)`, roleName).Scan(&exists)
		if err != nil {
			return err
		}
		if !exists {
			return ErrNotFound
		}
	}

	return nil
}

func (s *Store) GetRolesForUser(ctx context.Context, userID string) ([]string, error) {
	rows, err := s.db.QueryContext(
		ctx,
		`SELECT r.name
		 FROM roles r
		 JOIN user_roles ur ON ur.role_id = r.id
		 WHERE ur.user_id = $1
		 ORDER BY r.name`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles []string
	for rows.Next() {
		var roleName string
		if err := rows.Scan(&roleName); err != nil {
			return nil, err
		}
		roles = append(roles, roleName)
	}

	return roles, rows.Err()
}

func (s *Store) CreateSession(ctx context.Context, session Session) error {
	_, err := s.db.ExecContext(
		ctx,
		`INSERT INTO sessions (id, user_id, token_id, expires_at, revoked_at, created_at) VALUES ($1, $2, $3, $4, $5, $6)`,
		session.ID,
		session.UserID,
		session.TokenID,
		session.ExpiresAt,
		nullTime(session.RevokedAt),
		session.CreatedAt,
	)

	return err
}

func (s *Store) GetSessionByTokenID(ctx context.Context, tokenID string) (Session, error) {
	var session Session
	err := s.db.QueryRowContext(
		ctx,
		`SELECT id, user_id, token_id, expires_at, revoked_at, created_at FROM sessions WHERE token_id = $1`,
		tokenID,
	).Scan(
		&session.ID,
		&session.UserID,
		&session.TokenID,
		&session.ExpiresAt,
		&session.RevokedAt,
		&session.CreatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return Session{}, ErrNotFound
	}

	return session, err
}

func (s *Store) UpdateLastLogin(ctx context.Context, userID string, at time.Time) error {
	_, err := s.db.ExecContext(
		ctx,
		`UPDATE users SET last_login_at = $2, updated_at = $2 WHERE id = $1`,
		userID,
		at,
	)
	return err
}

func (s *Store) CreateAuditEvent(ctx context.Context, event AuditEvent) error {
	_, err := s.db.ExecContext(
		ctx,
		`INSERT INTO auth_audit_events (id, username, event_type, success, reason, remote_addr, user_agent, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		event.ID,
		nullString(event.Username),
		event.EventType,
		event.Success,
		nullString(event.Reason),
		nullString(event.RemoteAddr),
		nullString(event.UserAgent),
		event.CreatedAt,
	)
	return err
}

func (s *Store) EnsureCluster(ctx context.Context, name, clusterType, mode string) (Cluster, error) {
	now := time.Now().UTC()
	id := uuid.NewString()
	var cluster Cluster
	err := s.db.QueryRowContext(
		ctx,
		`INSERT INTO clusters (id, name, type, mode, status, created_at, updated_at, last_seen_at)
		 VALUES ($1, $2, $3, $4, 'active', $5, $5, $5)
		 ON CONFLICT (name) DO UPDATE SET
		   type = EXCLUDED.type,
		   mode = EXCLUDED.mode,
		   status = 'active',
		   updated_at = EXCLUDED.updated_at,
		   last_seen_at = EXCLUDED.last_seen_at
		 RETURNING id, name, type, mode, status, created_at, updated_at, last_seen_at`,
		id,
		name,
		clusterType,
		mode,
		now,
	).Scan(
		&cluster.ID,
		&cluster.Name,
		&cluster.Type,
		&cluster.Mode,
		&cluster.Status,
		&cluster.CreatedAt,
		&cluster.UpdatedAt,
		&cluster.LastSeenAt,
	)
	return cluster, err
}

func (s *Store) GetCluster(ctx context.Context, clusterID string) (Cluster, error) {
	var cluster Cluster
	err := s.db.QueryRowContext(
		ctx,
		`SELECT id, name, type, mode, status, created_at, updated_at, last_seen_at
		 FROM clusters
		 WHERE id = $1`,
		clusterID,
	).Scan(
		&cluster.ID,
		&cluster.Name,
		&cluster.Type,
		&cluster.Mode,
		&cluster.Status,
		&cluster.CreatedAt,
		&cluster.UpdatedAt,
		&cluster.LastSeenAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return Cluster{}, ErrNotFound
	}
	return cluster, err
}

func (s *Store) UpdateClusterLastSeen(ctx context.Context, clusterID string, at time.Time) error {
	_, err := s.db.ExecContext(
		ctx,
		`UPDATE clusters
		 SET updated_at = $2, last_seen_at = $2
		 WHERE id = $1`,
		clusterID,
		at,
	)
	return err
}

func (s *Store) UpsertKubernetesResource(ctx context.Context, params UpsertKubernetesResourceParams) (KubernetesResource, error) {
	if params.ID == "" {
		params.ID = uuid.NewString()
	}
	labelsJSON, err := marshalJSON(params.Labels)
	if err != nil {
		return KubernetesResource{}, err
	}
	annotationsJSON, err := marshalJSON(params.Annotations)
	if err != nil {
		return KubernetesResource{}, err
	}
	rawJSON, err := ensureJSONObject(params.RawJSON)
	if err != nil {
		return KubernetesResource{}, err
	}

	var resource KubernetesResource
	var labelsRaw, annotationsRaw, rawRaw []byte
	err = s.db.QueryRowContext(
		ctx,
		`INSERT INTO kubernetes_resources (
			id, cluster_id, api_version, kind, namespace, name, uid, resource_version,
			labels_json, annotations_json, status, raw_json, observed_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8,
			$9::jsonb, $10::jsonb, $11, $12::jsonb, $13
		)
		ON CONFLICT (cluster_id, uid) DO UPDATE SET
			api_version = EXCLUDED.api_version,
			kind = EXCLUDED.kind,
			namespace = EXCLUDED.namespace,
			name = EXCLUDED.name,
			resource_version = EXCLUDED.resource_version,
			labels_json = EXCLUDED.labels_json,
			annotations_json = EXCLUDED.annotations_json,
			status = EXCLUDED.status,
			raw_json = EXCLUDED.raw_json,
			observed_at = EXCLUDED.observed_at
		RETURNING id, cluster_id, api_version, kind, COALESCE(namespace, ''), name, uid, COALESCE(resource_version, ''), labels_json, annotations_json, COALESCE(status, ''), raw_json, observed_at`,
		params.ID,
		params.ClusterID,
		params.APIVersion,
		params.Kind,
		nullString(params.Namespace),
		params.Name,
		params.UID,
		nullString(params.ResourceVersion),
		labelsJSON,
		annotationsJSON,
		nullString(params.Status),
		string(rawJSON),
		params.ObservedAt,
	).Scan(
		&resource.ID,
		&resource.ClusterID,
		&resource.APIVersion,
		&resource.Kind,
		&resource.Namespace,
		&resource.Name,
		&resource.UID,
		&resource.ResourceVersion,
		&labelsRaw,
		&annotationsRaw,
		&resource.Status,
		&rawRaw,
		&resource.ObservedAt,
	)
	if err != nil {
		return KubernetesResource{}, err
	}

	resource.Labels = parseStringMap(labelsRaw)
	resource.Annotations = parseStringMap(annotationsRaw)
	resource.RawJSON = rawRaw
	return resource, nil
}

func (s *Store) UpsertKubernetesEvent(ctx context.Context, params UpsertKubernetesEventParams) (KubernetesEvent, error) {
	var existingID string
	err := s.db.QueryRowContext(
		ctx,
		`SELECT id FROM kubernetes_events
		 WHERE cluster_id = $1
		   AND COALESCE(involved_uid, '') = COALESCE($2, '')
		   AND COALESCE(namespace, '') = COALESCE($3, '')
		   AND COALESCE(involved_kind, '') = COALESCE($4, '')
		   AND COALESCE(involved_name, '') = COALESCE($5, '')
		   AND COALESCE(type, '') = COALESCE($6, '')
		   AND COALESCE(reason, '') = COALESCE($7, '')
		   AND COALESCE(message, '') = COALESCE($8, '')
		 ORDER BY observed_at DESC
		 LIMIT 1`,
		params.ClusterID,
		nullString(params.InvolvedUID),
		nullString(params.Namespace),
		nullString(params.InvolvedKind),
		nullString(params.InvolvedName),
		nullString(params.Type),
		nullString(params.Reason),
		nullString(params.Message),
	).Scan(&existingID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return KubernetesEvent{}, err
	}

	if errors.Is(err, sql.ErrNoRows) {
		if params.ID == "" {
			params.ID = uuid.NewString()
		}
		_, err = s.db.ExecContext(
			ctx,
			`INSERT INTO kubernetes_events (
				id, cluster_id, involved_uid, namespace, involved_kind, involved_name,
				type, reason, message, count, first_seen_at, last_seen_at, observed_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)`,
			params.ID,
			params.ClusterID,
			nullString(params.InvolvedUID),
			nullString(params.Namespace),
			nullString(params.InvolvedKind),
			nullString(params.InvolvedName),
			nullString(params.Type),
			nullString(params.Reason),
			nullString(params.Message),
			params.Count,
			nullTime(params.FirstSeenAt),
			nullTime(params.LastSeenAt),
			params.ObservedAt,
		)
		if err != nil {
			return KubernetesEvent{}, err
		}
		return KubernetesEvent{
			ID:           params.ID,
			ClusterID:    params.ClusterID,
			InvolvedUID:  params.InvolvedUID,
			Namespace:    params.Namespace,
			InvolvedKind: params.InvolvedKind,
			InvolvedName: params.InvolvedName,
			Type:         params.Type,
			Reason:       params.Reason,
			Message:      params.Message,
			Count:        params.Count,
			FirstSeenAt:  params.FirstSeenAt,
			LastSeenAt:   params.LastSeenAt,
			ObservedAt:   params.ObservedAt,
		}, nil
	}

	_, err = s.db.ExecContext(
		ctx,
		`UPDATE kubernetes_events
		 SET count = $2,
		     first_seen_at = $3,
		     last_seen_at = $4,
		     observed_at = $5
		 WHERE id = $1`,
		existingID,
		params.Count,
		nullTime(params.FirstSeenAt),
		nullTime(params.LastSeenAt),
		params.ObservedAt,
	)
	if err != nil {
		return KubernetesEvent{}, err
	}

	return s.getKubernetesEvent(ctx, existingID)
}

func (s *Store) InsertResourceMetric(ctx context.Context, params InsertResourceMetricParams) error {
	if params.ID == "" {
		params.ID = uuid.NewString()
	}
	_, err := s.db.ExecContext(
		ctx,
		`INSERT INTO resource_metrics (
			id, cluster_id, namespace, pod_name, container_name, cpu_millicores, memory_bytes, observed_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		params.ID,
		params.ClusterID,
		nullString(params.Namespace),
		nullString(params.PodName),
		nullString(params.ContainerName),
		nullNumericString(params.CPUMillicores),
		nullNumericString(params.MemoryBytes),
		params.ObservedAt,
	)
	return err
}

func (s *Store) UpsertFinding(ctx context.Context, params UpsertFindingParams) (Finding, error) {
	if params.Status == "" {
		params.Status = "open"
	}
	if params.ID == "" {
		params.ID = uuid.NewString()
	}
	evidenceJSON, err := ensureJSONObject(params.EvidenceJSON)
	if err != nil {
		return Finding{}, err
	}

	var existingID string
	err = s.db.QueryRowContext(
		ctx,
		`SELECT id FROM findings
		 WHERE cluster_id = $1
		   AND status = 'open'
		   AND category = $2
		   AND title = $3
		   AND ((resource_id = $4::uuid) OR (resource_id IS NULL AND $4::uuid IS NULL))
		 ORDER BY created_at DESC
		 LIMIT 1`,
		params.ClusterID,
		params.Category,
		params.Title,
		nullableUUID(params.ResourceID),
	).Scan(&existingID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return Finding{}, err
	}

	if errors.Is(err, sql.ErrNoRows) {
		_, err = s.db.ExecContext(
			ctx,
			`INSERT INTO findings (
				id, cluster_id, resource_id, severity, category, title, description, status,
				evidence_json, created_at, updated_at, resolved_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9::jsonb, $10, $11, NULL)`,
			params.ID,
			params.ClusterID,
			nullableUUID(params.ResourceID),
			params.Severity,
			params.Category,
			params.Title,
			params.Description,
			params.Status,
			string(evidenceJSON),
			params.CreatedAt,
			params.UpdatedAt,
		)
		if err != nil {
			return Finding{}, err
		}
		return s.GetFinding(ctx, params.ID)
	}

	_, err = s.db.ExecContext(
		ctx,
		`UPDATE findings
		 SET severity = $2,
		     description = $3,
		     evidence_json = $4::jsonb,
		     updated_at = $5,
		     status = $6,
		     resolved_at = NULL
		 WHERE id = $1`,
		existingID,
		params.Severity,
		params.Description,
		string(evidenceJSON),
		params.UpdatedAt,
		params.Status,
	)
	if err != nil {
		return Finding{}, err
	}
	return s.GetFinding(ctx, existingID)
}

func (s *Store) CreateEvidencePack(ctx context.Context, params CreateEvidencePackParams) (EvidencePack, error) {
	if params.ID == "" {
		params.ID = uuid.NewString()
	}
	if params.ScopeType == "" {
		params.ScopeType = "finding"
	}
	packJSON, err := ensureJSONObject(params.PackJSON)
	if err != nil {
		return EvidencePack{}, err
	}
	_, err = s.db.ExecContext(
		ctx,
		`INSERT INTO evidence_packs (id, finding_id, cluster_id, resource_id, scope_type, scope_namespace, scope_name, token_estimate, pack_json, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9::jsonb, $10)`,
		params.ID,
		nullableUUID(params.FindingID),
		params.ClusterID,
		nullableUUID(params.ResourceID),
		params.ScopeType,
		nullString(params.ScopeNamespace),
		nullString(params.ScopeName),
		params.TokenEstimate,
		string(packJSON),
		params.CreatedAt,
	)
	if err != nil {
		return EvidencePack{}, err
	}
	return s.GetEvidencePack(ctx, params.ID)
}

func (s *Store) CreateAdhocEvidencePack(ctx context.Context, params CreateEvidencePackParams) (EvidencePack, error) {
	if params.ScopeType == "" || params.ScopeType == "finding" {
		return EvidencePack{}, fmt.Errorf("adhoc evidence pack requires scope_type namespace or pod")
	}
	params.FindingID = ""
	return s.CreateEvidencePack(ctx, params)
}

func (s *Store) ListClusters(ctx context.Context) ([]Cluster, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id, name, type, mode, status, created_at, updated_at, last_seen_at FROM clusters ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var clusters []Cluster
	for rows.Next() {
		var cluster Cluster
		if err := rows.Scan(
			&cluster.ID,
			&cluster.Name,
			&cluster.Type,
			&cluster.Mode,
			&cluster.Status,
			&cluster.CreatedAt,
			&cluster.UpdatedAt,
			&cluster.LastSeenAt,
		); err != nil {
			return nil, err
		}
		clusters = append(clusters, cluster)
	}
	return clusters, rows.Err()
}

func (s *Store) ListKubernetesResources(ctx context.Context, filters ListResourceFilters) ([]KubernetesResource, error) {
	query := `SELECT id, cluster_id, api_version, kind, COALESCE(namespace, ''), name, uid, COALESCE(resource_version, ''), labels_json, annotations_json, COALESCE(status, ''), raw_json, observed_at FROM kubernetes_resources`
	var clauses []string
	var args []any
	if filters.Kind != "" {
		args = append(args, filters.Kind)
		clauses = append(clauses, fmt.Sprintf("kind = $%d", len(args)))
	}
	if filters.Namespace != "" {
		args = append(args, filters.Namespace)
		clauses = append(clauses, fmt.Sprintf("namespace = $%d", len(args)))
	}
	if filters.Name != "" {
		args = append(args, filters.Name)
		clauses = append(clauses, fmt.Sprintf("name = $%d", len(args)))
	}
	if len(clauses) > 0 {
		query += " WHERE " + strings.Join(clauses, " AND ")
	}
	query += " ORDER BY observed_at DESC, kind, namespace, name"

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var resources []KubernetesResource
	for rows.Next() {
		var resource KubernetesResource
		var labelsRaw, annotationsRaw, rawRaw []byte
		if err := rows.Scan(
			&resource.ID,
			&resource.ClusterID,
			&resource.APIVersion,
			&resource.Kind,
			&resource.Namespace,
			&resource.Name,
			&resource.UID,
			&resource.ResourceVersion,
			&labelsRaw,
			&annotationsRaw,
			&resource.Status,
			&rawRaw,
			&resource.ObservedAt,
		); err != nil {
			return nil, err
		}
		resource.Labels = parseStringMap(labelsRaw)
		resource.Annotations = parseStringMap(annotationsRaw)
		resource.RawJSON = rawRaw
		resources = append(resources, resource)
	}
	return resources, rows.Err()
}

func (s *Store) GetKubernetesResource(ctx context.Context, resourceID string) (KubernetesResource, error) {
	row := s.db.QueryRowContext(
		ctx,
		`SELECT id, cluster_id, api_version, kind, COALESCE(namespace, ''), name, uid, COALESCE(resource_version, ''), labels_json, annotations_json, COALESCE(status, ''), raw_json, observed_at
		 FROM kubernetes_resources
		 WHERE id = $1`,
		resourceID,
	)
	resource, err := scanKubernetesResource(row)
	if errors.Is(err, sql.ErrNoRows) {
		return KubernetesResource{}, ErrNotFound
	}
	return resource, err
}

func (s *Store) FindResourceByKindNamespaceName(ctx context.Context, clusterID, kind, namespace, name string) (KubernetesResource, error) {
	row := s.db.QueryRowContext(
		ctx,
		`SELECT id, cluster_id, api_version, kind, COALESCE(namespace, ''), name, uid, COALESCE(resource_version, ''), labels_json, annotations_json, COALESCE(status, ''), raw_json, observed_at
		 FROM kubernetes_resources
		 WHERE cluster_id = $1
		   AND kind = $2
		   AND COALESCE(namespace, '') = $3
		   AND name = $4
		 ORDER BY observed_at DESC
		 LIMIT 1`,
		clusterID,
		kind,
		namespace,
		name,
	)
	resource, err := scanKubernetesResource(row)
	if errors.Is(err, sql.ErrNoRows) {
		return KubernetesResource{}, ErrNotFound
	}
	return resource, err
}

func (s *Store) ListResourcesByNamespace(ctx context.Context, clusterID, namespace string) ([]KubernetesResource, error) {
	rows, err := s.db.QueryContext(
		ctx,
		`SELECT id, cluster_id, api_version, kind, COALESCE(namespace, ''), name, uid, COALESCE(resource_version, ''), labels_json, annotations_json, COALESCE(status, ''), raw_json, observed_at
		 FROM kubernetes_resources
		 WHERE cluster_id = $1
		   AND COALESCE(namespace, '') = $2
		 ORDER BY kind, name`,
		clusterID,
		namespace,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var resources []KubernetesResource
	for rows.Next() {
		resource, err := scanKubernetesResource(rows)
		if err != nil {
			return nil, err
		}
		resources = append(resources, resource)
	}
	return resources, rows.Err()
}

func (s *Store) ListFindings(ctx context.Context, filters ListFindingFilters) ([]Finding, error) {
	query := `SELECT f.id, f.cluster_id, COALESCE(f.resource_id::text, ''), f.severity, f.category, f.title, f.description, f.status, f.evidence_json, f.created_at, f.updated_at, f.resolved_at,
		COALESCE(r.kind, ''), COALESCE(r.namespace, ''), COALESCE(r.name, '')
		FROM findings f
		LEFT JOIN kubernetes_resources r ON r.id = f.resource_id`
	var clauses []string
	var args []any
	if filters.Severity != "" {
		args = append(args, filters.Severity)
		clauses = append(clauses, fmt.Sprintf("f.severity = $%d", len(args)))
	}
	if filters.Status != "" {
		args = append(args, filters.Status)
		clauses = append(clauses, fmt.Sprintf("f.status = $%d", len(args)))
	}
	if filters.Namespace != "" {
		args = append(args, filters.Namespace)
		clauses = append(clauses, fmt.Sprintf("COALESCE(r.namespace, '') = $%d", len(args)))
	}
	if filters.Kind != "" {
		args = append(args, filters.Kind)
		clauses = append(clauses, fmt.Sprintf("COALESCE(r.kind, '') = $%d", len(args)))
	}
	if len(clauses) > 0 {
		query += " WHERE " + strings.Join(clauses, " AND ")
	}
	query += " ORDER BY f.updated_at DESC, f.created_at DESC"

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var findings []Finding
	for rows.Next() {
		finding, err := scanFinding(rows)
		if err != nil {
			return nil, err
		}
		findings = append(findings, finding)
	}
	return findings, rows.Err()
}

func (s *Store) ListEventsByNamespace(ctx context.Context, clusterID, namespace string) ([]KubernetesEvent, error) {
	rows, err := s.db.QueryContext(
		ctx,
		`SELECT id, cluster_id, COALESCE(involved_uid, ''), COALESCE(namespace, ''), COALESCE(involved_kind, ''), COALESCE(involved_name, ''),
			COALESCE(type, ''), COALESCE(reason, ''), COALESCE(message, ''), COALESCE(count, 0), first_seen_at, last_seen_at, observed_at
		 FROM kubernetes_events
		 WHERE cluster_id = $1
		   AND COALESCE(namespace, '') = $2
		 ORDER BY COALESCE(last_seen_at, observed_at) DESC, observed_at DESC`,
		clusterID,
		namespace,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []KubernetesEvent
	for rows.Next() {
		event, err := scanKubernetesEvent(rows)
		if err != nil {
			return nil, err
		}
		events = append(events, event)
	}
	return events, rows.Err()
}

func (s *Store) ListEventsForResource(ctx context.Context, clusterID, resourceUID string) ([]KubernetesEvent, error) {
	rows, err := s.db.QueryContext(
		ctx,
		`SELECT id, cluster_id, COALESCE(involved_uid, ''), COALESCE(namespace, ''), COALESCE(involved_kind, ''), COALESCE(involved_name, ''),
			COALESCE(type, ''), COALESCE(reason, ''), COALESCE(message, ''), COALESCE(count, 0), first_seen_at, last_seen_at, observed_at
		 FROM kubernetes_events
		 WHERE cluster_id = $1
		   AND COALESCE(involved_uid, '') = $2
		 ORDER BY COALESCE(last_seen_at, observed_at) DESC, observed_at DESC`,
		clusterID,
		resourceUID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []KubernetesEvent
	for rows.Next() {
		event, err := scanKubernetesEvent(rows)
		if err != nil {
			return nil, err
		}
		events = append(events, event)
	}
	return events, rows.Err()
}

func (s *Store) ListLatestResourceMetrics(ctx context.Context, clusterID, namespace, podName string) ([]ResourceMetric, error) {
	query := `SELECT DISTINCT ON (pod_name, container_name)
		id, cluster_id, COALESCE(namespace, ''), COALESCE(pod_name, ''), COALESCE(container_name, ''),
		COALESCE(cpu_millicores::text, ''), COALESCE(memory_bytes::text, ''), observed_at
		FROM resource_metrics
		WHERE cluster_id = $1`
	args := []any{clusterID}
	if namespace != "" {
		args = append(args, namespace)
		query += fmt.Sprintf(" AND COALESCE(namespace, '') = $%d", len(args))
	}
	if podName != "" {
		args = append(args, podName)
		query += fmt.Sprintf(" AND COALESCE(pod_name, '') = $%d", len(args))
	}
	query += ` ORDER BY pod_name, container_name, observed_at DESC`
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var metrics []ResourceMetric
	for rows.Next() {
		var metric ResourceMetric
		if err := rows.Scan(
			&metric.ID,
			&metric.ClusterID,
			&metric.Namespace,
			&metric.PodName,
			&metric.ContainerName,
			&metric.CPUMillicores,
			&metric.MemoryBytes,
			&metric.ObservedAt,
		); err != nil {
			return nil, err
		}
		metrics = append(metrics, metric)
	}
	return metrics, rows.Err()
}

func (s *Store) GetFinding(ctx context.Context, findingID string) (Finding, error) {
	row := s.db.QueryRowContext(
		ctx,
		`SELECT f.id, f.cluster_id, COALESCE(f.resource_id::text, ''), f.severity, f.category, f.title, f.description, f.status, f.evidence_json, f.created_at, f.updated_at, f.resolved_at,
			COALESCE(r.kind, ''), COALESCE(r.namespace, ''), COALESCE(r.name, '')
		 FROM findings f
		 LEFT JOIN kubernetes_resources r ON r.id = f.resource_id
		 WHERE f.id = $1`,
		findingID,
	)
	finding, err := scanFinding(row)
	if errors.Is(err, sql.ErrNoRows) {
		return Finding{}, ErrNotFound
	}
	return finding, err
}

func (s *Store) ListFindingsByNamespace(ctx context.Context, clusterID, namespace string) ([]Finding, error) {
	rows, err := s.db.QueryContext(
		ctx,
		`SELECT f.id, f.cluster_id, COALESCE(f.resource_id::text, ''), f.severity, f.category, f.title, f.description, f.status, f.evidence_json, f.created_at, f.updated_at, f.resolved_at,
			COALESCE(r.kind, ''), COALESCE(r.namespace, ''), COALESCE(r.name, '')
		 FROM findings f
		 LEFT JOIN kubernetes_resources r ON r.id = f.resource_id
		 WHERE f.cluster_id = $1
		   AND COALESCE(r.namespace, '') = $2
		 ORDER BY f.updated_at DESC, f.created_at DESC`,
		clusterID,
		namespace,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var findings []Finding
	for rows.Next() {
		finding, err := scanFinding(rows)
		if err != nil {
			return nil, err
		}
		findings = append(findings, finding)
	}
	return findings, rows.Err()
}

func (s *Store) ListFindingsForResource(ctx context.Context, resourceID string) ([]Finding, error) {
	rows, err := s.db.QueryContext(
		ctx,
		`SELECT f.id, f.cluster_id, COALESCE(f.resource_id::text, ''), f.severity, f.category, f.title, f.description, f.status, f.evidence_json, f.created_at, f.updated_at, f.resolved_at,
			COALESCE(r.kind, ''), COALESCE(r.namespace, ''), COALESCE(r.name, '')
		 FROM findings f
		 LEFT JOIN kubernetes_resources r ON r.id = f.resource_id
		 WHERE COALESCE(f.resource_id::text, '') = $1
		 ORDER BY f.updated_at DESC, f.created_at DESC`,
		resourceID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var findings []Finding
	for rows.Next() {
		finding, err := scanFinding(rows)
		if err != nil {
			return nil, err
		}
		findings = append(findings, finding)
	}
	return findings, rows.Err()
}

func (s *Store) GetEvidencePackByFindingID(ctx context.Context, findingID string) (EvidencePack, error) {
	row := s.db.QueryRowContext(
		ctx,
		`SELECT id, COALESCE(finding_id::text, ''), cluster_id, COALESCE(resource_id::text, ''), COALESCE(scope_type, 'finding'), COALESCE(scope_namespace, ''), COALESCE(scope_name, ''), token_estimate, pack_json, created_at
		 FROM evidence_packs
		 WHERE finding_id = $1
		 ORDER BY created_at DESC
		 LIMIT 1`,
		findingID,
	)
	pack, err := scanEvidencePack(row)
	if errors.Is(err, sql.ErrNoRows) {
		return EvidencePack{}, ErrNotFound
	}
	return pack, err
}

func (s *Store) GetEvidencePack(ctx context.Context, evidencePackID string) (EvidencePack, error) {
	row := s.db.QueryRowContext(
		ctx,
		`SELECT id, COALESCE(finding_id::text, ''), cluster_id, COALESCE(resource_id::text, ''), COALESCE(scope_type, 'finding'), COALESCE(scope_namespace, ''), COALESCE(scope_name, ''), token_estimate, pack_json, created_at
		 FROM evidence_packs
		 WHERE id = $1`,
		evidencePackID,
	)
	pack, err := scanEvidencePack(row)
	if errors.Is(err, sql.ErrNoRows) {
		return EvidencePack{}, ErrNotFound
	}
	return pack, err
}

func (s *Store) ListEvidencePacks(ctx context.Context, filters ListEvidencePackFilters) ([]EvidencePack, error) {
	query := `SELECT id, COALESCE(finding_id::text, ''), cluster_id, COALESCE(resource_id::text, ''), COALESCE(scope_type, 'finding'), COALESCE(scope_namespace, ''), COALESCE(scope_name, ''), token_estimate, pack_json, created_at FROM evidence_packs`
	var clauses []string
	var args []any
	if filters.FindingID != "" {
		args = append(args, filters.FindingID)
		clauses = append(clauses, fmt.Sprintf("COALESCE(finding_id::text, '') = $%d", len(args)))
	}
	if filters.ScopeType != "" {
		args = append(args, filters.ScopeType)
		clauses = append(clauses, fmt.Sprintf("COALESCE(scope_type, 'finding') = $%d", len(args)))
	}
	if filters.Namespace != "" {
		args = append(args, filters.Namespace)
		clauses = append(clauses, fmt.Sprintf("COALESCE(scope_namespace, '') = $%d", len(args)))
	}
	if filters.Name != "" {
		args = append(args, filters.Name)
		clauses = append(clauses, fmt.Sprintf("COALESCE(scope_name, '') = $%d", len(args)))
	}
	if len(clauses) > 0 {
		query += " WHERE " + strings.Join(clauses, " AND ")
	}
	query += ` ORDER BY created_at DESC`
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var packs []EvidencePack
	for rows.Next() {
		pack, err := scanEvidencePack(rows)
		if err != nil {
			return nil, err
		}
		packs = append(packs, pack)
	}
	return packs, rows.Err()
}

func (s *Store) CreateLLMRun(ctx context.Context, params CreateLLMRunParams) (LLMRun, error) {
	if params.ID == "" {
		params.ID = uuid.NewString()
	}
	outputJSON, err := ensureJSONObject(params.OutputJSON)
	if err != nil {
		return LLMRun{}, err
	}
	_, err = s.db.ExecContext(
		ctx,
		`INSERT INTO llm_runs (id, finding_id, evidence_pack_id, provider, model, prompt_version, input_token_estimate, output_token_estimate, output_json, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9::jsonb, $10)`,
		params.ID,
		nullableUUID(params.FindingID),
		nullableUUID(params.EvidencePackID),
		params.Provider,
		params.Model,
		params.PromptVersion,
		nullInt(params.InputTokenEstimate),
		nullInt(params.OutputTokenEstimate),
		string(outputJSON),
		params.CreatedAt,
	)
	if err != nil {
		return LLMRun{}, err
	}
	return s.getLLMRun(ctx, params.ID)
}

func (s *Store) CreateActionPlanDraft(ctx context.Context, params CreateActionPlanDraftParams) (ActionPlan, error) {
	if params.ID == "" {
		params.ID = uuid.NewString()
	}
	if params.Status == "" {
		params.Status = "draft"
	}
	planJSON, err := ensureJSONObject(params.PlanJSON)
	if err != nil {
		return ActionPlan{}, err
	}
	_, err = s.db.ExecContext(
		ctx,
		`INSERT INTO action_plans (id, finding_id, evidence_pack_id, title, summary, risk_level, status, plan_json, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8::jsonb, $9, $10)`,
		params.ID,
		nullableUUID(params.FindingID),
		nullableUUID(params.EvidencePackID),
		params.Title,
		params.Summary,
		params.RiskLevel,
		params.Status,
		string(planJSON),
		params.CreatedAt,
		params.UpdatedAt,
	)
	if err != nil {
		return ActionPlan{}, err
	}
	return s.GetActionPlan(ctx, params.ID)
}

func (s *Store) ListActionPlans(ctx context.Context) ([]ActionPlan, error) {
	rows, err := s.db.QueryContext(
		ctx,
		`SELECT id, COALESCE(finding_id::text, ''), COALESCE(evidence_pack_id::text, ''), title, summary, risk_level, status, plan_json, created_at, updated_at
		 FROM action_plans
		 ORDER BY updated_at DESC, created_at DESC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var plans []ActionPlan
	for rows.Next() {
		plan, err := scanActionPlan(rows)
		if err != nil {
			return nil, err
		}
		plans = append(plans, plan)
	}
	return plans, rows.Err()
}

func (s *Store) GetActionPlan(ctx context.Context, actionPlanID string) (ActionPlan, error) {
	row := s.db.QueryRowContext(
		ctx,
		`SELECT id, COALESCE(finding_id::text, ''), COALESCE(evidence_pack_id::text, ''), title, summary, risk_level, status, plan_json, created_at, updated_at
		 FROM action_plans
		 WHERE id = $1`,
		actionPlanID,
	)
	plan, err := scanActionPlan(row)
	if errors.Is(err, sql.ErrNoRows) {
		return ActionPlan{}, ErrNotFound
	}
	return plan, err
}

func (s *Store) ControllerStatus(ctx context.Context) (ControllerStatus, error) {
	clusters, err := s.ListClusters(ctx)
	if err != nil {
		return ControllerStatus{}, err
	}
	if len(clusters) == 0 {
		return ControllerStatus{}, ErrNotFound
	}
	cluster := clusters[0]

	rows, err := s.db.QueryContext(
		ctx,
		`SELECT kind, COUNT(*)
		 FROM kubernetes_resources
		 WHERE cluster_id = $1
		 GROUP BY kind
		 ORDER BY kind`,
		cluster.ID,
	)
	if err != nil {
		return ControllerStatus{}, err
	}
	defer rows.Close()

	counts := map[string]int{}
	for rows.Next() {
		var kind string
		var count int
		if err := rows.Scan(&kind, &count); err != nil {
			return ControllerStatus{}, err
		}
		counts[kind] = count
	}
	if err := rows.Err(); err != nil {
		return ControllerStatus{}, err
	}

	var openFindings int
	if err := s.db.QueryRowContext(
		ctx,
		`SELECT COUNT(*)
		 FROM findings
		 WHERE cluster_id = $1
		   AND status = 'open'`,
		cluster.ID,
	).Scan(&openFindings); err != nil {
		return ControllerStatus{}, err
	}

	var evidenceCount int
	if err := s.db.QueryRowContext(
		ctx,
		`SELECT COUNT(*)
		 FROM evidence_packs
		 WHERE cluster_id = $1`,
		cluster.ID,
	).Scan(&evidenceCount); err != nil {
		return ControllerStatus{}, err
	}

	return ControllerStatus{
		ClusterName:             cluster.Name,
		Mode:                    cluster.Mode,
		LastSeenAt:              cluster.LastSeenAt,
		ResourceCountsByKind:    counts,
		OpenFindingCount:        openFindings,
		LatestEvidencePackCount: evidenceCount,
	}, nil
}

func (s *Store) getKubernetesEvent(ctx context.Context, eventID string) (KubernetesEvent, error) {
	row := s.db.QueryRowContext(
		ctx,
		`SELECT id, cluster_id, COALESCE(involved_uid, ''), COALESCE(namespace, ''), COALESCE(involved_kind, ''), COALESCE(involved_name, ''),
			COALESCE(type, ''), COALESCE(reason, ''), COALESCE(message, ''), COALESCE(count, 0), first_seen_at, last_seen_at, observed_at
		 FROM kubernetes_events WHERE id = $1`,
		eventID,
	)
	event, err := scanKubernetesEvent(row)
	if errors.Is(err, sql.ErrNoRows) {
		return KubernetesEvent{}, ErrNotFound
	}
	return event, err
}

func (s *Store) getEvidencePackByID(ctx context.Context, packID string) (EvidencePack, error) {
	return s.GetEvidencePack(ctx, packID)
}

func (s *Store) getLLMRun(ctx context.Context, llmRunID string) (LLMRun, error) {
	var run LLMRun
	var outputRaw []byte
	err := s.db.QueryRowContext(
		ctx,
		`SELECT id, COALESCE(finding_id::text, ''), COALESCE(evidence_pack_id::text, ''), provider, model, prompt_version,
			COALESCE(input_token_estimate, 0), COALESCE(output_token_estimate, 0), output_json, created_at
		 FROM llm_runs WHERE id = $1`,
		llmRunID,
	).Scan(
		&run.ID,
		&run.FindingID,
		&run.EvidencePackID,
		&run.Provider,
		&run.Model,
		&run.PromptVersion,
		&run.InputTokenEstimate,
		&run.OutputTokenEstimate,
		&outputRaw,
		&run.CreatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return LLMRun{}, ErrNotFound
	}
	run.OutputJSON = outputRaw
	return run, err
}

type scanner interface {
	Scan(dest ...any) error
}

func scanFinding(row scanner) (Finding, error) {
	var finding Finding
	var evidenceRaw []byte
	err := row.Scan(
		&finding.ID,
		&finding.ClusterID,
		&finding.ResourceID,
		&finding.Severity,
		&finding.Category,
		&finding.Title,
		&finding.Description,
		&finding.Status,
		&evidenceRaw,
		&finding.CreatedAt,
		&finding.UpdatedAt,
		&finding.ResolvedAt,
		&finding.ResourceKind,
		&finding.ResourceNamespace,
		&finding.ResourceName,
	)
	if err != nil {
		return Finding{}, err
	}
	finding.EvidenceJSON = evidenceRaw
	return finding, nil
}

func scanEvidencePack(row scanner) (EvidencePack, error) {
	var pack EvidencePack
	var packRaw []byte
	err := row.Scan(
		&pack.ID,
		&pack.FindingID,
		&pack.ClusterID,
		&pack.ResourceID,
		&pack.ScopeType,
		&pack.ScopeNamespace,
		&pack.ScopeName,
		&pack.TokenEstimate,
		&packRaw,
		&pack.CreatedAt,
	)
	if err != nil {
		return EvidencePack{}, err
	}
	pack.PackJSON = packRaw
	return pack, nil
}

func scanKubernetesResource(row scanner) (KubernetesResource, error) {
	var resource KubernetesResource
	var labelsRaw, annotationsRaw, rawRaw []byte
	err := row.Scan(
		&resource.ID,
		&resource.ClusterID,
		&resource.APIVersion,
		&resource.Kind,
		&resource.Namespace,
		&resource.Name,
		&resource.UID,
		&resource.ResourceVersion,
		&labelsRaw,
		&annotationsRaw,
		&resource.Status,
		&rawRaw,
		&resource.ObservedAt,
	)
	if err != nil {
		return KubernetesResource{}, err
	}
	resource.Labels = parseStringMap(labelsRaw)
	resource.Annotations = parseStringMap(annotationsRaw)
	resource.RawJSON = rawRaw
	return resource, nil
}

func scanKubernetesEvent(row scanner) (KubernetesEvent, error) {
	var event KubernetesEvent
	err := row.Scan(
		&event.ID,
		&event.ClusterID,
		&event.InvolvedUID,
		&event.Namespace,
		&event.InvolvedKind,
		&event.InvolvedName,
		&event.Type,
		&event.Reason,
		&event.Message,
		&event.Count,
		&event.FirstSeenAt,
		&event.LastSeenAt,
		&event.ObservedAt,
	)
	if err != nil {
		return KubernetesEvent{}, err
	}
	return event, nil
}

func scanActionPlan(row scanner) (ActionPlan, error) {
	var plan ActionPlan
	var planRaw []byte
	err := row.Scan(
		&plan.ID,
		&plan.FindingID,
		&plan.EvidencePackID,
		&plan.Title,
		&plan.Summary,
		&plan.RiskLevel,
		&plan.Status,
		&planRaw,
		&plan.CreatedAt,
		&plan.UpdatedAt,
	)
	if err != nil {
		return ActionPlan{}, err
	}
	plan.PlanJSON = planRaw
	return plan, nil
}

func marshalJSON(value any) (string, error) {
	if value == nil {
		return "{}", nil
	}
	payload, err := json.Marshal(value)
	if err != nil {
		return "", err
	}
	if len(payload) == 0 {
		return "{}", nil
	}
	return string(payload), nil
}

func ensureJSONObject(payload json.RawMessage) (json.RawMessage, error) {
	if len(payload) == 0 {
		return json.RawMessage(`{}`), nil
	}
	if !json.Valid(payload) {
		return nil, fmt.Errorf("invalid json payload")
	}
	return payload, nil
}

func parseStringMap(payload []byte) map[string]string {
	if len(payload) == 0 {
		return map[string]string{}
	}
	values := map[string]string{}
	if err := json.Unmarshal(payload, &values); err != nil {
		return map[string]string{}
	}
	return values
}

func nullString(value string) any {
	if value == "" {
		return nil
	}
	return value
}

func nullTime(value sql.NullTime) any {
	if !value.Valid {
		return nil
	}
	return value.Time
}

func nullableUUID(value string) any {
	if value == "" {
		return nil
	}
	return value
}

func nullInt(value int) any {
	if value == 0 {
		return nil
	}
	return value
}

func nullNumericString(value string) any {
	if value == "" {
		return nil
	}
	return value
}
