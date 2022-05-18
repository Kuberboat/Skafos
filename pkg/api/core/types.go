package core

// Config configures proxy rules.
//
// Each time a Config is applied,
// the original config will be completely overwritten.
//
// Maps names to rules.
type Config struct {
	RatioRules map[string]*RatioRuleGenerator
	ProxyRules map[string]*RegexRuleGenerator
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

// Kind specified the category of an rule object.
type Kind string

// These are valid kinds of an rule object.
const (
	// RatioType means it's a ratio rule.
	RatioType Kind = "ratio"
	// RegexType means it's a regular expression matching rule.
	RegexType Kind = "re"
)

// ObjectMeta is metadata that ratio rule must have.
type ObjectMeta struct {
	// The name of an ratio rule.
	// Must not be empty.
	Name string
}

type RatioSpec struct {
	// Ratio is the ratio of requests forwarding to selected pods.
	Ratio uint32
	// ServiceName is the name of the service this rule applies to.
	ServiceName string
	// Selector selects the pods whose labels match with the selector.
	Selector map[string]string
}

type Ratio struct {
	// The type of a ratio rule is ratio
	Kind
	// Standard object's metadata.
	ObjectMeta `yaml:"metadata"`
	// Specifictions of desired ratio routing rules.
	Spec RatioSpec
}

type Matcher struct {
	// Header is the name of the HTTP header.
	Header string
	// Regex is the regular expression of this matcher.
	Regex string
	// Selector selects the pods whose labels match with the selector.
	Selector map[string]string
}

type RegexSpec struct {
	// ServiceName is the name of the service this rule applies to.
	ServiceName string
	// Matchers is the collection of all the matcher this rule has.
	Matchers []Matcher
}

type Regex struct {
	// The type of a regex rule is regex.
	Kind
	// Standard object's metadata.
	ObjectMeta `yaml:"metadata"`
	// Specifictions of desired regex routing rules.
	Spec RegexSpec
}
