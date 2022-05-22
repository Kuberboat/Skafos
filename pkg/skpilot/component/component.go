package component

import (
	"fmt"
	"sync"

	kubeCore "p9t.io/kuberboat/pkg/api/core"
	"p9t.io/skafos/pkg/api/core"
)

// SkComponents contains the snapshot of all discovered pods and services in Kuberboat, as well
// as all the rules applied to services.
type SkComponents struct {
	Mtx sync.Mutex

	// Stores the mapping from pod name to pod.
	Pods map[string]*kubeCore.Pod
	// Stores the mapping from service name to service.
	Services map[string]*kubeCore.Service
	// Stores the mapping from the name of a service to the pods it selects by label.
	ServicesToPods map[string]*[]string

	// Stores the mapping from the name of a ratio rule to the rule.
	RatioRules map[string]*core.RatioRule
	// Stores the mapping from the name of a regex rule to the rule.
	RegexRules map[string]*core.RegexRule
	// Stores the mapping from the name of a service to metadata of the rule applied to it.
	ServiceToRule map[string]*core.RuleMeta
}

func NewSkComponents() *SkComponents {
	return &SkComponents{
		Mtx:            sync.Mutex{},
		Pods:           map[string]*kubeCore.Pod{},
		Services:       map[string]*kubeCore.Service{},
		ServicesToPods: map[string]*[]string{},
		RatioRules:     map[string]*core.RatioRule{},
		RegexRules:     map[string]*core.RegexRule{},
		ServiceToRule:  map[string]*core.RuleMeta{},
	}
}

// GetServiceAndServicePods gets service and its pods according to service name. If the service does not
// exist, it will return an error. If the service exists, then its pods are gauranteed to exist.
func (sc *SkComponents) GetServiceAndServicePods(serviceName string) (*kubeCore.Service, []*kubeCore.Pod, error) {
	service, ok := sc.Services[serviceName]
	if !ok {
		return nil, nil, fmt.Errorf("no such service: %s", serviceName)
	}
	servicePodNames, ok := sc.ServicesToPods[serviceName]
	if !ok {
		panic(fmt.Sprintf("expect pods for service %s", serviceName))
	}
	servicePods := make([]*kubeCore.Pod, 0)
	for _, podName := range *servicePodNames {
		pod, ok := sc.Pods[podName]
		if !ok {
			panic(fmt.Sprintf("expect pod %s", podName))
		}
		servicePods = append(servicePods, pod)
	}
	return service, servicePods, nil
}

// CheckRule checks whether a rule could be applied to a service.
func (sc *SkComponents) CheckRule(ruleName string, serviceName string) error {
	if _, ok := sc.RatioRules[ruleName]; ok {
		return fmt.Errorf("duplicate rule: %s", ruleName)
	}
	if _, ok := sc.RegexRules[ruleName]; ok {
		return fmt.Errorf("duplicate rule: %s", ruleName)
	}
	if _, ok := sc.ServiceToRule[serviceName]; ok {
		return fmt.Errorf(
			"service %s already has a rule applied to it",
			serviceName,
		)
	}
	return nil
}
