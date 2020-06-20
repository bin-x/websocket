package websocket

import (
	"encoding/json"
	"errors"
	pb "github.com/bin-x/websocket/proto"
	"github.com/gorilla/websocket"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"log"
	"net"
	"net/http"
	"net/url"
	"time"
)

type ServiceHub struct {
	rm *rpcMethods

	registerAddr string
	rpcPort      string
	lanIp        string

	clients    map[string]*Client
	uidClients map[string]map[*Client]bool
	groups     map[string]map[*Client]bool

	connect chan *Client
	close   chan *Client

	addServices   chan map[string]*serviceRpcClient
	deleteService chan string

	disbandGroup chan string
	joinGroup    chan map[*Client]string
	leaveGroup   chan map[*Client]string
	bindUid      chan map[*Client]string
	unbindUid    chan *Client

	application   Application
	otherAddress  map[string]bool
	otherServices map[string]*serviceRpcClient
}

func NewServiceHub(registerAddr, rpcPort, lanIp string, application Application) *ServiceHub {
	return &ServiceHub{
		registerAddr: registerAddr,
		rpcPort:      rpcPort,
		lanIp:        lanIp,

		clients: make(map[string]*Client),
		connect: make(chan *Client),
		close:   make(chan *Client),

		otherServices: make(map[string]*serviceRpcClient),
		addServices:   make(chan map[string]*serviceRpcClient),
		deleteService: make(chan string),

		uidClients:   make(map[string]map[*Client]bool),
		groups:       make(map[string]map[*Client]bool),
		disbandGroup: make(chan string),
		joinGroup:    make(chan map[*Client]string),
		leaveGroup:   make(chan map[*Client]string),
		bindUid:      make(chan map[*Client]string),
		unbindUid:    make(chan *Client),
		application:  application,
		otherAddress: make(map[string]bool),
	}
}

func (sh *ServiceHub) run() {
	for {
		select {
		//新链接
		case client := <-sh.connect:
			sh.clients[client.id] = client
		// 断开链接
		case client := <-sh.close:
			delete(sh.clients, client.id)
			// 从uid中删除
			delete(sh.uidClients[client.uid], client)
			// 从group中删除
			for group := range client.groups {
				delete(sh.groups[group], client)
			}
		//解散组,组中每个成员的group都要去掉该租
		case group := <-sh.disbandGroup:
			if clients, ok := sh.groups[group]; ok {
				for client := range clients {
					client.leaveGroup <- group
				}
				delete(sh.groups, group)
			}
		// 加入到某组
		case data := <-sh.joinGroup:
			for client, group := range data {
				client.joinGroup <- group
				// 分组不存在则创建分组
				if _, ok := sh.groups[group]; !ok {
					sh.groups[group] = make(map[*Client]bool)
				}
				sh.groups[group][client] = true

			}
		// 退出分组
		case data := <-sh.leaveGroup:
			for client, group := range data {
				client.leaveGroup <- group
				if _, ok := sh.groups[group]; !ok {
					break
				}
				delete(sh.groups[group], client)

				// 分组无成员则删除分组
				if len(sh.groups[group]) == 0 {
					delete(sh.groups, group)
				}
			}
		// 绑定uid
		case data := <-sh.bindUid:
			for client, uid := range data {
				oldUid := client.uid
				delete(sh.uidClients[oldUid], client)

				client.uid = uid
				if _, ok := sh.uidClients[uid]; !ok {
					sh.uidClients[uid] = make(map[*Client]bool)
				}
				sh.uidClients[uid][client] = true
			}
		// 解绑uid
		case client := <-sh.unbindUid:
			uid := client.uid
			client.uid = ""
			if _, ok := sh.uidClients[uid]; !ok {
				break
			}
			delete(sh.uidClients[uid], client)
			if len(sh.uidClients[uid]) == 0 {
				delete(sh.uidClients, uid)
			}
		case service := <-sh.addServices:
			for addr, client := range service {
				sh.otherServices[addr] = client
			}
		case addr := <-sh.deleteService:
			delete(sh.otherServices, addr)
		}
	}
}

func (sh *ServiceHub) Start(addr string) {
	go sh.run()
	go sh.checkRegisterConnection()
	go sh.StartRpc()

	Api = &ServiceApi{hub: sh}
	log.Println("starting Service...")

	http.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		ServeWs(sh, writer, request)
	})

	err := http.ListenAndServe(addr, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

// 开启rpc服务
func (sh *ServiceHub) StartRpc() {
	listen, err := net.Listen("tcp", ":"+sh.rpcPort)

	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	rm := rpcMethods{hub: sh}
	sh.rm = &rm
	enforcementPolicy := keepalive.EnforcementPolicy{
		MinTime:             60 * time.Second,
		PermitWithoutStream: true,
	}
	serverParameters := keepalive.ServerParameters{
		Time:    30 * time.Second,
		Timeout: 5 * time.Second,
	}

	s := grpc.NewServer(grpc.KeepaliveEnforcementPolicy(enforcementPolicy), grpc.KeepaliveParams(serverParameters))
	pb.RegisterServiceApiServer(s, sh.rm)
	log.Println("rpc服务已经开启")
	s.Serve(listen)
}

//链接到register
func (sh *ServiceHub) connectToRegister() error {
	u := url.URL{Scheme: "ws", Host: sh.registerAddr, Path: "/"}
	log.Printf("connecting to %s", u.String())

	//创建到register的链接
	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return err
	}
	defer c.Close()

	// 通过done判断read通道是否关闭，关闭则结束链接
	done := make(chan struct{})
	//获取register的消息。
	go func() {
		defer close(done)
		for {
			_, data, err := c.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				return
			}
			var message RegisterMessage
			err = json.Unmarshal(data, &message)
			if err != nil {
				log.Println("message err:", string(data[:]))
			}
			log.Println("read message from register:", message)
			switch message.Action {
			case registerActionBroadcastAddr:
				sh.otherAddress = map[string]bool{}
				for _, addr := range message.Addresses {
					sh.otherAddress[addr] = true
				}
				log.Println(sh.otherAddress)
			}
		}
	}()

	message := RegisterMessage{
		Action:  registerActionConnect,
		RpcAddr: sh.lanIp + ":" + sh.rpcPort,
	}

	// 发送注册信息给register
	err = c.WriteJSON(&message)
	if err != nil {
		return err
	}

	//保持链接
	for {
		select {
		// 如果read通道关闭，则结束链接
		case <-done:
			log.Println("done")
			return errors.New("done")
		}
	}
}

//保持和register的链接，断开连接后自动重连
func (sh *ServiceHub) checkRegisterConnection() {
	for {
		sh.connectToRegister()
		time.Sleep(10 * time.Second)
	}
}

func (sh *ServiceHub) getServiceConn(addr string) (*serviceRpcClient, error) {
	if client, ok := sh.otherServices[addr]; ok {
		return client, nil
	}

	clientParameters := keepalive.ClientParameters{
		Time:                10 * time.Second,
		Timeout:             time.Second,
		PermitWithoutStream: true,
	}

	conn, err := grpc.Dial(addr, grpc.WithInsecure(), grpc.WithKeepaliveParams(clientParameters))
	if err != nil {
		return nil, err
	}

	rpcClient := serviceRpcClient{hub: sh, conn: conn}
	client := make(map[string]*serviceRpcClient)
	client[addr] = &rpcClient
	sh.addServices <- client
	return &rpcClient, nil
}
