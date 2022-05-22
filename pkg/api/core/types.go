package core

// Kind specified the category of an rule object.
type Kind string

// These are valid kinds of an rule object.
const (
	// RatioType means it's a ratio rule.
	RatioType Kind = "ratio"
	// RegexType means it's a regular expression matching rule.
	RegexType Kind = "regex"
)

// RuleMeta contains the metadata of a rule.
type RuleMeta struct {
	// The type of a ratio rule is ratio
	Kind
	// The name of the rule.
	Name string
}

// RatioSpec contains the specifications of a ratio rule.
type RatioSpec struct {
	// ServiceName is the name of the service this rule applies to.
	ServiceName string `yaml:"serviceName"`
	// Ratio is the ratio of requests forwarding to selected pods.
	Ratio uint32
	// Selector selects the pods whose labels match with the selector.
	Selector map[string]string
}

// RatioRule is a rule defining the network traffic of a service in a ratio pattern.
type RatioRule struct {
	// RuleMeta contains the type and the name of a ratio rule.
	RuleMeta `yaml:",inline"`
	// Specifications of desired ratio routing rules.
	Spec RatioSpec
}

// Matcher contains the regex matching requirements of a regex rule.
type Matcher struct {
	// Header is the name of the HTTP header.
	Header string
	// Regex is the regular expression of this matcher.
	Regex string
	// Selector selects the pods whose labels match with the selector.
	Selector map[string]string
}

// RegexSpec contains the specifications of a regex rule.
type RegexSpec struct {
	// ServiceName is the name of the service this rule applies to.
	ServiceName string `yaml:"serviceName"`
	// Matchers is the collection of all the matcher this rule has.
	Matchers []Matcher
}

// RegexRule is a rule defining the network traffic of a service in a regex pattern.
type RegexRule struct {
	// RuleMeta contains the type and the name of a regex rule.
	RuleMeta `yaml:",inline"`
	// Specifications of desired regex routing rules.
	Spec RegexSpec
}

// SandboxInfo contains the basic information of the sandbox container in a pod.
type SandboxInfo struct {
	// SandboxName is the name of the sandbox container in a pod.
	SandboxName string
	// SandboxIP is the IP of the sandbox container in a pod, i.e. pod IP.
	SandboxIP string
	// HostIP is the IP of host of the pod.
	HostIP string
}
