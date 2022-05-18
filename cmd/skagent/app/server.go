package app

import (
	"fmt"
	"net"

	"github.com/golang/glog"
	"google.golang.org/grpc"
	"p9t.io/kuberboat/pkg/api/core"
	pb "p9t.io/skafos/pkg/proto"
	"p9t.io/skafos/pkg/skagent/client"
)

type server struct {
	pb.UnimplementedSkagentSkpilotServiceServer
}

func StartServer(address string, port uint16, skPilotAddress string, skPilotPort uint16) {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		glog.Fatal(err)
	}
	glog.Infof("skagent listening at %v", lis.Addr())

	grpcServer := grpc.NewServer()
	pb.RegisterSkagentSkpilotServiceServer(grpcServer, &server{})

	client, err := client.NewCtlClient(skPilotAddress, skPilotPort)
	if err != nil {
		glog.Fatal(err)
	}
	nodeSelf := core.Node{
		Status: core.NodeStatus{
			Address: address,
			Port:    port,
		},
	}
	_, err = client.RegisterSelf(&nodeSelf)
	if err != nil {
		glog.Fatalf("fail to register self to control plane: %v", err)
	}

	if err := grpcServer.Serve(lis); err != nil {
		glog.Fatal(err)
	}
}
