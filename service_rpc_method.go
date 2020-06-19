package websocket

import (
	"golang.org/x/net/context"
	"log"
	"strconv"
	pb "websocket/proto"
)

type rpcMethods struct {
	hub *ServiceHub
}

func (rm *rpcMethods) GetUidByClientId(ctx context.Context, request *pb.ServiceRequest) (*pb.ServiceResponse, error) {
	var uids []string
	if client, ok := rm.hub.clients[request.ClientId]; ok {
		uids = append(uids, client.uid)
	}
	return &pb.ServiceResponse{Uids: uids}, nil

}

func (rm *rpcMethods) GetClientIdsByUid(ctx context.Context, request *pb.ServiceRequest) (*pb.ServiceResponse, error) {
	var clientIds []string
	if clients, ok := rm.hub.uidClients[request.Uid]; ok {
		for client, _ := range clients {
			clientIds = append(clientIds, client.id)
		}
	}
	return &pb.ServiceResponse{ClientIds: clientIds}, nil
}

func (rm *rpcMethods) GetClientIdsByGroup(ctx context.Context, request *pb.ServiceRequest) (*pb.ServiceResponse, error) {
	var clientIds []string
	if clients, ok := rm.hub.groups[request.Group]; ok {
		clientIds = make([]string, len(clients))
		for client, _ := range clients {
			clientIds = append(clientIds, client.id)
		}
	}
	return &pb.ServiceResponse{ClientIds: clientIds}, nil
}

func (rm *rpcMethods) GetUidsByGroup(ctx context.Context, request *pb.ServiceRequest) (*pb.ServiceResponse, error) {
	var uidMaps map[string]bool
	if clients, ok := rm.hub.groups[request.Group]; ok {
		for client, _ := range clients {
			uidMaps[client.uid] = true
		}
	}
	length := len(uidMaps)
	var uids = make([]string, len(uidMaps), length)
	for uid, _ := range uidMaps {
		uids = append(uids, uid)
	}
	return &pb.ServiceResponse{Uids: uids}, nil
}

func (rm *rpcMethods) GetAllGroups(ctx context.Context, request *pb.ServiceRequest) (*pb.ServiceResponse, error) {
	groups := make([]string, len(rm.hub.groups))
	for group, _ := range rm.hub.groups {
		groups = append(groups, group)
	}
	return &pb.ServiceResponse{Groups: groups}, nil

}

func (rm *rpcMethods) GetInfo(ctx context.Context, request *pb.ServiceRequest) (*pb.ServiceResponse, error) {
	var clients []*pb.Client
	if c, ok := rm.hub.clients[request.ClientId]; ok {
		clients = append(clients, &pb.Client{Info: c.info})
	}
	return &pb.ServiceResponse{Clients: clients}, nil
}

func (rm *rpcMethods) GetClientCountByGroup(ctx context.Context, request *pb.ServiceRequest) (*pb.ServiceResponse, error) {
	count := len(rm.hub.groups[request.Group])
	return &pb.ServiceResponse{Count: int32(count)}, nil
}

func (rm *rpcMethods) GetUidCountByGroup(ctx context.Context, request *pb.ServiceRequest) (*pb.ServiceResponse, error) {
	var uidMaps map[string]bool
	if clients, ok := rm.hub.groups[request.Group]; ok {
		for client, _ := range clients {
			uidMaps[client.uid] = true
		}
	}
	count := len(uidMaps)
	return &pb.ServiceResponse{Count: int32(count)}, nil
}

func (rm *rpcMethods) SendToClient(ctx context.Context, request *pb.ServiceRequest) (*pb.ServiceResponse, error) {
	if client, ok := rm.hub.clients[request.ClientId]; ok {
		client.send <- request.Message
	}
	return &pb.ServiceResponse{}, nil
}

func (rm *rpcMethods) SendToUid(ctx context.Context, request *pb.ServiceRequest) (*pb.ServiceResponse, error) {
	if clients, ok := rm.hub.uidClients[request.Uid]; ok {
		for client := range clients {
			client.send <- request.Message
		}
	}
	return &pb.ServiceResponse{}, nil
}

func (rm *rpcMethods) SendToGroup(ctx context.Context, request *pb.ServiceRequest) (*pb.ServiceResponse, error) {
	if clients, ok := rm.hub.groups[request.Group]; ok {
		for client := range clients {
			client.send <- request.Message
		}
	}
	return &pb.ServiceResponse{}, nil
}

