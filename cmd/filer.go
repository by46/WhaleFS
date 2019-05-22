package cmd

import (
	"github.com/spf13/cobra"

	"github.com/by46/whalefs/server"
	"github.com/by46/whalefs/utils"
)

var (
	fileCmd = &cobra.Command{
		Use:   "filer",
		Short: "whalefs gateway ",
		Run:   executeFiler,
	}
	filerCpuProfile string
	filerMemProfile string
)

func init() {
	rootCmd.AddCommand(fileCmd)

	fileCmd.Flags().StringVarP(&filerCpuProfile, "cpuprofile", "", "", "cpu profile output file")
	fileCmd.Flags().StringVarP(&filerMemProfile, "memprofile", "", "", "memory profile output file")
}

func executeFiler(cmd *cobra.Command, args []string) {
	utils.SetupProfiling(filerCpuProfile, filerMemProfile)
	srv := server.NewServer()
	srv.ListenAndServe()
}
