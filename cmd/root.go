package cmd

import (
	"os"
	"strconv"
	"time"

	"github.com/briandowns/spinner"
	"github.com/fatih/color"
	"github.com/lemonsoul/jenkins-cli/api"
	"github.com/lemonsoul/jenkins-cli/config"
	"github.com/lemonsoul/jenkins-cli/util"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var rootCmd = &cobra.Command{
	Use:   "jenkins-cli",
	Short: "jcli",
	Long:  `jenkins-cli is a command line tool for managing Jenkins jobs and builds.`,
	Run: func(cmd *cobra.Command, args []string) {
		color.Cyan("🎈 Welcome to Jenkins CLI!")
		accountName, _ := cmd.Flags().GetString("account")
		viewName, _ := cmd.Flags().GetString("view")
		jobName, _ := cmd.Flags().GetString("job")

		var account config.JenkinsConfig
		var err error
		if accountName != "" {
			account, err = util.GetAccountByName(accountName)
		} else {
			account, err = util.PickAccount("")
		}
		if err != nil {
			color.Red("❌ Error loading account configuration: %v", err)
			return
		}

		workspaceCfg, err := util.GetWorkspaceFile(account.Name)
		if err != nil {
			color.Red("❌ Error loading workspace configuration: %v", err)
			return
		}
		if normalizeWorkspaceRecent(&workspaceCfg, account.Name) {
			color.Yellow("⚠️ Workspace recent updated.")
		}

		if jobName == "" {
			if viewName == "" {
				viewName = selectView(workspaceCfg)
				if viewName == "" {
					color.Red("🚨 No view selected, please try again.")
					return
				}
			}
			jobName = selectJob(workspaceCfg, viewName)
			if jobName == "" {
				color.Red("🚨 No job selected, please try again.")
				return
			}
		}

		choices, branches, err := api.GetJobParams(account, jobName)
		if err != nil {
			color.Red("❌ Error getting job parameters: %v", err)
			return
		}

		if updateWorkspaceParams(&workspaceCfg, account.Name, viewName, jobName, choices, branches) {
			color.Yellow("⚠️ Workspace params updated.")
		}

		choicesSelect := util.StrUISelectWithRecent("Select Choices", choices, getJobRecentChoices(workspaceCfg, viewName, jobName, choices))
		branchSelect := util.StrUISelectWithRecent("Select Branch", branches, getJobRecentBranches(workspaceCfg, viewName, jobName, branches))

		if jobName == "" || branchSelect == "" {
			color.Red("🚨 Selection incomplete, please try again.")
			return
		}

		queueId, err := api.BuildWithParameters(account, jobName, choicesSelect, branchSelect)
		if err != nil {
			color.Red("❌ Error starting build: %v", err)
			return
		}

		if queueId != "" {
			color.Cyan("🎉 Build " + jobName + choicesSelect + branchSelect + " success, queue id is " + queueId)
			if updateWorkspaceRecent(&workspaceCfg, account.Name, viewName, jobName, choicesSelect, branchSelect) {
				color.Yellow("⚠️ Workspace recent updated.")
			}
		}
		waitOperation(4)
		buildNumber := getBuildNumber(account, queueId, 8)
		if buildNumber == "" {
			color.Yellow("♻️ job maybe waiting to run, please check it later")
		} else {
			color.Cyan("🍻 Build " + jobName + " " + choicesSelect + " " + branchSelect + " success, build number is " + buildNumber)
		}
		buildInfo, err := api.GetBuildStatus(account, jobName, buildNumber)
		if err != nil {
			color.Yellow("⚠️ Error getting build status: %v", err)
			return
		}
		if len(buildInfo.ChangeSets) > 0 {
			for index, item := range buildInfo.ChangeSets {
				color.Cyan(strconv.Itoa(index+1) + "、" + item.Comment + " by " + item.AuthorFullName)
			}
		} else {
			color.Yellow("⚠️ No change sets")
		}
	},
}

func selectView(cfg config.Workspace) string {
	viewNames := make([]string, 0)
	for _, item := range cfg.Views {
		viewNames = append(viewNames, item.Name)
	}
	recent := util.FilterRecent(cfg.RecentViews, util.BuildAllowSet(viewNames), 3)
	return util.StrUISelectWithRecent("Select View", viewNames, recent)
}

func selectJob(cfg config.Workspace, viewResult string) string {
	jobNames := make([]string, 0)
	recent := make([]string, 0)
	for _, item := range cfg.Views {
		if item.Name == viewResult {
			for _, job := range item.Job {
				jobNames = append(jobNames, job.Name)
			}
			recent = util.FilterRecent(item.RecentJobs, util.BuildAllowSet(jobNames), 3)
			break
		}
	}
	return util.StrUISelectWithRecent("Select Job", jobNames, recent)
}

func updateWorkspaceParams(workspaceCfg *config.Workspace, accountName, viewName, jobName string, choices, branches []string) bool {
	if workspaceCfg == nil {
		return false
	}
	updated := false
	for viewIndex := range workspaceCfg.Views {
		if viewName != "" && workspaceCfg.Views[viewIndex].Name != viewName {
			continue
		}
		for jobIndex := range workspaceCfg.Views[viewIndex].Job {
			job := workspaceCfg.Views[viewIndex].Job[jobIndex]
			if job.Name != jobName {
				continue
			}
			current := workspaceCfg.Views[viewIndex].Job[jobIndex].JobParam
			if slicesEqual(current.Choices, choices) && slicesEqual(current.Branch, branches) {
				continue
			}
			workspaceCfg.Views[viewIndex].Job[jobIndex].JobParam.Choices = choices
			workspaceCfg.Views[viewIndex].Job[jobIndex].JobParam.Branch = branches
			updated = true
		}
	}
	if !updated {
		return false
	}
	workspacePath := util.GetWorkspaceFilePathByName(accountName)
	return writeWorkspaceFile(workspacePath, *workspaceCfg)
}

func updateWorkspaceRecent(workspaceCfg *config.Workspace, accountName, viewName, jobName, choice, branch string) bool {
	if workspaceCfg == nil {
		return false
	}
	updated := false
	recentViews := util.UpdateRecent(workspaceCfg.RecentViews, viewName, 3)
	if !slicesEqual(workspaceCfg.RecentViews, recentViews) {
		workspaceCfg.RecentViews = recentViews
		updated = true
	}
	for viewIndex := range workspaceCfg.Views {
		if viewName != "" && workspaceCfg.Views[viewIndex].Name != viewName {
			continue
		}
		recentJobs := util.UpdateRecent(workspaceCfg.Views[viewIndex].RecentJobs, jobName, 3)
		if !slicesEqual(workspaceCfg.Views[viewIndex].RecentJobs, recentJobs) {
			workspaceCfg.Views[viewIndex].RecentJobs = recentJobs
			updated = true
		}
		for jobIndex := range workspaceCfg.Views[viewIndex].Job {
			job := workspaceCfg.Views[viewIndex].Job[jobIndex]
			if job.Name != jobName {
				continue
			}
			recentChoices := util.UpdateRecent(job.RecentChoices, choice, 3)
			recentBranches := util.UpdateRecent(job.RecentBranches, branch, 3)
			if !slicesEqual(job.RecentChoices, recentChoices) || !slicesEqual(job.RecentBranches, recentBranches) {
				workspaceCfg.Views[viewIndex].Job[jobIndex].RecentChoices = recentChoices
				workspaceCfg.Views[viewIndex].Job[jobIndex].RecentBranches = recentBranches
				updated = true
			}
		}
	}
	if !updated {
		return false
	}
	workspacePath := util.GetWorkspaceFilePathByName(accountName)
	return writeWorkspaceFile(workspacePath, *workspaceCfg)
}

func getJobRecentChoices(cfg config.Workspace, viewName, jobName string, allow []string) []string {
	allowSet := util.BuildAllowSet(allow)
	for _, view := range cfg.Views {
		if viewName != "" && view.Name != viewName {
			continue
		}
		for _, job := range view.Job {
			if job.Name == jobName {
				return util.FilterRecent(job.RecentChoices, allowSet, 3)
			}
		}
	}
	return nil
}

func getJobRecentBranches(cfg config.Workspace, viewName, jobName string, allow []string) []string {
	allowSet := util.BuildAllowSet(allow)
	for _, view := range cfg.Views {
		if viewName != "" && view.Name != viewName {
			continue
		}
		for _, job := range view.Job {
			if job.Name == jobName {
				return util.FilterRecent(job.RecentBranches, allowSet, 3)
			}
		}
	}
	return nil
}

func slicesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func writeWorkspaceFile(path string, cfg config.Workspace) bool {
	data, err := yaml.Marshal(&cfg)
	if err != nil {
		color.Yellow("⚠️ Failed to marshal workspace config: %v", err)
		return false
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		color.Yellow("⚠️ Failed to write workspace config: %v", err)
		return false
	}
	return true
}

func normalizeWorkspaceRecent(workspaceCfg *config.Workspace, accountName string) bool {
	if workspaceCfg == nil {
		return false
	}
	viewNames := make([]string, 0)
	updated := false
	for _, view := range workspaceCfg.Views {
		viewNames = append(viewNames, view.Name)
		jobNames := make([]string, 0, len(view.Job))
		for _, job := range view.Job {
			jobNames = append(jobNames, job.Name)
			recentChoices := util.FilterRecent(job.RecentChoices, util.BuildAllowSet(job.JobParam.Choices), 3)
			recentBranches := util.FilterRecent(job.RecentBranches, util.BuildAllowSet(job.JobParam.Branch), 3)
			if !slicesEqual(job.RecentChoices, recentChoices) || !slicesEqual(job.RecentBranches, recentBranches) {
				for viewIndex := range workspaceCfg.Views {
					if workspaceCfg.Views[viewIndex].Name != view.Name {
						continue
					}
					for jobIndex := range workspaceCfg.Views[viewIndex].Job {
						if workspaceCfg.Views[viewIndex].Job[jobIndex].Name == job.Name {
							workspaceCfg.Views[viewIndex].Job[jobIndex].RecentChoices = recentChoices
							workspaceCfg.Views[viewIndex].Job[jobIndex].RecentBranches = recentBranches
							updated = true
							break
						}
					}
				}
			}
		}
		recentJobs := util.FilterRecent(view.RecentJobs, util.BuildAllowSet(jobNames), 3)
		if !slicesEqual(view.RecentJobs, recentJobs) {
			for viewIndex := range workspaceCfg.Views {
				if workspaceCfg.Views[viewIndex].Name == view.Name {
					workspaceCfg.Views[viewIndex].RecentJobs = recentJobs
					updated = true
					break
				}
			}
		}
	}
	recentViews := util.FilterRecent(workspaceCfg.RecentViews, util.BuildAllowSet(viewNames), 3)
	if slicesEqual(workspaceCfg.RecentViews, recentViews) && !updated {
		return false
	}
	workspaceCfg.RecentViews = recentViews
	workspacePath := util.GetWorkspaceFilePathByName(accountName)
	return writeWorkspaceFile(workspacePath, *workspaceCfg)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		color.Red(err.Error())
		os.Exit(1)
	}
}

func waitOperation(second time.Duration) {
	if second < 0 {
		second = 0
	}
	s := spinner.New(spinner.CharSets[9], 100*time.Millisecond)
	s.Start()
	time.Sleep(second * time.Second)
	s.Stop()
}

func getBuildNumber(cfg config.JenkinsConfig, queueId string, size int) string {
	if size <= 0 {
		return ""
	}
	buildNumber, err := api.GetBuildNumber(cfg, queueId)
	if err != nil {
		color.Yellow("⚠️ Error getting build number: %v", err)
		return ""
	}
	if buildNumber != "" {
		return buildNumber
	}
	size = size - 1
	waitOperation(3)
	return getBuildNumber(cfg, queueId, size)
}

func init() {
	rootCmd.Flags().String("account", "", "account name")
	rootCmd.Flags().String("view", "", "view name")
	rootCmd.Flags().String("job", "", "job name")
}
