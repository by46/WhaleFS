# 分布式文件系统
## 环境
目前分布式文件系统有两套环境: QA 和 PRD

### QA
服务地址： http://image.jc.yzw.cn.qa:8000
老JC服务地址： http://imagetest.yzw.cn.qa:8000

由于这几个域名都是自定义域名，所以需要在你的local DNS中添加解析规则

Windows：C:\Windows\system32\drivers\etc\hosts, 规则如下：
```text
172.16.0.253 image.jc.yzw.cn.qa
172.16.0.253 imagetest.yzw.cn.qa
```

MacOS: /etc/hosts, 规则如下：
```text
172.16.0.253 image.jc.yzw.cn.qa
172.16.0.253 imagetest.yzw.cn.qa
```

### PRD
PRD环境使用的是一个外部域名， 可以直接访问
服务地址： https://oss.yzw.cn



## API

API接口有两部分组成：

- 老集采系统原来接口， 目前