func (rm *rpcMethods) BindUid(ctx context.Context, request *pb.ServiceRequest) (*pb.ServiceResponse, error) {
	if client, ok := rm.hub.clients[request.ClientId]; ok {
		data := make(map[*Client]string)
		data[client] = request.Uid
		rm.hub.bindUid <- data
	}
	return &pb.ServiceResponse{}, nil
}

func (rm *rpcMethods) UnbindUid(ctx context.Context, request *pb.ServiceRequest) (*pb.ServiceResponse, error) {
	if client, ok := rm.hub.clients[request.ClientId]; ok {
		rm.hub.unbindUid <- client
	}
	return &pb.ServiceResponse{}, nil
}

func (rm *rpcMethods) IsUidOnline(ctx context.Context, request *pb.ServiceRequest) (*pb.ServiceResponse, error) {
	_, ok := rm.hub.uidClients[request.Uid]
	return &pb.ServiceResponse{Result: ok}, nil
}

func (rm *rpcMethods) JoinGroup(ctx context.Context, request *pb.ServiceRequest) (*pb.ServiceResponse, error) {
	if client, ok := rm.hub.clients[request.ClientId]; ok {
		data := make(map[*Client]string)
		data[client] = request.Group
		rm.hub.joinGroup <- data
	}
	return &pb.ServiceResponse{}, nil
}

func (rm *rpcMethods) LeaveGroup(ctx context.Context, request *pb.ServiceRequest) (*pb.ServiceResponse, error) {
	if client, ok := rm.hub.clients[request.ClientId]; ok {
		rm.hub.leaveGroup <- map[*Client]string{client: request.Group}
	}
	return &pb.ServiceResponse{}, nil
}

func (rm *rpcMethods) CloseClient(ctx context.Context, request *pb.ServiceRequest) (*pb.ServiceResponse, error) {
	if client, ok := rm.hub.clients[request.ClientId]; ok {
		rm.hub.close <- client
	}
	return &pb.ServiceResponse{}, nil
}

func (rm *rpcMethods) IsOnline(ctx context.Context, request *pb.ServiceRequest) (*pb.ServiceResponse, error) {
	_, ok := rm.hub.clients[request.ClientId]
	return &pb.ServiceResponse{Result: ok}, nil
}

func (rm *rpcMethods) UpdateInfo(ctx context.Context, request *pb.ServiceRequest) (*pb.ServiceResponse, error) {
	if client, ok := rm.hub.clients[request.ClientId]; ok {
		client.updateInfo <- request.Info
	}

	log.Println("call SetInfo.")

	return &pb.ServiceResponse{Success: true}, nil
}

func (rm *rpcMethods) SendToAll(ctx context.Context, request *pb.ServiceRequest) (*pb.ServiceResponse, error) {
	log.Println("call SendToAll. message: ", request.Message)

	for _, client := range rm.hub.clients {
		client.send <- request.Message
	}
	log.Println("client length:", len(rm.hub.clients))

	return &pb.ServiceResponse{Success: true}, nil
}

func (rm *rpcMethods) GetAllClientCount(ctx context.Context, request *pb.ServiceRequest) (*pb.ServiceResponse, error) {
	log.Println("call GetAllClientCount. count:" + strconv.Itoa(len(rm.hub.clients)))

	count := int32(len(rm.hub.clients))
	return &pb.ServiceResponse{Success: true, Count: count}, nil
}

func (rm *rpcMethods) GetAllUid(ctx context.Context, request *pb.ServiceRequest) (*pb.ServiceResponse, error) {
	uids := make([]string, len(rm.hub.uidClients))
	for uid, _ := range rm.hub.uidClients {
		uids = append(uids, uid)
	}
	return &pb.ServiceResponse{Uids: uids}, nil
}

func (rm *rpcMethods) SetInfo(ctx context.Context, request *pb.ServiceRequest) (*pb.ServiceResponse, error) {
	if client, ok := rm.hub.clients[request.ClientId]; ok {
		client.setInfo <- request.Info
	}

	log.Println("call SetInfo.")

	return &pb.ServiceResponse{Success: true}, nil
}

//
//func (rm *rpcMethods) GetAllInfos(ctx context.Context, request *pb.ServiceRequest) (*pb.ServiceResponse, error) {
//	log.Println("call GetAllInfos.")
//
//	var cs []*pb.Client
//	for _, client := range rm.hub.clients {
//		info, err := json.Marshal(client.info)
//		if err != nil {
//			continue
//		}
//		cs = append(cs, &pb.Client{Info: info})
//	}
//	return &pb.ServiceResponse{Success: true, Clients: cs}, nil
//}
