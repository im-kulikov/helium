# Helium

[![codecov](https://codecov.io/gh/im-kulikov/helium/branch/master/graph/badge.svg)](https://codecov.io/gh/im-kulikov/helium)
[![CircleCI](https://circleci.com/gh/im-kulikov/helium.svg?style=svg)](https://circleci.com/gh/im-kulikov/helium)
[![Report](https://goreportcard.com/badge/github.com/im-kulikov/helium)](https://goreportcard.com/report/github.com/im-kulikov/helium)
[![GitHub release](https://img.shields.io/github/release/im-kulikov/helium.svg)](https://github.com/im-kulikov/helium)

<img src="./.github/helium.jpg" width="350" alt="logo">

# Documentation

* [About](#about)

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

## Web Module

