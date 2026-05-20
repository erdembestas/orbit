package config

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	defaultConfigDir = "/etc/orbit/config"
	defaultAPIPort   = "8080"
)

type Config struct {
	Environment                 string
	LogLevel                    string
	APIPort                     string
	AuthMode                    string
	ClusterName                 string
	ClusterType                 string
	Mode                        string
	ControllerEnabled           bool
	ControllerIntervalSeconds   int
	EvidenceMaxEvents           int
	EvidenceMaxRelatedResources int
	EvidenceMaxLogLines         int
	EvidenceMaxTokenEstimate    int
}

func Load() Config {
	return LoadFromDir(defaultConfigDir)
}

func LoadFromDir(dir string) Config {
	return Config{
		Environment:                 readConfigValue(dir, "ORBIT_ENV", "local"),
		LogLevel:                    readConfigValue(dir, "ORBIT_LOG_LEVEL", "debug"),
		APIPort:                     readConfigValue(dir, "ORBIT_API_PORT", defaultAPIPort),
		AuthMode:                    readConfigValue(dir, "ORBIT_AUTH_MODE", "local"),
		ClusterName:                 readConfigValue(dir, "ORBIT_CLUSTER_NAME", "minikube"),
		ClusterType:                 readConfigValue(dir, "ORBIT_CLUSTER_TYPE", "kubernetes"),
		Mode:                        readConfigValue(dir, "ORBIT_MODE", "single-cluster"),
		ControllerEnabled:           readBoolValue(dir, "ORBIT_CONTROLLER_ENABLED", true),
		ControllerIntervalSeconds:   readIntValue(dir, "ORBIT_CONTROLLER_INTERVAL_SECONDS", 30),
		EvidenceMaxEvents:           readIntValue(dir, "ORBIT_EVIDENCE_MAX_EVENTS", 20),
		EvidenceMaxRelatedResources: readIntValue(dir, "ORBIT_EVIDENCE_MAX_RELATED_RESOURCES", 15),
		EvidenceMaxLogLines:         readIntValue(dir, "ORBIT_EVIDENCE_MAX_LOG_LINES", 80),
		EvidenceMaxTokenEstimate:    readIntValue(dir, "ORBIT_EVIDENCE_MAX_TOKEN_ESTIMATE", 4000),
	}
}

func (c Config) Address() string {
	return ":" + c.APIPort
}

func readConfigValue(dir, key, fallback string) string {
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

func readIntValue(dir, key string, fallback int) int {
	value := readConfigValue(dir, key, "")
	if value == "" {
		return fallback
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}

	return parsed
}

func readBoolValue(dir, key string, fallback bool) bool {
	value := strings.ToLower(readConfigValue(dir, key, ""))
	switch value {
	case "true", "1", "yes", "on":
		return true
	case "false", "0", "no", "off":
		return false
	case "":
		return fallback
	default:
		return fallback
	}
}
