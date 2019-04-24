<h1 align="center">Gotit</h1>
<p align="center">帮你获取 Golang 依赖</p>

<p align="center">
    <a href="https://raw.githubusercontent.com/faceair/gotit/master/LICENSE"><img src="https://img.shields.io/hexpm/l/plug.svg" alt="License"></a>
    <a href="https://travis-ci.org/faceair/gotit"><img src="https://img.shields.io/travis/faceair/gotit/master.svg?t=1529297795" alt="Travis branch"></a>
    <a href="https://coveralls.io/github/faceair/gotit?branch=master"><img src="https://coveralls.io/repos/github/faceair/gotit/badge.svg?branch=master&t=1529297795" alt="Coverage Status"></a>
    <a href="https://goreportcard.com/report/github.com/faceair/gotit"><img src="https://goreportcard.com/badge/github.com/faceair/gotit?t=1529297795" alt="Go Report Card"></a>
    <a href="https://godoc.org/github.com/faceair/gotit"><img src="https://godoc.org/github.com/faceair/gotit?status.svg" alt="GoDoc"></a>
</p>

[English DOC](README.md)

Gotit 是一个由 [betproxy](https://github.com/faceair/betproxy) 强力驱动的 Golang 的包缓存代理。

只需要将你的包管理工具的代理设置改到 Gotit，之后 Gotit 就能自动帮你拉取、缓存和更新所有的依赖。

## Gotit 有什么特性？

- **快** 命中缓存时拉取速度非常快！
- **可靠** 断网或源库被删后 Gotit 可以依靠缓存继续工作 （天天开十X大都不怕 🙌）
- **透明** 理论上支持所有 Go 包管理工具（需要关闭 HTTPS 证书校验）

## 部署

### 依赖

请确认 `git` 和 `go` 可执行文件在系统的 `PATH` 环境变量里。

### 安装

```
go get github.com/faceair/gotit
```

### 运行

将 Gotit 运行在本机的 8080 端口上
```
$GOPATH/bin/gotit -port 8080
```
直接运行 `gotit` 可以查看其他命令的使用帮助，默认 `gotit` 使用系统 `GOPATH` 保存依赖包。

### 配置

### CI 或个人使用直接信任证书 （推荐）

```
# 导入环境变量和证书
export HTTPS_PROXY=http://127.0.0.1:3128
curl -o /usr/local/share/ca-certificates/gotit.crt http://127.0.0.1:3128/ssl && update-ca-certificates

# 正常使用就好
go get -v github.com/golang/dep/cmd/dep
go get -v github.com/axw/gocov/gocov
go get -v github.com/AlekSi/gocov-xml
go get -v golang.org/x/lint/golint
dep ensure -v
go mod download
glide install -v

# 构建完毕可以删除证书
rm /usr/local/share/ca-certificates/gotit.crt && update-ca-certificates
```

### 配置包管理工具

如果你担心导入证书的安全问题，可以尝试绕过包管理工具的 HTTPS 校验。

#### dep

dep 不支持关闭 HTTPS 证书校验，我们需要打上自己的 [Patch](https://github.com/faceair/dep/commit/43c5e6bf4597bc644a9326d16849b986076b7921)。你可以自己在这份 [Fork 仓库](https://github.com/faceair/dep)构建，或者[下载修改后的二进制文件](https://github.com/faceair/dep/releases/latest)。

然后使用时指定 HTTPS_PROXY 到 Gotit 监听的端口上
```
HTTPS_PROXY=http://127.0.0.1:8080 dep ensure -v
```
或者
```
export HTTPS_PROXY=http://127.0.0.1:8080
dep ensure -v
```

#### glide

glide 也需要打 [Patch](https://github.com/faceair/glide/commit/54b926f67677ed3a5e1b6354c491907d8e2cfabf), [下载二进制文件](https://github.com/faceair/glide/releases/latest).

```
HTTPS_PROXY=http://127.0.0.1:8080 glide install
```

#### go get

```
HTTPS_PROXY=http://127.0.0.1:8080 GIT_SSL_NO_VERIFY=true go get -v -insecure github.com/faceair/gotit
```

## 常见问题

1. Gotit 什么时候更新仓库？

客户端拉取代码后 Gotit 会检查这个仓库的更新。所以如果你一次拉取没有更新到最新的版本，可以稍等再重试。

2. Gotit 跟带缓存的正向代理有什么区别？

Git HTTP 协议中拉取代码是 POST 请求，无法缓存。
