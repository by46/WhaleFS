package cmd

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"

	"github.com/by46/whalefs/server"
)

var (
	initializeCmd = &cobra.Command{
		Use:   "initialize",
		Short: "initialize system",
		Run:   initialize,
	}
	bucketsConfig string
)

func init() {
	rootCmd.AddCommand(initializeCmd)
	initializeCmd.Flags().StringVarP(&bucketsConfig, "buckets", "", "buckets.json", "指定buckets配置的json文件路径")
}

func initialize(cmd *cobra.Command, args []string) {
	config, err := server.BuildConfig()
	if err != nil {
		log.Fatalf("加载配置失败 %v", err)
	}
	meta := server.BuildDao(config.BucketMeta)

	f, err := os.Open(bucketsConfig)
	if err != nil {
		log.Fatalf("打开配置文件失败 %v", err)
	}
	defer func() {
		_ = f.Close()
	}()

	buckets := make(map[string]interface{})
	if err := json.NewDecoder(f).Decode(&buckets); err != nil {
		log.Fatalf("解析json配置文件失败 %v", err)
	}
	for key := range buckets {
		value := buckets[key]
		if err := meta.Set(key, value); err != nil {
			fmt.Printf("添加配置key: %s失败, %v\n", key, err)
		}
	}
}
