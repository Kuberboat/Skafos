package buffer

import (
	"sync"

	"github.com/golang/glog"
	"p9t.io/skafos/pkg/skpilot/client"
	"p9t.io/skafos/pkg/skproxy"
)

// RuleBuffer is the SkBuffer for rule updates.
type RuleBuffer struct {
	mtx   sync.Mutex
	rules map[string]skproxy.Config
}

func NewRuleBuffer() *RuleBuffer {
	return &RuleBuffer{
		mtx:   sync.Mutex{},
		rules: map[string]skproxy.Config{},
	}
}

func (rb *RuleBuffer) LockBuffer() {
	rb.mtx.Lock()
}

func (rb *RuleBuffer) UnlockBuffer() {
	rb.mtx.Unlock()
}

func (rb *RuleBuffer) IsEmpty(agentAddr string) bool {
	return len(rb.rules[agentAddr].RatioRules) == 0 &&
		len(rb.rules[agentAddr].RegexRules) == 0
}

func (rb *RuleBuffer) ResetAgentBuffer(agentAddr string) {
	rb.rules[agentAddr] = skproxy.Config{
		RatioRules: map[string]*skproxy.RatioRuleGenerator{},
		RegexRules: map[string]*skproxy.RegexRuleGenerator{},
	}
}

func (rb *RuleBuffer) AcceptAgent(agentAddr string, cli *client.SkClient, wg *sync.WaitGroup) {
	defer wg.Done()
	config := rb.rules[agentAddr]
	_, err := cli.UpdateRule(&config)
	if err == nil {
		glog.Infof("[RULE BUFFER] updated rules %v, now reset buffer for agent %s", config, agentAddr)
		rb.ResetAgentBuffer(agentAddr)
	} else {
		glog.Errorf("[RULE BUFFER] fail to inform agent %s: %v", agentAddr, err)
	}
}

func (rb *RuleBuffer) BufferType() string {
	return "rule"
}

func (rb *RuleBuffer) SetRatioRule(ruleName string, ratioRule *skproxy.RatioRuleGenerator) {
	for _, config := range rb.rules {
		config.RatioRules[ruleName] = ratioRule
	}
	glog.Infof("[RULE BUFFER] add ratio rule %s: %v", ruleName, ratioRule)
}

func (rb *RuleBuffer) SetRegexRule(ruleName string, regexRule *skproxy.RegexRuleGenerator) {
	for _, config := range rb.rules {
		config.RegexRules[ruleName] = regexRule
	}
	glog.Infof("[RULE BUFFER] add regex rule %s: %v", ruleName, regexRule)
	if regexRule != nil {
		for _, matcher := range regexRule.Matchers {
			glog.Infof("%v\n", *matcher)
		}
	}
}
