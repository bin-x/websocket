# websocket
一款支持分布式部署的websocket框架。
设计思路来源GatewayWorker框架：http://doc2.workerman.net/

## 依赖
* [gorilla/websocket](https://github.com/gorilla/websocket) : 底层使用gorilla/websocket进行连接。
* [grpc](https://github.com/grpc/grpc-go) : 不同服务之间通过grpc进行通信。


## 工作原理
整个系统包括两个部分：register和service。
1. 启动register和service
2. service向register发起长连接注册自己，将自己的rpc地址发送给register
3. register收到service消息后保存在内存中，并广播所有service的rpc地址给所有的service。
4. service收到register的广播后，将所有的service都保存到本地
5. service在需要连接其他service时创建rpc连接（目前使用grpc），并将连接保存到内存中（心跳机制保证长连接）。
6. service通过rpc和其他service进行交互
7. 业务逻辑代码需要实现websocket.Application接口。包括OnConnect(string)，OnMessage(string, []byte)，OnClose(string)，分别对应创立连接，收到消息，关闭连接时的操作

## 安装
使用go mod进行包管理，在项目中使用以下命令：
```
go get -u github.com/bin-x/websocket
```

## 使用：
复制 examples/default/ 文件夹中的所有文件到你的项目目录中。
```
register/start_register.go: 注册中心启动文件
app/start_register.go: 服务启动文件，可以修改监听端口及内网ip
app/App.go： 编写具体业务逻辑。你只需要关心该文件即可。
```

app/app.go 包含以下内容:
```
import (
	. "github.com/bin-x/websocket"
)

type App struct {
}

// called when new websocket client connected. Only called once per connection.
// @params clientId will create by system. this is unique in the entire distributed system
func (app *App) OnConnect(clientId string) {
	Api.SendToAll([]byte("hello"))
}

// Called every time the service receives a message from the client
func (app *App) OnMessage(clientId string, message []byte) {
}

// called before close the connection. Only called once per connection.
func (app *App) OnClose(clientId string) {
}
```
```
定义了一个结构体App，该结构体实现了接口 github.com/bin-x/websocket/Application。
接口中包含三个方法：
OnConnect: 客户端连接到服务端时调用，每个连接只会调用一次。可以做些初始化操作，如建立数据库连接等。
OnMessage: 每次服务端收到客户端发来消息时都会调用。最重要的业务单元。
OnClose: 客户端断开连接时调用，做些清理释放工作。

在OnConnect方法中包含了一个结构体的实例：Api。
该实例包含所有可在业务逻辑中使用的方法，下文有具体方法介绍。
```

### example
[chat-app](https://github.com/bin-x/websocket/tree/master/examples/chat-app)

## 业务接口
业务中可使用的接口放在api.go文件中，包含以下接口：
 
 | 接口 | 描述 |
 | ---- | --- |
 |func SendToAll(message []byte) | 发送消息给所有客户端 |
 |func SendToClient(clientId string, message []byte) | 发送消息给某个客户端|
 |func SendToUid(uid string, message []byte) |  发送消息给某个uid|
 |func SendToGroup(group string, message []byte) | 发送消息给某个分组|
 |func BindUid(clientId, uid string) | 绑定uid到某个client|
 |func UnbindUid(clientId string) |  解绑uid|
 |func IsUidOnline(uid string) bool |   判断某个uid是否在线|
 |func GetUidByClientId(clientId string) string |   通过clientId获取对应的uid|
 |func GetClientIdsByUid(uid string) []string |  通过uid获取对应的clientId，多个client可以绑定到同一个uid，所以该函数返回[]string|
 |func JoinGroup(clientId, group string) |  加入到某个分组|
 |func LeaveGroup(clientId, group string) |  离开某个分组|
 |func GetClientCountByGroup(group string) int | 获取某个分组的client数目|
 |func GetClientIdsByGroup(group string) string | 获取某个分组的所有的clientId|
 |func GetUidsByGroup(group string) string | 获取某个分组的所有uid|
 |func GetUidCountByGroup(group string) int | 获取某个分组的uid数目|
 |func GetAllUid() []string | 获取所有的uid|
 |func GetAllGroups() []string | 获取所有的分组|
 |func CloseClient(clientId string) |  关闭连接|
 |func IsOnline(clientId string) bool | 判断某个clientId 是否在线|
 |func GetAllClientCount() int | 获取所有client数目|
 |func GetInfo(clientId string) map[string]string | 获取某个client的info信息|
 |func SetInfo(clientId string, info map[string]string) | 全局替换某个client的info信息|
 |func UpdateInfo(clientId string, info map[string]string) | 局部更新某个client的info信息|