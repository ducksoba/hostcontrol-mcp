package tools

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

type Config struct {
	AllowedPaths        []string `json:"allowed_paths,omitempty"`
	DeniedPaths         []string `json:"denied_paths,omitempty"`
	BashAllowRE         []string `json:"bash_allow_re,omitempty"`
	BashDenyRE          []string `json:"bash_deny_re,omitempty"`
	BashStrict          bool     `json:"bash_strict,omitempty"`
	MaxBashTimeout      int      `json:"max_bash_timeout,omitempty"`
	KillRestrictToOwner bool     `json:"kill_restrict_to_owner,omitempty"`

	bashAllowCompiled []*regexp.Regexp
	bashDenyCompiled  []*regexp.Regexp
}

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	for _, pattern := range cfg.BashAllowRE {
		re, err := regexp.Compile(pattern)
		if err != nil {
			return nil, err
		}
		cfg.bashAllowCompiled = append(cfg.bashAllowCompiled, re)
	}

	for _, pattern := range cfg.BashDenyRE {
		re, err := regexp.Compile(pattern)
		if err != nil {
			return nil, err
		}
		cfg.bashDenyCompiled = append(cfg.bashDenyCompiled, re)
	}

	return &cfg, nil
}

func (c *Config) CheckPath(path string) (bool, string) {
	if c == nil {
		return true, ""
	}

	abs, err := filepath.Abs(path)
	if err != nil {
		return false, "failed to resolve path"
	}
	abs = filepath.Clean(abs)

	for _, denied := range c.DeniedPaths {
		if strings.HasPrefix(abs, filepath.Clean(denied)+"/") || abs == filepath.Clean(denied) {
			return false, "path " + path + " is denied"
		}
	}

	if len(c.AllowedPaths) > 0 {
		allowed := false
		for _, allowedPath := range c.AllowedPaths {
			if strings.HasPrefix(abs, filepath.Clean(allowedPath)+"/") || abs == filepath.Clean(allowedPath) {
				allowed = true
				break
			}
		}
		if !allowed {
			return false, "path " + path + " is not in allowed paths"
		}
	}

	return true, ""
}

func (c *Config) CheckBashCommand(command string) (bool, string) {
	if c == nil {
		return true, ""
	}

	if c.BashStrict {
		if strings.ContainsAny(command, ";|") {
			return false, "command chaining is not allowed in strict mode"
		}
		if strings.Contains(command, "$(") || strings.Contains(command, "`") {
			return false, "command substitution is not allowed in strict mode"
		}
		if strings.Contains(command, "\n") {
			return false, "multiline commands are not allowed in strict mode"
		}
	}

	for _, re := range c.bashDenyCompiled {
		if re.MatchString(command) {
			return false, "command matches deny rule: " + re.String()
		}
	}

	if len(c.bashAllowCompiled) > 0 {
		allowed := false
		for _, re := range c.bashAllowCompiled {
			if re.MatchString(command) {
				allowed = true
				break
			}
		}
		if !allowed {
			return false, "command does not match any allow rule"
		}
	}

	return true, ""
}

func (c *Config) CapTimeout(timeout int) int {
	if c == nil || c.MaxBashTimeout <= 0 {
		return timeout
	}

	if timeout > c.MaxBashTimeout {
		return c.MaxBashTimeout
	}

	return timeout
}

func (c *Config) CheckKillOwner(pid int) (bool, string) {
	if c == nil || !c.KillRestrictToOwner {
		return true, ""
	}

	statusPath := filepath.Join("/proc", strconv.Itoa(pid), "status")
	data, err := os.ReadFile(statusPath)
	if err != nil {
		return false, "failed to read process status"
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "Uid:") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				processUID := parts[1]
				currentUID := strconv.Itoa(os.Getuid())
				if processUID != currentUID {
					return false, "process " + strconv.Itoa(pid) + " is not owned by current user"
				}
				return true, ""
			}
		}
	}

	return false, "could not determine process owner"
}

func (c *Config) CheckPathForWalk(path string) bool {
	if c == nil {
		return true
	}

	abs, err := filepath.Abs(path)
	if err != nil {
		return false
	}
	abs = filepath.Clean(abs)

	for _, denied := range c.DeniedPaths {
		if strings.HasPrefix(abs, filepath.Clean(denied)+"/") || abs == filepath.Clean(denied) {
			return false
		}
	}

	if len(c.AllowedPaths) > 0 {
		for _, allowedPath := range c.AllowedPaths {
			if strings.HasPrefix(abs, filepath.Clean(allowedPath)+"/") || abs == filepath.Clean(allowedPath) {
				return true
			}
		}
		return false
	}

	return true
}
