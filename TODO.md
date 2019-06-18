 # TODO

 - archive upload
 - accept-encoding support - done
 - accept-encoding min-size - done
 - user management
 - bucket management
 - revalidation - done
 - [issue] download with query not found - done
 - [issue] mime type detect
 - favico revalidation
 - image support
 
 
 ## 分布式文件系统特性

 1. 系统高可用, 无数据丢失 - p1 - done
 3. 完全兼容旧文件系统的下载接口, 尽量兼容上传接口(差异通过SDK来解决) - p1 - done
 4. 支持 bucket - p1 - done 
 5. 支持上传/下载 - p1 - done
 6. 支持上传限制(文件类型限制, 大小限制, 宽高限制, 宽高比限制) - p1 - done
 8. 可能要考虑异步打包，有的包会比较大，打包时间达到分钟级别 - p1
 9. 支持图片操作(动态切图, 水印) - p1 - done
 10. 只是客户端缓存(缓存策略, 过期策越) - p1 - done
 11. 支持自定义文件名上传(例如 /benjamin/thumbnail/product/demo.jpg, 这样的路径) - p1 - done
 13. 提供SDK(Java, C#) - p1
 19. 通过外部URL上传资源 - p1 - done
 2. 简单部署 - done
 7. 支持大文件断点续传
 12. 支持目录(目前需要讨论需求) - done
 14. 视频处理(按照帧数来截图)
 15. 在线预览PDF,WORD,EXCEL,图片视频
 16. 支持秒传功能
 17. 文件索引支持模糊查找
 18. 权限控制(access key)
 20. 支持文件过期 - done
 21. 简单portal(管理bucket, 系统运行状态监控, 简单运维工作)
 22. 支持linux 文件系统原生挂载


1. ApiUploadHandler.ashx - done
2. BatchMergePdfHandler.ashx
3. DownloadHandler.ashx - done
4. DownloadSaveServerHandler.ashx - done
5. SliceUploadHandler.ashx 
6. UploadHandler.ashx - done
7. 错误提示多语言 - done
8. 水印支持文字
9. migrate 迁移支持大文件
10. multi-chunk file etag 如何生成
11. 支持默认图
12. 支持单图多水印
13. 如果通过remote 下载时, 如果文件过大,需要转换为multi-chunks
14. 接口兼容问题(handler存在多个系统中时, 有可能失败) - done
15. bucket 更新接口
16. resize cache 如何过期


80,110,160,240,220,600,375,750,1125

{
    "name": "p80",
    "width": 80,
    "height": 60,
    "mode": "stretch"
},
{
    "name": "p110",
    "width": 110,
    "height": 82,
    "mode": "stretch"
},
{
    "name": "p160",
    "width": 160,
    "height": 120,
    "mode": "stretch"
},
{
 "name": "p220",
 "width": 220,
 "height": 165,
 "mode": "stretch"
},
{
 "name": "p600",
 "width": 600,
 "height": 450,
 "mode": "stretch"
},
 {
  "name": "p375",
  "width": 375,
  "height": 281,
  "mode": "stretch"
 },
   {
    "name": "p750",
    "width": 750,
    "height": 563,
    "mode": "stretch"
   },
       {
        "name": "p1125",
        "width": 750,
        "height": 844,
        "mode": "stretch"
       }


目前根据facebook的一篇针对小文件优化的文件系统的论文, 找了一个golang的开源实现, 在此基础上进行开发, https://github.com/chrislusf/seaweedfs

### 打包下载
比如，一个压缩包包含的文件及结构：
{
  file: '',
  files:[{ source: "/benjamin/xxx/xxx/hello2.jpg", target: "供应商100/招标文件/xxx招标.doc" },
  { source: "100_101.doc", target: "供应商100/招标文件1/xxx招标.doc" },
  { source: "101_100.doc", target: "供应商101\招标文件\xxx招标.doc" },
  { source: "101_101.doc", target: "供应商101\投标文件\xxx投标.doc" }]
}

/usr/bin/kubelet  --bootstrap-kubeconfig=/etc/kubernetes/bootstrap-kubelet.conf --kubeconfig=/etc/kubernetes/kubelet.conf --config=/var/lib/kubelet/config.yaml --cgroup-driver=systemd --network-plugin=cni --pod-infra-container-image=k8s.gcr.io/pause:3.1