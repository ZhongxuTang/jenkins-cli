package cmd

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/briandowns/spinner"
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

type viewJobsResult struct {
	jobNames []string
	err      error
}

type paramResult struct {
	choices  []string
	branches []string
	err      error
}

type jobRef struct {
	viewIdx int
	jobIdx  int
	jobName string
}

const concurrency = 5

func syncWorkspaceForAccount(account config.JenkinsConfig) error {
	cfg, err := util.GetWorkspaceFile(account.Name)
	if err != nil {
		color.Yellow("⚠️ Workspace file not found, creating new one...")
		cfg = config.Workspace{Views: make([]config.View, 0)}
	}

	// Phase 1: Get all view names
	viewNames, err := api.GetViews(account)
	if err != nil {
		return err
	}

	oldCfg := cfg
	cfg.Views = make([]config.View, 0)
	cfg.RecentViews = util.FilterRecent(cfg.RecentViews, util.BuildAllowSet(viewNames), 3)

	// Phase 2: Fetch view jobs concurrently with spinner
	viewJobs := make([]viewJobsResult, len(viewNames))
	sem := make(chan struct{}, concurrency)
	var wg sync.WaitGroup

	if len(viewNames) > 0 {
		s := spinner.New(spinner.CharSets[9], 100*time.Millisecond)
		s.Suffix = fmt.Sprintf(" Syncing views... [0/%d]", len(viewNames))
		s.Start()

		var completed int
		var progressMu sync.Mutex

		for i, viewName := range viewNames {
			sem <- struct{}{}
			wg.Add(1)
			go func(idx int, vn string) {
				defer wg.Done()
				defer func() { <-sem }()
				jobNames, err := api.GetViewJob(account, vn)
				viewJobs[idx] = viewJobsResult{jobNames: jobNames, err: err}

				progressMu.Lock()
				completed++
				s.Suffix = fmt.Sprintf(" Syncing views... [%d/%d]", completed, len(viewNames))
				progressMu.Unlock()
			}(i, viewName)
		}
		wg.Wait()
		s.Stop()
	}

	for vi, vjr := range viewJobs {
		if vjr.err != nil {
			color.Yellow("⚠️ Error getting jobs for view %s: %v", viewNames[vi], vjr.err)
		}
	}

	// Phase 3: Build flat job list and pre-allocate param storage
	var allJobs []jobRef
	jobParams := make([][]paramResult, len(viewNames))
	for vi, vjr := range viewJobs {
		if vjr.err != nil {
			continue
		}
		jobParams[vi] = make([]paramResult, len(vjr.jobNames))
		for ji, jobName := range vjr.jobNames {
			allJobs = append(allJobs, jobRef{viewIdx: vi, jobIdx: ji, jobName: jobName})
		}
	}

	totalJobs := len(allJobs)

	// Phase 4: Fetch job params concurrently with spinner
	if totalJobs > 0 {
		s := spinner.New(spinner.CharSets[9], 100*time.Millisecond)
		s.Suffix = fmt.Sprintf(" Syncing jobs... [0/%d]", totalJobs)
		s.Start()

		var completed int
		var progressMu sync.Mutex

		for _, ref := range allJobs {
			sem <- struct{}{}
			wg.Add(1)
			go func(r jobRef) {
				defer wg.Done()
				defer func() { <-sem }()
				choices, branches, err := api.GetJobParams(account, r.jobName)
				jobParams[r.viewIdx][r.jobIdx] = paramResult{choices: choices, branches: branches, err: err}

				progressMu.Lock()
				completed++
				s.Suffix = fmt.Sprintf(" Syncing jobs... [%d/%d]", completed, totalJobs)
				progressMu.Unlock()
			}(ref)
		}
		wg.Wait()
		s.Stop()

		for _, ref := range allJobs {
			pr := jobParams[ref.viewIdx][ref.jobIdx]
			if pr.err != nil {
				color.Yellow("⚠️ Error getting job params for %s: %v", ref.jobName, pr.err)
			}
		}
	}

	// Phase 5: Ordered assembly
	for vi, vjr := range viewJobs {
		if vjr.err != nil {
			continue
		}
		view := config.View{}
		view.Name = viewNames[vi]
		view.Job = make([]config.Job, 0)
		view.RecentJobs = filterViewRecentJobs(oldCfg, viewNames[vi], vjr.jobNames)

		for ji, jobName := range vjr.jobNames {
			pr := jobParams[vi][ji]
			jobParam := config.JobParam{}
			existingJob, ok := findJob(oldCfg, viewNames[vi], jobName)

			if pr.err != nil {
				if ok {
					jobParam = existingJob.JobParam
				}
			} else {
				jobParam.Choices = pr.choices
				jobParam.Branch = pr.branches
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

	// Phase 6: Marshal and write YAML
	data, err := yaml.Marshal(&cfg)
	if err != nil {
		return err
	}

	workspacePath := util.GetWorkspaceFilePathByName(account.Name)
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
