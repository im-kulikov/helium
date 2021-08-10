module github.com/im-kulikov/helium

go 1.13

require (
	bou.ke/monkey v1.0.2
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.11.0
	github.com/spf13/viper v1.8.1
	github.com/stretchr/testify v1.7.0
	go.uber.org/atomic v1.9.0
	go.uber.org/dig v1.11.0
	go.uber.org/zap v1.19.0
	golang.org/x/net v0.0.0-20210405180319-a5a99cb37ef4
	google.golang.org/grpc v1.38.0
)

// Blocked in Russia
replace bou.ke/monkey v1.0.2 => github.com/bouk/monkey v1.0.2
