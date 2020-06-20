package websocket

import (
	pb "github.com/bin-x/websocket/proto"
	"golang.org/x/net/context"
	"reflect"
	"testing"
)

type testApp struct {
}

func (t *testApp) OnConnect(s string) {
}

func (t *testApp) OnMessage(s string, bytes []byte) {
}

func (t *testApp) OnClose(s string) {
}

var groupString = "group"
var uid1 = "uid1"
var uid2 = "uid2"

func CreateHub() *ServiceHub {
	hub := &ServiceHub{
		clients:       make(map[string]*Client),
		connect:       make(chan *Client),
		close:         make(chan *Client),
		otherServices: make(map[string]*serviceRpcClient),
		addServices:   make(chan map[string]*serviceRpcClient),
		deleteService: make(chan string),
		uidClients:    make(map[string]map[*Client]bool),
		groups:        make(map[string]map[*Client]bool),
		disbandGroup:  make(chan string),
		joinGroup:     make(chan map[*Client]string),
		leaveGroup:    make(chan map[*Client]string),
		bindUid:       make(chan map[*Client]string),
		unbindUid:     make(chan *Client),
		otherAddress:  make(map[string]bool),
		application:   &testApp{},
	}

	client1 := &Client{
		hub:        hub,
		id:         "1",
		uid:        uid1,
		groups:     map[string]bool{groupString: true},
		info:       map[string]string{"age": "11"},
		joinGroup:  make(chan string),
		leaveGroup: make(chan string),
		setInfo:    make(chan map[string]string),
		updateInfo: make(chan map[string]string),
		done:       make(chan bool),
	}
	client2 := &Client{
		hub:        hub,
		id:         "2",
		uid:        uid2,
		groups:     map[string]bool{},
		info:       map[string]string{"age": "21"},
		joinGroup:  make(chan string),
		leaveGroup: make(chan string),
		setInfo:    make(chan map[string]string),
		updateInfo: make(chan map[string]string),
		done:       make(chan bool),
	}

	hub.clients = map[string]*Client{"1": client1, "2": client2}
	hub.groups = map[string]map[*Client]bool{groupString: {client1: true}}
	hub.uidClients = map[string]map[*Client]bool{uid1: {client1: true}, uid2: {client2: true}}
	go hub.run()
	go client1.run()
	go client2.run()
	return hub
}

func Test_rpcMethods_BindUid(t *testing.T) {
	t.Parallel()

	clientId := "2"
	newUid := uid1

	request := &pb.ServiceRequest{
		ClientId: clientId,
		Uid:      newUid,
	}
	rm := &rpcMethods{
		hub: CreateHub(),
	}
	client := rm.hub.clients[clientId]
	oldUid := client.uid
	rm.BindUid(context.Background(), request)

	if _, ok := rm.hub.uidClients[newUid][client]; !ok {
		t.Errorf("BindUid() hub.uidClients's uid not contain this client")
	}

	if _, ok := rm.hub.uidClients[oldUid][client]; ok {
		t.Errorf("BindUid() hub.uidClients's old uid don't delete this client")
	}

	if !reflect.DeepEqual(client.uid, newUid) {
		t.Errorf("BindUid(), client's uid error, got = %v, want %v", client.uid, newUid)
	}
}

//
//func Test_rpcMethods_CloseClient(t *testing.T) {
//	t.Parallel()
//
//	clientId := "1"
//
//	requst := &pb.ServiceRequest{
//		ClientId: clientId,
//	}
//	rm := &rpcMethods{
//		hub: CreateHub(),
//	}
//	client := rm.hub.clients[clientId]
//	uid := client.uid
//	groups := client.groups
//	rm.CloseClient(context.Background(), requst)
//
//	if _, ok := rm.hub.clients[clientId]; ok {
//		t.Errorf("CloseClient() hub.clients not delete this client")
//	}
//
//	if _, ok := rm.hub.uidClients[uid][client]; ok {
//		t.Errorf("BindUid() hub.uidClients not delete this client")
//	}
//
//	for group := range groups {
//		if _, ok := rm.hub.groups[group][client]; ok {
//			t.Errorf("BindUid() hub.groups not delete this client")
//		}
//	}
//}

func Test_rpcMethods_GetAllClientCount(t *testing.T) {
	t.Parallel()
	request := &pb.ServiceRequest{}
	rm := &rpcMethods{
		hub: CreateHub(),
	}
	response, err := rm.GetAllClientCount(context.Background(), request)
	if err != nil {
		t.Errorf("GetAllClientCount() error = %v", err)
		return
	}
	count := len(rm.hub.clients)
	if !reflect.DeepEqual(count, int(response.Count)) {
		t.Errorf("GetAllClientCount(), got = %v, want %v", response.Count, count)
	}
}

