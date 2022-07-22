module github.com/im-kulikov/helium

go 1.13

require (
	bou.ke/monkey v1.0.2
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.12.2
	github.com/spf13/viper v1.12.0
	github.com/stretchr/testify v1.7.1
	go.uber.org/atomic v1.9.0
	go.uber.org/dig v1.14.1
	go.uber.org/zap v1.21.0
	golang.org/x/mod v0.5.0 // indirect
	golang.org/x/net v0.0.0-20220520000938-2e3eb7b945c2
	google.golang.org/grpc v1.46.2
)

// Blocked in Russia
replace bou.ke/monkey v1.0.2 => github.com/bouk/monkey v1.0.2
