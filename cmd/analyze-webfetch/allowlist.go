package main

import (
	"strings"

	"github.com/usadamasa/claude-config/internal/settings"
)

// AllowlistEntry represents a single domain permission entry in settings.json.
type AllowlistEntry struct {
	Tool   string
	Domain string
}

// LoadSettings reads and returns the parsed settings.json.
func LoadSettings(settingsPath string) (*settings.Settings, error) {
	return settings.Load(settingsPath)
}

// ExtractAllowlist extracts WebFetch/Fetch domain permissions from Settings.
func ExtractAllowlist(s *settings.Settings) []AllowlistEntry {
	var entries []AllowlistEntry
	for _, perm := range s.Permissions.Allow {
		entry, ok := parseDomainPermission(perm)
		if ok {
			entries = append(entries, entry)
		}
	}
	return entries
}

// ExtractSandboxDomains extracts sandbox.network.allowedDomains from Settings.
func ExtractSandboxDomains(s *settings.Settings) []string {
	return s.Sandbox.Network.AllowedDomains
}

// parseDomainPermission parses a permission string like "WebFetch(domain:example.com)".
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
