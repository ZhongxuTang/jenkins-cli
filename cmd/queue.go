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

		queueArray := api.GetQueue()

		queueJobArray := make([]util.QueueSelctItem, 0)

		if len(queueArray) == 0 {
			color.White("🥚  Build queue with no data")
		} else {
			for _, queue := range queueArray {
				color.White("----------------------------------------------------------------------------------")
				color.White("Queue ID:%s\n", queue.Id)
				color.White("Task Name:%s\n", queue.TaskName)
				color.White("Params:%s\n", queue.Params)
				color.White("Blocked:", queue.Blocked)
				color.White("Stuck:", queue.Stuck)
				color.White("In Queue Since:%s\n", queue.InQueueSince)
				color.White("----------------------------------------------------------------------------------")

				queueJobArray = append(queueJobArray, util.QueueSelctItem{Name: queue.TaskName, QueueInfo: queue})
			}
			util.QueueUISelect("Queue", queueJobArray)
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
