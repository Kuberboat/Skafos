package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"p9t.io/kuberboat/pkg/api/core"
	pb "p9t.io/skafos/pkg/proto"
)

var CONN_TIMEOUT time.Duration = time.Second

type SkPilotClient struct {
	connection *grpc.ClientConn
	client     pb.SkpilotSkagentServiceClient
}

func NewClient(skPilotIP string, skPilotPort uint16) (*SkPilotClient, error) {
	addr := fmt.Sprintf("%v:%v", skPilotIP, skPilotPort)
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, errors.New("skagent failed to connect control plane")
	}
	return &SkPilotClient{
		connection: conn,
		client:     pb.NewSkpilotSkagentServiceClient(conn),
	}, nil
}

func (c *SkPilotClient) RegisterSelf(node *core.Node) (*pb.DefaultResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), CONN_TIMEOUT)
	defer cancel()
	data, err := json.Marshal(node)
	if err != nil {
		return &pb.DefaultResponse{Status: -1}, err
	}
	return c.client.RegisterSelf(ctx, &pb.RegisterSelfRequest{
		Node: data,
	})
}
