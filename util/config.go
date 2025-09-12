package util

import (
	"fmt"
	"os"
	"strings"

	"github.com/lemonsoul/jenkins-cli/config"
	"gopkg.in/yaml.v3"
)

func GetConfigFilePath() string {
	return os.Getenv("HOME") + config.BASE_CONFIG_DIR
}

func GetWorkspaceFilePath() string {
	return os.Getenv("HOME") + config.WORKSPACE_INFO_DIR
}

func GetConfigFile() (config.JenkinsConfig, error) {
	baseConfigDir := GetConfigFilePath()

	// Check if config file exists
	if _, err := os.Stat(baseConfigDir); os.IsNotExist(err) {
		return config.JenkinsConfig{}, fmt.Errorf("config file not found at %s. Please run 'jenkins-cli init' to create configuration", baseConfigDir)
	}

	configFile, err := os.ReadFile(baseConfigDir)
	if err != nil {
		return config.JenkinsConfig{}, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg config.JenkinsConfig
	err = yaml.Unmarshal(configFile, &cfg)
	if err != nil {
		return config.JenkinsConfig{}, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Validate required fields
	if err := validateJenkinsConfig(cfg); err != nil {
		return config.JenkinsConfig{}, fmt.Errorf("invalid configuration: %w", err)
	}

	return cfg, nil
}

func GetWorkspaceFile() (config.Workspace, error) {
	workspaceInfoDir := GetWorkspaceFilePath()

	// Check if workspace file exists
	if _, err := os.Stat(workspaceInfoDir); os.IsNotExist(err) {
		return config.Workspace{}, fmt.Errorf("workspace file not found at %s. Please run 'jenkins-cli sync' to create workspace configuration", workspaceInfoDir)
	}

	configFile, err := os.ReadFile(workspaceInfoDir)
	if err != nil {
		return config.Workspace{}, fmt.Errorf("failed to read workspace file: %w", err)
	}

	var cfg config.Workspace
	err = yaml.Unmarshal(configFile, &cfg)
	if err != nil {
		return config.Workspace{}, fmt.Errorf("failed to parse workspace file: %w", err)
	}

	return cfg, nil
}

// validateJenkinsConfig validates the Jenkins configuration
func validateJenkinsConfig(cfg config.JenkinsConfig) error {
	var errors []string

	if strings.TrimSpace(cfg.Username) == "" {
		errors = append(errors, "username is required")
	}
	if strings.TrimSpace(cfg.Token) == "" {
		errors = append(errors, "token is required")
	}
	if strings.TrimSpace(cfg.BaseApi) == "" {
		errors = append(errors, "base_api is required")
	}

	if len(errors) > 0 {
		return fmt.Errorf("configuration validation failed: %s", strings.Join(errors, ", "))
	}
	return nil
}

// GetWprkspaceFile is deprecated: use GetWorkspaceFile instead
// Keeping for backward compatibility
func GetWprkspaceFile() config.Workspace {
	cfg, err := GetWorkspaceFile()
	if err != nil {
		// For backward compatibility, we panic as the original function did
		panic(err.Error())
	}
	return cfg
}
