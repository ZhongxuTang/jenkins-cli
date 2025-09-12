package cmd

import (
	"os"

	"github.com/fatih/color"
	"github.com/lemonsoul/jenkins-cli/api"
	"github.com/lemonsoul/jenkins-cli/config"
	"github.com/lemonsoul/jenkins-cli/util"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "sync config",
	Long:  `sync jenkins data to config file`,
	Run: func(cmd *cobra.Command, args []string) {
		baseConfig, err := util.GetConfigFile()
		if err != nil {
			color.Red("❌ Error loading base configuration: %v", err)
			return
		}

		cfg, err := util.GetWorkspaceFile()
		if err != nil {
			// If workspace file doesn't exist, create empty workspace
			color.Yellow("⚠️ Workspace file not found, creating new one...")
			cfg = config.Workspace{Views: make([]config.View, 0)}
		}

		viewNames, err := api.GetViews()
		if err != nil {
			color.Red("❌ Error getting views from Jenkins: %v", err)
			return
		}

		// Reset views to get fresh data
		cfg.Views = make([]config.View, 0)

		for _, viewName := range viewNames {
			view := config.View{}
			view.Name = viewName
			view.Job = make([]config.Job, 0)

			jobNames, err := api.GetViewJob(baseConfig.Username, baseConfig.Token, baseConfig.BaseApi, viewName)
			if err != nil {
				color.Yellow("⚠️ Error getting jobs for view %s: %v", viewName, err)
				continue
			}

			for _, jobName := range jobNames {
				view.Job = append(view.Job, config.Job{Name: jobName, JobParam: config.JobParam{}})
			}
			cfg.Views = append(cfg.Views, view)
		}

		data, err := yaml.Marshal(&cfg)
		if err != nil {
			color.Red("❌ Error marshaling workspace config: %v", err)
			return
		}

		// Write the updated config back to the file
		err = os.WriteFile(util.GetWorkspaceFilePath(), data, 0644)
		if err != nil {
			color.Red("❌ Error writing workspace config file: %v", err)
			return
		}

		color.Green("✅ Workspace configuration synced successfully!")
	},
}

func init() {
	rootCmd.AddCommand(syncCmd)
}
