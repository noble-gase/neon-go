# neno-go

[![golang](https://img.shields.io/badge/Language-Go-green.svg?style=flat)](https://golang.org)
[![pkg.go.dev](https://img.shields.io/badge/dev-reference-007d9c?logo=go&logoColor=white&style=flat)](https://pkg.go.dev/github.com/noble-gase/neon)
[![MIT](http://img.shields.io/badge/license-MIT-brightgreen.svg)](http://opensource.org/licenses/MIT)

[氖-Neon] Go开发工具包

```shell
go get -u github.com/noble-gase/neon
```

| 模块      | 说明                                                                                         |
| --------- | -------------------------------------------------------------------------------------------- |
| array     | 切片常用操作                                                                                 |
| conv      | 类型转换                                                                                     |
| coord     | 距离、方位角、经纬度与平面直角坐标系的相互转化                                               |
| cryptokit | 封装 Crypto 常用方法，支持： `AES` 和 `RSA`                                                  |
| helper    | 常用的辅助方法合集，包含：HTTP、IP、VersionCompare 等                                        |
| httpzip   | 远程获取 `ZIP` 压缩包中的文件内容                                                            |
| hashkit   | 封装 Hash 常用方法                                                                           |
| imgkit    | 图片处理，如：缩略图、裁切、标注等                                                           |
| treekit   | 基于泛型的树形结构，可用于：菜单和组织关系等                                                 |
| pbkit     | 实现 `url.Values` 和 `proto.Message` 的相互转换                                              |
| redkit    | 基于 `singleflight` 封装 Redis 常用操作                                                      |
| redlock   | 基于 Redis 的分布式锁                                                                        |
| retry     | 重试操作                                                                                     |
| sqlkit    | 包含DB初始化和事务等封装 和 基于 [`Jet`](https://github.com/go-jet/jet) 的 curd 封装         |
| stepkit   | 分批次处理切片                                                                               |
| kvkit     | 用于处理 `k-v` 格式化的场景，如：生成签名串等                                                |
| validkit  | 验证器（基于 [`validator`](https://github.com/go-playground/validator)）支持汉化和自定义规则 |

**Enjoy 😊**
