package cmd

import (
	"github.com/spf13/cobra"
	"fmt"
)

var (
	versionCmd = &cobra.Command{
		Use:   "version",
		Short: "show version",
		Run: func(cmd *cobra.Command, args []string) {
			// TODO(benjamin): add release version
			fmt.Printf("0.0.1")
		},
	}
)

func init() {
	rootCmd.AddCommand(versionCmd)
}
