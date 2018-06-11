package cmd

import (
	"whalefs/server"

	"github.com/spf13/cobra"
)

var (
	fileCmd = &cobra.Command{
		Use:   "filer",
		Short: "whalefs gateway ",
		Run:   executeFiler,
	}
)

func init() {
	rootCmd.AddCommand(fileCmd)
}

func executeFiler(cmd *cobra.Command, args []string) {
	srv := server.NewServer()
	srv.ListenAndServe()
}