func Test_rpcMethods_GetAllGroups(t *testing.T) {
	t.Parallel()
	request := &pb.ServiceRequest{}
	rm := &rpcMethods{
		hub: CreateHub(),
	}
	response, err := rm.GetAllGroups(context.Background(), request)
	if err != nil {
		t.Errorf("GetAllGroups() error = %v", err)
		return
	}
	want := []string{groupString}
	if !EqualWithoutIndex(want, response.Groups) {
		t.Errorf("GetAllGroups(), got = %v, want %v", response.Groups, want)
	}
}

func Test_rpcMethods_GetAllUid(t *testing.T) {
	t.Parallel()
	request := &pb.ServiceRequest{}
	rm := &rpcMethods{
		hub: CreateHub(),
	}
	response, err := rm.GetAllUid(context.Background(), request)
	if err != nil {
		t.Errorf("GetAllUid() error = %v", err)
		return
	}
	want := []string{uid1, uid2}
	if !EqualWithoutIndex(want, response.Uids) {
		t.Errorf("GetAllUid(), got = %v, want %v", response.Uids, want)
	}
}

func Test_rpcMethods_GetClientCountByGroup(t *testing.T) {
	t.Parallel()
	tests := []struct {
		request *pb.ServiceRequest
		want    int32
	}{
		{
			request: &pb.ServiceRequest{Group: groupString},
			want:    1,
		},
		{
			request: &pb.ServiceRequest{Group: "notexist"},
			want:    0,
		},
	}

	for _, tt := range tests {
		rm := &rpcMethods{
			hub: CreateHub(),
		}
		response, err := rm.GetClientCountByGroup(context.Background(), tt.request)
		if err != nil {
			t.Errorf("GetClientCountByGroup() error = %v", err)
			return
		}
		if !reflect.DeepEqual(tt.want, response.Count) {
			t.Errorf("GetClientCountByGroup(), got = %v, want %v", response.Count, tt.want)
		}
	}
}

func Test_rpcMethods_GetClientIdsByGroup(t *testing.T) {
	t.Parallel()
	tests := []struct {
		request *pb.ServiceRequest
		want    []string
	}{
		{
			request: &pb.ServiceRequest{Group: groupString},
			want:    []string{"1"},
		},
		{
			request: &pb.ServiceRequest{Group: "nogroup"},
			want:    []string{},
		},
	}

	for _, tt := range tests {
		rm := &rpcMethods{
			hub: CreateHub(),
		}
		response, err := rm.GetClientIdsByGroup(context.Background(), tt.request)
		if err != nil {
			t.Errorf("GetClientIdsByGroup() error = %v", err)
			return
		}
		if !EqualWithoutIndex(tt.want, response.ClientIds) {
			t.Errorf("GetClientIdsByGroup(), got = %v, want %v", response.ClientIds, tt.want)
		}
	}

}

func Test_rpcMethods_GetClientIdsByUid(t *testing.T) {
	t.Parallel()
	tests := []struct {
		request *pb.ServiceRequest
		want    []string
	}{
		{
			request: &pb.ServiceRequest{Uid: uid1},
			want:    []string{"1"},
		},
		{
			request: &pb.ServiceRequest{Uid: uid2},
			want:    []string{"2"},
		},
	}

	for _, tt := range tests {
		rm := &rpcMethods{
			hub: CreateHub(),
		}
		response, err := rm.GetClientIdsByUid(context.Background(), tt.request)
		if err != nil {
			t.Errorf("GetClientIdsByUid() error = %v", err)
			return
		}
		if !EqualWithoutIndex(tt.want, response.ClientIds) {
			t.Errorf("GetClientIdsByUid(), got = %v, want %v", response.ClientIds, tt.want)
		}
	}
}

func Test_rpcMethods_GetInfo(t *testing.T) {
	t.Parallel()
	tests := []struct {
		request *pb.ServiceRequest
		want    []*pb.Client
	}{
		{
			request: &pb.ServiceRequest{ClientId: "1"},
			want:    []*pb.Client{{Info: map[string]string{"age": "11"}}},
		},
		{
			request: &pb.ServiceRequest{ClientId: "2"},
			want:    []*pb.Client{{Info: map[string]string{"age": "21"}}},
		},
	}

	for _, tt := range tests {
		rm := &rpcMethods{
			hub: CreateHub(),
		}
		response, err := rm.GetInfo(context.Background(), tt.request)
		if err != nil {
			t.Errorf("GetInfo() error = %v", err)
			return
		}

		if !reflect.DeepEqual(response.Clients, tt.want) {
			t.Errorf("GetInfo() got = %v, want %v", response.Clients, tt.want)
		}
	}
}

