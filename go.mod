module github.com/by46/whalefs

go 1.12

require (
	github.com/BurntSushi/toml v0.3.1
	github.com/aliyun/aliyun-oss-go-sdk v2.0.1+incompatible
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/disintegration/imaging v1.6.0
	github.com/golang/snappy v0.0.1 // indirect
	github.com/google/uuid v1.1.1
	github.com/hhrutter/pdfcpu v0.1.23
	github.com/inconshreveable/mousetrap v1.0.0 // indirect
	github.com/konsorten/go-windows-terminal-sequences v1.0.2 // indirect
	github.com/labstack/echo v3.3.10+incompatible
	github.com/labstack/gommon v0.0.0-20190125185610-82ef680aef51
	github.com/mattn/go-colorable v0.1.1 // indirect
	github.com/nicksnyder/go-i18n/v2 v2.0.0
	github.com/opentracing/opentracing-go v1.1.0 // indirect
	github.com/pelletier/go-toml v1.4.0 // indirect
	github.com/pkg/errors v0.8.1
	github.com/prometheus/client_golang v0.9.3
	github.com/rafaeljesus/rabbus v2.3.0+incompatible
	github.com/rafaeljesus/retry-go v0.0.0-20171214204623-5981a380a879 // indirect
	github.com/robfig/cron v1.1.0
	github.com/sirupsen/logrus v1.4.1
	github.com/sony/gobreaker v0.4.1 // indirect
	github.com/spf13/afero v1.2.2 // indirect
	github.com/spf13/cobra v0.0.3
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/spf13/viper v1.3.2
	github.com/streadway/amqp v0.0.0-20190404075320-75d898a42a94
	github.com/stretchr/testify v1.3.0
	github.com/syndtr/goleveldb v1.0.0
	github.com/valyala/fasttemplate v1.0.1 // indirect
	golang.org/x/image v0.0.0-20181116024801-cd38e8056d9b
	golang.org/x/net v0.0.0-20190509222800-a4d6f7feada5 // indirect
	golang.org/x/text v0.3.2
	golang.org/x/time v0.0.0-20190308202827-9d24e82272b4 // indirect
	gopkg.in/couchbase/gocb.v1 v1.6.1
	gopkg.in/couchbase/gocbcore.v7 v7.1.13
	gopkg.in/couchbaselabs/gocbconnstr.v1 v1.0.2 // indirect
	gopkg.in/couchbaselabs/gojcbmock.v1 v1.0.3 // indirect
	gopkg.in/couchbaselabs/jsonx.v1 v1.0.0 // indirect
)

