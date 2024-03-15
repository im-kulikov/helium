# Helium

![Codecov](https://img.shields.io/codecov/c/github/im-kulikov/helium.svg?style=flat-square)
[![Maintainability](https://api.codeclimate.com/v1/badges/01507c1d7c4b9649b1a7/maintainability)](https://codeclimate.com/github/im-kulikov/helium/maintainability)
[![Build Status](https://github.com/im-kulikov/helium/workflows/Go/badge.svg)](https://github.com/im-kulikov/helium/actions)
[![Report](https://goreportcard.com/badge/github.com/im-kulikov/helium)](https://goreportcard.com/report/github.com/im-kulikov/helium)
[![GitHub release](https://img.shields.io/github/release/im-kulikov/helium.svg)](https://github.com/im-kulikov/helium)
![GitHub](https://img.shields.io/github/license/im-kulikov/helium.svg?style=popout)
[![Sourcegraph](https://sourcegraph.com/github.com/im-kulikov/helium/-/badge.svg)](https://sourcegraph.com/github.com/im-kulikov/helium?badge)
[![Dependabot Status](https://api.dependabot.com/badges/status?host=github&repo=im-kulikov/helium)](https://dependabot.com)
[![FOSSA Status](https://app.fossa.io/api/projects/git%2Bgithub.com%2Fim-kulikov%2Fhelium.svg?type=shield)](https://app.fossa.io/projects/git%2Bgithub.com%2Fim-kulikov%2Fhelium?ref=badge_shield)

<img src="./.github/helium.jpg" width="350" alt="logo">

# There's no changes for multiple months

DEPRECATED

Due to the high workload at my main job, I have almost no time left to maintain two projects: Helium and [GoBones](https://github.com/im-kulikov/go-bones)

In this regard, after consulting with those who use Helium and who switched to GoBones, it was decided to curtail the development and support of Helium in favor of [GoBones](https://github.com/im-kulikov/go-bones)

# Documentation

* [Why Helium](#why-helium)
* [About Helium and modules](#about-helium-and-modules)
  + [Defaults and preconfigure](#defaults-and-preconfigure)
* [Group (services)](#group--services-)
  + [Service module](#service-module)
  + [Logger module](#logger-module)
* [NATS Module](#nats-module)
* [PostgreSQL Module](#postgresql-module)
* [Redis Module](#redis-module)
* [Settings module](#settings-module)
* [Web Module](#web-module)
* [Project Examples](#project-examples)
* [Example](#example)
* [Supported Go versions](#supported-go-versions)
* [Contribute](#contribute)
* [Credits](#credits)
* [License](#license)

<small><i><a href='http://ecotrust-canada.github.io/markdown-toc/'>Table of contents generated with markdown-toc</a></i></small>

## Why Helium

When building a modern application or prototype proof of concept, the last thing you want to worry about is boilerplate code or pass dependencies.
All this is a routine that each of us faces.
Helium lets you get rid of this routine and focus on your code.
Helium provides to you Convention over configuration.

The modular structure of helium consists of the following concepts
- **Module** is a set of providers
- **Provider** is what you put in the DI container. It teaches the container how to build values of one or more types and expresses their dependencies. It consists of two components of the **constructor** and **options**.
- **Constructor** is a function that accepts zero or more parameters and returns one or more results. The function may optionally return an error to indicate that it failed to build the value. This function will be treated as the constructor for all the types it returns. This function will be called AT MOST ONCE when a type produced by it, or a type that consumes this function's output, is requested via Invoke. If the same types are requested multiple times, the previously produced value will be reused. In addition to accepting constructors that accept dependencies as separate arguments and produce results as separate return values, Provide also accepts constructors that specify dependencies as dig.In structs and/or specify results as dig.Out structs.
- **Options** modifies the default behavior of dig.Provide

## About Helium and modules

*Helium is a small, simple, modular constructor with some pre-built components for your convenience.*
 
It contains the following components for rapid prototyping of your projects:
- Grace - [context](https://golang.org/pkg/context/) that helps you gracefully shutdown your application
- Group - collects an services and runs them concurrently, [see examples](#group-services).
- Logger - [zap](https://go.uber.org/zap) is blazing fast, structured, leveled logging in Go
- DI - based on [DIG](https://go.uber.org/dig). A reflection based dependency injection toolkit for Go.
- Module - set of tools for working with the DI component
- [NATS](https://github.com/go-helium/nats) - [nats](https://github.com/nats-io/go-nats) and [NSS](https://github.com/nats-io/nats-streaming-server), client for the cloud native messaging system
- [PostgreSQL](https://github.com/go-helium/postgres) - client module for [ORM](https://github.com/go-pg/pg) with focus on PostgreSQL features and performance
- [Redis](https://github.com/go-helium/redis) - module for type-safe [Redis](https://github.com/go-redis/redis) client for Golang  
- Settings - based on [Viper](https://github.com/spf13/viper). A complete configuration solution for Go applications including 12-Factor apps. It is designed to work within an application, and can handle all types of configuration needs and formats
- Web - [see more](#web-module)

### Defaults and preconfigure

*Helium* allows passing `defaults` (`settings.Defaults`) handler, which allows configuring application before it will be run
or do something with DI.

**Example:**

```go
package main

import (
  "github.com/spf13/viper"
  
  "github.com/im-kulikov/helium"
  "github.com/im-kulikov/helium/settings"
)

func defaults(v *viper.Viper) {
    v.SetDefault("some-key", "default-value")
}

func main() {
    _, err := helium.New(&helium.Settings{
        Name: "Abc",
        Defaults: defaults,
    }, settings.Module)
    helium.Catch(err)
}
``` 

## Group (services)

*Helium* provides primitive to run group of services (callback and shutdown functions) concurrently and stop when
- context will be canceled
- context will be deadlined
- any of service will be done (return from callback)

*Example*

```go
package main

import (
    "context"
    "fmt"
    "time"

    "github.com/im-kulikov/helium/group"
    "github.com/im-kulikov/helium/service"
)

func prepare(svc service.Service) (group.Callback, group.Shutdown) {
    fmt.Println("added service", svc.Name())
    return func(ctx context.Context) error {
        fmt.Println("start service", svc.Name())
        return svc.Start(ctx)
    },
    func(ctx context.Context) {
        fmt.Println("stop service", svc.Name())
        svc.Stop(ctx)
    }
}

func runner(ctx context.Context, services []service.Service) error {
    run := group.New(
        group.WithIgnoreErrors(context.Canceled),
        group.WithShutdownTimeout(time.Second * 30))
    for _, svc := range services {
        run.Add(prepare(svc))
    }

    // - wait until any service will be stopped
    // - wait until context will be canceled or deadlined
    return run.Run(ctx)
}
```

### Service module

*Helium* provide primitive for runnable services. That can be web-servers, workers, etc.


*Settings (used for all services)*
```yaml
shutdown_timeout: 30s
```

*Examples*

```go
package service

import "context"

type Service interface {
    Start(context.Context) error
    Stop(context.Context)

    Name() string 
}

type Group interface {
  Run(context.Context) error
}
```

You can pass into DI group of services and use them in `app.Run` method, for example:

```go
package main

import (
  "context"

  "github.com/im-kulikov/helium/service"
  "github.com/im-kulikov/helium/group"
)

type app struct {}

func (a *app) Run(ctx context.Context, group service.Group) error {
  return group.Run(ctx)
}
```

To provide single service:

```go
package some_pkg

import (
    "context"

    "github.com/im-kulikov/helium/module"
    "github.com/im-kulikov/helium/service"
    "go.uber.org/dig"
)

type OutParams struct {
    dig.Out
    Service service.Service `group:"services"`
}

type testWorker struct {
    name string
}

var _ = module.Module{
    // when you expose OutParams with `group:"services"`
    {Constructor: NewSingleOutService()},
    // when you use Options (dig.Group("services,flatten") and directly expose service.Service.
    {Constructor: NewSingleService, Options: []dig.ProvideOption{dig.Group("services,flatten")}},
}

func (w *testWorker) Start(context.Context) error { return nil }
func (w *testWorker) Stop(context.Context) { }
func (w *testWorker) Name() string { return w.name }

// NewSingleOutService used with OutParams and  module.New(NewSingleOutService).
func NewSingleOutService() OutParams {
    return OutParams{ Service: &testWorker{name: "worker1"} }
}

// NewSingleService used with module.New(NewSingleService, dig.Group("services,flatten")).
func NewSingleService() service.Service {
    return &testWorker{name: "worker1"}
}
```

To provide multiple services:

```go
package some_pkg

import (
    "context"
    
    "go.uber.org/dig"

    "github.com/im-kulikov/helium/module"
    "github.com/im-kulikov/helium/service"
)

// for multiple services use `group:"services,flatten"`
type OutParams struct {
    dig.Out
    Service []service.Service `group:"services,flatten"`
}

type testWorker struct {
    name string
}

var _ = module.Module{
    {Constructor: NewMultipleOut},
    {Constructor: NewMultiple, Options: []dig.ProvideOption{dig.Group("services,flatten")},
}

func (w *testWorker) Start(context.Context) error { return nil }
func (w *testWorker) Stop(context.Context) { return nil }
func (w *testWorker) Name() string { return w.name }

func NewMultipleOut() OutParams {
    return OutParams{
        Service: []service.Service{
            &testWorker{name: "worker1"},
            &testWorker{name: "worker2"},
        },
    }
}

// or using dig.Group("services,flatten")
// module.New(NewMultiple, dig.Group("services,flatten"))
func NewMultiple() []service.Service {
    return &testWorker{name: "worker1"}
}
```

### Logger module

Module provides you with the following things:
- `*zap.Logger` instance of [Logger](https://godoc.org/go.uber.org/zap#Logger)

A Logger provides fast, leveled, structured logging. All methods are safe for concurrent use.
The Logger is designed for contexts in which every microsecond and every allocation matters, so its API intentionally favors performance and type safety over brevity. For most applications, the SugaredLogger strikes a better balance between performance and ergonomics.
- `*zap.SugaredLogger` instance of [SugaredLogger](https://godoc.org/go.uber.org/zap#SugaredLogger)

Unlike the Logger, the SugaredLogger doesn't insist on structured logging. For each log level, it exposes three methods: one for loosely-typed structured logging, one for println-style formatting, and one for printf-style formatting. For example, SugaredLoggers can produce InfoLevel output with Infow ("info with" structured context), Info, or Infof.
A SugaredLogger wraps the base Logger functionality in a slower, but less verbose, API. Any Logger can be converted to a SugaredLogger with its Sugar method.
- `logger.StdLogger` provides simple interface that pass calls to **zap.SugaredLogger**

```
StdLogger interface {
    Fatal(v ...interface{})
    Fatalf(format string, v ...interface{})
    Print(v ...interface{})
    Printf(format string, v ...interface{})
}
```

Logger levels:
- DebugLevel logs are typically voluminous, and are usually disabled in production
- InfoLevel is the default logging priority
- WarnLevel logs are more important than Info, but don't need individual human review
- ErrorLevel logs are high-priority. If an application is running smoothly, it shouldn't generate any error-level logs
- DPanicLevel logs are particularly important errors. In development the logger panics after writing the message
- PanicLevel logs a message, then panics.
- FatalLevel logs a message, then calls os.Exit(1)

Logger formats:
- console:
```
2019-02-19T20:22:28.239+0300	info	web/servers.go:80	Create metrics http server, bind address: :8090	{"app_name": "Test", "app_version": "dev"}
2019-02-19T20:22:28.239+0300	info	web/servers.go:80	Create pprof http server, bind address: :6060	{"app_name": "Test", "app_version": "dev"}
2019-02-19T20:22:28.239+0300	info	app.go:26	init	{"app_name": "Test", "app_version": "dev"}
2019-02-19T20:22:28.239+0300	info	app.go:28	run workers	{"app_name": "Test", "app_version": "dev"}
2019-02-19T20:22:28.239+0300	info	app.go:31	run web-servers	{"app_name": "Test", "app_version": "dev"}
```
- json:
```
{"level":"info","msg":"Failed to fetch URL.","url":"http://example.com","attempt":3,"backoff":"1s"}
{"level":"info","msg":"Failed to fetch URL: http://example.com"}
{"level":"info","msg":"Failed to fetch URL.","url":"http://example.com","attempt":3,"backoff":"1s"}
```

Configuration for logger
- yaml example
```yaml
debug: true

logger:
    format: console
    level: info
    trace_level: fatal
    no_disclaimer: false
    color: true
    no_caller: false
    full_caller: true
    sampling:
      initial: 100
      thereafter: 100
```
- env example
```
DEBUG=true
LOGGER_NO_DISCLAIMER=true
LOGGER_COLOR=true
LOGGER_NO_CALLER=false
LOGGER_FULL_CALLER=true
LOGGER_FORMAT=console
LOGGER_LEVEL=info
LOGGER_TRACE_LEVEL=fatal
LOGGER_SAMPLING_INITIAL=100
LOGGER_SAMPLING_THEREAFTER=100
```

- `debug` - with this option you can enable `zap.DevelopmentConfig()`
- `logger.no_disclaimer` - with this option, you can disable `app_name` and `app_version` for any reason (not recommended in production)
- `logger.trace_level` - configures the Logger to record a stack trace for all messages at or above a given level
- `logger.color` - serializes a Level to an all-caps string and adds color
- `logger.no_caller` - disable serialization of a caller 
- `logger.full_caller` - serializes a caller in /full/path/to/package/file:line format
- `logger.sampling.initial` and `logger.sampling.thereafter` to setup [logger sampling](https://godoc.org/go.uber.org/zap#SamplingConfig). SamplingConfig sets a sampling strategy for the logger. Sampling caps the global CPU and I/O load that logging puts on your process while attempting to preserve a representative subset of your logs. Values configured here are per-second. See [zapcore.NewSampler](https://godoc.org/go.uber.org/zap/zapcore#NewSampler) for details.

## NATS Module

[Module](https://github.com/go-helium/nats) provides you with the following things:
- [`*nats.Conn`](https://godoc.org/github.com/nats-io/go-nats#Conn) represents a bare connection to a nats-server. It can send and receive []byte payloads
- [`stan.Conn`](https://godoc.org/github.com/nats-io/go-nats-streaming#Conn) represents a connection to the NATS Streaming subsystem. It can Publish and Subscribe to messages within the NATS Streaming cluster.

Configuration:
- yaml example
```yaml
nats:
  url: nats://<host>:<port>
  cluster_id: string
  client_id: string
  servers: [...server slice...]
  no_randomize: bool
  name: string
  verbose: bool
  pedantic: bool
  secure: bool
  allow_reconnect: bool
  max_reconnect: int
  reconnect_wait: duration
  timeout: duration
  flusher_timeout: duration
  ping_interval: duration
  max_pings_out: int
  reconnect_buf_size: int
  sub_chan_len: int
  user: string
  password: string
  token: string
```
- env example
```
NATS_URL=nats://<host>:<port>
NATS_CLUSTER_ID=string
NATS_CLIENT_ID=string
NATS_SERVERS=[...server slice...]
NATS_NO_RANDOMIZE=bool
NATS_NAME=string
NATS_VERBOSE=bool
NATS_PEDANTIC=bool
NATS_SECURE=bool
NATS_ALLOW_RECONNECT=bool
NATS_MAX_RECONNECT=int
NATS_RECONNECT_WAIT=duration
NATS_TIMEOUT=duration
NATS_FLUSHER_TIMEOUT=duration
NATS_PING_INTERVAL=duration
NATS_MAX_PINGS_OUT=int
NATS_RECONNECT_BUF_SIZE=int
NATS_SUB_CHAN_LEN=int
NATS_USER=string
NATS_PASSWORD=string
NATS_TOKEN=string
```

## PostgreSQL Module

[Module](https://github.com/go-helium/postgres) provides you connection to PostgreSQL server
- `*pg.DB` is a database handle representing a pool of zero or more underlying connections. It's safe for concurrent use by multiple goroutines

Configuration:
- yaml example
```yaml
posgres:
    address: string
    username: string
    password: string
    database: string
    debug: bool
    pool_size: int
```
- env example
```
POSTGRES_ADDRESS=string
POSTGRES_USERNAME=string
POSTGRES_PASSWORD=string
POSTGRES_DATABASE=string
POSTGRES_DEBUG=bool
POSTGRES_POOL_SIZE=int
```

## Redis Module

[Module](https://github.com/go-helium/redis) provides you connection to Redis server
- `*redis.Client` is a Redis client representing a pool of zero or more underlying connections. It's safe for concurrent use by multiple goroutines

Configuration:
- yaml example
```yaml
redis:
  address: string
  password: string
  db: int
  max_retries: int
  min_retry_backoff: duration
  max_retry_backoff: duration
  dial_timeout: duration
  read_timeout: duration
  write_timeout: duration
  pool_size: int
  pool_timeout: duration
  idle_timeout: duration
  idle_check_frequency: duration
```
- env example
```
REDIS_ADDRESS=string
REDIS_PASSWORD=string
REDIS_DB=int
REDIS_MAX_RETRIES=int
REDIS_MIN_RETRY_BACKOFF=duration
REDIS_MAX_RETRY_BACKOFF=duration
REDIS_DIAL_TIMEOUT=duration
REDIS_READ_TIMEOUT=duration
REDIS_WRITE_TIMEOUT=duration
REDIS_POOL_SIZE=int
REDIS_POOL_TIMEOUT=duration
REDIS_IDLE_TIMEOUT=duration
REDIS_IDLE_CHECK_FREQUENCY=duration
```

## Settings module

Module provides you [`*viper.Viper`](https://godoc.org/github.com/spf13/viper#Viper), a complete configuration solution for Go applications including 12-Factor apps. It is designed to work within an application, and can handle all types of configuration needs and formats.

Viper is a prioritized configuration registry. It maintains a set of configuration sources, fetches values to populate those, and provides them according to the source's priority. The priority of the sources is the following:
1. overrides
2. flags
3. env. variables
4. config file
5. key/value store 6. defaults

**Environments:**
```
<PREFIX>_CONFIG=/path/to/config
<PREFIX>_CONFIG_TYPE=<format>
```

## Web Module

- `ServersModule` puts into container [web.Service](https://github.com/im-kulikov/web/service.go):
    - [gRPC](https://github.com/golang/protobuf) endpoint
    - [Listener](https://github.com/im-kulikov/web/listener.go) allows provide custom web service and run it in scope. 
    - You can pass `pprof_handler` and/or `metric_handler`, that will be embedded into common handler,
      and will be available to call them
    - You can pass `api_listener`, `pprof_listener`, `metric_listener` to use them instead of network
      and address from settings
    - **api** endpoint by passing http.Handler from DI
- `OpsModule` puts into container [web.Service](https://github.com/im-kulikov/web/service.go):
  - [pprof](https://pkg.go.dev/net/http/pprof) `/debug/pprof` endpoints
  - [expvar](https://pkg.go.dev/expvar#Handler) `/debug/vars` endpoint
  - [metrics](https://pkg.go.dev/github.com/prometheus/client_golang) `/metrics` endpoint
  - health and ready endpoints
  
- [`echo.Module`](https://github.com/go-helium/echo) boilerplate that preconfigures echo.Engine for you
    - with custom Binder / Logger / Validator / ErrorHandler
    - bind - simple replacement for echo.Binder
    - validate - simple replacement for echo.Validate
    - logger - provides echo.Logger that pass calls to **zap.Logger**

Configuration:
```yaml
ops:
  address: :6060
  network: string
  disable_metrics: bool
  disable_pprof: bool
  disable_healthy: bool
  read_timeout: duration
  read_header_timeout: duration
  write_timeout: duration
  idle_timeout: duration
  max_header_bytes: int

api:
  address: :8080
  network: string
  disable_metrics: bool
  disable_pprof: bool
  disable_healthy: bool
  read_timeout: duration
  read_header_timeout: duration
  write_timeout: duration
  idle_timeout: duration
  max_header_bytes: int
```

```dotenv
OPS_ADDRESS=string
OPS_NETWORK=string
OPS_DISABLE_METRICS=bool
OPS_DISABLE_PPROF=bool
OPS_DISABLE_HEALTHY=bool
OPS_READ_TIMEOUT=duration
OPS_READ_HEADER_TIMEOUT=duration
OPS_WRITE_TIMEOUT=duration
OPS_IDLE_TIMEOUT=duration
OPS_MAX_HEADER_BYTES=int

API_ADDRESS=string
API_NETWORK=string
API_DISABLE_METRICS=bool
API_DISABLE_PPROF=bool
API_DISABLE_HEALTHY=bool
API_READ_TIMEOUT=duration
API_READ_HEADER_TIMEOUT=duration
API_WRITE_TIMEOUT=duration
API_IDLE_TIMEOUT=duration
API_MAX_HEADER_BYTES=int
```

**Possible options for HTTP server**:
- `address` - (string) host and port
- `network` - (string) tcp, udp, etc
- `read_timeout` - (duration) is the maximum duration for reading the entire request, including the body
- `read_header_timeout` - (duration) is the amount of time allowed to read request headers
- `write_timeout` - (duration) is the maximum duration before timing out writes of the response
- `idle_timeout` - (duration) is the maximum amount of time to wait for the next request when keep-alives are enabled
- `max_header_bytes` - (int) controls the maximum number of bytes the server will read parsing the request header's keys and values, including the request line

**Possible options for gRPC server**:
- `address` - (string) host and port
- `network` - (string) tcp, udp, etc
- `skip_errors` - allows ignore all errors
- `disabled` - (bool) to disable server

**OPS server configuration**
```yaml
ops:
  address: ":8081"
  network: "tcp"
  name: "ops-server" # by default
  disable_healthy: false
  disable_metrics: false
  disable_pprof: false
  idle_timeout: 0s
  max_header_bytes: 0
  read_header_timeout: 0s
  read_timeout: 0s
  write_timeout: 0s
```

```dotenv
OPS_ADDRESS=string
OPS_NETWORK=string
OPS_READ_TIMEOUT=duration
OPS_READ_HEADER_TIMEOUT=duration
OPS_WRITE_TIMEOUT=duration
OPS_IDLE_TIMEOUT=duration
OPS_MAX_HEADER_BYTES=int
OPS_DISABLE_METRICS=bool
OPS_DISABLE_PROFILE=bool
OPS_DISABLE_HEALTHY=bool
```

**Listener example:**
```go
package my

import (
  "github.com/im-kulikov/helium/module"
  "github.com/im-kulikov/helium/web"
  "github.com/k-sone/snmpgo"
  "github.com/spf13/viper"
  "go.uber.org/zap"
)

type SNMPListener struct {
  serve *snmpgo.TrapServer
}

var _ = module.Module{
  {Constructor: NewSNMPServer},
}

func NewSNMPServer(v *viper.Viper, l *zap.Logger) (web.ServerResult, error) {
  var res web.ServerResult

  switch {
  case v.GetBool("snmp.disabled"):
		l.Warn("SNMP server is disabled")
    return res, nil
  case !v.IsSet("snmp.address"):
    l.Warn("SNMP server shoud have address")
    return res, nil

  }

  lis, err := snmpgo.NewTrapServer(snmpgo.ServerArguments{
		LocalAddr: v.GetString("snmp.address"),
  })
  
  // if something went wrong, return ServerResult and error:
  if err != nil {
    return res, err
  }
  
  opts := []web.ListenerOption{
    // If you want to ignore all errors
    web.ListenerSkipErrors(),

    // Ignore shutdown error
    web.ListenerIgnoreErrors(errors.New("snmp: server shutdown")),
  }
  
  res.serve, err = web.NewListener(lis, opts...)
  return res, err
}

func (l *SNMPListener) ListenAndServe() error {
  return l.serve.Serve(l) // because SNMPListener is TrapHandler
}

func (l *SNMPListener) Shutdown(context.Context) error {
	return l.server.Close()
}

func (l *SNMPListener) OnTRAP(trap *snmpgo.TrapRequest) {
  // do something with received message
}
```

**Use default gRPC example:**
```go
package my

import (
  "github.com/im-kulikov/helium/module"
  "github.com/im-kulikov/helium/web"
  "go.uber.org/dig"
  "go.uber.org/zap"
  "google.golang.org/grpc"
)

type gRPCResult struct {
    dig.Out
    
    Key    string       `name:"grpc_config"`
    Server *grpc.Server `name:"grpc_server"`
}

var _ = module.Module{
  {Constructor: newDefaultGRPCServer},
}

// newDefaultGRPCServer returns gRPCResult that would be used
// to create gRPC Service and run it with other web-services.
//
// See web.newDefaultGRPCServer for more information.
func newDefaultGRPCServer() gRPCResult {
  return gRPCResult{
    // config key that would be used
    // For exmpale :
    // - <key>.address
    // - <key>.network
    // - etc
    Key:    "grpc",
    Server: grpc.NewServer(),
  }
}
```

## Project Examples

- [Atlant.io Test Task](https://github.com/im-kulikov/atlantio-task) 
- [Simplinic Test Task](https://github.com/im-kulikov/simplinic-task) 
- [Golang documentation Telegram Bot](https://github.com/im-kulikov/doc-bot)
- [Potter](https://github.com/im-kulikov/potter) is simple fixture based API service
- [Image processing service](https://github.com/archaron/secimage)

## Example

**config.yml**
```yaml
api:
  address: :8080
  debug: true

logger:
  level: debug
  format: console
```

**main.go**
```go
package main

import (
	"context"
    "net/http"
	

	"go.uber.org/dig"
	"go.uber.org/zap"
	
	"github.com/im-kulikov/helium"
	"github.com/im-kulikov/helium/grace"
	"github.com/im-kulikov/helium/logger"
	"github.com/im-kulikov/helium/module"
	"github.com/im-kulikov/helium/service"
	"github.com/im-kulikov/helium/settings"
	"github.com/im-kulikov/helium/web"
)

func main() {
	h, err := helium.New(&helium.Settings{
		Name:         "demo3",
		Prefix:       "DM3",
		File:         "config.yml",
		BuildVersion: "dev",
	}, module.Module{
		{Constructor: handler},
	}.Append(
		grace.Module,
		logger.Module,
		settings.Module,
		web.DefaultServersModule,
	))
	err = dig.RootCause(err)
	helium.Catch(err)
	err = h.Invoke(runner)
	err = dig.RootCause(err)
	helium.Catch(err)
}

func handler() http.Handler {
	h := http.NewServeMux()
	h.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})
	return h
}

func runner(ctx context.Context, svc service.Group) error {
	return svc.Run(ctx)
}
```

## Supported Go versions

Helium is available as a [Go module](https://github.com/golang/go/wiki/Modules).
- 1.13+

## Contribute

**Use issues for everything**

- For a small change, just send a PR.
- For bigger changes open an issue for discussion before sending a PR.
- PR should have:
  - Test case
  - Documentation
  - Example (If it makes sense)
- You can also contribute by:
  - Reporting issues
  - Suggesting new features or enhancements
  - Improve/fix documentation

## Credits

- [Evgeniy Kulikov](https://github.com/im-kulikov) - Author
- [Alexander Tischenko](https://github.com/archaron) - Consultant
- [Contributors](https://github.com/im-kulikov/helium/graphs/contributors)

## License

[MIT](LICENSE)

[![FOSSA Status](https://app.fossa.io/api/projects/git%2Bgithub.com%2Fim-kulikov%2Fhelium.svg?type=large)](https://app.fossa.io/projects/git%2Bgithub.com%2Fim-kulikov%2Fhelium?ref=badge_large)
