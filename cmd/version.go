package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/by46/whalefs/constant"
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
	fmt.Println(constant.VERSION)
}
