module github.com/uptrace/go-realworld-example-app

go 1.15

require (
	github.com/benbjohnson/clock v1.1.0
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/go-pg/migrations/v8 v8.0.1
	github.com/go-pg/pg/extra/pgdebug v0.2.0
	github.com/go-pg/pg/extra/pgotel v0.2.0
	github.com/go-pg/pg/v10 v10.7.3
	github.com/go-pg/urlstruct v1.0.0
	github.com/go-redis/cache/v8 v8.2.1
	github.com/go-redis/redis/extra/redisotel v0.2.0
	github.com/go-redis/redis/v8 v8.4.0
	github.com/go-redis/redis_rate/v9 v9.1.0
	github.com/google/uuid v1.1.2 // indirect
	github.com/gosimple/slug v1.9.0
	github.com/onsi/ginkgo v1.14.2
	github.com/onsi/gomega v1.10.3
	github.com/sirupsen/logrus v1.7.0
	github.com/uptrace/uptrace-go v0.5.5
	github.com/vmihailenco/msgpack/v5 v5.1.0 // indirect
	github.com/vmihailenco/treemux v0.4.1
	github.com/vmihailenco/treemux/extra/reqlog v0.4.1
	github.com/vmihailenco/treemux/extra/treemuxgzip v0.4.1
	github.com/vmihailenco/treemux/extra/treemuxotel v0.4.1
	go.opentelemetry.io/otel v0.14.0
	go.opentelemetry.io/otel/sdk v0.14.0
	golang.org/x/crypto v0.0.0-20201124201722-c8d3bf9c5392
	golang.org/x/exp v0.0.0-20201008143054-e3b2a7f2fdc7
	golang.org/x/sys v0.0.0-20201201145000-ef89a241ccb3 // indirect
	golang.org/x/text v0.3.4 // indirect
	golang.org/x/xerrors v0.0.0-20200804184101-5ec99f83aff1 // indirect
	gopkg.in/yaml.v2 v2.4.0
)
