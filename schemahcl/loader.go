package schemahcl

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

// loadSchema loads a Schema from an HCL file. The ctx may be nil, in which case default values are assumed.
func loadSchema(src []byte, name string, ctx *hcl.EvalContext) (*Schema, error) {
	f, diags := hclsyntax.ParseConfig(src, name, hcl.Pos{Line: 1, Column: 1})
	if diags.HasErrors() {
		return nil, diags
	}
	ctx = evalCtxWithDefaults(ctx)

	// Look for providers block first
	schema := &Schema{ctx: ctx}
	content, _, diags := f.Body.PartialContent(providersSchema())
	if !diags.HasErrors() && len(content.Blocks) > 0 {
		providers, diags := ParseProviders(content.Blocks[0].Body)
		if diags.HasErrors() {
			return nil, diags
		}
		// Assign the ProviderConfig directly.
		schema.Providers = providers
	}

	// Load the rest of the schema
	s, err := newSchema(f.Body, ctx)
	if err != nil {
		return nil, err
	}

	// Merge the providers information if it was found
	if schema.Providers != nil {
		s.Providers = schema.Providers
	}

	return s, nil
}

// providersSchema returns a schema for decoding the providers block.
func providersSchema() *hcl.BodySchema {
	return &hcl.BodySchema{
		Blocks: []hcl.BlockHeaderSchema{
			{
				Type: "providers",
			},
		},
	}
}

// Stub: evalCtxWithDefaults returns the provided context or a new one.
func evalCtxWithDefaults(ctx *hcl.EvalContext) *hcl.EvalContext {
	if ctx != nil {
		return ctx
	}
	return &hcl.EvalContext{}
}

// Stub: newSchema returns a new empty Schema for now.
func newSchema(body hcl.Body, ctx *hcl.EvalContext) (*Schema, error) {
	// TODO: Replace with actual schema parsing logic.
	return &Schema{ctx: ctx}, nil
}
