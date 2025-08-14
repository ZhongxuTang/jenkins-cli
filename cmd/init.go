package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/fatih/color"
	"github.com/lemonsoul/jenkins-cli/config"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "init config",
	Long:  `init config for jenkins-cli`,
	Run: func(cmd *cobra.Command, args []string) {
		baseConfigDir, _ := createFile(config.BASE_CONFIG_DIR)
		createFile(config.WORKSPACE_INFO_DIR)
		var (
			username string
			token    string
			baseApi  string
		)
		color.Yellow("plase config jenkuins username:")
		fmt.Scanln(&username)
		color.Yellow("plase config jenkuins token:")
		fmt.Scanln(&token)
		color.Yellow("plase config jenkuins base api url:")
		fmt.Scanln(&baseApi)

		config := config.JenkinsConfig{
			Username: username,
			Token:    token,
			BaseApi:  baseApi,
		}
		data, _ := yaml.Marshal(&config)
		if err := os.WriteFile(baseConfigDir, data, 0644); err != nil {
			color.Red("failed to write config file", err)
			return
		}

	},
}

func init() {
	rootCmd.AddCommand(initCmd)
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
