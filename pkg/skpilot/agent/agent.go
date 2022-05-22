package agent

import (
	"fmt"

	"github.com/golang/glog"
	"p9t.io/skafos/pkg/skpilot/buffer"
	"p9t.io/skafos/pkg/skpilot/client"
)

// AgentManager accepts the registration of an SkAgent and contains the information
// of all the SkAgents.
type AgentManager interface {
	// AddAgent adds an SkAgent into the cluster.
	AddAgent(address string, port uint16) error
	// ListAgent lists all the registered SkAgents.
	ListAllAgent() map[string]*client.SkClient
}

func NewAgentManager(
	ruleBuffer *buffer.RuleBuffer,
	proxyBuffer *buffer.ProxyBuffer,
) AgentManager {
	return &agentManagerInner{
		skClients: map[string]*client.SkClient{},
		buffers:   []buffer.SkBuffer{ruleBuffer, proxyBuffer},
	}
}

type agentManagerInner struct {
	skClients map[string]*client.SkClient
	buffers   []buffer.SkBuffer
}

func (am *agentManagerInner) AddAgent(address string, port uint16) error {
	client, err := client.NewSkClient(address, port)
	if err != nil {
		return fmt.Errorf("fail to create client with skagent: %v", err)
	}
	am.skClients[address] = client

	// Based on the assumption that all agents joins before the work starts, we just
	// initialize the buffer for the client without writing anything into it.
	for _, buf := range am.buffers {
		buf.LockBuffer()
		buf.ResetAgentBuffer(address)
		buf.UnlockBuffer()
	}

	glog.Infof("[AGENT MANAGER] agent %s registered", address)

	return nil
}

func (am *agentManagerInner) ListAllAgent() map[string]*client.SkClient {
	return am.skClients
}
