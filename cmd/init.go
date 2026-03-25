package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/fatih/color"
	"github.com/lemonsoul/jenkins-cli/config"
	"github.com/lemonsoul/jenkins-cli/util"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "init config",
	Long:  `init config for jenkins-cli`,
	Run: func(cmd *cobra.Command, args []string) {
		baseConfigDir, _ := createFile(config.BASE_CONFIG_DIR)
		var (
			accountName string
			username    string
			token       string
			baseApi     string
		)
		accountName, _ = cmd.Flags().GetString("account")
		username, _ = cmd.Flags().GetString("username")
		token, _ = cmd.Flags().GetString("token")
		baseApi, _ = cmd.Flags().GetString("base-api")

		cfgFile, err := readConfigFileShared(baseConfigDir)
		if err != nil {
			color.Red("failed to read config file", err)
			return
		}
		for {
			if accountName == "" {
				color.Yellow("Account name: ")
				fmt.Print("> ")
				fmt.Scanln(&accountName)
			}
			if accountName == "" {
				accountName = config.DEFAULT_ACCOUNT_NAME
			}
			if hasAccountNameShared(cfgFile.Accounts, accountName) {
				color.Yellow("account name already exists: %s", accountName)
				if cmd.Flags().Changed("account") {
					return
				}
				accountName = ""
				continue
			}
			break
		}
		if username == "" {
			color.Yellow("Jenkins username: ")
			fmt.Print("> ")
			fmt.Scanln(&username)
		}
		if token == "" {
			color.Yellow("Jenkins token: ")
			fmt.Print("> ")
			fmt.Scanln(&token)
		}
		if baseApi == "" {
			color.Yellow("Jenkins base api url: ")
			fmt.Print("> ")
			fmt.Scanln(&baseApi)
		}

		cfgFile.Accounts = append(cfgFile.Accounts, config.JenkinsConfig{
			Name:     accountName,
			Username: username,
			Token:    token,
			BaseApi:  baseApi,
		})
		data, _ := yaml.Marshal(&cfgFile)
		if err := os.WriteFile(baseConfigDir, data, 0644); err != nil {
			color.Red("failed to write config file", err)
			return
		}

		workspacePath := util.GetWorkspaceFilePathByName(accountName)
		if err := createFileByFullPath(workspacePath); err != nil {
			color.Red("failed to create workspace file", err)
			return
		}
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
	initCmd.Flags().String("account", "", "account name")
	initCmd.Flags().String("username", "", "jenkins username")
	initCmd.Flags().String("token", "", "jenkins token")
	initCmd.Flags().String("base-api", "", "jenkins base api url")
}

func createFile(path string) (dir string, err error) {
	fileDir := os.Getenv("HOME") + path
	if _, err := os.Stat(fileDir); err == nil {
		color.Yellow("config file already exists")
	} else {
		if err := os.MkdirAll(filepath.Dir(fileDir), os.ModePerm); err != nil {
			color.Red("failed to create config dir", err)
			return "", err
		}
		if _, err := os.Create(fileDir); err != nil {
			color.Red("failed to create config file")
			return "", err
		}
	}
	return fileDir, nil
}

func createFileByFullPath(filePath string) error {
	if _, err := os.Stat(filePath); err == nil {
		color.Yellow("config file already exists")
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(filePath), os.ModePerm); err != nil {
		color.Red("failed to create config dir", err)
		return err
	}
	if _, err := os.Create(filePath); err != nil {
		color.Red("failed to create config file")
		return err
	}
	return nil
}

// helpers moved to config_helpers.go
