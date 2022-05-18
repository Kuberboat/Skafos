package app

import (
	"context"
	"encoding/json"
	"fmt"
	"net"

	"github.com/golang/glog"
	"google.golang.org/grpc"
	"gopkg.in/yaml.v2"
	kuberboatCore "p9t.io/kuberboat/pkg/api/core"
	"p9t.io/skafos/pkg/api/core"
	pb "p9t.io/skafos/pkg/proto"
	"p9t.io/skafos/pkg/skpilot/client"
)

type server struct {
	pb.UnimplementedSkpilotCtlServiceServer
	pb.UnimplementedSkpilotSkagentServiceServer
}

func (s *server) ApplyRule(ctx context.Context, req *pb.ApplyRuleRequest) (*pb.DefaultResponse, error) {
	var ruleKind core.RuleKind
	err := yaml.Unmarshal(req.Rule, &ruleKind)
	if err != nil {
		return &pb.DefaultResponse{Status: -1}, err
	}
	switch ruleKind.Kind {
	case string(core.RatioType):
		var ratioRule core.Ratio
		err := yaml.Unmarshal(req.Rule, &ratioRule)
		if err != nil {
			glog.Errorf("unmarshal yaml failed: %v", err)
			return &pb.DefaultResponse{Status: -1}, err
		}
		// TODO: add ratio rule
	case string(core.RegexType):
		var reRule core.Regex
		err := yaml.Unmarshal(req.Rule, &reRule)
		if err != nil {
			glog.Errorf("unmarshal yaml failed: %v", err)
			return &pb.DefaultResponse{Status: -1}, err
		}
		// TODO: add regex rule
	}
	return &pb.DefaultResponse{Status: 1}, nil
}

func (s *server) RegisterSelf(ctx context.Context, req *pb.RegisterSelfRequest) (*pb.DefaultResponse, error) {
	var node kuberboatCore.Node
	json.Unmarshal(req.Node, &node)
	_, err := client.NewCtlClient(node.Status.Address, node.Status.Port)
	if err != nil {
		glog.Errorf("fail to create client with skagent: %v", err)
		return &pb.DefaultResponse{Status: -1}, err
	}
	return &pb.DefaultResponse{Status: 1}, nil
}

func StartServer() {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", core.SKPILOT_PORT))
	if err != nil {
		glog.Fatal("Api server failed to connect!")
	}

	skPilot := grpc.NewServer()
	pb.RegisterSkpilotCtlServiceServer(skPilot, &server{})
	pb.RegisterSkpilotSkagentServiceServer(skPilot, &server{})

	glog.Infof("skpilot listening at %v", lis.Addr())

	if err := skPilot.Serve(lis); err != nil {
		glog.Fatal(err)
	}
}
