# Syncer
[![go-report](https://goreportcard.com/badge/github.com/MR5356/syncer)](https://goreportcard.com/report/github.com/MR5356/syncer)
[![release](https://img.shields.io/github/v/release/MR5356/syncer)](https://github.com/MR5356/syncer/releases)

## Usage
```shell
[root@toodo ~] ./syncer -h

Usage:
  syncer [command]

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  help        Help about any command
  image       A registry image sync tool

Flags:
  -d, --debug     enable debug mode
  -h, --help      help for syncer
  -v, --version   version for syncer

Use "syncer [command] --help" for more information about a command.
```

## Commands
### image
```shell
[root@toodo ~] ./syncer image -h

A registry image sync tool implement by Go.

Complete code is available at https://github.com/Mr5356/syncer

Usage:
  syncer image [flags]

Flags:
  -c, --config string   config file path
  -d, --debug           enable debug mode
  -h, --help            help for image
  -p, --proc int        process num (default 10)
  -r, --retries int     retries num (default 3)
  -v, --version         version for image
```
#### config file example
```yaml
auth:
  registry.cn-hangzhou.aliyuncs.com:
    username: your_name
    password: your_password
    insecure: false
  docker.io:
    username: your_name
    password: your_password
    insecure: false
# 镜像同步任务列表
images:
  # 该镜像的所有标签将会进行同步
  registry.cn-hangzhou.aliyuncs.com/toodo/alpine: registry.cn-hangzhou.aliyuncs.com/toodo/test
  # 该镜像会同步到目标仓库，并使用新的tag
  alpine@sha256:1fd62556954250bac80d601a196bb7fd480ceba7c10e94dd8fd4c6d1c08783d5: registry.cn-hangzhou.aliyuncs.com/toodo/test:alpine-latest
  # 该镜像会同步至多个目标仓库，如果目标镜像没有填写tag，将会使用源镜像tag
  alpine:latest:
    - hub1.test.com/library/alpine
    - hub2.test.com/library/alpine
# 最大并行数量
proc: 3
# 最大失败重试次数
retries: 3
```

```shell
[root@toodo ~] ./syncer image -c config.yaml
```

## Star History

[![Star History Chart](https://api.star-history.com/svg?repos=Mr5356/syncer&type=Date)](https://star-history.com/#Mr5356/syncer&Date)
