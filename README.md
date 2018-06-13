# whalefs

## seaweedfs

```bash
weed master -port=9001
weed master -port=9002 -peers="localhost:9001"

weed volume -port=9081 -mserver="localhost:9001" -dir="data"
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
  "alias": ["benjamin1", "benjamin2"],
  "expires": 20,
  "extends":[{"key":"keepdate", "value":"21"}],
  "memo":"mo bucket",
  "last_edit_date": 123143,
  "last_edit_user":"by46"
}
```