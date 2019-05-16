package cmd

import (
	"sync"

	"github.com/by46/whalefs/task"

	"github.com/spf13/cobra"
)

var (
	cronCmd = &cobra.Command{
		Use:   "cron",
		Short: "whalefs cron job ",
		Run:   runScheduler,
	}
)

func init() {
	rootCmd.AddCommand(cronCmd)
}

func runScheduler(cmd *cobra.Command, args []string) {
	waitGroup := new(sync.WaitGroup)
	waitGroup.Add(1)

	scheduler := task.NewScheduler()
	scheduler.Cron.Start()

	waitGroup.Wait()
}
