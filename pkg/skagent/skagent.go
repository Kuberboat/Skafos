package skagent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"time"

	"github.com/docker/distribution/context"

	dockertypes "github.com/docker/docker/api/types"
	dockercontainer "github.com/docker/docker/api/types/container"
	dockerclient "github.com/docker/docker/client"
	"github.com/golang/glog"
	"p9t.io/skafos/pkg/skagent/proxy"
	"p9t.io/skafos/pkg/skproxy"
)

const (
	proxyImageName     = "gun9nir/skproxy"
	proxyUserId        = "1234"
	iptablesScriptPath = "/usr/local/bin/skafos-iptables.sh"
	gcInterval         = 5
)

type Agent interface {
	// ConfigureProxy changes the iptables of the network namespace of the container
	// identified by sandboxName.
	SetupProxy(sandboxName string, ip string) error
	// ApplyProxyConfig updates proxy rules sends the lastest rules to all proxies on this node.
	ApplyProxyConfig(config *skproxy.Config) error
}

type agent struct {
	// Docker client to access docker apis.
	dockerClient *dockerclient.Client
	// ruleCache handles incremental changes of proxy rules from skpilot.
	ruleCache *proxy.RuleGeneratorCache
	// proxyManager manages all proxies.
	proxyManager *proxy.ProxyManager
}

func NewAgent() Agent {
	// Create docker client.
	cli, err := dockerclient.NewClientWithOpts(dockerclient.FromEnv, dockerclient.WithAPIVersionNegotiation())
	if err != nil {
		glog.Fatal(err)
	}

	agent := &agent{
		dockerClient: cli,
		ruleCache:    proxy.NewRuleGeneratorCache(),
		proxyManager: proxy.NewProxyManager(),
	}

	go func() {
		for range time.Tick(time.Second * gcInterval) {
			agent.cleanDeadProxy()
		}
	}()

	return agent
}

func (a *agent) SetupProxy(sandboxName string, ip string) error {
	cli := a.dockerClient

	// Create proxy container.
	resp, err := cli.ContainerCreate(context.Background(), &dockercontainer.Config{
		Image: proxyImageName,
		User:  proxyUserId,
	}, &dockercontainer.HostConfig{
		NetworkMode: dockercontainer.NetworkMode(fmt.Sprintf("container:%v", sandboxName)),
	}, nil, nil, getProxyContainerName(sandboxName))
	if err != nil {
		return err
	}

	// Start proxy container.
	if err := cli.ContainerStart(
		context.Background(),
		resp.ID,
		dockertypes.ContainerStartOptions{}); err != nil {
		return err
	}

	// Configure iptables.
	cmd := exec.Command(iptablesScriptPath, sandboxName)
	if err := cmd.Run(); err != nil {
		return err
	}

	// Update metadata.
	a.proxyManager.SetProxy(resp.ID, &proxy.ProxyContainer{
		IP:          ip,
		ID:          resp.ID,
		SandboxName: sandboxName,
	})

	// If there are rules currently, sync the rule to that proxy.
	if a.ruleCache.HasRules() {
		err := a.applyConfigToOneProxy(ip, skproxy.ConfigPort, a.ruleCache.DumpConfig())
		if err != nil {
			glog.Errorf("failed to apply proxy rule to skproxy at %v: %v", ip, err.Error())
		}
	}

	return nil
}

func (a *agent) ApplyProxyConfig(config *skproxy.Config) error {
	// Update rules incrementally.
	for name, rule := range config.RatioRules {
		if rule == nil {
			a.ruleCache.DeleteRule(name)
		} else {
			a.ruleCache.SetRatioRule(name, rule)
		}
	}
	for name, rule := range config.RegexRules {
		if rule == nil {
			a.ruleCache.DeleteRule(name)
		} else {
			a.ruleCache.SetRegexRule(name, rule)
		}
	}

	newConfig := a.ruleCache.DumpConfig()
	var ret error = nil
	for _, proxy := range a.proxyManager.GetProxies() {
		err := a.applyConfigToOneProxy(proxy.IP, skproxy.ConfigPort, newConfig)
		if err != nil {
			ret = err
			glog.Errorf("failed to apply proxy rule to skproxy %v: %v", proxy.ID, err.Error())
		}
	}
	return ret
}

func (a *agent) applyConfigToOneProxy(ip string, port uint16, config *skproxy.Config) error {
	configJson, err := json.Marshal(config)
	if err != nil {
		return err
	}
	_, err = http.Post(
		fmt.Sprintf("http://%v:%v", ip, port),
		"application/json", bytes.NewBuffer(configJson))
	if err != nil {
		return err
	}
	return nil
}

func getProxyContainerName(sandboxName string) string {
	return fmt.Sprintf("skproxy_%v", sandboxName)
}

func (a *agent) cleanDeadProxy() {
	cli := a.dockerClient
	proxies := a.proxyManager.GetProxies()
	for _, p := range proxies {
		_, err := cli.ContainerInspect(context.Background(), p.SandboxName)
		// Delete proxy when the pod is removed.
		if err != nil {
			a.proxyManager.DeleteProxy(p.ID)
			glog.Infof("clean up proxy for %v", p.SandboxName)
			err := cli.ContainerStop(context.Background(), p.ID, nil)
			if err != nil {
				glog.Errorf("failed to stop proxy: %v", err.Error())
				continue
			}
			err = cli.ContainerRemove(context.Background(), p.ID, dockertypes.ContainerRemoveOptions{})
			if err != nil {
				glog.Errorf("failed to remove proxy: %v", err.Error())
			}
		}
	}
}
