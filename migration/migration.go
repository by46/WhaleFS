package migration

import (
	"context"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"github.com/pkg/errors"

	"github.com/by46/whalefs/client"
	"github.com/by46/whalefs/common"
	"github.com/by46/whalefs/utils"
)

const (
	NameOriginal = "original"
)

var (
	ReIsProduct = regexp.MustCompile("^P[0-9]+$")
)

type MigrationOptions struct {
	Location     string
	Target       string
	Includes     []string
	Excludes     []string
	WorkerCount  uint8
	Override     bool
	AppId        string
	AppSecretKey string
}

type migration struct {
	options         *MigrationOptions
	c               chan string
	client          client.Client
	cache           *Cache
	productAppNames map[string]bool
	common.Logger
}

func Migrate(options *MigrationOptions) {
	if utils.DirExists(options.Location) == false {
		log.Fatalf("需要迁移的文件夹不存在, 或者没有读权限 %s\n", options.Location)
	}

	location, err := filepath.Abs(options.Location)
	if err != nil {
		log.Fatalf("读取需要迁移的文件夹 %s %v\n", options.Location, err)
	}
	options.Location = location

	m := &migration{
		options: options,
		c:       make(chan string, 10),
		cache:   NewCache(""),
		client: client.NewClient(&client.ClientOptions{
			Base: options.Target,
		}),
		Logger: utils.BuildLogger("logs", "INFO"),
	}
	m.start()
}

func (m *migration) start() {
	ctx, _ := context.WithCancel(context.Background())
	wg := new(sync.WaitGroup)
	wg.Add(int(m.options.WorkerCount))
	for i := uint8(0); i < m.options.WorkerCount; i++ {
		go m.upload(ctx, wg, i)
	}
	m.listFiles(ctx)
	close(m.c)
	wg.Wait()
	_ = m.cache.db.Close()
}

func (m *migration) prepare() []string {
	options := m.options
	if len(options.Includes) > 0 {
		return options.Includes
	}
	names := make([]string, 0)
	if entities, err := ioutil.ReadDir(options.Location); err != nil {
		m.Fatalf("读取迁移路径[%s]出现错误, %v", options.Location, err)
		return nil
	} else {
		ignores := make(map[string]bool)
		for _, ignore := range options.Excludes {
			ignores[ignore] = true
		}
		for _, entity := range entities {
			if entity.IsDir() == false {
				continue
			}
			if _, exists := ignores[strings.ToLower(entity.Name())]; !exists {
				names = append(names, entity.Name())
			}
		}
	}
	return names;
}

func (m *migration) detectAppType(names []string) map[string]bool {
	return map[string]bool{"pdt": true}

	//options := m.options
	//mapping := make(map[string]bool)
	//for _, name := range names {
	//	if m.isProduct(filepath.Join(options.Location, name)) {
	//		mapping[strings.ToLower(name)] = true
	//	}
	//}
	//return mapping
}

func (m *migration) isProduct(fullPath string) bool {
	if files, err := ioutil.ReadDir(fullPath); err != nil {
		m.Infof("读取文件夹[%s]失败, %v", errors.WithStack(err))
	} else {
		for _, file := range files {
			if ReIsProduct.MatchString(file.Name()) {
				return true
			}
		}
	}
	return false
}

func (m *migration) splitBucketNameAndFileName(name string) (bucketName, filename string) {
	name = strings.ToLower(name)
	segments := strings.Split(name, string(os.PathSeparator))
	bucketName = segments[0]
	if _, exists := m.productAppNames[strings.ToLower(bucketName)]; exists {
		if NameOriginal == strings.ToLower(segments[1]) {
			filename = filepath.Base(name)
			return
		}
	} else {
		if NameOriginal == strings.ToLower(segments[1]) {
			filename = path.Join(segments[2:]...)
			return
		}
	}
	filename = path.Join(segments[1:]...)
	return
}

func (m *migration) listFiles(ctx context.Context) {
	var parent string
	options := m.options
	queue := m.prepare()
	m.productAppNames = m.detectAppType(queue)
	for len(queue) > 0 {
		parent, queue = queue[0], queue[1:]
		files, err := ioutil.ReadDir(filepath.Join(options.Location, parent))
		if err != nil {
			m.Infof("读取文件件信息失败, %v\n", err)
			continue
		}
		for _, file := range files {
			if strings.HasPrefix(file.Name(), ".") {
				continue
			}
			filename := filepath.Join(parent, file.Name())
			if file.IsDir() {
				if !strings.ContainsRune(parent, os.PathSeparator) && ReIsProduct.MatchString(file.Name()) {
					continue
				}
				queue = append(queue, filename)
			} else {
				m.c <- filename
			}
		}
	}
}

func (m *migration) upload(ctx context.Context, wg *sync.WaitGroup, workerNo uint8) {
	options := m.options
	m.Infof("Worker[%d]:启动\n", workerNo)
	for {
		select {
		case <-ctx.Done():
			m.Infof("worker[%d]:终止\n", workerNo)
			return
		case filename, more := <-m.c:
			if !more {
				wg.Done()
				return
			}
			fullPath := filepath.Join(options.Location, filename)
			if utils.FileExists(fullPath) == false {
				m.Infof("worker[%d]: 文件不存在 %s\n", workerNo, fullPath)
				continue
			}
			if m.cache.Exists(filename) {
				m.Infof("worker[%d], 文件已经上传不需要再次上传 %s", workerNo, fullPath)
				continue
			}

			m.uploadFile(ctx, fullPath, filename, workerNo)
		}
	}
}

func (m *migration) uploadFile(ctx context.Context, fullPath, filename string, workerNo uint8) {
	f, err := os.Open(fullPath)
	if err != nil {
		m.Errorf("worker[%d]: 读取文件失败 %s\n", workerNo, fullPath)
		return
	}
	defer func() {
		_ = f.Close()
	}()
	bucketName, key := m.splitBucketNameAndFileName(filename)
	opt := &client.Options{
		Bucket:   bucketName,
		FileName: key,
		Override: m.options.Override,
		Content:  f,
	}
	m.Infof("worker[%d]: 上传文件 %s, bucket: %s, key: %s ", workerNo, fullPath, bucketName, key)
	if _, err = m.client.Upload(ctx, opt); err != nil {
		m.Errorf("worker[%d]: 上传文件失败 %s %v\n", workerNo, fullPath, err)
	} else {
		m.cache.Put(filename)
	}
}
