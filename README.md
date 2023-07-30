#### 适用于监听本地端口，并转发至目标端口

``` shell
[tcp/udp] [本地监听端口] [目标地址]
go run main.go tcp 8080 localhost:9090
```
