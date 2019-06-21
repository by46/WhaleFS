package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/by46/whalefs/rabbitmq"
	"github.com/by46/whalefs/server"
)

var (
	syncCmd = &cobra.Command{
		Use:   "sync",
		Short: "sync file from whale-fs",
		Run:   runSync,
	}
)

func init() {
	rootCmd.AddCommand(syncCmd)
}

func runSync(cmd *cobra.Command, args []string) {
	config, err := server.BuildConfig()
	if err != nil {
		panic(fmt.Errorf("Load config fatal: %s\n", err))
	}
	consumer := rabbitmq.NewConsumer(config)
	consumer.Run2()
}
