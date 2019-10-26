module github.com/im-kulikov/helium

go 1.11

require (
	bou.ke/monkey v1.0.1
	github.com/chapsuk/mserv v1.1.0
	github.com/chapsuk/worker v1.0.0
	github.com/pkg/errors v0.8.1
	github.com/prometheus/client_golang v1.1.0
	github.com/spf13/viper v1.4.0
	github.com/stretchr/testify v1.4.0
	go.uber.org/atomic v1.4.0
	go.uber.org/dig v1.7.0
	go.uber.org/zap v1.10.0
)

// Blocked in Russia
replace bou.ke/monkey v1.0.1 => github.com/bouk/monkey v1.0.1
