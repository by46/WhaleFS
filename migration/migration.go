package migration

import (
	"context"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"

	"github.com/by46/whalefs/client"
	"github.com/by46/whalefs/utils"
)

type MigrationOptions struct {
	Location    string
	Target      string
	IsImage     bool
	WorkerCount uint8
	c           chan string
	client      client.Client
	name        string
	cache       *Cache
}

func Migrate(options *MigrationOptions) {
	if utils.DirExists(options.Location) == false {
		log.Fatalf("需要迁移的文件夹不存在, 或者没有读权限 %s\n", options.Location)
	}

	location, err := filepath.Abs(options.Location)
	if err != nil {
		log.Fatalf("读取需要迁移的文件夹 %s %v\n", options.Location, err)
	}
	options.Location, options.name = filepath.Split(location)
	options.client = client.NewClient(&client.ClientOptions{
		Base: options.Target,
	})
	options.cache = NewCache("")
	options.c = make(chan string, 10)
	ctx, _ := context.WithCancel(context.Background())
	wg := new(sync.WaitGroup)
	wg.Add(int(options.WorkerCount))
	for i := uint8(0); i < options.WorkerCount; i++ {
		go upload(ctx, wg, options, i)
	}
	listFiles(ctx, options)
	close(options.c)
	wg.Wait()
	_ = options.cache.db.Close()
}

func listFiles(ctx context.Context, options *MigrationOptions) {
	var parent string
	queue := make([]string, 0)
	queue = append(queue, options.name)
	for len(queue) > 0 {
		parent, queue = queue[0], queue[1:]
		files, err := ioutil.ReadDir(filepath.Join(options.Location, parent))
		if err != nil {
			log.Printf("读取文件件信息失败, %v\n", err)
			continue
		}
		for _, file := range files {
			if strings.HasPrefix(file.Name(), ".") {
				continue
			}
			filename := filepath.Join(parent, file.Name())
			filename = strings.ToLower(filename)
			if file.IsDir() {
				queue = append(queue, filename)
			} else {
				options.c <- filename
			}
		}
	}
}

func buildFileName(filename string, isImage bool) string {
	if isImage {
		return filepath.Base(filename)
	}
	segments := strings.Split(filename, string(os.PathSeparator))
	return path.Join(segments[1:]...)
}

func upload(ctx context.Context, wg *sync.WaitGroup, options *MigrationOptions, workerNo uint8) {
	log.Printf("Worker[%d]:启动\n", workerNo)
	for {
		select {
		case <-ctx.Done():
			log.Printf("worker[%d]:终止\n", workerNo)
			return
		case filename, more := <-options.c:
			if !more {
				wg.Done()
				return
			}
			fullPath := filepath.Join(options.Location, filename)
			if utils.FileExists(fullPath) == false {
				log.Printf("worker[%d]: 文件不存在 %s\n", workerNo, fullPath)
				continue
			}
			if options.cache.Exists(filename) {
				log.Printf("worker[%d], 文件已经上传不需要再次上传 %s", workerNo, fullPath)
				continue
			}
			f, err := os.Open(fullPath)
			if err != nil {
				log.Printf("worker[%d]: 读取文件失败 %s\n", workerNo, fullPath)
				continue
			}
			opt := &client.Options{
				Bucket:   options.name,
				FileName: buildFileName(filename, options.IsImage),
				Override: false,
				Content:  f,
			}
			log.Printf("worker[%d]: 上传文件 %s", workerNo, fullPath)
			if _, err = options.client.Upload(ctx, opt); err != nil {
				log.Printf("worker[%d]: 上传文件失败 %s %v\n", workerNo, fullPath, err)
			} else {
				options.cache.Put(filename)
			}
		}
	}
}
