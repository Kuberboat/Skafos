package client

import (
	"errors"
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	pb "p9t.io/skafos/pkg/proto"
)

var CONN_TIMEOUT time.Duration = time.Second

type SkpilotClient struct {
	connection    *grpc.ClientConn
	skAgentClient pb.SkagentSkpilotServiceClient
}

func NewCtlClient(url string, skAgentPort uint16) (*SkpilotClient, error) {
	addr := fmt.Sprintf("%v:%v", url, skAgentPort)
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, errors.New("apiserver client failed to connect to worker node")
	}
	return &SkpilotClient{
		connection:    conn,
		skAgentClient: pb.NewSkagentSkpilotServiceClient(conn),
	}, nil
}
