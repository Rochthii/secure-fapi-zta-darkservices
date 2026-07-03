package policy

import (
	"encoding/json"
	"fmt"
	"os"
)

type PolicyRule struct {
	Role     string   `json:"role"`
	Resource string   `json:"resource"`
	Actions  []string `json:"actions"`
}

type PolicyEngine struct {
	rules []PolicyRule
}

var globalEngine *PolicyEngine

// LoadPolicies initializes the global policy engine from policies.json
func LoadPolicies() error {
	paths := []string{
		"config/policies.json",
		"../gateway/config/policies.json",
		"../../gateway/config/policies.json",
		"../config/policies.json",
	}

	var data []byte
	var err error
	var foundPath string

	for _, p := range paths {
		data, err = os.ReadFile(p)
		if err == nil {
			foundPath = p
			break
		}
	}

	if err != nil {
		return fmt.Errorf("could not find policies.json: %w", err)
	}

	var rules []PolicyRule
	if err := json.Unmarshal(data, &rules); err != nil {
		return fmt.Errorf("failed to parse policies.json from %s: %w", foundPath, err)
	}

	globalEngine = &PolicyEngine{rules: rules}
	fmt.Printf("Policy Engine (PDP) initialized successfully using rules from: %s\n", foundPath)
	return nil
}

// Evaluate checks if the given role is allowed to perform action on resource
func Evaluate(role, resource, action string) bool {
	if globalEngine == nil {
		// Fallback: If policy engine not initialized, fail closed
		return false
	}

	for _, rule := range globalEngine.rules {
		if rule.Role == role && rule.Resource == resource {
			for _, act := range rule.Actions {
				if act == action {
					return true
				}
			}
		}
	}

	return false
}
