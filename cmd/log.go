package cmd

import (
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/lemonsoul/jenkins-cli/api"
	"github.com/lemonsoul/jenkins-cli/util"
	"github.com/spf13/cobra"
)

var logCmd = &cobra.Command{
	Use:   "log",
	Short: "log <jobName> <buildNumber>",
	Long:  `task log`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 2 {
			color.White("Please provide the job name and build number as arguments.")
			return
		}
		account, err := util.PickAccount("")
		if err != nil {
			color.Red("❌ Error loading account configuration: %v", err)
			return
		}
		var logText string
		var moreData bool
		var textSize int

		logText, moreData, textSize, err = api.GetTextLog(account, args[0], args[1], nil)
		if err != nil {
			color.Red("❌ Error getting log: %v", err)
			return
		}

		if textSize == -1 || !moreData {
			// log complete
			printLogLine(logText, 20*time.Millisecond)
			color.White("Log output is completed!")
		} else {
			printLogLine(logText, 20*time.Millisecond)
			for moreData {
				logText, moreData, textSize, err = api.GetTextLog(account, args[0], args[1], &textSize)
				if err != nil {
					color.Red("❌ Error getting more log data: %v", err)
					break
				}
				printLogLine(logText, 20*time.Millisecond)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(logCmd)
}

func printLogLine(logText string, delay time.Duration) {
	if logText != "" {
		for line := range strings.SplitSeq(logText, "\n") {
			switch {
			case strings.Contains(line, "INFO"):
				color.White(line)
			case strings.Contains(line, "WARN"):
				color.Yellow(line)
			case strings.Contains(line, "ERROR"):
				color.Red(line)
			default:
				color.White(line)
			}
			time.Sleep(delay)
		}
		color.Unset()

	}
}
