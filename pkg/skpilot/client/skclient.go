package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"p9t.io/skafos/pkg/api/core"
	pb "p9t.io/skafos/pkg/proto"
	"p9t.io/skafos/pkg/skproxy"
)

var SK_CONN_TIMEOUT time.Duration = time.Second * 6

type SkClient struct {
	connection *grpc.ClientConn
	client     pb.SkagentSkpilotServiceClient
}

func NewSkClient(url string, skAgentPort uint16) (*SkClient, error) {
	addr := fmt.Sprintf("%v:%v", url, skAgentPort)
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, errors.New("apiserver client failed to connect to worker node")
	}
	return &SkClient{
		connection: conn,
		client:     pb.NewSkagentSkpilotServiceClient(conn),
	}, nil
}

func (c *SkClient) CreateProxy(infos []core.SandboxInfo) (*pb.DefaultResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), SK_CONN_TIMEOUT)
	defer cancel()

	containerNames := make([]string, 0, len(infos))
	sandboxIPs := make([]string, 0, len(infos))
	for _, info := range infos {
		containerNames = append(containerNames, info.SandboxName)
		sandboxIPs = append(sandboxIPs, info.SandboxIP)
	}

	return c.client.CreateProxy(ctx, &pb.CreateProxyRequest{
		ContainerNames: containerNames,
		SandboxIps:     sandboxIPs,
	})
}

func (c *SkClient) UpdateRule(config *skproxy.Config) (*pb.DefaultResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), SK_CONN_TIMEOUT)
	defer cancel()
	data, err := json.Marshal(config)
	if err != nil {
		return &pb.DefaultResponse{Status: -1}, err
	}
	return c.client.UpdateRule(ctx, &pb.UpdateRulesRequest{
		Config: data,
	})
}
