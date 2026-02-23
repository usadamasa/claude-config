package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// AllowlistEntry represents a single domain permission entry in settings.json.
type AllowlistEntry struct {
	Tool   string // "WebFetch" or "Fetch"
	Domain string // e.g. "github.com" or "*.databricks.com"
}

// settingsJSON represents the relevant parts of a Claude settings.json file.
type settingsJSON struct {
	Permissions struct {
		Allow []string `json:"allow"`
	} `json:"permissions"`
	Sandbox struct {
		Network struct {
			AllowedDomains []string `json:"allowedDomains"`
		} `json:"network"`
	} `json:"sandbox"`
}

// LoadAllowlist reads settings.json and extracts WebFetch/Fetch domain permissions.
func LoadAllowlist(settingsPath string) ([]AllowlistEntry, error) {
	data, err := os.ReadFile(settingsPath)
	if err != nil {
		return nil, fmt.Errorf("reading settings file: %w", err)
	}

	var settings settingsJSON
	if err := json.Unmarshal(data, &settings); err != nil {
		return nil, fmt.Errorf("parsing settings JSON: %w", err)
	}

	var entries []AllowlistEntry
	for _, perm := range settings.Permissions.Allow {
		entry, ok := parseDomainPermission(perm)
		if ok {
			entries = append(entries, entry)
		}
	}
	return entries, nil
}

// LoadSandboxDomains reads settings.json and extracts sandbox.network.allowedDomains.
// キーが存在しない場合は空スライスを返す(エラーにしない)。
func LoadSandboxDomains(settingsPath string) ([]string, error) {
	data, err := os.ReadFile(settingsPath)
	if err != nil {
		return nil, fmt.Errorf("reading settings file: %w", err)
	}

	var settings settingsJSON
	if err := json.Unmarshal(data, &settings); err != nil {
		return nil, fmt.Errorf("parsing settings JSON: %w", err)
	}

	return settings.Sandbox.Network.AllowedDomains, nil
}

// parseDomainPermission parses a permission string like "WebFetch(domain:example.com)"
// and returns an AllowlistEntry. Returns false if the string is not a domain permission.
func parseDomainPermission(perm string) (AllowlistEntry, bool) {
	for _, tool := range []string{"WebFetch", "Fetch"} {
		prefix := tool + "(domain:"
		if strings.HasPrefix(perm, prefix) && strings.HasSuffix(perm, ")") {
			domain := perm[len(prefix) : len(perm)-1]
			return AllowlistEntry{Tool: tool, Domain: domain}, true
		}
	}
	return AllowlistEntry{}, false
}
