// Copyright 2021-present The Atlas Authors. All rights reserved.
// This source code is licensed under the Apache 2.0 license found
// in the LICENSE file in the root directory of this source tree.

package schemahcl

import (
	"testing"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/stretchr/testify/require"
)

func TestParseProviders(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected interface{} // Can be either *ProvidersBlock or *ProviderConfig
		wantErr  bool
	}{
		{
			name: "basic providers",
			input: `
providers {
  atlas = "1.0.0"
  postgres = "14.0.0"
}`,
			expected: &ProviderConfig{
				Atlas:    "1.0.0",
				Postgres: "14.0.0",
				Extra:    map[string]string{},
			},
			wantErr: false,
		},
		{
			name: "with custom provider",
			input: `
providers {
  atlas = "1.0.0"
  postgres = "14.0.0"
  mysql = "8.0.0"
  cockroachdb = "22.1.0"
}`,
			expected: &ProviderConfig{
				Atlas:    "1.0.0",
				Postgres: "14.0.0",
				MySQL:    "8.0.0",
				Extra: map[string]string{
					"cockroachdb": "22.1.0",
				},
			},
			wantErr: false,
		},
		{
			name: "empty providers block",
			input: `
providers {
}`,
			expected: &ProviderConfig{
				Extra: map[string]string{},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f, diags := hclsyntax.ParseConfig([]byte(tt.input), "test.hcl", hcl.Pos{Line: 1, Column: 1})
			require.False(t, diags.HasErrors(), "parsing failed")

			content, _, diags := f.Body.PartialContent(&hcl.BodySchema{
				Blocks: []hcl.BlockHeaderSchema{
					{
						Type: "providers",
					},
				},
			})
			require.False(t, diags.HasErrors(), "partial content failed")
			require.Len(t, content.Blocks, 1, "expected 1 providers block")

			providers, diags := ParseProviders(content.Blocks[0].Body)
			if tt.wantErr {
				require.True(t, diags.HasErrors())
				return
			}

			require.False(t, diags.HasErrors())

			// Extract values from expected based on its type
			var expectedAtlas, expectedPostgres, expectedMySQL string
			var expectedExtra map[string]string

			switch e := tt.expected.(type) {
			case *ProvidersBlock:
				expectedAtlas = e.Atlas
				expectedPostgres = e.Postgres
				expectedMySQL = e.MySQL
				expectedExtra = e.Extra
			case *ProviderConfig:
				expectedAtlas = e.Atlas
				expectedPostgres = e.Postgres
				expectedMySQL = e.MySQL
				expectedExtra = e.Extra
			}

			require.Equal(t, expectedAtlas, providers.Atlas)
			require.Equal(t, expectedPostgres, providers.Postgres)
			require.Equal(t, expectedMySQL, providers.MySQL)

			// Check extra providers
			for k, v := range expectedExtra {
				actual, ok := providers.Extra[k]
				require.True(t, ok, "expected extra provider %s", k)
				require.Equal(t, v, actual)
			}
		})
	}
}

