package websocket

import (
	"context"
	"errors"
	pb "github.com/bin-x/websocket/proto"
	"log"
	"reflect"
)

var Api *ServiceApi

type ServiceApi struct {
	hub *ServiceHub
}

// 判断是否为本地服务，如果为本地服务则不使用rpc
func (s *ServiceApi) isLocal(addr string) bool {
	locaAddr := s.hub.lanIp + ":" + s.hub.rpcPort
	return locaAddr == addr
}

// 调用分布式系统中的服务，并将返回结果合并
func (s *ServiceApi) call(method string, ctx context.Context, request *pb.ServiceRequest) ([]*pb.ServiceResponse, error) {
	//log.Println("call", method)
	var responses []*pb.ServiceResponse
	for addr, _ := range s.hub.otherAddress {
		// 本地服务则直接调用，减少rpc的开销
		if s.isLocal(addr) {
			response, err := call(s.hub.rm, method, ctx, request)
			if err != nil {
				log.Println("call local method error:", err)
				continue
			}
			responses = append(responses, response)
			continue
		}

		client, err := s.hub.getServiceConn(addr)
		if err != nil {
			log.Println("get service conn error:", err)
			continue
		}
		c := pb.NewServiceApiClient(client.conn)
		response, err := call(c, method, ctx, request)
		if err != nil {
			log.Println("call remote method error:", err)
			continue
		}
		responses = append(responses, response)
	}
	return responses, nil
}

// 方法调用
func call(value interface{}, method string, ctx context.Context, request *pb.ServiceRequest) (*pb.ServiceResponse, error) {
	v := reflect.ValueOf(value)
	m := v.MethodByName(method)

	if !m.IsValid() {
		return nil, errors.New("the method " + method + " not exist in" + v.String())
	}
	in := []reflect.Value{
		reflect.ValueOf(ctx),
		reflect.ValueOf(request),
	}

	out := m.Call(in)

	if len(out) != 2 {
		return nil, errors.New("call error")
	}
	response, ok := out[0].Interface().(*pb.ServiceResponse)
	if !ok {
		return nil, errors.New("call error")
	}
	return response, nil
}

// 发送消息给所有客户端
func (s *ServiceApi) SendToAll(message []byte) {
	s.call("SendToAll", context.Background(), &pb.ServiceRequest{Message: message})
}

// 发送消息给某个客户端
func (s *ServiceApi) SendToClient(clientId string, message []byte) {
	s.call("SendToClient", context.Background(), &pb.ServiceRequest{Message: message, ClientId: clientId})
}

// 发送消息给某个uid
func (s *ServiceApi) SendToUid(uid string, message []byte) {
	s.call("SendToUid", context.Background(), &pb.ServiceRequest{Message: message, Uid: uid})
}

// 发送消息给某个分组
func (s *ServiceApi) SendToGroup(group string, message []byte) {
	s.call("SendToGroup", context.Background(), &pb.ServiceRequest{Message: message, Group: group})
}

// 绑定uid
func (s *ServiceApi) BindUid(clientId, uid string) {
	s.call("BindUid", context.Background(), &pb.ServiceRequest{ClientId: clientId, Uid: uid})
}

// 解绑uid
func (s *ServiceApi) UnbindUid(clientId string) {
	s.call("UnbindUid", context.Background(), &pb.ServiceRequest{ClientId: clientId})
}

// 判断某个uid是否在线
func (s *ServiceApi) IsUidOnline(uid string) bool {
	responses, _ := s.call("IsUidOnline", context.Background(), &pb.ServiceRequest{Uid: uid})
	for _, response := range responses {
		if response.Result {
			return true
		}
	}
	return false
}

func (s *ServiceApi) GetUidByClientId(clientId string) string {
	responses, _ := s.call("GetUidByClientId", context.Background(), &pb.ServiceRequest{ClientId: clientId})
	for _, response := range responses {
		if len(response.Uids) > 0 {
			return response.Uids[0]
		}
	}
	return ""
}
func (s *ServiceApi) GetClientIdsByUid(uid string) []string {
	var clientIds []string
	responses, _ := s.call("GetClientIdsByUid", context.Background(), &pb.ServiceRequest{Uid: uid})
	for _, response := range responses {
		for _, clientId := range response.ClientIds {
			clientIds = append(clientIds, clientId)
		}
	}
	return clientIds
}

