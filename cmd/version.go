package cmd

import (
	"github.com/coreos/etcd/raft"
	"github.com/spf13/cobra"
	"fmt"
)

var (
	versionCmd = &cobra.Command{
		Use:   "version",
		Short: "show version",
		Run:   runVersion,
	}
)

func init() {
	rootCmd.AddCommand(versionCmd)
}

func runVersion(cmd *cobra.Command, args []string) {
	// TODO(benjamin): add release version
	fmt.Printf("0.0.1")

	storage := raft.NewMemoryStorage()
	c := &raft.Config{
		ID:              0x01,
		ElectionTick:    10,
		HeartbeatTick:   1,
		Storage:         storage,
		MaxSizePerMsg:   4096,
		MaxInflightMsgs: 256,
	}

	raft.StartNode(c, []raft.Peer{{ID: 0x02}, {ID: 0x03}})
}
