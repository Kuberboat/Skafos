package client

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"p9t.io/skafos/pkg/api/core"
	pb "p9t.io/skafos/pkg/proto"
)

var CONN_TIMEOUT time.Duration = time.Second
var SKPILOT_URL string = "localhost"
var SKPILOT_PORT uint16 = core.SKPILOT_PORT

type ctlClient struct {
	connection *grpc.ClientConn
	client     pb.SkpilotCtlServiceClient
}

func NewCtlClient() *ctlClient {
	addr := fmt.Sprintf("%v:%v", SKPILOT_URL, SKPILOT_PORT)
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal("skctl client failed to connect to skpilot")
	}
	return &ctlClient{
		connection: conn,
		client:     pb.NewSkpilotCtlServiceClient(conn),
	}
}

func (c *ctlClient) ApplyRatioRule(rule *core.RatioRule) (*pb.DefaultResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), CONN_TIMEOUT)
	defer cancel()
	data, err := json.Marshal(rule)
	if err != nil {
		return &pb.DefaultResponse{Status: 1}, err
	}
	return c.client.ApplyRatioRule(ctx, &pb.ApplyRatioRuleRequest{
		RatioRule: data,
	})
}

func (c *ctlClient) ApplyRegexRule(rule *core.RegexRule) (*pb.DefaultResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), CONN_TIMEOUT)
	defer cancel()
	data, err := json.Marshal(rule)
	if err != nil {
		return &pb.DefaultResponse{Status: 1}, err
	}
	return c.client.ApplyRegexRule(ctx, &pb.ApplyRegexRuleRequest{
		RegexRule: data,
	})
}
