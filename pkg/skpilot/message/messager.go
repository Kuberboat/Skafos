package message

import (
	"sync"
	"time"

	"p9t.io/skafos/pkg/skpilot/agent"
	"p9t.io/skafos/pkg/skpilot/buffer"
)

// Messager informs SkAgent of the rule changes and proxy updates at set interval by reading the
// buffers and send the data in buffers to SkAgent.
type Messager struct {
	// buffers are all the buffers from which Messager retrieves data and sends to SkAgent.
	buffers []buffer.SkBuffer
	// agentManager contains information of all SkAgents.
	agentManager agent.AgentManager
}

func NewMessager(
	ruleBuffer *buffer.RuleBuffer,
	proxyBuffer *buffer.ProxyBuffer,
	agentManager agent.AgentManager,
) *Messager {
	return &Messager{
		buffers:      []buffer.SkBuffer{ruleBuffer, proxyBuffer},
		agentManager: agentManager,
	}
}

// DoProbingAndMessaging probes the buffers and messages SkAgent rule changes and proxy updates.
func (m *Messager) DoProbingAndMessaging(probeInterval time.Duration) {
	for range time.Tick(probeInterval) {
		for _, buf := range m.buffers {
			go m.probeAndMessage(buf)
		}
	}
}

// probeAndMessage checks whether there are new data in buffers. If so, send the data to SkAgent
// and empty the buffer on success.
func (m *Messager) probeAndMessage(buf buffer.SkBuffer) {
	clients := m.agentManager.ListAllAgent()
	{
		buf.LockBuffer()
		var wg sync.WaitGroup
		wg.Add(len(clients))
		for addr, cli := range clients {
			if !buf.IsEmpty(addr) {
				go buf.AcceptAgent(addr, cli, &wg)
			} else {
				wg.Done()
			}
		}
		wg.Wait()
		buf.UnlockBuffer()
	}
}
