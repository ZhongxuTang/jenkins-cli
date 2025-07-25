package cmd

import (
	"os"
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
		cfg := util.GetConfigFile()
		selectView := selectView(cfg)
		selectJob := selectJob(cfg, selectView)
		choices, branchs := api.GetJobPararm(selectJob)

		choicesSelect := util.StrUISelect("Select Choices", choices)
		branchsSelect := util.StrUISelect("Select Branch", branchs)

		if selectView == "" || selectJob == "" || branchsSelect == "" {
			color.White("🚨 Selection incomplete, please try again.")
			return
		}

		queueId := api.BuildWithParameters(selectJob, choicesSelect, branchsSelect)
		if queueId != "" {
			color.White("🎉 Build " + selectJob + choicesSelect + branchsSelect + " success, queue id is " + queueId)
		}
		s := spinner.New(spinner.CharSets[9], 100*time.Millisecond)
		s.Start()
		time.Sleep(4 * time.Second)
		s.Stop()
		buildNumber := api.GetBuildNumber(queueId)
		if buildNumber == "" {
			color.White("♻️ job mabe waiting to run, please check it later")
		} else {
			color.White("🍻 Build " + selectJob + " " + choicesSelect + " " + branchsSelect + " success, build number is " + buildNumber)
		}

	},
}

func selectView(cfg config.JenkinsConfig) string {
	viewNames := make([]string, 0)
	for _, item := range cfg.Views {
		viewNames = append(viewNames, item.Name)
	}
	//viewNames = append(viewNames, "return")
	return util.StrUISelect("Select View", viewNames)
}

func selectJob(cfg config.JenkinsConfig, viewResult string) string {
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
