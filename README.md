# whalefs

## seaweedfs

```bash

weed master -port=9001

weed master -port=9002 -peers="localhost:9001"

weed volume -port=9081 -mserver="localhost:9001" -dir="data"

/opt/weed/weed master -mdir="/opt/dfs/master"

/opt/weed/weed volume -ip=192.168.1.9 -port=18081 -mserver="localhost:9333" -dir="/opt/dfs/data1" 

/opt/weed/weed volume -ip=192.168.1.9 -port=18082 -mserver="localhost:9333" -dir="/opt/dfs/data2"

/opt/weed/weed volume -ip=192.168.1.9 -port=18083 -mserver="localhost:9333" -dir="/opt/dfs/data3"

```

### buckets
system.buckets
```json
{
  "buckets": [
    "system.bucket.benjamin"
  ]
}
```
### bucket

system.bucket.benjamin
```json
{
  "name": "benjamin",
  "memo":"mo bucket",
  "basis": {
    "alias": "pdt",
    "collection": "",
    "replication": "100",
    "expires": 20,
    "prepare_thumbnail_min_width": 1024,
    "prepare_thumbnail": ""
  },
  "expires": 20,
  "extends":[{"key":"keepdate", "value":"21"}],
  "limit": {
    "min_size": null,
    "max_size": 102400,
    "width": null,
    "height": null,
    "mime_types": ["image/png", "image/jpeg", "image/png"]
  },
  "last_edit_date": 123143,
  "last_edit_user":"by46",
  "overlays": [
    {"name": "demo1", "default": true, "position": "TopLeft", "image": "7,15154f3ef7", "opacity":  0.8},
    {"name": "demo2", "default": false, "position": "TopRight", "image": "7,15154f3ef7", "opacity":  0.8},
    {"name": "demo3", "default": false, "position": "BottomLeft", "image": "7,15154f3ef7", "opacity":  0.8},
    {"name": "demo4", "default": false, "position": "BottomRight", "image": "7,15154f3ef7", "opacity":  0.8},
    {"name": "demo5", "default": false, "position": "{\"top\":null, \"right\":0,\"bottom\":0, \"left\":0}", "image": "7,15154f3ef7", "opacity":  0.8}
  ],
  "sizes": [
      {"name": "p200", "width":200, "height": 150, "mode": "stretch"},
      {"name": "p60", "width":60, "height": 45, "mode": "fit"},
      {"name": "p160", "width":160, "height": 120, "mode": "thumbnail"}
  ]
}
```

system.bucket.package
```json
{
  "name": "package",
  "memo":"mo bucket",
  "basis": {
    "collection": "",
    "replication": "100",
    "expires": 20,
    "prepare_thumbnail_min_width": 1024,
    "prepare_thumbnail": ""
  },
  "expires": 20,
  "extends":[{"key":"keepdate", "value":"21"}],
  "limit": {
    "min_size": null,
    "max_size": 102400,
    "width": null,
    "height": null,
    "mime_types": ["image/png", "image/jpeg", "image/png"]
  },
  "last_edit_date": 123143,
  "last_edit_user":"by46",
  "overlays": [
    {"name": "demo1", "default": true, "position": "TopLeft", "image": "7,15154f3ef7", "opacity":  0.8},
    {"name": "demo2", "default": false, "position": "TopRight", "image": "7,15154f3ef7", "opacity":  0.8},
    {"name": "demo3", "default": false, "position": "BottomLeft", "image": "7,15154f3ef7", "opacity":  0.8},
    {"name": "demo4", "default": false, "position": "BottomRight", "image": "7,15154f3ef7", "opacity":  0.8},
    {"name": "demo5", "default": false, "position": "{\"top\":null, \"right\":0,\"bottom\":0, \"left\":0}", "image": "7,15154f3ef7", "opacity":  0.8}
  ],
  "sizes": [
      {"name": "p200", "width":200, "height": 150, "mode": "stretch"},
      {"name": "p60", "width":60, "height": 45, "mode": "fit"},
      {"name": "p160", "width":160, "height": 120, "mode": "thumbnail"}
  ]
}
```

system.bucket.product

```json
{
  "name": "product",
  "memo":"mo bucket",
  "basis": {
    "alias": "pdt",
    "collection": "",
    "replication": "100",
    "expires": 20,
    "prepare_thumbnail_min_width": 1024,
    "prepare_thumbnail": ""
  },
  "expires": 20,
  "extends":[{"key":"keepdate", "value":"21"}],
  "limit": {
    "min_size": null,
    "max_size": 102400,
    "width": null,
    "height": null,
    "mime_types": ["image/png", "image/jpeg", "image/png"]
  },
  "last_edit_date": 123143,
  "last_edit_user":"by46",
  "overlays": [
    {"name": "demo1", "default": true, "position": "TopLeft", "image": "7,15154f3ef7", "opacity":  0.8},
    {"name": "demo2", "default": false, "position": "TopRight", "image": "7,15154f3ef7", "opacity":  0.8},
    {"name": "demo3", "default": false, "position": "BottomLeft", "image": "7,15154f3ef7", "opacity":  0.8},
    {"name": "demo4", "default": false, "position": "BottomRight", "image": "7,15154f3ef7", "opacity":  0.8},
    {"name": "demo5", "default": false, "position": "{\"top\":null, \"right\":0,\"bottom\":0, \"left\":0}", "image": "7,15154f3ef7", "opacity":  0.8}
  ],
  "sizes": [
      {"name": "p200", "width":200, "height": 150, "mode": "stretch"},
      {"name": "p60", "width":60, "height": 45, "mode": "fit"},
      {"name": "p160", "width":160, "height": 120, "mode": "thumbnail"}
  ]
}

```



```
package: github.com/by46/whalefs
homepage: https://github.com/by46/whalefs
license: MIT
owners:
  - name: benjamin.c.yan
    email: ycs_ctbu_2010@126.com
import:
- package: github.com/spf13/viper
  version: ^1.0.2
- package: github.com/sirupsen/logrus
  version: ^1.0.5
- package: github.com/couchbase/go-couchbase
- package: github.com/couchbase/gomemcached
- package: github.com/couchbase/goutils
- package: github.com/mholt/binding
  version: ^0.3.0
- package: github.com/spf13/cobra
  version: 0.0.2
- package: github.com/pkg/errors
  version: ^0.8.0

```

### 普罗米修斯
http://172.16.0.158:9090/graph

### golang 编程规范

### 接口 大文件上传
#### 初始化上传
POST /benjamin/demo/hello.jpg?uploads

{
    "upload_id":"uuid1"
}

#### 上传chunk
PUT /benjamin/demo/hello.jpg?uploadId=uuid1&partNumber=partNumber1

<multipart-form>
</multipart-form>

#### 完成上传
POST /benjamin/demo/hello.jpg?uploadId=uuid1

[{
  "part_number": "part number 1",
  "etag": "etag1"   
}]



whalefs.exe migrate --location="D:\application\ImageServer" --target="192.168.1.9:8000" --includes="banner,banner_01"
whalefs.exe migrate --location="D:\application\ImageServer" --target="192.168.1.9:8000" --includes="bond,bond_01,contract,contract_01,eInvoice,eInvoice_01"