package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/fatih/color"
	"github.com/lemonsoul/jenkins-cli/api"
	"github.com/lemonsoul/jenkins-cli/config"
	"github.com/lemonsoul/jenkins-cli/util"
	"github.com/manifoldco/promptui"
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

		choicesSelect := baseSelect("Select Choices", choices)
		branchsSelect := baseSelect("Select Branch", branchs)

		if selectView == "" || selectJob == "" || branchsSelect == "" {
			color.White("🚨 Selection incomplete, please try again.")
			return
		}

		queueId := api.BuildWithParameters(selectJob, choicesSelect, branchsSelect)
		if queueId != "" {
			color.White("🎉 Build " + selectJob + choicesSelect + branchsSelect + " success, queue id is " + queueId)
		}
		time.Sleep(5 * time.Second)
		buildNumber := api.GetBuildNumber(queueId)
		if buildNumber == "" {
			color.White("♻️ job mabe waiting to run, please check it later")
		} else {
			color.White("🍻 Build " + selectJob + " " + choicesSelect + " " + branchsSelect + " success, build number is " + buildNumber)
		}

	},
}

func baseSelect(label string, items []string) string {
	if len(items) == 0 {
		//color.Red("No items to select")
		return ""
	}
	fmt.Println(items)
	selectPrompt := promptui.Select{
		Label: label,
		Items: items,
		Size:  10,
	}
	_, result, err := selectPrompt.Run()
	if err != nil {
		color.White("failed to select", err)
	}
	fmt.Println(result)
	return result
}

func selectView(cfg config.JenkinsConfig) string {
	viewNames := make([]string, 0)
	for _, item := range cfg.Views {
		viewNames = append(viewNames, item.Name)
	}
	//viewNames = append(viewNames, "return")
	return baseSelect("Select View", viewNames)
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
	return baseSelect("Select Job", jobNames)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		color.Red(err.Error())
		os.Exit(1)
	}
}
