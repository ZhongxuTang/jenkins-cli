package cmd

import (
	"github.com/fatih/color"
	"github.com/lemonsoul/jenkins-cli/api"
	"github.com/lemonsoul/jenkins-cli/util"
	"github.com/spf13/cobra"
)

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "stop <jobName> <buildNumber>",
	Long:  `jenkins-cli job stop`,
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
		flag, err := api.Stop(account, args[0], args[1])
		if flag {
			color.Yellow("job [%s] stopped successfully, build number is %s", args[0], args[1])
		}
		if err != nil {
			color.Red("❌ %s", err.Error())
		}
	},
}

func init() {
	rootCmd.AddCommand(stopCmd)
}