func TestValidateProviderVersions(t *testing.T) {
	tests := []struct {
		name      string
		providers interface{} // Can be either *ProvidersBlock or *ProviderConfig
		current   map[string]string
		wantErr   bool
		errMsg    string
	}{
		{
			name: "matching versions",
			providers: &ProviderConfig{
				Atlas:    "1.0.0",
				Postgres: "14.0.0",
				Extra:    map[string]string{},
			},
			current: map[string]string{
				"atlas":    "1.0.0",
				"postgres": "14.0.0",
			},
			wantErr: false,
		},
		{
			name: "higher current versions",
			providers: &ProvidersBlock{
				Atlas:    "1.0.0",
				Postgres: "14.0.0",
				Extra:    map[string]string{},
			},
			current: map[string]string{
				"atlas":    "1.1.0",
				"postgres": "15.0.0",
			},
			wantErr: false,
		},
		{
			name: "lower current atlas version",
			providers: &ProvidersBlock{
				Atlas:    "1.1.0",
				Postgres: "14.0.0",
				Extra:    map[string]string{},
			},
			current: map[string]string{
				"atlas":    "1.0.0",
				"postgres": "14.0.0",
			},
			wantErr: true,
			errMsg:  "provider \"atlas\" version v1.0.0 is less than required version v1.1.0",
		},
		{
			name: "lower current postgres version",
			providers: &ProvidersBlock{
				Atlas:    "1.0.0",
				Postgres: "15.0.0",
				MySQL:    "",
				SQLite:   "",
				Extra:    map[string]string{},
			},
			current: map[string]string{
				"atlas":    "1.0.0",
				"postgres": "14.0.0",
			},
			wantErr: true,
			errMsg:  "provider \"postgres\" version v14.0.0 is less than required version v15.0.0",
		},
		{
			name: "missing provider",
			providers: &ProvidersBlock{
				Atlas:    "1.0.0",
				Postgres: "14.0.0",
				MySQL:    "8.0.0",
				Extra:    map[string]string{},
			},
			current: map[string]string{
				"atlas":    "1.0.0",
				"postgres": "14.0.0",
			},
			wantErr: true,
			errMsg:  "provider \"mysql\" is required but not available",
		},
		{
			name: "extra providers",
			providers: &ProvidersBlock{
				Atlas: "1.0.0",
				Extra: map[string]string{
					"cockroachdb": "22.1.0",
				},
			},
			current: map[string]string{
				"atlas":       "1.0.0",
				"cockroachdb": "22.1.0",
			},
			wantErr: false,
		},
		{
			name:      "nil providers",
			providers: nil,
			current: map[string]string{
				"atlas":    "1.0.0",
				"postgres": "14.0.0",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var providerConfig *ProviderConfig
			switch p := tt.providers.(type) {
			case *ProvidersBlock:
				providerConfig = ProvidersBlockToConfig(p)
			case *ProviderConfig:
				providerConfig = p
			case nil:
				providerConfig = nil
			}

			err := ValidateProviderVersions(providerConfig, tt.current)
			if tt.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestSchemaProviderValidation(t *testing.T) {
	const input = `
providers {
  atlas = "1.0.0"
  postgres = "14.0.0"
}

table "users" {
  schema = schema.public
  column "id" {
    type = int
  }
}
`
	f, diags := hclsyntax.ParseConfig([]byte(input), "test.hcl", hcl.Pos{Line: 1, Column: 1})
	require.False(t, diags.HasErrors(), "parsing failed")

	content, _, diags := f.Body.PartialContent(&hcl.BodySchema{
		Blocks: []hcl.BlockHeaderSchema{
			{
				Type: "providers",
			},
		},
	})
	require.False(t, diags.HasErrors(), "partial content failed")

	schema := &struct {
		Providers *ProviderConfig
	}{}

	if len(content.Blocks) > 0 {
		providers, diags := ParseProviders(content.Blocks[0].Body)
		require.False(t, diags.HasErrors(), "parsing providers failed")
		schema.Providers = providers
	}

	err := error(nil)
	require.NoError(t, err)
	require.NotNil(t, schema.Providers)
	require.Equal(t, "1.0.0", schema.Providers.Atlas)
	require.Equal(t, "14.0.0", schema.Providers.Postgres)

	// Test validation
	err = ValidateProviderVersions(schema.Providers, map[string]string{
		"atlas":    "1.0.0",
		"postgres": "14.0.0",
	})
	require.NoError(t, err)

	// Test validation with lower version
	err = ValidateProviderVersions(schema.Providers, map[string]string{
		"atlas":    "0.9.0",
		"postgres": "14.0.0",
	})
	require.Error(t, err)
	if err != nil {
		require.Contains(t, err.Error(), "provider \"atlas\" version v0.9.0 is less than required version v1.0.0")
	}
}
