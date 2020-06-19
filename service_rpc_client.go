package websocket

import (
	"google.golang.org/grpc"
)

type serviceRpcClient struct {
	hub  *ServiceHub
	conn *grpc.ClientConn
}

//
//func (client *serviceRpcClient) Call(method string, request pb.ServiceRequest)(pb.ServiceResponse, error) {
//}
