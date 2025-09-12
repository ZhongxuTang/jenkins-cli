package cmd

import (
	"strconv"
	"strings"
	"time"

	"github.com/briandowns/spinner"
	"github.com/fatih/color"
	"github.com/lemonsoul/jenkins-cli/api"
	"github.com/lemonsoul/jenkins-cli/config"
	"github.com/spf13/cobra"
)

var stagesCmd = &cobra.Command{
	Use:   "stages",
	Short: "stages",
	Long:  `task stages`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 2 {
			color.White("Please provide the job name and build number as arguments.")
			return
		}
		//color.White("Fetching stages for job:", args[0], "and build number:", args[1])
		//pipelineInfo := api.GetPipelineConfig(args[0])
		wFDescribe, err := api.GetWFDescribe(args[0], args[1])
		if err != nil {
			color.Red("❌ Error getting workflow description: %v", err)
			return
		}

		color.Cyan("📦 Project Name:%s", args[0])
		color.Cyan("🔁 Build Number:%s", wFDescribe.QueueId)
		color.Cyan("🕒 Begin Time: %s", time.UnixMilli(wFDescribe.StartTimeMillis).Format("2006-01-02 15:04:05"))

		if strings.Compare(wFDescribe.Status, "SUCCESS") == 0 {
			color.Cyan("⏳ Status:%s", "✅ Build is sucess")
			//return
		} else if strings.Compare(wFDescribe.Status, "FAILURE") == 0 {
			color.Red("⏳ Status:%s", "❌ Build is failure")
			return
		} else if strings.Compare(wFDescribe.Status, "ABORTED") == 0 {
			color.Yellow("⏳ Status:%s", "🛑 Build is aborted")
			return
		} else if strings.Compare(wFDescribe.Status, "IN_PROGRESS") == 0 {
			color.Cyan("⏳ Status:%s", "🔄 Build is in progress")
		}

		// build stages progress bar
		pipelineInfo, _ := api.GetPipelineConfig(args[0])

		complate := false
		for index, stage := range pipelineInfo.Stages {
			s := spinner.New(spinner.CharSets[9], 100*time.Millisecond)
			s.Suffix = "[" + strconv.Itoa((index + 1)) + "/" + strconv.Itoa(len(pipelineInfo.Stages)) + "]" + stage.Name
			s.Start()
			wFDescribe := config.WFDescribe{}
		Loop:
			for {
				if !complate {
					wFDescribe, _ = api.GetWFDescribe(args[0], args[1])
				}
				if complate || strings.Compare(wFDescribe.Status, "SUCCESS") == 0 {
					complate = true
					s.Stop()
					color.Cyan("✔ [" + strconv.Itoa((index + 1)) + "/" + strconv.Itoa(len(pipelineInfo.Stages)) + "]" + stage.Name)
					break
				}

				for _, wFstages := range wFDescribe.Stages {
					if wFstages.Id == stage.Id && wFstages.Status == "SUCCESS" {
						s.Stop()
						color.Cyan("✔ [" + strconv.Itoa((index + 1)) + "/" + strconv.Itoa(len(pipelineInfo.Stages)) + "]" + stage.Name +
							(time.Duration((wFstages.DurationMillis / int64(1000))) * time.Second).String())
						break Loop
					}
				}
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(stagesCmd)
}
