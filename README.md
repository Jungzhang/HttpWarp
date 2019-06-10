# HttpWarp

一个简单的websocket代理，可对TCP应用进行代理，达到隐藏数据包特征和使用CDN进行加速的目的。

### QuickStart

#### 环境要求

```
golang 版本 >= 1.9.1
```

#### 相关概念

定义如下概念，下文中名词请自行对号入座

> ws服务端 ：本服务对应的服务端
> 
> TCP服务端：被代理的TCP服务的服务端
> 
> ws客户端：本服务对应的客户端
> 
> TCP客户端：被代理的TCP服务的客户端

#### 安装

下载

```
go get -u github.com/Jungzhang/HttpWarp
```

编译

```
cd $GOPATH/src/github.com/Jungzhang/HttpWarp
go build -o server 或 go build -o client
```
> 注意：当部署ws服务端时可执行文件名称必须为`server`

启动服务端

```
./server [-p 8080] [-u /images/upload]
```
> 参数说明：
> 
>- p：ws服务端监听端口。该选项可选, 默认为80端口
>- u：用于ws代理服务收发数据的url地址, ws服务端和ws客户端需保证该地址一致。该选项可选，默认为：/data/put

启动客户端

```
./client -p 8001 -d github.com [-i 192.168.0.120] [-l 10087] [-u /images/upload]
```

> 参数说明：
> 
>- p：需要转发到的`TCP服务端`的服务端口，即ws服务端代理的后端TCP服务端口。该选项必选
>- d：ws服务端所在域名或ip地址。该选项必选
>- i：TCP服务端对应地址，默认为`127.0.0.1`
>- l：ws客户端监听的本地代理端口，默认为`10086`
>- u：用于ws客户端和ws服务端之间传输数据的url地址，默认为`/data/put`。注意该地址必须和ws服务端地址保持一致。

TCP客服端配置

> 将TCP客户端对应的TCP服务端地址修改为`127.0.0.1:本地监听端口(默认为10086)`

或

> 在本地TCP客户端中配置http代理到：`127.0.0.1:本地监听端口(默认为10086)`

#### 具体案例

- 使用HttpWarp&CDN复活被墙ip：[待补充]() 

