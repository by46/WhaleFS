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
 2. 支持上传, 浏览器下载, 支持获取元数据, 支持是否覆盖 - p1
 3. 支持 bucket - p1 - done 
 4. 支持文件过期
 5. 只是客户端缓存(缓存策略, 过期策越) - p1
 6. 支持自定义文件名上传 - p1
 7. 支持目录(目前需要讨论需求)
 8. 提供SDK(Java, C#)
 9. 支持表单上传 - p1
 10. 支持上传限制(文件类型限制, 大小限制) - p1
 11. 支持图片操作(动态切图, 水印) - p1
 12. 视频处理(按照帧数来截图)
 13. 在线预览PDF,WORD,EXCEL,图片视频
 14. 可能要考虑异步打包，有的包会比较大，打包时间达到分钟级别 - p1
 15. 支持大文件断点续传
 16. 支持秒传功能
 17. 文件索引支持模糊查找
 18. 权限控制(access key)


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