package cmd

import (
	"github.com/spf13/cobra"
)

var (
	raftCmd = &cobra.Command{
		Use: "raft",
		Run: runRaft,
	}
)

func init() {
	rootCmd.AddCommand(raftCmd)
}

func runRaft(cmd *cobra.Command, args []string) {
}
