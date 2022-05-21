package proxy

import (
	"sync"

	"p9t.io/skafos/pkg/skproxy"
)

// RuleGeneratorCache handles incremental changes of proxy rules, so that
// new proxies can quickly catch up with the others.
type RuleGeneratorCache struct {
	mtx                 sync.RWMutex
	ratioRuleGenerators map[string]*skproxy.RatioRuleGenerator
	regexRuleGenerators map[string]*skproxy.RegexRuleGenerator
}

func NewRuleGeneratorCache() *RuleGeneratorCache {
	return &RuleGeneratorCache{
		ratioRuleGenerators: map[string]*skproxy.RatioRuleGenerator{},
		regexRuleGenerators: map[string]*skproxy.RegexRuleGenerator{},
	}
}

func (c *RuleGeneratorCache) SetRatioRule(name string, generator *skproxy.RatioRuleGenerator) {
	if generator != nil {
		c.mtx.Lock()
		defer c.mtx.Unlock()
		c.ratioRuleGenerators[name] = generator
	}
}

func (c *RuleGeneratorCache) SetRegexRule(name string, generator *skproxy.RegexRuleGenerator) {
	if generator != nil {
		c.mtx.Lock()
		defer c.mtx.Unlock()
		c.regexRuleGenerators[name] = generator
	}
}

func (c *RuleGeneratorCache) DeleteRule(name string) {
	c.mtx.Lock()
	defer c.mtx.Unlock()
	delete(c.ratioRuleGenerators, name)
	delete(c.regexRuleGenerators, name)
}

func (c *RuleGeneratorCache) HasRules() bool {
	c.mtx.RLock()
	defer c.mtx.RUnlock()
	return len(c.ratioRuleGenerators) != 0 || len(c.regexRuleGenerators) != 0
}

func (c *RuleGeneratorCache) DumpConfig() *skproxy.Config {
	c.mtx.RLock()
	defer c.mtx.RUnlock()
	return &skproxy.Config{
		RatioRules: c.ratioRuleGenerators,
		RegexRules: c.regexRuleGenerators,
	}
}

type ProxyContainer struct {
	IP          string
	ID          string
	SandboxName string
}

type ProxyManager struct {
	mtx       sync.RWMutex
	idToProxy map[string]*ProxyContainer
}

func NewProxyManager() *ProxyManager {
	return &ProxyManager{
		idToProxy: map[string]*ProxyContainer{},
	}
}

func (m *ProxyManager) GetProxies() []*ProxyContainer {
	m.mtx.RLock()
	defer m.mtx.RUnlock()
	ret := make([]*ProxyContainer, 0, len(m.idToProxy))
	for _, p := range m.idToProxy {
		ret = append(ret, p)
	}
	return ret
}

func (m *ProxyManager) SetProxy(id string, proxy *ProxyContainer) {
	m.mtx.Lock()
	defer m.mtx.Unlock()
	m.idToProxy[id] = proxy
}

func (m *ProxyManager) DeleteProxy(id string) {
	m.mtx.Lock()
	defer m.mtx.Unlock()
	delete(m.idToProxy, id)
}
