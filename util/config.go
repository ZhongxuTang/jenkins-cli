package util

import (
	"os"

	"github.com/lemonsoul/jenkins-cli/config"
	"gopkg.in/yaml.v3"
)

func GetConfigFilePath() string {
	return os.Getenv("HOME") + config.BASE_CONFIG_DIR
}

func GetConfigFile() config.JenkinsConfig {
	base_config_dir := GetConfigFilePath()

	config_file, err := os.ReadFile(base_config_dir)
	if err != nil {
		panic("failed to read config file")
	}
	var cfg config.JenkinsConfig
	err = yaml.Unmarshal(config_file, &cfg)
	if err != nil {
		panic("failed to unmarshal config file")
	}
	return cfg
}
