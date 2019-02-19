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
- *Module* is a set of providers
- *Provider* is what you put in the DI container. It teaches the container how to build values of one or more types and expresses their dependencies. It consists of two components of the **constructor** and **options**.
- *Constructor* is a function that accepts zero or more parameters and returns one or more results. The function may optionally return an error to indicate that it failed to build the value. This function will be treated as the constructor for all the types it returns. This function will be called AT MOST ONCE when a type produced by it, or a type that consumes this function's output, is requested via Invoke. If the same types are requested multiple times, the previously produced value will be reused. In addition to accepting constructors that accept dependencies as separate arguments and produce results as separate return values, Provide also accepts constructors that specify dependencies as dig.In structs and/or specify results as dig.Out structs.
- *Options* modifies the default behavior of dig.Provide

## Web Module



## Credits

- [Evgeniy Kulikov](https://github.com/im-kulikov) - Author
- [Alexander Tischenko](https://github.com/archaron) - Consultant
- [Contributors](https://github.com/im-kulikov/helium/graphs/contributors)