package cmd

import (
	"strings"

	"github.com/spf13/cobra"

	"github.com/by46/whalefs/migration"
)

const (
	Comma = ","
)

var (
	migrateCmd = &cobra.Command{
		Use:   "migrate",
		Short: "migrate fs data",
		Run:   migrate,
	}
	location    string
	target      string
	includes    string
	excludes    string
	workerCount uint8
)

func init() {
	rootCmd.AddCommand(migrateCmd)
	migrateCmd.Flags().StringVarP(&location, "location", "", "", "old file system folder")
	migrateCmd.Flags().StringVarP(&target, "target", "", "", "file system url, ie http://172.16.1.9:8089")
	migrateCmd.Flags().Uint8VarP(&workerCount, "count", "", 10, "where")
	migrateCmd.Flags().StringVarP(&includes, "includes", "", "", "which app name should migrate, separate by ','")
	migrateCmd.Flags().StringVarP(&excludes, "excludes", "", "", "which app name should be ignore, separate by ','")
}

func splitByComma(name string) []string {
	if name == "" {
		return nil
	}
	segments := strings.Split(name, Comma)
	if len(segments) == 0 {
		return nil
	}
	for i := 0; i < len(segments); i++ {
		segments[i] = strings.ToLower(strings.TrimSpace(segments[i]))
	}
	return segments
}

func migrate(cmd *cobra.Command, args []string) {
	options := &migration.MigrationOptions{
		Location:    location,
		Includes:    splitByComma(includes),
		Excludes:    splitByComma(excludes),
		Target:      target,
		WorkerCount: workerCount,
	}
	migration.Migrate(options)
}
