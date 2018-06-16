<h1 align="center">Gotit</h1>
<p align="center">帮你获取 Golang 依赖</p>

<p align="center">
    <a href="https://raw.githubusercontent.com/faceair/gotit/master/LICENSE"><img src="https://img.shields.io/hexpm/l/plug.svg" alt="License"></a>
    <a href="https://travis-ci.org/faceair/gotit"><img src="https://img.shields.io/travis/faceair/gotit/master.svg" alt="Travis branch"></a>
    <a href="https://coveralls.io/github/faceair/gotit?branch=master"><img src="https://coveralls.io/repos/github/faceair/gotit/badge.svg?branch=master" alt="Coverage Status"></a>
    <a href="https://goreportcard.com/report/github.com/faceair/gotit"><img src="https://goreportcard.com/badge/github.com/faceair/gotit" alt="Go Report Card"></a>
    <a href="https://godoc.org/github.com/faceair/gotit"><img src="https://godoc.org/github.com/faceair/gotit?status.svg" alt="GoDoc"></a>
</p>

[English DOC](README.md)

Gotit 是一个 Golang 的包缓存代理。只需要将你的包管理工具的代理设置改到 Gotit，之后 Gotit 就能自动帮你拉取、缓存和更新所有的依赖。

## Gotit 有什么特性？

- 将 Gotit 部署在企业内网中可以加快拉取依赖的速度，同时减少外网带宽使用
- 越多人使用 Gotit，缓存下来的的依赖包会越全，加速效果越明显
- Gotit 有缓存，极端情况下断外网后还可以继续提供服务，比依赖外网构建更可靠（FXXK GFW）
- Gotit 能自动更新依赖包，不用担心缓存版本过旧的问题
- 对包管理工具透明，理论支持所有 Go 包管理工具（需要关闭 HTTPS 证书校验）

## 部署

### 依赖

请确认 `git` 和 `go` 可执行文件在系统的 `PATH` 环境变量里。

### 安装

```
go install github.com/faceair/gotit
```

### 运行

将 Gotit 运行在本机的 8080 端口上
```
$GOPATH/bin/gotit -port 8080
```
直接运行 `gotit` 可以查看其他命令行参数的使用帮助，默认 `gotit` 使用系统 `GOPATH` 保存依赖包。

### 配置包管理工具

#### dep

dep 不支持关闭 HTTPS 证书校验，我们需要打上自己的 [Patch](https://github.com/faceair/dep/commit/19ef30fc8abae44709d5a732f34065d2919d8377)。你可以自己在这份 [Fork 源码](https://github.com/faceair/dep)构建，或者[下载修改后的二进制文件](https://github.com/faceair/dep/releases/tag/v0.4.2)。

然后使用时指定 HTTPS_PROXY 到 Gotit 监听的端口上
```
HTTPS_PROXY=http://127.0.0.1:8080 dep ensure -v
```
或者
```
export HTTPS_PROXY=http://127.0.0.1:8080
dep ensure -v
```

#### 其他包管理工具

TODO

## 常见问题

1. Gotit 跟带缓存的正向代理有什么区别？

Git HTTP 协议中同步代码是 POST 请求，无法缓存。有的依赖走 SSH 协议，无法使用普通 HTTPS 正向代理，Gotit 中使用 go get 绕过了这个问题。

2. Gotit 什么时候更新依赖？

客户端每次拉取代码后 Gotit 会检查这个仓库的更新，所以如果你一次拉取没有更新到最新的代码，可以稍等再重试一次。
