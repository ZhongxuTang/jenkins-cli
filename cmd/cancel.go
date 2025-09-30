package cmd

import (
	"github.com/fatih/color"
	"github.com/lemonsoul/jenkins-cli/api"
	"github.com/spf13/cobra"
)

var cancelCmd = &cobra.Command{
	Use:   "cancel",
	Short: "cancel <queueId>",
	Long:  `jenkins-cli job cancel`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			color.White("Please provide the job name and build number as arguments.")
			return
		}
		flag, _ := api.CancelItem(args[0])
		if flag {
			color.Green("✅ Queue item %s cancelled successfully", args[0])
		}
	},
}

func init() {
	rootCmd.AddCommand(cancelCmd)
}
