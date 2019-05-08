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
 
 1. 系统高可用, 无数据丢失
 2. 支持上传, 浏览器下载, 支持获取元数据
 3. 支持 bucket
 4. 支持文件过期
 5. 只是客户端缓存(缓存策略, 过期策越)
 6. 支持自定义文件名上传
 7. 支持目录(目前需要讨论需求)
 8. 提供SDK(Java, C#)
 9. 支持表单上传
 10. 支持上传限制(文件类型限制, 大小限制)
 11. 支持图片操作(动态切图, 水印)

目前根据facebook的一篇针对小文件优化的文件系统的论文, 找了一个golang的开源实现, 在此基础上进行开发, 