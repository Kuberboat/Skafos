package skproxy

import (
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"regexp"
	"sync"
)

// IMPORTANT: Any IP address or domain name must not contain http:// or ending slash.
// e.g. http://www.github.com/ must be www.github.com.
// Because the IP/domain name will be written in Host http header, and this header
// should not contain protocol name or trailing slash.

// =============================================================================
//
// Proxy rule:
//
// What skproxy uses to determine how to forward a request.
//
// =============================================================================

// ProxyRule represents the proxy rule for a service, i.e. one service should have at most one rule.
//
// When handling a request, the caller should first use CanProxyRequest to determine a ProxyRule
// can be applied to a request. If so, call GetProxiedAddress to get the new address.
type ProxyRule interface {
	// CanProxyRequest returns true if this proxy rule can forward host:port to some other address.
	CanProxyRequest(host string, port uint16) bool
	// GetProxiedAddress returns the proxied address. We still pass in host and port, because
	// they cannot be easily extracted from the original request object. The caller must compute host and port
	// and ensure they are valid.
	GetProxiedAddress(req *http.Request, host string, port uint16) (string, error)
}

// ruleBase is the part of a ProxyRule that corresponds to service info.
type ruleBase struct {
	serviceIP   string
	portMapping map[uint16]uint16
}

func newRuleBase(serviceIP string, portMapping map[uint16]uint16) *ruleBase {
	return &ruleBase{
		serviceIP:   serviceIP,
		portMapping: portMapping,
	}
}

// CanProxyRequest determines whether a rule can be applied to a request.
func (r *ruleBase) CanProxyRequest(host string, port uint16) bool {
	_, ok := r.portMapping[port]
	return host == r.serviceIP && ok
}

type ratioRule struct {
	// base is used to determine if host:port can be proxied.
	base *ruleBase
	// ratio is the proportion of requests that
	// will be redirected to IPs in proxiedIPs
	ratio int
	// proxiedIPs is the set of IPs from which the new IP will be selected
	// for the ratio of requests.
	proxiedIPs *ipRRSelector
	// otherIPs is the set of IPs from which the new IP will be selected
	// for (1 - ratio%) of requests.
	otherIPs *ipRRSelector
}

func (r *ratioRule) CanProxyRequest(host string, port uint16) bool {
	return r.base.CanProxyRequest(host, port)
}

func (r *ratioRule) GetProxiedAddress(req *http.Request, host string, port uint16) (string, error) {
	if !r.CanProxyRequest(host, port) {
		return "", errors.New("%v:%p cannot be proxied by this rule")
	}

	rand := rand.Intn(100)
	if rand < r.ratio {
		return r.proxiedIPs.NextIP()
	} else {
		return r.otherIPs.NextIP()
	}
}

type regexRule struct {
	// base is used to determine if host:port can be proxied.
	base *ruleBase
	// matchers are the regex matchers of this rule.
	//
	// When applying a regexRule to a request,
	// all the headers will be matched against every matcher,
	// and the first match will be chosen.
	matchers []*headerRegexMatcher
	// otherIPs is the set of IPs from which the new IP will be selected
	// if none of the matchers are matched.
	otherIPs *ipRRSelector
}

func (r *regexRule) CanProxyRequest(host string, port uint16) bool {
	return r.base.CanProxyRequest(host, port)
}

func (r *regexRule) GetProxiedAddress(req *http.Request, host string, port uint16) (string, error) {
	if !r.CanProxyRequest(host, port) {
		return "", errors.New("%v:%p cannot be proxied by this rule")
	}

	for k, vs := range req.Header {
		for _, v := range vs {
			for _, matcher := range r.matchers {
				if ip, ok := matcher.MatchAndGetIP(k, v); ok {
					return ip, nil
				}
			}
		}
	}

	return r.otherIPs.NextIP()
}

// =============================================================================
//
// Proxy rule generator:
//
// The interface exposed to other components for easy serialization.
//
// =============================================================================

// Config configures proxy rules. Each time a Config is applied,
// the original config will be completely overwritten.
type Config struct {
	RatioRules map[string]*RatioRuleGenerator
	RegexRules map[string]*RegexRuleGenerator
}

// ProxyRuleGenerator can be used to generate ProxyRule, which includes members that cannot be
// serialized to/from json.
type ProxyRuleGenerator interface {
	GenerateRule() (ProxyRule, error)
}

// RatioRuleGenerator is the exported generator of ratio rule.
// This struct should be used in configuration for easy serialization.
type RatioRuleGenerator struct {
	// ServiceIP is the service IP this rule applies to.
	ServiceIP string
	// PortMapping is the port mapping of that service.
	PortMapping map[uint16]uint16
	// Ratio is the proportion of requests that
	// will be redirected to IPs in ProxiedIPs.
	Ratio int
	// ProxiedIPs is the set of IPs from which the new IP will be selected
	// for the Ratio of requests.
	// The pod IPs of the service should either fall in ProxiedIPs or OtherIPs.
	ProxiedIPs []string
	// OtherIPs is the set of IPs from which the new IP will be selected
	// for (1 - Ratio%) of requests.
	OtherIPs []string
}

