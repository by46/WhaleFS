package cmd

import (
	"whalefs/raftexample"

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
	raftexample.RaftDemo()
}
