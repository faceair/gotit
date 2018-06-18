<h1 align="center">Gotit</h1>
<p align="center">Help You Got It (golang dependencies)</p>

<p align="center">
    <a href="https://raw.githubusercontent.com/faceair/gotit/master/LICENSE"><img src="https://img.shields.io/hexpm/l/plug.svg" alt="License"></a>
    <a href="https://travis-ci.org/faceair/gotit"><img src="https://img.shields.io/travis/faceair/gotit/master.svg?t=1529307051" alt="Travis branch"></a>
    <a href="https://coveralls.io/github/faceair/gotit?branch=master"><img src="https://coveralls.io/repos/github/faceair/gotit/badge.svg?branch=master&t=1529307051" alt="Coverage Status"></a>
    <a href="https://goreportcard.com/report/github.com/faceair/gotit"><img src="https://goreportcard.com/badge/github.com/faceair/gotit?t=1529307051" alt="Go Report Card"></a>
    <a href="https://godoc.org/github.com/faceair/gotit"><img src="https://godoc.org/github.com/faceair/gotit?status.svg" alt="GoDoc"></a>
</p>

[中文文档](README.zh.md)

Gotit is a Golang package cache proxy, proudly powered by [betproxy](https://github.com/faceair/betproxy).

Just change the proxy settings of your package management tool to Gotit, and Gotit will automatically pull, cache and update all dependencies for you.

## Features

- **Faster** Pulling is very fast when hitting the cache.
- **Reliable** Gotit can continue working on the cache after disconnecting or deleting the origin repository.
- **Transparency** In theory, Gotit can work with all Go package management tools. (needs to skip HTTPS certificate verification)

## Deployment

### Requirements

Make sure the `git` and `go` executable is on your `PATH` variable.

### Installation

```
go get github.com/faceair/gotit
```

### Run

Run Gotit on port 8080
```
$GOPATH/bin/gotit -port 8080
```
Run `gotit` directly see help for other commands. Gotit use system `GOPATH` to save dependencies by default.

### Configure dependency management tool

#### dep

dep don't support skip HTTPS certificate verification, we need [patch](https://github.com/faceair/dep/commit/43c5e6bf4597bc644a9326d16849b986076b7921) dep. You can build it yourself in this [fork repository](https://github.com/faceair/dep) or [download modified binary files](https://github.com/faceair/dep/releases/latest).

Then set HTTPS_PROXY to Gotit address
```
HTTPS_PROXY=http://127.0.0.1:8080 dep ensure -v
```
or
```
export HTTPS_PROXY=http://127.0.0.1:8080
dep ensure -v
```

#### go get

```
HTTPS_PROXY=http://127.0.0.1:8080 GIT_SSL_NO_VERIFY=true go get -v -insecure github.com/faceair/gotit
```

#### other

TODO

## FAQ

1. When does Gotit update the repository?

After the client pulls the code, Gotit checks the repository for updates. So if you do not update to the latest version at a time, you can wait and try again.

2. What is the difference between Gotit and the forward proxy with cache?

Pull code in git http protocol is a post request, it cannot be cached.
