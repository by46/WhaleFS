package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
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
}
