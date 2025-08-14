package util

import (
	"os"

	"github.com/lemonsoul/jenkins-cli/config"
	"gopkg.in/yaml.v3"
)

func GetConfigFilePath() string {
	return os.Getenv("HOME") + config.BASE_CONFIG_DIR
}

func GetWorkspaceFilePath() string {
	return os.Getenv("HOME") + config.WORKSPACE_INFO_DIR
}

func GetConfigFile() config.JenkinsConfig {
	baseConfigDir := GetConfigFilePath()

	configFile, err := os.ReadFile(baseConfigDir)
	if err != nil {
		panic("failed to read config file")
	}
	var cfg config.JenkinsConfig
	err = yaml.Unmarshal(configFile, &cfg)
	if err != nil {
		panic("failed to unmarshal config file")
	}
	return cfg
}

func GetWprkspaceFile() config.Workspace {
	workspaceInfoDir := GetWorkspaceFilePath()

	config_file, err := os.ReadFile(workspaceInfoDir)
	if err != nil {
		panic("failed to read config file")
	}
	var cfg config.Workspace
	err = yaml.Unmarshal(config_file, &cfg)
	if err != nil {
		panic("failed to unmarshal config file")
	}
	return cfg
}
