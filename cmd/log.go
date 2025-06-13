package cmd

import (
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/lemonsoul/jenkins-cli/api"
	"github.com/spf13/cobra"
)

var logCmd = &cobra.Command{
	Use:   "log",
	Short: "log",
	Long:  `task log`,
	Run: func(cmd *cobra.Command, args []string) {

		if len(args) < 2 {
			color.White("Please provide the job name and build number as arguments.")
			return
		}
		var logText string
		var moreDate bool
		var textSize int
		logText, moreDate, textSize = api.GetTextLog(args[0], args[1], nil)
		printLogLine(logText, 20*time.Millisecond)
		for moreDate {
			//time.Sleep(500 * time.Millisecond)
			logText, moreDate, textSize = api.GetTextLog(args[0], args[1], &textSize)
			printLogLine(logText, 20*time.Millisecond)
		}
	},
}

func init() {
	rootCmd.AddCommand(logCmd)
}

func printLogLine(logText string, delay time.Duration) {
	lines := strings.Split(logText, "\n")
	for _, line := range lines {
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
