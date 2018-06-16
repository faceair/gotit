<h1 align="center">Gotit</h1>
<p align="center">Help You Got It (golang dependencies)</p>

<p align="center">
    <a href="https://raw.githubusercontent.com/faceair/gotit/master/LICENSE"><img src="https://img.shields.io/hexpm/l/plug.svg" alt="License"></a>
    <a href="https://travis-ci.org/faceair/gotit"><img src="https://img.shields.io/travis/faceair/gotit/master.svg" alt="Travis branch"></a>
    <a href="https://coveralls.io/github/faceair/gotit?branch=master"><img src="https://coveralls.io/repos/github/faceair/gotit/badge.svg?branch=master" alt="Coverage Status"></a>
    <a href="https://goreportcard.com/report/github.com/faceair/gotit"><img src="https://goreportcard.com/badge/github.com/faceair/gotit" alt="Go Report Card"></a>
    <a href="https://godoc.org/github.com/faceair/gotit"><img src="https://godoc.org/github.com/faceair/gotit?status.svg" alt="GoDoc"></a>
</p>

[中文文档](README.zh.md)

Gotit is a Golang package caching proxy. Just change the proxy settings of your package management tool to Gotit, and Gotit will automatically pull, cache, and update all dependencies for you.

## Features

- Deploying Gotit in the corporate intranet can speed up the pull of dependencies while reducing external network bandwidth usage
- The more people use Gotit, the more the cached dependencies will be, the more obvious the acceleration effect will be
- Gotit has caching. In extreme cases, it can continue to provide services after disconnecting the external network. It is more reliable than relying on the external network
- Gotit can automatically update dependencies without worrying about old versions of the cache
- Transparency to package management tools, theoretically supporting for all Go package management tools (needs to turn off HTTPS certificate verification)

## Deployment

### Requirements

Please confirm that the `git` and `go` executable files in the `PATH` environment variable.

### Installation

```
go install github.com/faceair/gotit
```

### Run

Run Gotti on port 8080
```
$GOPATH/bin/gotti -port 8080
```
Run `gotit` directly to view other command-line param's usage help. Gotit use system `GOPATH` to save dependencies by default.

### Configure dependency management tool

#### dep

dep does not support turning off https certificate verification, we need [patch](https://github.com/faceair/dep/commit/19ef30fc8abae44709d5a732f34065d2919d8377) dep. You can build it yourself in this [fork source](https://github.com/faceair/dep) or [download modified binary files](https://github.com/faceair/dep/releases/tag/v0.4.2).

Then set HTTPS_PROXY to Gotit address
```
HTTPS_PROXY=http://127.0.0.1:8080 dep ensure -v
```
or
```
export HTTPS_PROXY=http://127.0.0.1:8080
dep ensure -v
```

#### other
Todo

## FAQ

1. What is the difference between Gotit and the forward proxy with cache?

Synchronization code in git http protocol is a post request so cannot be cached. Some package host on ssh protocol and cannot use normal https forward proxy, `go get` is used in Gotit to bypass this problem.

2. When does Gotit update its dependency?

Each time the client pulls the code, the Gotit will check the update of this repository, so if you do not pull the latest code at a time, you can wait a moment and try again.
