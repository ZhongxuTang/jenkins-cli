package cmd

import (
	"github.com/fatih/color"
	"github.com/lemonsoul/jenkins-cli/api"
	"github.com/lemonsoul/jenkins-cli/util"
	"github.com/spf13/cobra"
)

var queueCmd = &cobra.Command{
	Use:   "queue",
	Short: "queue",
	Long:  `task queue`,
	Run: func(cmd *cobra.Command, args []string) {
		queueArray, err := api.GetQueue()
		if err != nil {
			color.Red("❌ Error getting queue information: %v", err)
			return
		}

		queueJobArray := make([]util.QueueSelectItem, 0)

		if len(queueArray) == 0 {
			color.White("🥚  Build queue with no data")
		} else {
			for _, queue := range queueArray {
				color.White("----------------------------------------------------------------------------------")
				color.White("Queue ID:%s\n", queue.Id)
				color.White("Task Name:%s\n", queue.TaskName)
				color.White("Params:%s\n", queue.Params)
				color.White("Blocked:%t\n", queue.Blocked)
				color.White("Stuck:%t\n", queue.Stuck)
				color.White("In Queue Since:%s\n", queue.InQueueSince)
				color.White("----------------------------------------------------------------------------------")

				queueJobArray = append(queueJobArray, util.QueueSelectItem{Name: queue.TaskName, QueueInfo: queue})
			}
			util.QueueUISelect("Queue", queueJobArray)
		}

		computerArray, err := api.GetComputer()
		if err != nil {
			color.Red("❌ Error getting computer information: %v", err)
			return
		}

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
