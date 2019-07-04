# 权限策略

app_id
app_secret_key

```json
{
 "scope": "mro_item",
 "deadline": 141311
}

```

encodedPutPolicy = urlsafe_base64_encode(putPolicy)
#实际值为：
encodedPutPolicy = "eyJzY29wZSI6Im15LWJ1Y2tldDpzdW5mbG93ZXIuanBnIiwiZGVhZGxpbmUiOjE0NTE0OTEyMDAsInJldHVybkJvZHkiOiJ7XCJuYW1lXCI6JChmbmFtZSksXCJzaXplXCI6JChmc2l6ZSksXCJ3XCI6JChpbWFnZUluZm8ud2lkdGgpLFwiaFwiOiQoaW1hZ2VJbmZvLmhlaWdodCksXCJoYXNoXCI6JChldGFnKX0ifQ=="

sign = hmac_sha1(encodedPutPolicy, "<SecretKey>")
#假设 SecretKey 为 MY_SECRET_KEY，实际签名为：
sign = "c10e287f2b1e7f547b20a9ebce2aada26ab20ef2"

encodedSign = urlsafe_base64_encode(sign)
#最终签名值为：
encodedSign = "wQ4ofysef1R7IKnrziqtomqyDvI="

uploadToken = AccessKey + ':' + encodedSign + ':' + encodedPutPolicy
#假设用户的 AccessKey 为 MY_ACCESS_KEY ，则最后得到的上传凭证应为：
uploadToken = "MY_ACCESS_KEY:wQ4ofysef1R7IKnrziqtomqyDvI=:eyJzY29wZSI6Im15LWJ1Y2tldDpzdW5mbG93ZXIuanBnIiwiZGVhZGxpbmUiOjE0NTE0OTEyMDAsInJldHVybkJvZHkiOiJ7XCJuYW1lXCI6JChmbmFtZSksXCJzaXplXCI6JChmc2l6ZSksXCJ3XCI6JChpbWFnZUluZm8ud2lkdGgpLFwiaFwiOiQoaW1hZ2VJbmZvLmhlaWdodCksXCJoYXNoXCI6JChldGFnKX0ifQ=="