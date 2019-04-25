module github.com/im-kulikov/helium

require (
	bou.ke/monkey v1.0.1
	github.com/bsm/redis-lock v8.0.0+incompatible // indirect
	github.com/chapsuk/mserv v0.4.1
	github.com/chapsuk/wait v0.3.1 // indirect
	github.com/chapsuk/worker v0.4.0
	github.com/go-redis/redis v6.15.2+incompatible
	github.com/golang/protobuf v1.3.1 // indirect
	github.com/onsi/ginkgo v1.8.0 // indirect
	github.com/onsi/gomega v1.5.0 // indirect
	github.com/pkg/errors v0.8.1
	github.com/prometheus/client_golang v0.9.2
	github.com/prometheus/common v0.3.0 // indirect
	github.com/robfig/cron v0.0.0-20180505203441-b41be1df6967 // indirect
	github.com/smartystreets/goconvey v0.0.0-20190222223459-a17d461953aa
	github.com/spf13/viper v1.3.2
	github.com/stretchr/testify v1.3.0
	go.uber.org/atomic v1.3.2
	go.uber.org/dig v1.7.0
	go.uber.org/multierr v1.1.0 // indirect
	go.uber.org/zap v1.9.1
	google.golang.org/grpc v1.20.1 // indirect
)

// Blocked in Russia
replace bou.ke/monkey v1.0.1 => github.com/bouk/monkey v1.0.1
