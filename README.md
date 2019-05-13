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
  "alias": ["benjamin1", "benjamin2"],
  "expires": 20,
  "extends":[{"key":"keepdate", "value":"21"}],
  "limit": {
    "min_size": -1,
    "max_size": 102400,
    "width": 10,
    "height": 10,
    "mime_types": ["image/png", "image/jpeg", "image/png"]
  },
  "last_edit_date": 123143,
  "last_edit_user":"by46",
  "overlays": [
    {"name": "demo1", "default": true, "position": "TopLeft", "image": "7,15154f3ef7"},
    {"name": "demo2", "default": false, "position": "TopRight", "image": "7,15154f3ef7"},
    {"name": "demo3", "default": false, "position": "BottomLeft", "image": "7,15154f3ef7"},
    {"name": "demo4", "default": false, "position": "BottomRight", "image": "7,15154f3ef7"},
    {"name": "demo5", "default": false, "position": "{\"top\":null, \"right\":0,\"bottom\":0, \"left\":0}", "image": "7,15154f3ef7"}
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

### golang 编程规范