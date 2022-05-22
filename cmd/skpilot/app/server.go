package app

import (
	"context"
	"encoding/json"
	"fmt"
	"net"

	"github.com/golang/glog"
	"google.golang.org/grpc"
	kuberboatCore "p9t.io/kuberboat/pkg/api/core"
	"p9t.io/skafos/pkg/api/core"
	pb "p9t.io/skafos/pkg/proto"
	"p9t.io/skafos/pkg/skpilot"
	"p9t.io/skafos/pkg/skpilot/agent"
	"p9t.io/skafos/pkg/skpilot/buffer"
	"p9t.io/skafos/pkg/skpilot/component"
)

var kubeEndpoint = "localhost"

var skPilot skpilot.SkPilot
var agentManager agent.AgentManager

type server struct {
	pb.UnimplementedSkpilotCtlServiceServer
	pb.UnimplementedSkpilotSkagentServiceServer
}

func (s *server) ApplyRatioRule(
	ctx context.Context,
	req *pb.ApplyRatioRuleRequest,
) (*pb.DefaultResponse, error) {
	var rule core.RatioRule
	if err := json.Unmarshal(req.RatioRule, &rule); err != nil {
		glog.Errorf("unmarshal rule failed: %v", err)
		return &pb.DefaultResponse{Status: -1}, err
	}
	if err := skPilot.ApplyRatioRule(&rule); err != nil {
		return &pb.DefaultResponse{Status: -1}, err
	}
	return &pb.DefaultResponse{Status: 0}, nil
}

func (s *server) ApplyRegexRule(
	ctx context.Context,
	req *pb.ApplyRegexRuleRequest,
) (*pb.DefaultResponse, error) {
	var rule core.RegexRule
	if err := json.Unmarshal(req.RegexRule, &rule); err != nil {
		glog.Errorf("unmarshal rule failed: %v", err)
		return &pb.DefaultResponse{Status: -1}, err
	}
	if err := skPilot.ApplyRegexRule(&rule); err != nil {
		return &pb.DefaultResponse{Status: -1}, err
	}
	return &pb.DefaultResponse{Status: 0}, nil
}

func (s *server) RegisterSelf(
	ctx context.Context,
	req *pb.RegisterSelfRequest,
) (*pb.DefaultResponse, error) {
	var node kuberboatCore.Node
	json.Unmarshal(req.Node, &node)
	err := agentManager.AddAgent(node.Status.Address, node.Status.Port)
	if err != nil {
		glog.Errorf("fail to create client with skagent: %v", err)
		return &pb.DefaultResponse{Status: -1}, err
	}
	return &pb.DefaultResponse{Status: 0}, nil
}

func StartServer() {
	components := component.NewSkComponents()
	ruleBuffer := buffer.NewRuleBuffer()
	proxyBuffer := buffer.NewProxyBuffer()
	agentManager = agent.NewAgentManager(ruleBuffer, proxyBuffer)
	skPilot = skpilot.NewSkPilot(
		kubeEndpoint,
		components,
		ruleBuffer,
		proxyBuffer,
		agentManager,
	)

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
