package cmd

import (
	"github.com/fatih/color"
	"github.com/lemonsoul/jenkins-cli/api"
	"github.com/spf13/cobra"
)

var queueCmd = &cobra.Command{
	Use:   "queue",
	Short: "queue",
	Long:  `task queue`,
	Run: func(cmd *cobra.Command, args []string) {

		queueArray := api.GetQueue()

		if len(queueArray) == 0 {
			color.White("🥚  Build queue with no data")
		} else {
			for _, queue := range queueArray {
				color.White("----------------------------------------------------------------------------------")
				color.White("Queue ID:", queue.Id)
				color.White("Task Name:", queue.TaskName)
				color.White("Params:", queue.Params)
				color.White("Blocked:", queue.Blocked)
				color.White("Stuck:", queue.Stuck)
				color.White("In Queue Since:", queue.InQueueSince)
				color.White("----------------------------------------------------------------------------------")
			}
		}

		computerArray := api.GetComputer()
		if len(computerArray) == 0 {
			color.White("🥚  No tasks are being built.")
		} else {
			for _, item := range computerArray {
				color.White("🚀  JobName: %s, BuildNumber: %d", item.JobName, item.BuildNumber)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(queueCmd)
}
