package accesscontrol

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"
)

type Policy struct {
	Default string          `json:"default"`
	Tools   map[string]Tool `json:"tools"`
}

type Tool struct {
	Action     string                `json:"-"`
	Parameters map[string]ParamRules `json:"-"`
}

type ParamRules struct {
	Allow   []string `json:"allow,omitempty"`
	AllowRE []string `json:"allow_re,omitempty"`
	Deny    []string `json:"deny,omitempty"`
	DenyRE  []string `json:"deny_re,omitempty"`
}

func (p *Policy) IsDefaultAllow() bool {
	return p.Default != "deny"
}

func (p *Policy) CheckTool(toolName string, args map[string]any) (bool, string) {
	rule, exists := p.Tools[toolName]
	if !exists {
		return p.IsDefaultAllow(), ""
	}

	if rule.Action == "deny" {
		return false, "tool " + toolName + " is denied"
	}

	for param, rules := range rule.Parameters {
		if val, ok := args[param]; ok {
			allowed, reason := rules.Match(fmt.Sprintf("%v", val))
			if !allowed {
				return false, "parameter " + param + "=" + fmt.Sprintf("%v", val) + " denied for tool " + toolName + reason
			}
		}
	}

	return true, ""
}

func (p *ParamRules) Match(value string) (bool, string) {
	for _, pattern := range p.Deny {
		if value == pattern || strings.HasPrefix(value, pattern) {
			return false, ""
		}
	}
	for _, pattern := range p.DenyRE {
		if matched, _ := regexp.MatchString(pattern, value); matched {
			return false, ""
		}
	}
	if len(p.Allow) == 0 && len(p.AllowRE) == 0 {
		return true, ""
	}
	for _, pattern := range p.Allow {
		if value == pattern || strings.HasPrefix(value, pattern) {
			return true, ""
		}
	}
	for _, pattern := range p.AllowRE {
		if matched, _ := regexp.MatchString(pattern, value); matched {
			return true, ""
		}
	}
	return false, " (no matching allow rule)"
}

func (t *Tool) UnmarshalJSON(data []byte) error {
	var raw string
	if err := json.Unmarshal(data, &raw); err == nil {
		t.Action = raw
		return nil
	}

	var obj struct {
		Action     string                `json:"action"`
		Parameters map[string]ParamRules `json:"parameters"`
	}
	if err := json.Unmarshal(data, &obj); err != nil {
		return err
	}

	if obj.Action == "" {
		obj.Action = "allow"
	}
	t.Action = obj.Action
	t.Parameters = obj.Parameters
	return nil
}

func LoadPolicyFromURL(url string) (*Policy, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("failed to fetch policy: " + resp.Status)
	}

	var policy Policy
	if err := json.NewDecoder(resp.Body).Decode(&policy); err != nil {
		return nil, err
	}
	return &policy, nil
}

func LoadPolicyFromFile(path string) (*Policy, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var policy Policy
	if err := json.Unmarshal(data, &policy); err != nil {
		return nil, err
	}
	return &policy, nil
}
