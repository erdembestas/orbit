package secrets

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const defaultSecretsDir = "/etc/orbit/secrets"

type BootstrapConfig struct {
	AdminUsername string
	AdminPassword string
	JWTSecret     string
}

func Load() (BootstrapConfig, error) {
	return LoadFromDir(defaultSecretsDir)
}

func LoadFromDir(dir string) (BootstrapConfig, error) {
	adminUsername, err := readRequiredSecretValue(dir, "ORBIT_BOOTSTRAP_ADMIN_USERNAME")
	if err != nil {
		return BootstrapConfig{}, err
	}
	adminPassword, err := readRequiredSecretValue(dir, "ORBIT_BOOTSTRAP_ADMIN_PASSWORD")
	if err != nil {
		return BootstrapConfig{}, err
	}
	jwtSecret, err := readRequiredSecretValue(dir, "ORBIT_JWT_SECRET")
	if err != nil {
		return BootstrapConfig{}, err
	}

	return BootstrapConfig{
		AdminUsername: adminUsername,
		AdminPassword: adminPassword,
		JWTSecret:     jwtSecret,
	}, nil
}

func readRequiredSecretValue(dir, key string) (string, error) {
	value, err := os.ReadFile(filepath.Join(dir, key))
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("required secret file missing: %s", key)
		}
		return "", fmt.Errorf("failed to read secret file %s: %w", key, err)
	}

	trimmed := strings.TrimSpace(string(value))
	if trimmed == "" {
		return "", fmt.Errorf("required secret file empty: %s", key)
	}

	return trimmed, nil
}
