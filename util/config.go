package util

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/lemonsoul/jenkins-cli/config"
	"gopkg.in/yaml.v3"
)

func GetConfigFilePath() string {
	return os.Getenv("HOME") + config.BASE_CONFIG_DIR
}

func GetWorkspaceFilePathByName(accountName string) string {
	if strings.TrimSpace(accountName) == "" || accountName == config.DEFAULT_ACCOUNT_NAME {
		return os.Getenv("HOME") + config.WORKSPACE_INFO_DIR
	}
	safeName := sanitizeAccountName(accountName)
	return filepath.Join(os.Getenv("HOME"), ".config", config.BASE_NAME, config.WORKSPACE_INFO+"_"+safeName+".yaml")
}

func GetAccountByName(accountName string) (config.JenkinsConfig, error) {
	cfgFile, err := loadConfigFile()
	if err != nil {
		return config.JenkinsConfig{}, err
	}

	accounts, err := getAccountsFromConfigFile(cfgFile)
	if err != nil {
		return config.JenkinsConfig{}, err
	}

	if strings.TrimSpace(accountName) == "" {
		return config.JenkinsConfig{}, fmt.Errorf("account name is required")
	}

	account, ok := accounts[accountName]
	if !ok {
		return config.JenkinsConfig{}, fmt.Errorf("account not found: %s", accountName)
	}
	return account, nil
}

func PickAccount(accountName string) (config.JenkinsConfig, error) {
	cfgFile, err := loadConfigFile()
	if err != nil {
		return config.JenkinsConfig{}, err
	}

	accounts, err := getAccountsFromConfigFile(cfgFile)
	if err != nil {
		return config.JenkinsConfig{}, err
	}

	if strings.TrimSpace(accountName) != "" {
		account, ok := accounts[accountName]
		if !ok {
			return config.JenkinsConfig{}, fmt.Errorf("account not found: %s", accountName)
		}
		cfgFile.RecentAccounts = UpdateRecent(cfgFile.RecentAccounts, accountName, 3)
		_ = writeConfigFile(GetConfigFilePath(), cfgFile)
		return account, nil
	}

	if len(accounts) == 1 {
		for name, account := range accounts {
			cfgFile.RecentAccounts = UpdateRecent(cfgFile.RecentAccounts, name, 3)
			_ = writeConfigFile(GetConfigFilePath(), cfgFile)
			return account, nil
		}
	}

	accountNames := make([]string, 0, len(accounts))
	for name := range accounts {
		accountNames = append(accountNames, name)
	}
	sort.Strings(accountNames)
	recent := FilterRecent(cfgFile.RecentAccounts, BuildAllowSet(accountNames), 3)
	selected := StrUISelectWithRecent("Select Account", accountNames, recent)
	if strings.TrimSpace(selected) == "" {
		return config.JenkinsConfig{}, fmt.Errorf("no account selected")
	}

	account, ok := accounts[selected]
	if !ok {
		return config.JenkinsConfig{}, fmt.Errorf("selected account not found: %s", selected)
	}
	cfgFile.RecentAccounts = UpdateRecent(cfgFile.RecentAccounts, selected, 3)
	_ = writeConfigFile(GetConfigFilePath(), cfgFile)
	return account, nil
}

