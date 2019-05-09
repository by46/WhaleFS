package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	rootCmd = &cobra.Command{
		Use:   "whalefs",
		Short: "whalefs is a distribution file system",
		Long: `
		whalefs is a distribution file system`,
	}
)

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
