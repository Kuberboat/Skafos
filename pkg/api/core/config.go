package core

// Port reference: https://istio.io/latest/docs/ops/deployment/requirements/#ports-used-by-istio
const SKPILOT_PORT = 15017
const SKAGENT_PORT = 15000

type RuleKind struct {
	Kind string
}
