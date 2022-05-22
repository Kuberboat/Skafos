package buffer

import (
	"sync"

	"github.com/golang/glog"
	"p9t.io/skafos/pkg/api/core"
	"p9t.io/skafos/pkg/skpilot/client"
)

// ProxyBuffer is the SkBuffer for proxy updates.
type ProxyBuffer struct {
	mtx          sync.Mutex
	sandboxInfos map[string][]core.SandboxInfo
}

func NewProxyBuffer() *ProxyBuffer {
	return &ProxyBuffer{
		mtx:          sync.Mutex{},
		sandboxInfos: map[string][]core.SandboxInfo{},
	}
}

func (pb *ProxyBuffer) LockBuffer() {
	pb.mtx.Lock()
}

func (pb *ProxyBuffer) UnlockBuffer() {
	pb.mtx.Unlock()
}

func (pb *ProxyBuffer) IsEmpty(agentAddr string) bool {
	return len(pb.sandboxInfos[agentAddr]) == 0
}

func (pb *ProxyBuffer) ResetAgentBuffer(agentAddr string) {
	pb.sandboxInfos[agentAddr] = make([]core.SandboxInfo, 0)
}

func (pb *ProxyBuffer) AcceptAgent(agentAddr string, cli *client.SkClient, wg *sync.WaitGroup) {
	defer wg.Done()
	infos := pb.sandboxInfos[agentAddr]
	_, err := cli.CreateProxy(infos)
	if err == nil {
		glog.Infof("[PROXY BUFFER] created proxies %v, now reset buffer for agent %s", infos, agentAddr)
		pb.ResetAgentBuffer(agentAddr)
	} else {
		glog.Errorf("[PROXY BUFFER] fail to inform agent %s: %v", agentAddr, err)
	}
}

func (pb *ProxyBuffer) BufferType() string {
	return "proxy"
}

func (pb *ProxyBuffer) SetSandboxInfo(info *core.SandboxInfo) {
	for addr := range pb.sandboxInfos {
		if core.IsSameHostAddr(addr, info.HostIP) {
			pb.sandboxInfos[addr] = append(pb.sandboxInfos[addr], *info)
			glog.Infof("[PROXY BUFFER] add proxy with ip %s for agent %s", info.SandboxIP, addr)
		}
	}
}
