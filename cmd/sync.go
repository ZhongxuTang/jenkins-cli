package cmd

import (
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/lemonsoul/jenkins-cli/api"
	"github.com/lemonsoul/jenkins-cli/config"
	"github.com/lemonsoul/jenkins-cli/util"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var syncCmd = &cobra.Command{
	Use:   "sync [account]",
	Short: "sync config",
	Long:  `sync jenkins data to config file`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) > 1 {
			color.Red("❌ Too many arguments, at most one account name is allowed")
			return
		}
		if len(args) == 1 {
			accountName := strings.TrimSpace(args[0])
			if accountName == "" {
				color.Red("❌ Account name cannot be empty")
				return
			}
			account, err := util.GetAccountByName(accountName)
			if err != nil {
				color.Red("❌ Error loading account %s: %v", accountName, err)
				return
			}
			if err := syncWorkspaceForAccount(account); err != nil {
				color.Red("❌ Sync failed for account %s: %v", accountName, err)
			}
			return
		}

		accounts, err := util.ListAccounts()
		if err != nil {
			color.Red("❌ Error loading accounts: %v", err)
			return
		}
		for _, account := range accounts {
			if err := syncWorkspaceForAccount(account); err != nil {
				color.Red("❌ Sync failed for account %s: %v", account.Name, err)
			}
		}
	},
}

func syncWorkspaceForAccount(account config.JenkinsConfig) error {
	cfg, err := util.GetWorkspaceFile(account.Name)
	if err != nil {
		// If workspace file doesn't exist, create empty workspace
		color.Yellow("⚠️ Workspace file not found, creating new one...")
		cfg = config.Workspace{Views: make([]config.View, 0)}
	}

	viewNames, err := api.GetViews(account)
	if err != nil {
		return err
	}

	// Reset views to get fresh data
	oldCfg := cfg
	cfg.Views = make([]config.View, 0)
	cfg.RecentViews = util.FilterRecent(cfg.RecentViews, util.BuildAllowSet(viewNames), 3)

	for _, viewName := range viewNames {
		view := config.View{}
		view.Name = viewName
		view.Job = make([]config.Job, 0)

		jobNames, err := api.GetViewJob(account, viewName)
		if err != nil {
			color.Yellow("⚠️ Error getting jobs for view %s: %v", viewName, err)
			continue
		}
		view.RecentJobs = filterViewRecentJobs(oldCfg, viewName, jobNames)

		for _, jobName := range jobNames {
			jobParam := config.JobParam{}
			existingJob, ok := findJob(oldCfg, viewName, jobName)
			choices, branches, err := api.GetJobParams(account, jobName)
			if err != nil {
				color.Yellow("⚠️ Error getting job params for %s: %v", jobName, err)
				if ok {
					jobParam = existingJob.JobParam
				}
			} else {
				jobParam.Choices = choices
				jobParam.Branch = branches
			}
			job := config.Job{Name: jobName, JobParam: jobParam}
			if ok {
				choiceSet := util.BuildAllowSet(jobParam.Choices)
				branchSet := util.BuildAllowSet(jobParam.Branch)
				job.RecentChoices = util.FilterRecent(existingJob.RecentChoices, choiceSet, 3)
				job.RecentBranches = util.FilterRecent(existingJob.RecentBranches, branchSet, 3)
			}
			view.Job = append(view.Job, job)
		}
		cfg.Views = append(cfg.Views, view)
	}

	data, err := yaml.Marshal(&cfg)
	if err != nil {
		return err
	}

	workspacePath := util.GetWorkspaceFilePathByName(account.Name)

	// Write the updated config back to the file
	err = os.WriteFile(workspacePath, data, 0644)
	if err != nil {
		return err
	}

	color.Green("✅ Workspace configuration synced successfully!")
	return nil
}

func init() {
	rootCmd.AddCommand(syncCmd)
}

func findJob(cfg config.Workspace, viewName, jobName string) (config.Job, bool) {
	for _, view := range cfg.Views {
		if view.Name != viewName {
			continue
		}
		for _, job := range view.Job {
			if job.Name == jobName {
				return job, true
			}
		}
	}
	return config.Job{}, false
}

func filterViewRecentJobs(cfg config.Workspace, viewName string, allow []string) []string {
	allowSet := util.BuildAllowSet(allow)
	for _, view := range cfg.Views {
		if view.Name == viewName {
			return util.FilterRecent(view.RecentJobs, allowSet, 3)
		}
	}
	return nil
}
