package cmd

import (
	"fmt"
	"sync"

	"github.com/pkg/errors"
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

	syncWorkerCount uint8
)

func init() {
	rootCmd.AddCommand(syncCmd)

	syncCmd.Flags().Uint8VarP(&syncWorkerCount, "count", "", 10, "count")
}

func runSync(cmd *cobra.Command, args []string) {
	config, err := server.BuildConfig()
	if err != nil {
		panic(fmt.Errorf("Load config fatal: %s\n", errors.WithStack(err)))
	}
	wg := &sync.WaitGroup{}
	for i := 0; i < int(syncWorkerCount); i++ {
		wg.Add(1)
		consumer := rabbitmq.NewConsumer(config)
		go consumer.Run()
	}
	wg.Wait()
}
