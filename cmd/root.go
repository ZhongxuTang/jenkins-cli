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
)

var rootCmd = &cobra.Command{
	Use:   "jenkins-cli",
	Short: "jcli",
	Long:  `jenkins-cli is a command line tool for managing Jenkins jobs and builds.`,
	Run: func(cmd *cobra.Command, args []string) {
		workspaceCfg, err := util.GetWorkspaceFile()
		if err != nil {
			color.Red("❌ Error loading workspace configuration: %v", err)
			return
		}

		selectView := selectView(workspaceCfg)
		if selectView == "" {
			color.Red("🚨 No view selected, please try again.")
			return
		}

		selectJob := selectJob(workspaceCfg, selectView)
		if selectJob == "" {
			color.Red("🚨 No job selected, please try again.")
			return
		}

		choices, branchs, err := api.GetJobPararm(selectJob)
		if err != nil {
			color.Red("❌ Error getting job parameters: %v", err)
			return
		}

		choicesSelect := util.StrUISelect("Select Choices", choices)
		branchsSelect := util.StrUISelect("Select Branch", branchs)

		if selectView == "" || selectJob == "" || branchsSelect == "" {
			color.Red("🚨 Selection incomplete, please try again.")
			return
		}

		queueId, err := api.BuildWithParameters(selectJob, choicesSelect, branchsSelect)
		if err != nil {
			color.Red("❌ Error starting build: %v", err)
			return
		}

		if queueId != "" {
			color.Cyan("🎉 Build " + selectJob + choicesSelect + branchsSelect + " success, queue id is " + queueId)
		}
		waitOperation(4)
		buildNumber := getBudilNumber(queueId, 8)
		if buildNumber == "" {
			color.Yellow("♻️ job maybe waiting to run, please check it later")
		} else {
			color.Cyan("🍻 Build " + selectJob + " " + choicesSelect + " " + branchsSelect + " success, build number is " + buildNumber)
		}
		buildInfo, err := api.GetBuildStatus(selectJob, buildNumber)
		if err != nil {
			color.Yellow("⚠️ Error getting build status: %v", err)
			return
		}
		if len(buildInfo.ChangeSets) > 0 {
			for index, item := range buildInfo.ChangeSets {
				color.Cyan(strconv.Itoa(index+1) + "、 " + item.Comment + " by " + item.AuthorFullName)
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
	//viewNames = append(viewNames, "return")
	return util.StrUISelect("Select View", viewNames)
}

func selectJob(cfg config.Workspace, viewResult string) string {
	jobNames := make([]string, 0)
	for _, item := range cfg.Views {
		if item.Name == viewResult {
			for _, job := range item.Job {
				jobNames = append(jobNames, job.Name)
			}
			break
		}
	}
	return util.StrUISelect("Select Job", jobNames)
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

func getBudilNumber(queueId string, size int) string {
	if size <= 0 {
		return ""
	}
	buildNumber, err := api.GetBuildNumber(queueId)
	if err != nil {
		color.Yellow("⚠️ Error getting build number: %v", err)
		return ""
	}
	if buildNumber != "" {
		return buildNumber
	}
	size = size - 1
	waitOperation(3)
	return getBudilNumber(queueId, size)
}
