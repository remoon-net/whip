## 简介

通过 WebSocket 反代暴露本地或浏览器中的 http 服务

## 安装

```sh
go install remoon.net/wslink@latest
```

## 启动服务

```sh
wslink serve --listen :2234
```

## 使用

```sh
# 注意: peer_id 需要是长度为64位的hex字符串
wslink c ws://127.0.0.1:2234 peer_id http://127.0.0.1:80
# 访问
curl http://peer_id@link.host
```
