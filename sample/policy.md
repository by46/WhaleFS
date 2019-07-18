# 权限验证

权限是一个系统比较重要的部分， 我们系统采用比较常见的AK/SK验证模式，用户在我们的管理Portal上申请一个对AK/SK密钥，用于生成上传/下载凭证， 依赖该凭证进行权限验证和授权。

首先需要一个上传策略对象，结构如下

```json
{
    "scope":               "<Bucket                   string>",
    "deadline":            "<UnixTimestamp            uint32>"
}
```

| 名称     | 是否必填 | 说明                                                         |
| -------- | -------- | ------------------------------------------------------------ |
| scope    | 是       | 表示允许用户上传文件到指定的 bucket                          |
| deadline | 是       | 上传凭证有效截止时间。[Unix时间戳](https://developer.qiniu.com/kodo/glossary/1647/u#unixtime)，单位为秒。如果凭证过期，验证将失效，所以需要做好token的刷新策略, 需要主要的是， 时间戳是UTC时间 |

## 算法

1. 构造上传策略

用户根据业务需求，确定上传策略要素，构造出具体的上传策略。例如用户要向空间 ``test`` 上传文件，授权有效期截止到 2018-01-01 00:00:00。那么相应的上传策略各字段分别为：

```json
{
 "scope": "test",
 "deadline": 1514764800
}
```

2. 对 JSON 编码的上传策略进行[URL 安全的 Base64 编码](https://developer.qiniu.com/kodo/manual/1231/appendix#urlsafe-base64)，得到待签名字符串：

```json
encodedPutPolicy = urlsafe_base64_encode(putPolicy)
#实际值为：
encodedPutPolicy = "eyJzY29wZSI6Im15LWJ1Y2tldDpzdW5mbG93ZXIuanBnIiwiZGVhZGxpbmUiOjE0NTE0OTEyMDAsInJldHVybkJvZHkiOiJ7XCJuYW1lXCI6JChmbmFtZSksXCJzaXplXCI6JChmc2l6ZSksXCJ3XCI6JChpbWFnZUluZm8ud2lkdGgpLFwiaFwiOiQoaW1hZ2VJbmZvLmhlaWdodCksXCJoYXNoXCI6JChldGFnKX0ifQ=="
```

> URL安全的Base64编码适用于以URL方式传递Base64编码结果的场景。该编码方式的基本过程是先将内容以Base64格式编码为字符串，然后检查该结果字符串，将字符串中的加号`+`换成中划线`-`，并且将斜杠`/`换成下划线`_`。
>
> 详细编码规范请参考[RFC4648](http://www.ietf.org/rfc/rfc4648.txt)标准中的相关描述。

3. 使用[访问密钥（AK/SK）](https://developer.qiniu.com/kodo/manual/1277/product-introduction#ak-sk)对上一步生成的待签名字符串计算[HMAC-SHA1](https://developer.qiniu.com/kodo/glossary/1643/h#hmac-sha1)签名：

```json
sign = hmac_sha1(encodedPutPolicy, "<SecretKey>")
#假设 SecretKey 为 MY_SECRET_KEY，实际签名为：
sign = "c10e287f2b1e7f547b20a9ebce2aada26ab20ef2"
```

4. 对签名进行[URL安全的Base64编码](https://developer.qiniu.com/kodo/manual/1231/appendix#urlsafe-base64)：

```json
encodedSign = urlsafe_base64_encode(sign)
#最终签名值为：
encodedSign = "wQ4ofysef1R7IKnrziqtomqyDvI="
```

5. 将[访问密钥（AK/SK）](https://developer.qiniu.com/kodo/manual/1277/product-introduction#ak-sk)、encodedSign 和 encodedPutPolicy 用英文符号 : 连接起来：

```json
uploadToken = AccessKey + ':' + encodedSign + ':' + encodedPutPolicy
#假设用户的 AccessKey 为 MY_ACCESS_KEY ，则最后得到的上传凭证应为：
uploadToken = "MY_ACCESS_KEY:wQ4ofysef1R7IKnrziqtomqyDvI=:eyJzY29wZSI6Im15LWJ1Y2tldDpzdW5mbG93ZXIuanBnIiwiZGVhZGxpbmUiOjE0NTE0OTEyMDAsInJldHVybkJvZHkiOiJ7XCJuYW1lXCI6JChmbmFtZSksXCJzaXplXCI6JChmc2l6ZSksXCJ3XCI6JChpbWFnZUluZm8ud2lkdGgpLFwiaFwiOiQoaW1hZ2VJbmZvLmhlaWdodCksXCJoYXNoXCI6JChldGFnKX0ifQ=="
```

主要是参照七牛云做的， [文档地址](https://developer.qiniu.com/kodo/manual/1208/upload-token)

给出一段Golang的示例代码

```go
package utils

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"
 
  "github.com/stretchr/testify/assert"
)

type Policy struct {
	Bucket   string `json:"bucket"`
	Deadline int64  `json:"deadline"`
}

func (p *Policy) Decode(sign, appSecretKey, payload string) (err error) {
	content, err := base64.StdEncoding.DecodeString(payload)
	if err != nil {
		return fmt.Errorf("策略格式错误")
	}
	sign2 := Encode(content, appSecretKey)
	if sign != sign2 {
		return fmt.Errorf("权限凭证错误")
	}
	if err = json.Unmarshal(content, p); err != nil {
		return fmt.Errorf("策略格式JSON错误")
	}
	if p.Deadline <= time.Now().UTC().Unix() {
		return fmt.Errorf("Token过期")
	}
	return nil
}

func (p *Policy) Encode(appId, appSecretKey string) string {
	content, _ := json.Marshal(p)
	encodedSign := Encode(content, appSecretKey)
	return fmt.Sprintf("%s:%s:%s", appId, encodedSign, base64.StdEncoding.EncodeToString(content))
}

func Encode(policy []byte, appSecretKey string) string {
	encodingPolicy := base64.StdEncoding.EncodeToString(policy)
	mac := hmac.New(sha1.New, []byte(appSecretKey))
	mac.Write([]byte(encodingPolicy))
	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}


func TestPolicy_Encode(t *testing.T) {
	p := &Policy{
		Bucket:   "test",
		Deadline: time.Now().UTC().Add(10 * time.Minute).Unix(),
	}
	appId := "app_id"
	appSecretKey := "app_secret_key"
	sign := p.Encode(appId, appSecretKey)
	assert.Equal(t, "app_id:TfCgmTIDp4fL69TeQO0WXMjnfPU=:eyJidWNrZXQiOiJpdGVtIiwiZGVhZGxpbmUiOjE1NjIxNzA5ODh9", sign)
}

func TestPolicy_Decode(t *testing.T) {
	p := new(Policy)
	err := p.Decode("TfCgmTIDp4fL69TeQO0WXMjnfPU=", "app_secret_key", "eyJidWNrZXQiOiJpdGVtIiwiZGVhZGxpbmUiOjE1NjIxNzA5ODh9")
	assert.Nil(t, err)
	assert.Equal(t, "item", p.Bucket)
}
```

