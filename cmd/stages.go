package cmd

import (
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/lemonsoul/jenkins-cli/api"
	"github.com/lemonsoul/jenkins-cli/config"
	"github.com/schollz/progressbar/v3"
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
		wFDescribe := api.GetWFDescribe(args[0], args[1])

		color.Green("📦 Project Name:%s", args[0])
		color.Green("🔁 Build Number:%s", wFDescribe.QueueId)
		color.Green("🕒 Begin Time: %s", time.UnixMilli(wFDescribe.StartTimeMillis).Format("2006-01-02 15:04:05"))

		if strings.Compare(wFDescribe.Status, "SUCCESS") == 0 {
			color.Green("⏳ Status:%s", "✅ Build is SUCCESS")
			return
		} else if strings.Compare(wFDescribe.Status, "FAILURE") == 0 {
			color.Red("⏳ Status:%s", "❌ Build is FAILURE")
			return
		} else if strings.Compare(wFDescribe.Status, "ABORTED") == 0 {
			color.Yellow("⏳ Status:%s", "🛑 Build is ABORTED")
			return
		} else if strings.Compare(wFDescribe.Status, "IN_PROGRESS") == 0 {
			color.Green("⏳ Status:%s", "🔄 Build is IN_PROGRESS")
		}

		// build stages progress bar
		/*bar := stageProgressbars()
		processFloat := progress(len(pipelineInfo.Stages), &wFDescribe)
		bar.Add(int(processFloat * 100))
		for {
			wFDescribe := api.GetWFDescribe(args[0], args[1])
			processFloatTemp := progress(len(pipelineInfo.Stages), &wFDescribe)
			if processFloatTemp > processFloat {
				bar.Add(int((processFloatTemp - processFloat) * 100))
			}
			time.Sleep(2 * time.Second)

			if strings.Compare(wFDescribe.Status, "IN_PROGRESS") != 0 {
				break
			}
		}*/
	},
}

func init() {
	rootCmd.AddCommand(stagesCmd)
}

func progress(totalCount int, wFDescribe *config.WFDescribe) float64 {
	complateCount := 0
	runningCount := 0
	for _, stage := range wFDescribe.Stages {
		if stage.Status == "IN_PROGRESS" {
			runningCount++
			continue
		}
		if stage.Status == "SUCCESS" {
			complateCount++
			continue
		}
	}
	baseProgress := float64(complateCount) / float64(totalCount)
	inProgressBonus := 0.0
	if runningCount > 0 {
		inProgressBonus = 0.5 / float64(totalCount)
	}
	return baseProgress + inProgressBonus
}

func stageProgressbars() *progressbar.ProgressBar {
	return progressbar.NewOptions(100,
		progressbar.OptionEnableColorCodes(true),
		progressbar.OptionSetWidth(15),
		progressbar.OptionSetDescription("building..."),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "[green]=[reset]",
			SaucerHead:    "[green]>[reset]",
			SaucerPadding: " ",
			BarStart:      "[",
			BarEnd:        "]",
		}))
}
