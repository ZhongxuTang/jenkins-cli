package cmd

import (
	"os"
	"strings"

	"github.com/lemonsoul/jenkins-cli/config"
	"gopkg.in/yaml.v3"
)

func readConfigFileShared(path string) (config.JenkinsConfigFile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return config.JenkinsConfigFile{}, err
	}
	if len(strings.TrimSpace(string(data))) == 0 {
		return config.JenkinsConfigFile{Accounts: make([]config.JenkinsConfig, 0)}, nil
	}
	var cfg config.JenkinsConfigFile
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return config.JenkinsConfigFile{}, err
	}
	if cfg.Accounts == nil {
		cfg.Accounts = make([]config.JenkinsConfig, 0)
	}
	return cfg, nil
}

func hasAccountNameShared(accounts []config.JenkinsConfig, name string) bool {
	for _, account := range accounts {
		if account.Name == name {
			return true
		}
	}
	return false
}
