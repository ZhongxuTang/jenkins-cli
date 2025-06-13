package cmd

import (
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
		base_config_dir := os.Getenv("HOME") + config.BASE_CONFIG_DIR
		if _, err := os.Stat(base_config_dir); err == nil {
			color.White("config file already exists")
		} else {
			if err := os.MkdirAll(filepath.Dir(base_config_dir), os.ModePerm); err != nil {
				color.White("failed to create config dir", err)
				return
			}
			if _, err := os.Create(base_config_dir); err != nil {
				color.White("failed to create config file")
			}

			config := config.JenkinsConfig{
				Username: "",
				Token:    "",
				BaseApi:  "",
			}
			data, _ := yaml.Marshal(&config)
			if err := os.WriteFile(base_config_dir, data, 0644); err != nil {
				color.White("failed to write config file", err)
				return
			}

		}
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
