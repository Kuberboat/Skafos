package app

import (
	"context"
	"encoding/json"
	"fmt"
	"net"

	"github.com/golang/glog"
	"google.golang.org/grpc"
	"p9t.io/kuberboat/pkg/api/core"
	pb "p9t.io/skafos/pkg/proto"
	"p9t.io/skafos/pkg/skagent"
	"p9t.io/skafos/pkg/skagent/client"
	"p9t.io/skafos/pkg/skproxy"
)

type server struct {
	pb.UnimplementedSkagentSkpilotServiceServer
}

var agent = skagent.NewAgent()

func (*server) CreateProxy(ctx context.Context, req *pb.CreateProxyRequest) (*pb.DefaultResponse, error) {
	var retErr error = nil
	var retStatus int32 = 0
	for idx, sandboxName := range req.ContainerNames {
		ip := req.SandboxIps[idx]
		err := agent.SetupProxy(sandboxName, ip)
		if err != nil {
			retErr = err
			retStatus = -1
			glog.Errorf("failed to create proxy for container %v: %v", sandboxName, err.Error())
		}
	}
	return &pb.DefaultResponse{
		Status: retStatus,
	}, retErr
}

func (*server) UpdateRule(ctx context.Context, req *pb.UpdateRulesRequest) (*pb.DefaultResponse, error) {
	var config skproxy.Config
	if err := json.Unmarshal(req.Config, &config); err != nil {
		return &pb.DefaultResponse{
			Status: -1,
		}, err
	}
	err := agent.ApplyProxyConfig(&config)
	if err != nil {
		return &pb.DefaultResponse{
			Status: -1,
		}, err
	}
	return &pb.DefaultResponse{
		Status: 0,
	}, nil
}

func StartServer(ip string, port uint16, skPilotIP string, skPilotPort uint16) {
	grpcServer := grpc.NewServer()
	pb.RegisterSkagentSkpilotServiceServer(grpcServer, &server{})

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		glog.Fatal(err)
	}

	client, err := client.NewClient(skPilotIP, skPilotPort)
	if err != nil {
		glog.Fatal(err)
	}
	nodeSelf := core.Node{
		Status: core.NodeStatus{
			Address: ip,
			Port:    port,
		},
	}
	_, err = client.RegisterSelf(&nodeSelf)
	if err != nil {
		glog.Fatalf("failed to register skagent to skpilot: %v", err.Error())
	}

	glog.Infof("skagent listening at %v", lis.Addr())
	if err := grpcServer.Serve(lis); err != nil {
		glog.Fatal(err)
	}
}