func (g *RatioRuleGenerator) GenerateRule() (ProxyRule, error) {
	return &ratioRule{
		base:       newRuleBase(g.ServiceIP, g.PortMapping),
		ratio:      g.Ratio,
		proxiedIPs: newIPRRSelector(g.ProxiedIPs),
		otherIPs:   newIPRRSelector(g.OtherIPs),
	}, nil
}

// RegexRuleGenerator is the exported generator of regex rule.
// This struct should be used in configuration for easy serialization.
type RegexRuleGenerator struct {
	// ServiceIP is the service IP this rule applies to.
	ServiceIP string
	// PortMapping is the port mapping of that service.
	PortMapping map[uint16]uint16
	// Matchers are the regex matchers of this rule.
	//
	// When applying a regex rule to a request,
	// all the headers will be matched against every matcher,
	// and the first match will be chosen.
	Matchers []*HeaderRegexMatcher
	// OtherIPs is the set of IPs from which the new IP will be selected
	// if none of the matchers are matched.
	OtherIPs []string
}

func (g *RegexRuleGenerator) GenerateRule() (ProxyRule, error) {
	actualMatchers := make([]*headerRegexMatcher, 0, len(g.Matchers))
	for _, m := range g.Matchers {
		matcher, err := newHeaderRegexMatcher(m)
		if err != nil {
			return nil, err
		}
		actualMatchers = append(actualMatchers, matcher)
	}
	return &regexRule{
		base:     newRuleBase(g.ServiceIP, g.PortMapping),
		matchers: actualMatchers,
		otherIPs: newIPRRSelector(g.OtherIPs),
	}, nil
}

// =============================================================================
//
// Proxy rule manager:
//
// What skproxy uses to manage multiple rules.
//
// =============================================================================

// ProxyManager manages all the proxy rules.
type ProxyRuleManager struct {
	// mtx ensures safe concurrent access.
	mtx sync.RWMutex
	// rules maps rule name to the rule.
	rules map[string]ProxyRule
}

// SetRule sets a rule with given name.
func (m *ProxyRuleManager) SetRule(name string, generator ProxyRuleGenerator) error {
	m.mtx.Lock()
	defer m.mtx.Unlock()
	if rule, err := generator.GenerateRule(); err != nil {
		return err
	} else {
		m.rules[name] = rule
		return nil
	}
}

// ClearRules removes all rules.
func (m *ProxyRuleManager) ClearRules() {
	m.mtx.RLock()
	defer m.mtx.RUnlock()
	for k := range m.rules {
		delete(m.rules, k)
	}
}

// GetProxiedAddress tries to match the request against every rule. If no rule can be matched,
// just return host:port.
func (m *ProxyRuleManager) GetProxiedAddress(req *http.Request, host string, port uint16) string {
	m.mtx.RLock()
	defer m.mtx.RUnlock()
	for _, rule := range m.rules {
		if rule.CanProxyRequest(host, port) {
			addr, err := rule.GetProxiedAddress(req, host, port)
			if err == nil {
				return addr
			}
		}
	}
	return fmt.Sprintf("%v:%v", host, port)
}

func NewProxyRuleManager() *ProxyRuleManager {
	return &ProxyRuleManager{
		rules: map[string]ProxyRule{},
	}
}

// =============================================================================
//
// Utilities
//
// =============================================================================

// ipRRSelector selects IP addresses in a round robin fashion.
// IPs cannot be modified after construction.
type ipRRSelector struct {
	ips []string
	idx int
}

func newIPRRSelector(ips []string) *ipRRSelector {
	return &ipRRSelector{
		ips: ips,
		idx: 0,
	}
}

// NextIP selects the next IP from the selector.
func (s *ipRRSelector) NextIP() (string, error) {
	if len(s.ips) == 0 {
		return "", errors.New("no IP to select")
	}
	ret := s.ips[s.idx]
	if s.idx == len(s.ips)-1 {
		s.idx = 0
	} else {
		s.idx++
	}
	return ret, nil
}

// headerRegexMatcher tells whether an HTTP header of a request matches a regex.
// If so, it also tells which IP this request should be redirected to.
type headerRegexMatcher struct {
	header string
	regex  *regexp.Regexp
	ips    *ipRRSelector
}

// MatchAndGetIP is the matching logic for headerRegexMatcher.
// If there is a match, the second return value will be true, and the first return value will be the IP
// this request should be redirected to.
func (m *headerRegexMatcher) MatchAndGetIP(k string, v string) (string, bool) {
	if k == m.header && m.regex.MatchString(v) {
		ip, err := m.ips.NextIP()
		if err != nil {
			return "", false
		}
		return ip, true
	}
	return "", false
}

// HeaderRegexMatcher is the exported version of headerRegexMatcher
// for easy serialization.
type HeaderRegexMatcher struct {
	// Header is the name of the HTTP header.
	Header string
	// Regex is the regexp against which the value of the header will be matched.
	Regex string
	// IPs is the set of IPs from which the new IP will be chosen
	// if there is a match.
	IPs []string
}

func newHeaderRegexMatcher(m *HeaderRegexMatcher) (*headerRegexMatcher, error) {
	compiledRegex, err := regexp.Compile(m.Regex)
	if err != nil {
		return nil, err
	}
	return &headerRegexMatcher{
		header: m.Header,
		regex:  compiledRegex,
		ips:    newIPRRSelector(m.IPs),
	}, nil
}
