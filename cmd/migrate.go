package cmd

import (
	"github.com/spf13/cobra"

	"github.com/by46/whalefs/migration"
)

var (
	migrateCmd = &cobra.Command{
		Use:   "migrate",
		Short: "migrate fs data",
		Run:   migrate,
	}
	location    string
	target      string
	isImage     bool
	workerCount uint8
)

func init() {
	rootCmd.AddCommand(migrateCmd)
	migrateCmd.Flags().StringVarP(&location, "location", "", "", "old file system folder")
	migrateCmd.Flags().StringVarP(&target, "target", "", "", "file system url, ie http://172.16.1.9:8089")
	migrateCmd.Flags().BoolVarP(&isImage, "is-image", "", false, "whether image")
	migrateCmd.Flags().Uint8VarP(&workerCount, "count", "", 10, "where")
}

func migrate(cmd *cobra.Command, args []string) {
	options := &migration.MigrationOptions{
		Location:    location,
		Target:      target,
		IsImage:     isImage,
		WorkerCount: workerCount,
	}
	migration.Migrate(options)
}
