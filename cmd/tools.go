package cmd

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var (
	toolsCmd = &cobra.Command{
		Use:   "tools",
		Short: "tools kits",
		Run:   runTools,
	}
	mode string
)

func init() {
	rootCmd.AddCommand(toolsCmd)
	toolsCmd.Flags().StringVarP(&mode, "mode", "", "extension", "工具名")
}

func runTools(cmd *cobra.Command, args []string) {
	if mode == "extension" {
		fullPath, _ := filepath.Abs(args[0])
		detectExtension(fullPath)
	}
}

func detectExtension(fullPath string) {
	extensions := make(map[string]int)
	queue := []string{fullPath}
	parent := ""
	for len(queue) > 0 {
		parent, queue = queue[0], queue[1:]
		files, err := ioutil.ReadDir(parent)
		if err != nil {
			fmt.Printf("Read dir error: %v", err)
			continue
		}
		for _, file := range (files) {
			if file.IsDir() {
				if strings.ToLower(file.Name()) == "original" {
					continue
				}
				queue = append(queue, filepath.Join(parent, file.Name()))
				continue
			}

			if strings.HasPrefix(file.Name(), ".") {
				continue
			}

			extension := filepath.Ext(file.Name())
			extension = strings.ToLower(extension)
			extensions[extension] += 1
		}
	}
	for key := range extensions {
		fmt.Printf("%s %d\n", key, extensions[key])
	}
}
