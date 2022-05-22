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
	kubePb "p9t.io/kuberboat/pkg/proto"
)

var KUBE_CONN_TIMEOUT time.Duration = time.Second

type KubeClient struct {
	connection *grpc.ClientConn
	client     kubePb.ApiServerCtlServiceClient
}

func NewKubeClient(url string, KubePort uint16) (*KubeClient, error) {
	addr := fmt.Sprintf("%v:%v", url, KubePort)
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, errors.New("client failed to connect to worker node")
	}
	return &KubeClient{
		connection: conn,
		client:     kubePb.NewApiServerCtlServiceClient(conn),
	}, nil
}

func (c *KubeClient) GetAllPods() ([]*core.Pod, error) {
	ctx, cancel := context.WithTimeout(context.Background(), KUBE_CONN_TIMEOUT)
	defer cancel()

	resp, err := c.client.DescribePods(ctx, &kubePb.DescribePodsRequest{
		All:      true,
		PodNames: []string{},
	})
	if err != nil {
		return []*core.Pod{}, err
	}

	var pods []*core.Pod
	err = json.Unmarshal(resp.Pods, &pods)
	if err != nil {
		return []*core.Pod{}, err
	}
	return pods, nil
}

func (c *KubeClient) GetAllServices() ([]*core.Service, [][]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), KUBE_CONN_TIMEOUT)
	defer cancel()

	resp, err := c.client.DescribeServices(ctx, &kubePb.DescribeServicesRequest{
		All:          true,
		ServiceNames: []string{},
	})
	if err != nil {
		return []*core.Service{}, [][]string{}, err
	}

	var services []*core.Service
	var servicePods [][]string
	err = json.Unmarshal(resp.Services, &services)
	if err != nil {
		return []*core.Service{}, [][]string{}, err
	}
	err = json.Unmarshal(resp.ServicePodNames, &servicePods)
	if err != nil {
		return []*core.Service{}, [][]string{}, err
	}
	return services, servicePods, nil
}
