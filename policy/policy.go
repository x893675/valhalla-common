package policy

type Principal struct {
	IAM       []string `json:"IAM,omitempty"`
	Service   []string `json:"Service,omitempty"`
	Federated []string `json:"Federated,omitempty"`
}

type PolicyStatement struct {
	Version    string    `json:"version,omitempty"`
	Effect     string    `json:"effect,omitempty"`
	Resources  []string  `json:"resources,omitempty"`
	Actions    []string  `json:"actions,omitempty"`
	Principal  Principal `json:"principal,omitempty"`
	Conditions Condition `json:"conditions,omitempty"`
}

/*
Conditions: {
	IpAddress: {
		acs:SourceIp: ["203.0.113.2"]
	},
	Bool: {
		acs:MFAPresent: ["true"]
	},
}
*/

type Condition map[string]ConditionValue

type ConditionValue map[string][]string
