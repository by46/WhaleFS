# 分布式文件系统
## 环境
目前分布式文件系统有两套环境: QA 和 PRD

### QA
服务地址： http://oss.yzw.cn.qa http://oss-portal.yzw.cn.qa

portal的测试账号test/admin

由于这几个域名都是自定义域名，所以需要在你的local DNS中添加解析规则

Windows：C:\Windows\system32\drivers\etc\hosts, 规则如下：
```text
172.16.0.143 oss.yzw.cn.qa
172.16.0.143 oss-portal.yzw.cn.qa
```

MacOS: /etc/hosts, 规则如下：
```text
172.16.0.143 oss.yzw.cn.qa
172.16.0.143 oss-portal.yzw.cn.qa
```

### PRD
PRD环境使用的是一个外部域名， 可以直接访问
服务地址： https://oss.yzw.cn

但是Portal需要通过VPN访问, http://oss-portal.yzw.cn

portal的测试账号test/admin

Windows：C:\Windows\system32\drivers\etc\hosts, 规则如下：

```text
172.168.220.65 oss-portal.yzw.cn
```

MacOS: /etc/hosts, 规则如下：

```text
172.168.220.65 oss-portal.yzw.cn
```

### 

## API

API接口有两部分组成：

- 老集采系统原来接口， 目前已经基本兼容
- 新系统的接口， 主要是增强原有文件系统的一些未显示的功能

### UploadHandler.ashx
上传接口



### DownloadSaveServerHandler.ashx
通过URL下载文件并保存到系统中



### DownloadHandler.ashx
下载接口



### ApiUploadHandler.ashx
应该是指定上传到tender app里面



### BatchMergePdfHandler.ashx
多个PDF合并成一个PDF



### SliceUploadHandler.ashx
分块上传大文件



### BatchDownloadHandler.ashx 

批量下载文件，生成tar或这zip格式的文件



### 上传（表单上传）

*since: 1.0.0*

上传请求格式, 需要构建一个表单，内容如下

```text
POST / HTTP/1.1
Host: 192.168.1.9:8089
Content-Length: 386
Content-Type: multipart/form-data; boundary=----WebKitFormBoundaryFRIRSZsF3BVUQlU9

------WebKitFormBoundaryFRIRSZsF3BVUQlU9
Content-Disposition: form-data; name="key"

/benjamin
------WebKitFormBoundaryFRIRSZsF3BVUQlU9
Content-Disposition: form-data; name="file"; filename="file.txt"
Content-Type: text/plain

filecontent

------WebKitFormBoundaryFRIRSZsF3BVUQlU9
Content-Disposition: form-data; name="source"


------WebKitFormBoundaryFRIRSZsF3BVUQlU9--
```



