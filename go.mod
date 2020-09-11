module github.com/seknox/trasadbproxy

go 1.14

require (
	cloud.google.com/go/storage v1.10.0
	github.com/Azure/azure-storage-blob-go v0.10.0
	github.com/GeertJohan/go.rice v1.0.0
	github.com/PuerkitoBio/goquery v1.5.1
	github.com/armon/consul-api v0.0.0-20180202201655-eb2c6b5be1b6
	github.com/armon/go-metrics v0.3.1
	github.com/aws/aws-sdk-go v1.34.20
	github.com/buger/jsonparser v1.0.0
	github.com/cespare/xxhash/v2 v2.1.1
	github.com/coreos/etcd v3.3.13+incompatible
	github.com/corpix/uarand v0.1.1 // indirect
	github.com/cyberdelia/go-metrics-graphite v0.0.0-20161219230853-39f87cc3b432
	github.com/fortytw2/leaktest v1.3.0 // indirect
	github.com/go-ini/ini v1.61.0 // indirect
	github.com/go-martini/martini v0.0.0-20170121215854-22fa46961aab
	github.com/go-sql-driver/mysql v1.5.0
	github.com/golang/glog v0.0.0-20160126235308-23def4e6c14b
	github.com/golang/mock v1.4.4
	github.com/golang/protobuf v1.4.2
	github.com/golang/snappy v0.0.1
	github.com/google/go-cmp v0.5.1
	github.com/gorilla/websocket v1.4.2
	github.com/grpc-ecosystem/go-grpc-middleware v1.0.1-0.20190118093823-f849b5445de4
	github.com/grpc-ecosystem/go-grpc-prometheus v1.2.0
	github.com/hashicorp/consul/api v1.4.0
	github.com/hashicorp/go-msgpack v0.5.5
	github.com/howeyc/gopass v0.0.0-20190910152052-7cb4b85ec19c
	github.com/icrowley/fake v0.0.0-20180203215853-4178557ae428
	github.com/klauspost/pgzip v1.2.5
	github.com/magiconair/properties v1.8.1
	github.com/mailru/easyjson v0.7.6 // indirect
	github.com/martini-contrib/auth v0.0.0-20150219114609-fa62c19b7ae8
	github.com/martini-contrib/gzip v0.0.0-20151124214156-6c035326b43f
	github.com/martini-contrib/render v0.0.0-20150707142108-ec18f8345a11
	github.com/mattn/go-sqlite3 v1.14.2
	github.com/minio/minio-go v6.0.14+incompatible
	github.com/minio/minio-go/v7 v7.0.5
	github.com/montanaflynn/stats v0.6.3
	github.com/olekukonko/tablewriter v0.0.4
	github.com/olivere/elastic v6.2.35+incompatible
	github.com/opentracing-contrib/go-grpc v0.0.0-20200813121455-4a6760c71486
	github.com/opentracing/opentracing-go v1.2.0
	github.com/oschwald/geoip2-golang v1.4.0
	github.com/patrickmn/go-cache v2.1.0+incompatible
	github.com/pborman/uuid v1.2.1
	github.com/philhofer/fwd v1.0.0 // indirect
	github.com/pires/go-proxyproto v0.1.3
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.7.1
	github.com/prometheus/common v0.10.0
	github.com/rcrowley/go-metrics v0.0.0-20200313005456-10cdbea86bc0
	github.com/samuel/go-zookeeper v0.0.0-20200724154423-2164a8ac840e
	github.com/satori/go.uuid v1.2.1-0.20181028125025-b2ce2384e17b
	github.com/seknox/trasa v0.0.0-20200906054606-a1c0faa025e3
	github.com/sirupsen/logrus v1.6.0
	github.com/sjmudd/stopwatch v0.0.0-20170613150411-f380bf8a9be1
	github.com/spf13/viper v1.7.1
	github.com/stretchr/testify v1.6.1
	github.com/tchap/go-patricia v2.3.0+incompatible
	github.com/tebeka/selenium v0.9.9
	github.com/uber/jaeger-client-go v2.25.0+incompatible
	github.com/uber/jaeger-lib v2.2.0+incompatible // indirect
	github.com/z-division/go-zookeeper v1.0.0
	golang.org/x/crypto v0.0.0-20200728195943-123391ffb6de
	golang.org/x/net v0.0.0-20200904194848-62affa334b73
	golang.org/x/oauth2 v0.0.0-20200107190931-bf48bf16ab8d
	golang.org/x/sync v0.0.0-20200625203802-6e8e738ad208
	golang.org/x/text v0.3.3
	golang.org/x/time v0.0.0-20191024005414-555d28b269f0
	google.golang.org/api v0.30.0
	google.golang.org/grpc v1.32.0
	google.golang.org/grpc/examples v0.0.0-20200909223711-52029da14822 // indirect
	gopkg.in/DataDog/dd-trace-go.v1 v1.26.0
	gopkg.in/gcfg.v1 v1.2.3
	gopkg.in/ldap.v2 v2.5.1
	gopkg.in/warnings.v0 v0.1.2 // indirect
	gotest.tools v2.2.0+incompatible
	k8s.io/apiextensions-apiserver v0.19.1
	k8s.io/apimachinery v0.19.1
	k8s.io/client-go v11.0.0+incompatible
	k8s.io/utils v0.0.0-20200821003339-5e75c0163111
	sigs.k8s.io/yaml v1.2.0
)
