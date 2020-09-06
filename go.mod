module github.com/im-kulikov/helium

go 1.13

require (
	bou.ke/monkey v1.0.2
	github.com/chapsuk/worker v1.0.0
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.7.1
	github.com/spf13/viper v1.7.1
	github.com/stretchr/testify v1.6.1
	go.uber.org/atomic v1.6.0
	go.uber.org/dig v1.10.0
	go.uber.org/zap v1.16.0
	golang.org/x/net v0.0.0-20190620200207-3b0461eec859
	google.golang.org/grpc v1.31.1
)

// Blocked in Russia
replace bou.ke/monkey v1.0.2 => github.com/bouk/monkey v1.0.2