func loadConfigFile() (config.JenkinsConfigFile, error) {
	baseConfigDir := GetConfigFilePath()

	// Check if config file exists
	if _, err := os.Stat(baseConfigDir); os.IsNotExist(err) {
		return config.JenkinsConfigFile{}, fmt.Errorf("config file not found at %s. Please run 'jenkins-cli init' to create configuration", baseConfigDir)
	}

	configFile, err := os.ReadFile(baseConfigDir)
	if err != nil {
		return config.JenkinsConfigFile{}, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg config.JenkinsConfigFile
	err = yaml.Unmarshal(configFile, &cfg)
	if err != nil {
		return config.JenkinsConfigFile{}, fmt.Errorf("failed to parse config file: %w", err)
	}
	if cfg.Accounts == nil {
		cfg.Accounts = make([]config.JenkinsConfig, 0)
	}
	if cfg.RecentAccounts == nil {
		cfg.RecentAccounts = make([]string, 0)
	}

	return cfg, nil
}

func writeConfigFile(path string, cfg config.JenkinsConfigFile) error {
	data, err := yaml.Marshal(&cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func ListAccounts() ([]config.JenkinsConfig, error) {
	cfgFile, err := loadConfigFile()
	if err != nil {
		return nil, err
	}
	accounts, err := getAccountsFromConfigFile(cfgFile)
	if err != nil {
		return nil, err
	}
	result := make([]config.JenkinsConfig, 0, len(accounts))
	for _, account := range accounts {
		result = append(result, account)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})
	return result, nil
}

func GetWorkspaceFile(accountName string) (config.Workspace, error) {
	workspaceInfoDir := GetWorkspaceFilePathByName(accountName)

	// Check if workspace file exists
	if _, err := os.Stat(workspaceInfoDir); os.IsNotExist(err) {
		return config.Workspace{}, fmt.Errorf("workspace file not found at %s. Please run 'jenkins-cli sync' to create workspace configuration", workspaceInfoDir)
	}

	configFile, err := os.ReadFile(workspaceInfoDir)
	if err != nil {
		return config.Workspace{}, fmt.Errorf("failed to read workspace file: %w", err)
	}
	needsMigration := workspaceNeedsMigration(configFile)

	var cfg config.Workspace
	err = yaml.Unmarshal(configFile, &cfg)
	if err != nil {
		return config.Workspace{}, fmt.Errorf("failed to parse workspace file: %w", err)
	}
	if needsMigration || ensureWorkspaceRecent(&cfg) {
		if err := writeWorkspaceFile(workspaceInfoDir, cfg); err != nil {
			return config.Workspace{}, fmt.Errorf("failed to migrate workspace file: %w", err)
		}
	}

	return cfg, nil
}

func writeWorkspaceFile(path string, cfg config.Workspace) error {
	data, err := yaml.Marshal(&cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func ensureWorkspaceRecent(cfg *config.Workspace) bool {
	if cfg == nil {
		return false
	}
	updated := false
	if cfg.RecentViews == nil {
		cfg.RecentViews = make([]string, 0)
		updated = true
	}
	for viewIndex := range cfg.Views {
		if cfg.Views[viewIndex].RecentJobs == nil {
			cfg.Views[viewIndex].RecentJobs = make([]string, 0)
			updated = true
		}
		if len(cfg.LegacyRecentJobs) > 0 && len(cfg.Views[viewIndex].RecentJobs) == 0 {
			cfg.Views[viewIndex].RecentJobs = append([]string(nil), cfg.LegacyRecentJobs...)
			updated = true
		}
		for jobIndex := range cfg.Views[viewIndex].Job {
			job := &cfg.Views[viewIndex].Job[jobIndex]
			if job.RecentChoices == nil {
				job.RecentChoices = make([]string, 0)
				updated = true
			}
			if job.RecentBranches == nil {
				job.RecentBranches = make([]string, 0)
				updated = true
			}
			if len(job.JobParam.LegacyRecentChoices) > 0 && len(job.RecentChoices) == 0 {
				job.RecentChoices = append([]string(nil), job.JobParam.LegacyRecentChoices...)
				updated = true
			}
			if len(job.JobParam.LegacyRecentBranches) > 0 && len(job.RecentBranches) == 0 {
				job.RecentBranches = append([]string(nil), job.JobParam.LegacyRecentBranches...)
				updated = true
			}
			if job.JobParam.LegacyRecentChoices != nil {
				job.JobParam.LegacyRecentChoices = nil
				updated = true
			}
			if job.JobParam.LegacyRecentBranches != nil {
				job.JobParam.LegacyRecentBranches = nil
				updated = true
			}
		}
	}
	if cfg.LegacyRecentJobs != nil {
		cfg.LegacyRecentJobs = nil
		updated = true
	}
	return updated
}

func workspaceNeedsMigration(data []byte) bool {
	var node yaml.Node
	if err := yaml.Unmarshal(data, &node); err != nil {
		return false
	}
	return nodeHasLegacyParams(&node)
}

func nodeHasLegacyParams(node *yaml.Node) bool {
	if node == nil {
		return false
	}
	if node.Kind == yaml.MappingNode {
		for i := 0; i < len(node.Content)-1; i += 2 {
			key := node.Content[i]
			value := node.Content[i+1]
			if key != nil && (key.Value == "choices" || key.Value == "branch") {
				if value != nil && value.Kind == yaml.SequenceNode {
					for _, item := range value.Content {
						if item != nil && (item.Kind == yaml.MappingNode || item.Kind == yaml.SequenceNode) {
							return true
						}
					}
				}
			}
			if nodeHasLegacyParams(value) {
				return true
			}
		}
	}
	if node.Kind == yaml.SequenceNode {
		for _, item := range node.Content {
			if nodeHasLegacyParams(item) {
				return true
			}
		}
	}
	return false
}

func getAccountsFromConfigFile(cfg config.JenkinsConfigFile) (map[string]config.JenkinsConfig, error) {
	accounts := make([]config.JenkinsConfig, 0)
	if len(cfg.Accounts) > 0 {
		accounts = append(accounts, cfg.Accounts...)
	}

	if len(accounts) == 0 {
		return nil, fmt.Errorf("no account configured in %s", GetConfigFilePath())
	}

	accountMap := make(map[string]config.JenkinsConfig, len(accounts))
	for _, account := range accounts {
		requireName := len(accounts) > 1
		if strings.TrimSpace(account.Name) == "" && !requireName {
			account.Name = config.DEFAULT_ACCOUNT_NAME
		}
		if err := validateJenkinsAccount(account, requireName); err != nil {
			return nil, err
		}
		if _, exists := accountMap[account.Name]; exists {
			return nil, fmt.Errorf("duplicate account name: %s", account.Name)
		}
		accountMap[account.Name] = account
	}

	return accountMap, nil
}

// validateJenkinsAccount validates the Jenkins account configuration
func validateJenkinsAccount(cfg config.JenkinsConfig, requireName bool) error {
	var errors []string

	if requireName && strings.TrimSpace(cfg.Name) == "" {
		errors = append(errors, "name is required")
	}
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

func sanitizeAccountName(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return config.DEFAULT_ACCOUNT_NAME
	}
	return strings.Map(func(r rune) rune {
		switch {
		case r >= 'a' && r <= 'z':
			return r
		case r >= 'A' && r <= 'Z':
			return r
		case r >= '0' && r <= '9':
			return r
		case r == '-' || r == '_' || r == '.':
			return r
		default:
			return '_'
		}
	}, name)
}
