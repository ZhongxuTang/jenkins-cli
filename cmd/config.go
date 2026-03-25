package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
	"github.com/lemonsoul/jenkins-cli/config"
	"github.com/lemonsoul/jenkins-cli/util"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "config account",
	Long:  `manage jenkins-cli accounts`,
	Run: func(cmd *cobra.Command, args []string) {
		action := util.StrUISelect("Select Action", []string{"Add Account", "Delete Account"})
		switch action {
		case "Add Account":
			if err := addAccount(); err != nil {
				color.Red("❌ %v", err)
			}
		case "Delete Account":
			if err := deleteAccount(); err != nil {
				color.Red("❌ %v", err)
			}
		default:
			color.Yellow("no action selected")
		}
	},
}

func init() {
	rootCmd.AddCommand(configCmd)
}

func addAccount() error {
	baseConfigPath := util.GetConfigFilePath()
	if err := ensureFile(baseConfigPath); err != nil {
		return fmt.Errorf("failed to ensure config file: %w", err)
	}
	cfgFile, err := readConfigFileShared(baseConfigPath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	accountName := promptAccountName(cfgFile.Accounts)

	var (
		username string
		token    string
		baseApi  string
	)
	color.Yellow("Jenkins username: ")
	fmt.Print("> ")
	fmt.Scanln(&username)
	color.Yellow("Jenkins token: ")
	fmt.Print("> ")
	fmt.Scanln(&token)
	color.Yellow("Jenkins base api url: ")
	fmt.Print("> ")
	fmt.Scanln(&baseApi)

	cfgFile.Accounts = append(cfgFile.Accounts, config.JenkinsConfig{
		Name:     accountName,
		Username: username,
		Token:    token,
		BaseApi:  baseApi,
	})

	if err := writeConfigFile(baseConfigPath, cfgFile); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	workspacePath := util.GetWorkspaceFilePathByName(accountName)
	if err := ensureFile(workspacePath); err != nil {
		return fmt.Errorf("failed to create workspace file: %w", err)
	}

	color.Green("✅ Account added successfully!")
	return nil
}

func deleteAccount() error {
	baseConfigPath := util.GetConfigFilePath()
	if err := ensureFile(baseConfigPath); err != nil {
		return fmt.Errorf("failed to ensure config file: %w", err)
	}
	cfgFile, err := readConfigFileShared(baseConfigPath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}
	if len(cfgFile.Accounts) == 0 {
		return fmt.Errorf("no accounts found")
	}

	accountNames := make([]string, 0, len(cfgFile.Accounts))
	for _, account := range cfgFile.Accounts {
		accountNames = append(accountNames, account.Name)
	}
	selected := util.StrUISelect("Select Account", accountNames)
	if strings.TrimSpace(selected) == "" {
		return fmt.Errorf("no account selected")
	}

	updated := make([]config.JenkinsConfig, 0, len(cfgFile.Accounts))
	for _, account := range cfgFile.Accounts {
		if account.Name != selected {
			updated = append(updated, account)
		}
	}
	cfgFile.Accounts = updated

	if err := writeConfigFile(baseConfigPath, cfgFile); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	workspacePath := util.GetWorkspaceFilePathByName(selected)
	if err := os.Remove(workspacePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove workspace file: %w", err)
	}

	color.Green("✅ Account deleted successfully!")
	return nil
}

func writeConfigFile(path string, cfg config.JenkinsConfigFile) error {
	data, err := yaml.Marshal(&cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func promptAccountName(accounts []config.JenkinsConfig) string {
	for {
		var accountName string
		color.Yellow("Account name: ")
		fmt.Print("> ")
		fmt.Scanln(&accountName)
		if accountName == "" {
			accountName = config.DEFAULT_ACCOUNT_NAME
		}
		if hasAccountNameShared(accounts, accountName) {
			color.Yellow("account name already exists: %s", accountName)
			continue
		}
		return accountName
	}
}

func ensureFile(filePath string) error {
	if _, err := os.Stat(filePath); err == nil {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(filePath), os.ModePerm); err != nil {
		return err
	}
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	return file.Close()
}