replace (
	golang.org/x/crypto v0.0.0-20180904163835-0709b304e793 => github.com/golang/crypto v0.0.0-20180904163835-0709b304e793
	golang.org/x/crypto v0.0.0-20181203042331-505ab145d0a9 => github.com/golang/crypto v0.0.0-20181203042331-505ab145d0a9
	golang.org/x/crypto v0.0.0-20190308221718-c2843e01d9a2 => github.com/golang/crypto v0.0.0-20190308221718-c2843e01d9a2
	golang.org/x/crypto v0.0.0-20190325154230-a5d413f7728c => github.com/golang/crypto v0.0.0-20190325154230-a5d413f7728c
	golang.org/x/crypto v0.0.0-20190506204251-e1dfcc566284 => github.com/golang/crypto v0.0.0-20190506204251-e1dfcc566284
	golang.org/x/image v0.0.0-20180708004352-c73c2afc3b81 => github.com/golang/image v0.0.0-20180708004352-c73c2afc3b81
	golang.org/x/image v0.0.0-20181116024801-cd38e8056d9b => github.com/golang/image v0.0.0-20181116024801-cd38e8056d9b
	golang.org/x/net v0.0.0-20180906233101-161cd47e91fd => github.com/golang/net v0.0.0-20180906233101-161cd47e91fd
	golang.org/x/net v0.0.0-20181114220301-adae6a3d119a => github.com/golang/net v0.0.0-20181114220301-adae6a3d119a
	golang.org/x/net v0.0.0-20190311183353-d8887717615a => github.com/golang/net v0.0.0-20190311183353-d8887717615a
	golang.org/x/net v0.0.0-20190404232315-eb5bcb51f2a3 => github.com/golang/net v0.0.0-20190404232315-eb5bcb51f2a3
	golang.org/x/net v0.0.0-20190503192946-f4e77d36d62c => github.com/golang/net v0.0.0-20190503192946-f4e77d36d62c
	golang.org/x/net v0.0.0-20190509222800-a4d6f7feada5 => github.com/golang/net v0.0.0-20190509222800-a4d6f7feada5
	golang.org/x/sync v0.0.0-20180314180146-1d60e4601c6f => github.com/golang/sync v0.0.0-20180314180146-1d60e4601c6f
	golang.org/x/sync v0.0.0-20181108010431-42b317875d0f => github.com/golang/sync v0.0.0-20181108010431-42b317875d0f
	golang.org/x/sync v0.0.0-20181221193216-37e7f081c4d4 => github.com/golang/sync v0.0.0-20181221193216-37e7f081c4d4
	golang.org/x/sync v0.0.0-20190423024810-112230192c58 => github.com/golang/sync v0.0.0-20190423024810-112230192c58
	golang.org/x/sys v0.0.0-20180905080454-ebe1bf3edb33 => github.com/golang/sys v0.0.0-20180905080454-ebe1bf3edb33
	golang.org/x/sys v0.0.0-20180909124046-d0be0721c37e => github.com/golang/sys v0.0.0-20180909124046-d0be0721c37e
	golang.org/x/sys v0.0.0-20181107165924-66b7b1311ac8 => github.com/golang/sys v0.0.0-20181107165924-66b7b1311ac8
	golang.org/x/sys v0.0.0-20181116152217-5ac8a444bdc5 => github.com/golang/sys v0.0.0-20181116152217-5ac8a444bdc5
	golang.org/x/sys v0.0.0-20181205085412-a5c9d58dba9a => github.com/golang/sys v0.0.0-20181205085412-a5c9d58dba9a
	golang.org/x/sys v0.0.0-20181217223516-dcdaa6325bcb => github.com/golang/sys v0.0.0-20181217223516-dcdaa6325bcb
	golang.org/x/sys v0.0.0-20190215142949-d0b11bdaac8a => github.com/golang/sys v0.0.0-20190215142949-d0b11bdaac8a
	golang.org/x/sys v0.0.0-20190222072716-a9d3bda3a223 => github.com/golang/sys v0.0.0-20190222072716-a9d3bda3a223
	golang.org/x/sys v0.0.0-20190412213103-97732733099d => github.com/golang/sys v0.0.0-20190412213103-97732733099d
	golang.org/x/sys v0.0.0-20190507160741-ecd444e8653b => github.com/golang/sys v0.0.0-20190507160741-ecd444e8653b
	golang.org/x/text v0.0.0-20181114220301-adae6a3d119a => github.com/golang/text v0.0.0-20181114220301-adae6a3d119a
	golang.org/x/text v0.3.0 => github.com/golang/text v0.3.0
	golang.org/x/text v0.3.2 => github.com/golang/text v0.3.2
	golang.org/x/time v0.0.0-20190308202827-9d24e82272b4 => github.com/golang/time v0.0.0-20190308202827-9d24e82272b4
	golang.org/x/tools v0.0.0-20180917221912-90fa682c2a6e => github.com/golang/tools v0.0.0-20180917221912-90fa682c2a6e
	golang.org/x/tools v0.0.0-20190328211700-ab21143f2384 => github.com/golang/tools v0.0.0-20190328211700-ab21143f2384
	golang.org/x/tools v0.0.0-20190506145303-2d16b83fe98c => github.com/golang/tools v0.0.0-20190506145303-2d16b83fe98c
)