func Test_rpcMethods_GetUidByClientId(t *testing.T) {
	t.Parallel()
	tests := []struct {
		request *pb.ServiceRequest
		want    []string
	}{
		{
			request: &pb.ServiceRequest{ClientId: "1"},
			want:    []string{uid1},
		},
		{
			request: &pb.ServiceRequest{ClientId: "2"},
			want:    []string{uid2},
		},
		{
			request: &pb.ServiceRequest{ClientId: "notexist"},
			want:    []string{},
		},
	}

	for _, tt := range tests {
		rm := &rpcMethods{
			hub: CreateHub(),
		}
		response, err := rm.GetUidByClientId(context.Background(), tt.request)
		if err != nil {
			t.Errorf("GetUidByClientId() error = %v", err)
			return
		}

		if !EqualWithoutIndex(response.Uids, tt.want) {
			t.Errorf("GetUidByClientId() got = %v, want %v", response.Uids, tt.want)
		}
	}
}

func Test_rpcMethods_GetUidsByGroup(t *testing.T) {
	t.Parallel()
	tests := []struct {
		request *pb.ServiceRequest
		want    []string
	}{
		{
			request: &pb.ServiceRequest{Group: groupString},
			want:    []string{uid1},
		},
		{
			request: &pb.ServiceRequest{Group: "notexist"},
			want:    []string{},
		},
	}

	for _, tt := range tests {
		rm := &rpcMethods{
			hub: CreateHub(),
		}
		response, err := rm.GetUidsByGroup(context.Background(), tt.request)
		if err != nil {
			t.Errorf("GetUidsByGroup() error = %v", err)
			return
		}

		if !EqualWithoutIndex(response.Uids, tt.want) {
			t.Errorf("GetUidsByGroup() got = %v, want %v", response.Uids, tt.want)
		}
	}
}

func Test_rpcMethods_IsOnline(t *testing.T) {
	t.Parallel()
	tests := []struct {
		request *pb.ServiceRequest
		want    bool
	}{
		{
			request: &pb.ServiceRequest{ClientId: "1"},
			want:    true,
		},
		{
			request: &pb.ServiceRequest{ClientId: "2"},
			want:    true,
		},
		{
			request: &pb.ServiceRequest{ClientId: "notexist"},
			want:    false,
		},
	}

	for _, tt := range tests {
		rm := &rpcMethods{
			hub: CreateHub(),
		}
		response, err := rm.IsOnline(context.Background(), tt.request)
		if err != nil {
			t.Errorf("IsOnline() error = %v", err)
			return
		}

		if response.Result != tt.want {
			t.Errorf("IsOnline() got = %v, want %v", response.Result, tt.want)
		}
	}
}

func Test_rpcMethods_IsUidOnline(t *testing.T) {
	t.Parallel()
	tests := []struct {
		request *pb.ServiceRequest
		want    bool
	}{
		{
			request: &pb.ServiceRequest{Uid: uid1},
			want:    true,
		},
		{
			request: &pb.ServiceRequest{Uid: uid2},
			want:    true,
		},
		{
			request: &pb.ServiceRequest{Uid: "notexist"},
			want:    false,
		},
	}

	for _, tt := range tests {
		rm := &rpcMethods{
			hub: CreateHub(),
		}
		response, err := rm.IsUidOnline(context.Background(), tt.request)
		if err != nil {
			t.Errorf("IsUidOnline() error = %v", err)
			return
		}

		if response.Result != tt.want {
			t.Errorf("IsUidOnline() got = %v, want %v", response.Result, tt.want)
		}
	}
}

//func Test_rpcMethods_JoinGroup(t *testing.T) {
//}
//
//func Test_rpcMethods_LeaveGroup(t *testing.T) {
//}
//
//func Test_rpcMethods_SendToAll(t *testing.T) {
//}
//
//func Test_rpcMethods_SendToClient(t *testing.T) {
//}
//
//func Test_rpcMethods_SendToGroup(t *testing.T) {
//}
//
//func Test_rpcMethods_SendToUid(t *testing.T) {
//}
//
//func Test_rpcMethods_SetInfo(t *testing.T) {
//}
//
//func Test_rpcMethods_UnbindUid(t *testing.T) {
//}
//
//func Test_rpcMethods_UpdateInfo(t *testing.T) {
//}