func (s *ServiceApi) JoinGroup(clientId, group string) {
	s.call("JoinGroup", context.Background(), &pb.ServiceRequest{ClientId: clientId, Group: group})
}
func (s *ServiceApi) LeaveGroup(clientId, group string) {
	s.call("LeaveGroup", context.Background(), &pb.ServiceRequest{ClientId: clientId, Group: group})

}
func (s *ServiceApi) GetClientCountByGroup(group string) int {
	count := 0
	responses, _ := s.call("GetClientCountByGroup", context.Background(), &pb.ServiceRequest{Group: group})
	for _, response := range responses {
		count += int(response.Count)
	}
	return count
}

func (s *ServiceApi) GetClientIdsByGroup(group string) []string {
	var clientIds []string
	responses, _ := s.call("GetClientIdsByGroup", context.Background(), &pb.ServiceRequest{Group: group})
	for _, response := range responses {
		for _, clientId := range response.ClientIds {
			clientIds = append(clientIds, clientId)
		}
	}
	return clientIds

}
func (s *ServiceApi) GetUidsByGroup(group string) []string {
	// 使用map去重
	uidMaps := make(map[string]bool)
	responses, _ := s.call("GetUidsByGroup", context.Background(), &pb.ServiceRequest{Group: group})
	for _, response := range responses {
		for _, uid := range response.Uids {
			uidMaps[uid] = true
		}
	}
	uids := make([]string, 0, len(uidMaps))
	for uid := range uidMaps {
		uids = append(uids, uid)
	}
	return uids
}

func (s *ServiceApi) GetUidCountByGroup(group string) int {
	uids := s.GetUidsByGroup(group)
	return len(uids)
}

//
func (s *ServiceApi) GetAllUid() []string {
	uidMaps := make(map[string]bool)
	responses, _ := s.call("GetAllUid", context.Background(), &pb.ServiceRequest{})
	for _, response := range responses {
		for _, uid := range response.Uids {
			uidMaps[uid] = true
		}
	}
	uids := make([]string, 0, len(uidMaps))
	for uid := range uidMaps {
		uids = append(uids, uid)
	}
	return uids
}

func (s *ServiceApi) GetAllGroups() []string {
	groupMaps := make(map[string]bool)
	responses, _ := s.call("GetAllGroups", context.Background(), &pb.ServiceRequest{})
	for _, response := range responses {
		for _, group := range response.Groups {
			groupMaps[group] = true
		}
	}
	groups := make([]string, 0, len(groupMaps))
	for group := range groupMaps {
		groups = append(groups, group)
	}
	return groups
}
func (s *ServiceApi) CloseClient(clientId string) {
	s.call("CloseClient", context.Background(), &pb.ServiceRequest{ClientId: clientId})
}
func (s *ServiceApi) IsOnline(clientId string) bool {
	responses, _ := s.call("IsOnline", context.Background(), &pb.ServiceRequest{ClientId: clientId})
	for _, response := range responses {
		if response.Result {
			return true
		}
	}
	return false
}
func (s *ServiceApi) GetAllClientCount() int {
	count := 0
	responses, _ := s.call("GetAllClientCount", context.Background(), &pb.ServiceRequest{})
	for _, response := range responses {
		count += int(response.Count)
	}
	return count
}

//
func (s *ServiceApi) GetInfo(clientId string) map[string]string {
	responses, _ := s.call("GetInfo", context.Background(), &pb.ServiceRequest{ClientId: clientId})
	for _, response := range responses {
		for _, client := range response.Clients {
			return client.Info
		}
	}
	return make(map[string]string)
}

// 全局替换
func (s *ServiceApi) SetInfo(clientId string, info map[string]string) {
	s.call("SetInfo", context.Background(), &pb.ServiceRequest{ClientId: clientId, Info: info})
}

// 局部更新
func (s *ServiceApi) UpdateInfo(clientId string, info map[string]string) {
	s.call("UpdateInfo", context.Background(), &pb.ServiceRequest{ClientId: clientId, Info: info})
}
