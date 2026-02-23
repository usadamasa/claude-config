package main

import (
	"testing"
)

func TestCategorizeDomain(t *testing.T) {
	tests := []struct {
		domain   string
		expected Category
	}{
		// safe: official documentation
		{"docs.example.com", CategorySafe},
		{"developer.mozilla.org", CategorySafe},
		{"mylib.readthedocs.io", CategorySafe},

		// safe: code hosting
		{"github.com", CategorySafe},
		{"gitlab.com", CategorySafe},
		{"raw.githubusercontent.com", CategorySafe},

		// safe: package registries
		{"pkg.go.dev", CategorySafe},
		{"go.dev", CategorySafe},
		{"www.npmjs.com", CategorySafe},
		{"registry.terraform.io", CategorySafe},

		// safe: cloud providers
		{"cloud.google.com", CategorySafe},
		{"console.databricks.com", CategorySafe},
		{"docs.anthropic.com", CategorySafe},

		// medium: community
		{"stackoverflow.com", CategoryMedium},
		{"medium.com", CategoryMedium},
		{"dev.to", CategoryMedium},

		// medium: learning
		{"learning.oreilly.com", CategoryMedium},

		// review: unknown domains
		{"some-random-site.xyz", CategoryReview},
		{"malicious-site.ru", CategoryReview},
	}

	for _, tt := range tests {
		t.Run(tt.domain, func(t *testing.T) {
			result := CategorizeDomain(tt.domain)
			if result.Category != tt.expected {
				t.Errorf("CategorizeDomain(%q) = %q (reason: %s), want %q",
					tt.domain, result.Category, result.Reason, tt.expected)
			}
		})
	}
}

func TestCategorizeDomain_HasReason(t *testing.T) {
	result := CategorizeDomain("github.com")
	if result.Reason == "" {
		t.Error("expected non-empty reason for categorization")
	}
}

func TestExtractDomain(t *testing.T) {
	tests := []struct {
		rawURL   string
		expected string
		wantErr  bool
	}{
		{"https://github.com/foo/bar", "github.com", false},
		{"https://docs.example.com/api/v1?key=value", "docs.example.com", false},
		{"http://localhost:8080/path", "localhost", false},
		{"https://sub.domain.example.com", "sub.domain.example.com", false},
		{"", "", true},
		{"not-a-url", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.rawURL, func(t *testing.T) {
			got, err := ExtractDomain(tt.rawURL)
			if (err != nil) != tt.wantErr {
				t.Errorf("ExtractDomain(%q) error = %v, wantErr %v", tt.rawURL, err, tt.wantErr)
				return
			}
			if got != tt.expected {
				t.Errorf("ExtractDomain(%q) = %q, want %q", tt.rawURL, got, tt.expected)
			}
		})
	}
}
