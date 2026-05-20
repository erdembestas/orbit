package db

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const defaultDBDir = "/etc/orbit/db"

type Config struct {
	Host     string
	Port     string
	Name     string
	User     string
	Password string
	SSLMode  string
}

func Load() (Config, error) {
	return LoadFromDir(defaultDBDir)
}

func LoadFromDir(dir string) (Config, error) {
	password, err := readRequiredDBValue(dir, "ORBIT_DB_PASSWORD")
	if err != nil {
		return Config{}, err
	}

	return Config{
		Host:     readDBValue(dir, "ORBIT_DB_HOST", "localhost"),
		Port:     readDBValue(dir, "ORBIT_DB_PORT", "5432"),
		Name:     readDBValue(dir, "ORBIT_DB_NAME", "orbit"),
		User:     readDBValue(dir, "ORBIT_DB_USER", "orbit"),
		Password: password,
		SSLMode:  readDBValue(dir, "ORBIT_DB_SSLMODE", "disable"),
	}, nil
}

func (c Config) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%s dbname=%s user=%s password=%s sslmode=%s",
		c.Host,
		c.Port,
		c.Name,
		c.User,
		c.Password,
		c.SSLMode,
	)
}

func readDBValue(dir, key, fallback string) string {
	value, err := os.ReadFile(filepath.Join(dir, key))
	if err != nil {
		return fallback
	}

	trimmed := strings.TrimSpace(string(value))
	if trimmed == "" {
		return fallback
	}

	return trimmed
}

func readRequiredDBValue(dir, key string) (string, error) {
	value, err := os.ReadFile(filepath.Join(dir, key))
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("required db config file missing: %s", key)
		}
		return "", fmt.Errorf("failed to read db config file %s: %w", key, err)
	}

	trimmed := strings.TrimSpace(string(value))
	if trimmed == "" {
		return "", fmt.Errorf("required db config file empty: %s", key)
	}

	return trimmed, nil
}
