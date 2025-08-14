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
		baseConfig := util.GetConfigFile()
		cfg := util.GetWprkspaceFile()
		viewNames := api.GetViews()
		for _, viewName := range viewNames {
			view := config.View{}
			view.Name = viewName
			view.Job = make([]config.Job, 0)

			jobNames := api.GetViewJob(baseConfig.Username, baseConfig.Token, baseConfig.BaseApi, viewName)
			for _, jobName := range jobNames {
				//api.GetJobPararm(cfg.Username, cfg.Token, cfg.BaseApi, jobName)
				view.Job = append(view.Job, config.Job{Name: jobName, JobParam: config.JobParam{}})
			}
			cfg.Views = append(cfg.Views, view)
		}

		data, err := yaml.Marshal(&cfg)
		if err != nil {
			color.White("failed to marshal config file", err)
			return
		}
		// Write the updated config back to the file
		err = os.WriteFile(util.GetWorkspaceFilePath(), data, 0644)
		if err != nil {
			color.White("failed to write config file", err)
			return
		}
	},
}

func init() {
	rootCmd.AddCommand(syncCmd)
}
