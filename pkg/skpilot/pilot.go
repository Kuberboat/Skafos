package skpilot

import (
	"time"

	"p9t.io/skafos/pkg/api/core"
	"p9t.io/skafos/pkg/skpilot/agent"
	"p9t.io/skafos/pkg/skpilot/buffer"
	"p9t.io/skafos/pkg/skpilot/component"
	"p9t.io/skafos/pkg/skpilot/discover"
	"p9t.io/skafos/pkg/skpilot/message"
	"p9t.io/skafos/pkg/skpilot/util"
)

const (
	discoverInterval = time.Second * 10
	probeInterval    = time.Second * 8
)

// SkPilot handles user's requests of applying ratio rules and regex rules. It also
// starts the discoverer, which discovers all the pods and services from Kuberboat, and
// the messager, which informs SkAgent of the rule changes and proxy updates.
type SkPilot interface {
	// ApplyRatioRule handles user's requests of applying a ratio rule. It will write
	// the rule to the buffer if it is valid.
	ApplyRatioRule(rule *core.RatioRule) error
	// ApplyRegexRule handles user's requests of applying a regex rule. It will write
	// the rule to the buffer if it is valid.
	ApplyRegexRule(rule *core.RegexRule) error
}

func NewSkPilot(
	kubeEndpoint string,
	components *component.SkComponents,
	ruleBuffer *buffer.RuleBuffer,
	proxyBuffer *buffer.ProxyBuffer,
	agentManager agent.AgentManager,
) SkPilot {

	// Start discoverer
	discoverer := discover.NewDiscoverer(
		kubeEndpoint,
		components,
		ruleBuffer,
		proxyBuffer,
	)
	go discoverer.DoDiscovering(discoverInterval)

	// Start messager
	messager := message.NewMessager(ruleBuffer, proxyBuffer, agentManager)
	go messager.DoProbingAndMessaging(probeInterval)

	return &skPilotInner{
		components: components,
		ruleBuffer: ruleBuffer,
	}
}

type skPilotInner struct {
	// components stores metadata of all pods, services and rules.
	components *component.SkComponents
	// ruleBuffer is the buffer where rules to update are stored.
	ruleBuffer *buffer.RuleBuffer
}

func (sp *skPilotInner) ApplyRatioRule(rule *core.RatioRule) error {
	sp.components.Mtx.Lock()
	defer sp.components.Mtx.Unlock()

	if err := sp.components.CheckRule(rule.Name, rule.Spec.ServiceName); err != nil {
		return err
	}

	service, servicePods, err := sp.components.GetServiceAndServicePods(rule.Spec.ServiceName)
	if err != nil {
		return err
	}

	// Add the rule
	ruleGenerator := util.GenerateRatioRule(rule, service, servicePods)
	{
		sp.ruleBuffer.LockBuffer()
		sp.ruleBuffer.SetRatioRule(rule.Name, ruleGenerator)
		sp.ruleBuffer.UnlockBuffer()
	}

	// Update metadata
	sp.components.RatioRules[rule.Name] = rule
	sp.components.ServiceToRule[rule.Spec.ServiceName] = &rule.RuleMeta

	return nil
}

func (sp *skPilotInner) ApplyRegexRule(rule *core.RegexRule) error {
	sp.components.Mtx.Lock()
	defer sp.components.Mtx.Unlock()

	if err := sp.components.CheckRule(rule.Name, rule.Spec.ServiceName); err != nil {
		return err
	}

	service, servicePods, err := sp.components.GetServiceAndServicePods(rule.Spec.ServiceName)
	if err != nil {
		return err
	}

	// Add the rule
	ruleGenerator := util.GenerateRegexRule(rule, service, servicePods)
	{
		sp.ruleBuffer.LockBuffer()
		sp.ruleBuffer.SetRegexRule(rule.Name, ruleGenerator)
		sp.ruleBuffer.UnlockBuffer()
	}

	// Update metadata
	sp.components.RegexRules[rule.Name] = rule
	sp.components.ServiceToRule[rule.Spec.ServiceName] = &rule.RuleMeta

	return nil
}
