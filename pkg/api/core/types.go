package core

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
