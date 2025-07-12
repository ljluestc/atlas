package schemahcl

import (
	"encoding/json"
	"github.com/hashicorp/hcl/v2"
)

// --- Stubs for missing types ---
type Table struct{}
type View struct{}

// Schema represents an HCL document that configures an Atlas schema.
type Schema struct {
	Tables         []*Table
	Views          []*View
	Schemas        []*Schema
	Name           string
	Attrs          []*Attr
	Children       []Type          // Tables, Views and nested Schema.
	Qualifiers     []string        // contains a path of names from the root to the current schema.
	Comments       []string        // SchemaComments and associated line comments.
	Providers      *ProviderConfig // Specified provider versions
	json.Marshaler `json:"-"`
	Remain         hcl.Body                 `json:"-"`
	ctx            *hcl.EvalContext         `json:"-"`
	Refs           map[interface{}]struct{} `json:"-"` // references to data source variables
}
