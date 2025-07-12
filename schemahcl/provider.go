// Copyright 2021-present The Atlas Authors. All rights reserved.
// This source code is licensed under the Apache 2.0 license found
// in the LICENSE file in the root directory of this source tree.

package schemahcl

import (
	"fmt"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"golang.org/x/mod/semver"
)

// ProvidersBlock holds the provider versions required by the schema.
// This is kept for backward compatibility with existing tests.
type ProvidersBlock struct {
	Atlas    string            `hcl:"atlas,optional"`
	Postgres string            `hcl:"postgres,optional"`
	MySQL    string            `hcl:"mysql,optional"`
	SQLite   string            `hcl:"sqlite,optional"`
	Extra    map[string]string `hcl:",remain"`
}

// ProviderConfig represents the providers configuration block in an Atlas schema.
type ProviderConfig struct {
	Atlas    string            `hcl:"atlas,optional"`
	Postgres string            `hcl:"postgres,optional"`
	MySQL    string            `hcl:"mysql,optional"`
	SQLite   string            `hcl:"sqlite,optional"`
	Extra    map[string]string `hcl:",remain"`
}

// ProvidersBlockToConfig converts a ProvidersBlock to a ProviderConfig.
// Additional provider-related functionality can be added here
func ProvidersBlockToConfig(p *ProvidersBlock) *ProviderConfig {
	if p == nil {
		return nil
	}
	return &ProviderConfig{
		Atlas:    p.Atlas,
		Postgres: p.Postgres,
		MySQL:    p.MySQL,
		SQLite:   p.SQLite,
		Extra:    p.Extra,
	}
}

// ParseProviders parses the providers block from HCL.
func ParseProviders(b hcl.Body) (*ProviderConfig, hcl.Diagnostics) {
	providers := &ProviderConfig{
		Extra: make(map[string]string),
	}
	diags := gohcl.DecodeBody(b, nil, providers)
	return providers, diags
}

// ValidateProviderVersions validates the specified provider versions against the current versions.
// It returns an error if any required provider version doesn't match the current version.
func ValidateProviderVersions(providers *ProviderConfig, current map[string]string) error {
	if providers == nil {
		return nil
	}

	// Check Atlas version if specified
	if providers.Atlas != "" {
		if err := validateVersion("atlas", providers.Atlas, current); err != nil {
			return err
		}
	}

	// Check database provider versions
	for provider, version := range map[string]string{
		"postgres": providers.Postgres,
		"mysql":    providers.MySQL,
		"sqlite":   providers.SQLite,
	} {
		if version != "" {
			if err := validateVersion(provider, version, current); err != nil {
				return err
			}
		}
	}

	// Check any additional providers
	for provider, version := range providers.Extra {
		if err := validateVersion(provider, version, current); err != nil {
			return err
		}
	}

	return nil
}

// validateVersion checks if the required version is compatible with the current version.
func validateVersion(provider, required string, current map[string]string) error {
	currentVersion, exists := current[provider]
	if !exists {
		return fmt.Errorf("provider %q is required but not available", provider)
	}

	// Normalize versions by adding v prefix if missing
	if !strings.HasPrefix(required, "v") {
		required = "v" + required
	}
	if !strings.HasPrefix(currentVersion, "v") {
		currentVersion = "v" + currentVersion
	}

	// Check if versions are valid
	if !semver.IsValid(required) {
		return fmt.Errorf("invalid version %q for provider %q", required, provider)
	}
	if !semver.IsValid(currentVersion) {
		return fmt.Errorf("invalid current version %q for provider %q", currentVersion, provider)
	}

	// Compare versions
	if semver.Compare(currentVersion, required) < 0 {
		return fmt.Errorf("provider %q version %s is less than required version %s",
			provider, currentVersion, required)
	}

	return nil
}
