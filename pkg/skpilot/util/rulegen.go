package util

import (
	"reflect"

	kubeCore "p9t.io/kuberboat/pkg/api/core"
	"p9t.io/skafos/pkg/api/core"
	"p9t.io/skafos/pkg/skproxy"
)

// GenerateRatioRule generates a ratio rule that could be recognized by SkAgent and SkProxy
// based on the rule info, the service it applied to and pods of the service.
func GenerateRatioRule(
	rule *core.RatioRule,
	service *kubeCore.Service,
	pods []*kubeCore.Pod,
) *skproxy.RatioRuleGenerator {

	portMapping := make(map[uint16]uint16)
	for _, portPair := range service.Spec.Ports {
		portMapping[portPair.Port] = portMapping[portPair.TargetPort]
	}

	proxiedIPs := make([]string, 0)
	otherIPs := make([]string, 0)
	for _, pod := range pods {
		if reflect.DeepEqual(rule.Spec.Selector, pod.Labels) {
			proxiedIPs = append(proxiedIPs, pod.Status.PodIP)
		} else {
			otherIPs = append(otherIPs, pod.Status.PodIP)
		}
	}

	return &skproxy.RatioRuleGenerator{
		ServiceIP:   service.Spec.ClusterIP,
		PortMapping: portMapping,
		Ratio:       int(rule.Spec.Ratio),
		ProxiedIPs:  proxiedIPs,
		OtherIPs:    otherIPs,
	}
}

// GenerateRegexRule generates a regex rule that could be recognized by SkAgent and SkProxy
// based on the rule info, the service it applied to and pods of the service.
func GenerateRegexRule(
	rule *core.RegexRule,
	service *kubeCore.Service,
	pods []*kubeCore.Pod,
) *skproxy.RegexRuleGenerator {

	portMapping := make(map[uint16]uint16)
	for _, portPair := range service.Spec.Ports {
		portMapping[portPair.Port] = portMapping[portPair.TargetPort]
	}

	matchers := make([]*skproxy.HeaderRegexMatcher, 0, len(rule.Spec.Matchers))
	for _, matcher := range rule.Spec.Matchers {
		matchers = append(matchers, &skproxy.HeaderRegexMatcher{
			Header: matcher.Header,
			Regex:  matcher.Regex,
			IPs:    []string{},
		})
	}
	otherIPs := make([]string, 0)

	for _, pod := range pods {
		matched := false
		for i, matcher := range rule.Spec.Matchers {
			if reflect.DeepEqual(matcher.Selector, pod.Labels) {
				matchers[i].IPs = append(matchers[i].IPs, pod.Status.PodIP)
				matched = true
				break
			}
		}
		if !matched {
			otherIPs = append(otherIPs, pod.Status.PodIP)
		}
	}

	return &skproxy.RegexRuleGenerator{
		ServiceIP:   service.Spec.ClusterIP,
		PortMapping: portMapping,
		Matchers:    matchers,
		OtherIPs:    otherIPs,
	}
}
