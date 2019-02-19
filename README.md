# Helium

[![codecov](https://codecov.io/gh/im-kulikov/helium/branch/master/graph/badge.svg)](https://codecov.io/gh/im-kulikov/helium)
[![CircleCI](https://circleci.com/gh/im-kulikov/helium.svg?style=svg)](https://circleci.com/gh/im-kulikov/helium)
[![Report](https://goreportcard.com/badge/github.com/im-kulikov/helium)](https://goreportcard.com/report/github.com/im-kulikov/helium)
[![GitHub release](https://img.shields.io/github/release/im-kulikov/helium.svg)](https://github.com/im-kulikov/helium)

<img src="./.github/helium.jpg" width="350" alt="logo">

# Documentation

* [About](#about)
* [Why Helium](#why-helium)
* [Credits](#credits)

## About

*Helium is small, simple and modular constructor with building-blocks.*
 
It contains the following components for rapid prototyping of your projects:
- Grace - [context](https://golang.org/pkg/context/) that helps you graceful shutdown of your application
- Logger - [zap](https://go.uber.org/zap) is blazing fast, structured, leveled logging in Go
- DI - based on [DIG](https://go.uber.org/dig). A reflection based dependency injection toolkit for Go.
- Module - set of tools for working with DI component
- NATS - [nats](https://github.com/nats-io/go-nats) and [NSS](https://github.com/nats-io/nats-streaming-server), client for the cloud native messaging system
- ORM - client module for [ORM](https://github.com/go-pg/pg) with focus on PostgreSQL features and performance
- redis - module for type-safe [Redis](https://github.com/go-redis/redis) client for Golang  
- Settings - based on [Viper](https://github.com/spf13/viper). A complete configuration solution for Go applications including 12-Factor apps. It is designed to work within an application, and can handle all types of configuration needs and formats
- Web - [see more](#web-module)
- Workers - are tools to run goroutine and do some work on scheduling with a safe stop of their work. Based on [chapsuk/worker](https://github.com/chapsuk/worker)

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

## Logger

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
logger:
    format: console
    level: info
```
- env example
```
LOGGER_FORMAT=console
LOGGER_LEVEL=info
```

## Web Module



## Credits

- [Evgeniy Kulikov](https://github.com/im-kulikov) - Author
- [Alexander Tischenko](https://github.com/archaron) - Consultant
- [Contributors](https://github.com/im-kulikov/helium/graphs/contributors)