| 参数   | 默认值 | 是否必须 | 说明                                                         |
| ------ | ------ | -------- | ------------------------------------------------------------ |
| key    | N/A    | 是       | 指定上传的文件Key，Key=/<bucket name>/<object name>, 如果未包含object name，object name将由服务器自动生成，<br /> 例如： /benjamin, /benjamin/hello.txt |
| file   | N/A    | 是       | 文件内容                                                     |
| source | N/A    | 否       | 由source 指定的url资源作为文件内容,<br /> 例如：https://www.firstarriving.com/wp-content/uploads/2017/02/google-logo-1200x630.jpg |
| token  | N/A    | 否       | 通过AK/SK生成上传凭着，参考：[权限验证](https://www.tapd.cn/61498708/markdown_wikis/view/#1161498708001001924)<br/>since: *1.0.3* |





### 上传（HTTP Request Body）

since: *1.0.0*

Request Sample

```
POST /benjamin HTTP/1.1
Host: 192.168.1.9:8089
Content-Length: 12

file content
```

| 参数            | 默认值 | 是否必须 | 说明                                                         |
| --------------- | ------ | -------- | ------------------------------------------------------------ |
| key             | N/A    | 是       | 指定上传的文件Key，Key=/<bucket name>/<object name>, 如果未包含object name，object name将由服务器自动生成，<br /> 例如： /benjamin, /benjamin/hello.txt，于表单上传地方是， key参数来自HTTP request URL path |
| HTTP   Body     | N/A    | 是       | 内容是整个HTTP request Body                                  |
| X-WhaleFS-Token | N/A    | 否       | 通过AK/SK生成上传凭着，参考：[权限验证](https://www.tapd.cn/61498708/markdown_wikis/view/#1161498708001001924)<br/>since: *1.0.3* |



Response body

```json
{
    "key": "benjamin/Original/b0f3502d-490c-4708-8002-fa328db5ca48.txt",
    "size": 12,
    "url": "benjamin/Original/b0f3502d-490c-4708-8002-fa328db5ca48.txt",
    "title": "b0f3502d-490c-4708-8002-fa328db5ca48.txt",
    "message": "上传成功",
    "state": "SUCCESS",
    "original": ""
}
```

| 参数     | 说明                                                         |      |      |
| -------- | ------------------------------------------------------------ | ---- | ---- |
| key      | 用于下载的url相对地址，随机生成的文件名，会固定携带Original的路径 |      |      |
| size     | 文件实际大小                                                 |      |      |
| url      | 用于下载的url相对地址，主要是用于兼容老集采文件系统          |      |      |
| title    | 文件名                                                       |      |      |
| original | 上传的原始文件名                                             |      |      |
| message  | 消息                                                         |      |      |
| state    | 状态， SUCCESS\|FAILED                                       |      |      |



### 上传 （分块上传）

*since: 1.0.0*

分块上传/妙传都是通过该组接口实现支持

1. 初始化上传上下文

   该接口返回一个包含uploadId，key的对象, 在接下来的请求中需要用到这两个参数

   该请求中的Content-Type决定了文件的mime-type，需要设置

```text
POST /mro_item?uploads HTTP/1.1
Host: 192.168.1.9:8089
Content-Type: text/plain

HTTP/1.1 200 OK
Content-Length: 131
Content-Type: application/json; charset=UTF-8

{"bucket":"mro_item","key":"/mro_item/b26eb28b-408f-4c11-ab00-3d7f341ddde4.txt","uploadId":"b894724b-9a8c-44d2-981d-08641c5ae455"}
```

| 参数         | 默认值                   | 是否必须 | 说明                                                         |
| ------------ | ------------------------ | -------- | ------------------------------------------------------------ |
| bucket       | N/A                      | 是       | 需要上传的bucket名， 该示例为*mro_item*                      |
| uploads      | true                     | 是       | 表示要开启一次分段上传                                       |
| Content-Type | application/octet-stream | 否       | 表示该次上传的文件的mime-type， 会影响到生成随机文件名的后缀，所以需要设置 |



2. 上传分块

分块上传文件内容， 块的大小固定为16M，只允许最后一个块不足16M，

```text
POST /mro_item/b26eb28b-408f-4c11-ab00-3d7f341ddde4.txt?uploadId=b894724b-9a8c-44d2-981d-08641c5ae455&partNumber=1 HTTP/1.1
Host: 192.168.1.9:8089
Content-Length: 10

0123456789
```

| 参数       | 默认值 | 是否必须 | 说明                              |
| ---------- | ------ | -------- | --------------------------------- |
| key        | N/A    | 是       | 初始化上传之后返回的key           |
| uploadId   | N/A    | 是       | 分块上传ID                        |
| partNumber | N/A    | 是       | 分块编号，1开始的连续，自增的整数 |
| body       | N/A    | 是       | 文件内容，http request body       |



3. 完成上传

   所有分块上传完成之后，就需要调用该接口完成上传

```text
POST /mro_item/b26eb28b-408f-4c11-ab00-3d7f341ddde4.txt?uploadId=b894724b-9a8c-44d2-981d-08641c5ae455 HTTP/1.1
Host: image.jc.yzw.cn.qa:8000
Content-Length: 126
Content-Type: application/json

[{"partNumber": 1, "fid": "64,0b11a141191e2a", "size": 16777216}, {"partNumber": 2, "fid": "24,0b11a27df214cf", "size": 1024}]

HTTP/1.1 200 OK
Content-Length: 260
Content-Type: application/json; charset=UTF-8

{"key":"mro_item/Original/b26eb28b-408f-4c11-ab00-3d7f341ddde4.txt","size":16778240,"url":"mro_item/Original/b26eb28b-408f-4c11-ab00-3d7f341ddde4.txt","title":"b26eb28b-408f-4c11-ab00-3d7f341ddde4.txt","message":"............","state":"SUCCESS","original":""}

```

4. 终止上传

   如果你需呀终止上传

```text
DELETE /mro_item/b26eb28b-408f-4c11-ab00-3d7f341ddde4.txt?uploadId=b894724b-9a8c-44d2-981d-08641c5ae455 HTTP/1.1
Host: image.jc.yzw.cn.qa:8000


```



5. 检查分块

   为了支持妙传功能，我们添加检查分块是否的api, 首先你需要在客户端进行16M进行分块计算sha1，然后把分块信息发送whalefs进行验证， 哪些块已经存在，哪些块不存在， 就可以针对性的上传为存在的块

```text
POST /mro_item/402bc33e-fbda-45a6-bdc4-5d60e76360ec.txt?uploadId=02a50ab9-17f6-487d-aaa1-8a8778c6bec9&check= HTTP/1.1
Content-Length: 70
Content-Type: application/json

[{"chunkNo": 1, "digest": "2651d1fdd2a442bdbe1c38cef9fbd2329a281a08"},{"chunkNo": 2, "digest": "971c419dd609331343dee105fffd0f4608dc0bf2"}]
HTTP/1.1 200 OK
Via: NGINX
Content-Length: 67
Content-Type: application/json; charset=UTF-8

{"parts":[{"partNumber":1,"fid":"62,0b12304b111f9f","size":1024}],"missedDigest":[{"chunkNo": 2, "digest": "971c419dd609331343dee105fffd0f4608dc0bf2"}]}
```





### 下载接口

since: *1.0.0*

下载接口比较简单，只需要通过HTTP GET请求就能获取到文件内容

```
GET /benjamin/Original/b0f3502d-490c-4708-8002-fa328db5ca48.txt HTTP/1.1
Host: 192.168.1.9:8089
```

这样就能获取到上传的原始文件内容





### 下载接口（图片Resize， 水印效果）

如果有图片变化和水印效果的需求， 需要先在我们Portal上设置Resize和水印效果，然后通过本接口使用， 使用Resize的名称替换Original特殊字符, 而水印也是在bucket上设置， 目前支持一个默认水印图

```
GET /benjamin/p200/b0f3502d-490c-4708-8002-fa328db5ca48.txt HTTP/1.1
Host: 192.168.1.9:8089
```





### 下载接口 （触发浏览器的保存功能）

```
GET /benjamin/p200/b0f3502d-490c-4708-8002-fa328db5ca48.txt?attachmentName=file.txt HTTP/1.1
Host: 192.168.1.9:8089
```





### 视频截图

since：1.0.1

该接口主要用于视频文件获取第一帧图像，并返回图片内容。

目前该接口支持格式：mp4，生成的图片支持：jpeg

首先需要上传是视屏文件， 得到视频文件的URL，然后通过如下url就可以获取视屏的截图
```text
https://oss.yzw.cn/test/Original/0d27885f-b32e-4a85-ba29-7ccb03a3d572.mp4?preview&size=p200
```



| 参数    | 默认值 | 是否必须 | 说明                                                       |
| ------- | ------ | -------- | ---------------------------------------------------------- |
| preview | N/A    | 是       | 启用预览模式                                               |
| size    | N/A    | 否       | 该参数用于图片变化， 来源于bucket提前设置的size，例如：p60 |

