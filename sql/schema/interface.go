// Copyright 2021-present The Atlas Authors. All rights reserved.
// This source code is licensed under the Apache 2.0 license found
// in the LICENSE file in the root directory of this source tree.

package schema

import (
	"context"
	"strings"
)

// ChangeKind defines the kind of change performed on a resource.
type ChangeKind uint

const (
	// NoChange means no change.
	NoChange ChangeKind = iota

	// SchemaAdded means adding a new schema.
	SchemaAdded

	// ModifySchema means modifying a schema.
	ModifySchema

	// SchemaDropped means dropping a schema.
	SchemaDropped
)

// String implements the fmt.Stringer interface.
func (k ChangeKind) String() string {
	switch k {
	case NoChange:
		return "NoChange"
	case SchemaAdded:
		return "SchemaAdded"
	case ModifySchema:
		return "ModifySchema"
	case SchemaDropped:
		return "SchemaDropped"
	default:
		return "Unknown"
	}
}

// DiffOption configures the schema diff behavior.
type DiffOption func(*DiffOptions)

// DiffOptions holds the options for diffing two schemas.
type DiffOptions struct {
	// Tables defines a list of tables to diff. Empty means all tables.
	Tables []string

	// SkipTables defines a list of tables to skip when diffing schemas.
	SkipTables []string

	// SkipChanges defines a list of changes to skip when diffing schemas.
	SkipChanges []Change

	// Mode defines the diffing mode.
	Mode DiffMode

	// Extra defines per-driver configuration. If not
	// nil, should be set to schemahcl.Extension.
	Extra any // avoid circular dependency with schemahcl.

	// AskFunc can be implemented by the caller to
	// make diff process interactive.
	AskFunc func(string, []string) (string, error)
}

// Change represents a schema change.
type Change interface {
	Kind() ChangeKind
}

// InspectMode defines the inspection mode.
type InspectMode uint

// Inspection mode options.
const (
	// Default inspection mode.
	InspectModeDefault InspectMode = iota
	// Skip inspection of table contents.
	InspectModeNoContents
)

// Inspection mode options.
const (
	InspectSchemas InspectMode = 1 << iota
	InspectTables
	InspectViews
)

// InspectOptions holds schema inspection options.
type InspectOptions struct {
	// Mode defines the inspection behavior.
	Mode InspectMode
	// Tables filters inspection by table names.
	Tables []string
}

// InspectRealmOption allows configuring RealmInspectOptions.
type InspectRealmOption func(*RealmInspectOptions)

// RealmInspectOptions configures realm inspection.
type RealmInspectOptions struct {
	// Schemas filters inspection by schema names.
	Schemas []string
	// Tables holds table inspection options.
	Tables *InspectOptions
}

// Inspector is the interface implemented by the different drivers for inspecting schema.
type Inspector interface {
	// InspectSchema returns the schema description by its name.
	// If the schema does not exist, an error is returned.
	InspectSchema(ctx context.Context, name string, opts *InspectOptions) (*Schema, error)

	// InspectRealm returns the description of a database (realm).
	InspectRealm(ctx context.Context, opts *RealmInspectOptions) (*Realm, error)
}

// WithSchemas configures the schemas to inspect.
func WithSchemas(schemas ...string) InspectRealmOption {
	return func(o *RealmInspectOptions) {
		o.Schemas = append(o.Schemas, schemas...)
	}
}

// WithTables configures the tables to inspect. If options is nil, all tables are inspected.
func WithTables(opts *InspectOptions) InspectRealmOption {
	return func(o *RealmInspectOptions) {
		o.Tables = opts
	}
}

// A Parser is responsible for parsing schema types.
type Parser interface {
	// ParseType parses the given database type and returns a Type.
	ParseType(string) (Type, error)
}

// A TypeFormatter is responsible for formatting schema types.
type TypeFormatter interface {
	// FormatType returns a string representing the database type of a column.
	FormatType(Type) (string, error)
}

// TypeParseFormatter is a merger of the Parser and TypeFormatter
// interfaces that is implemented by most dialects.
type TypeParseFormatter interface {
	Parser
	TypeFormatter
}

// A Differ provides the functionality for computing the difference
// between two schemas and generating a migrate.Plan for aligning them.
type Differ interface {
	// SchemaDiff computes the difference between two schemas.
	SchemaDiff(from, to *Schema, opts ...*DiffOptions) ([]Change, error)

	// RealmDiff computes the difference between two realms.
	RealmDiff(from, to *Realm, opts ...*DiffOptions) ([]Change, error)
}

// A Planner provides the functionality for generating database
// changes from a Differ output in a topological order.
type Planner interface {
	// PlanChanges returns an ordered list of Changes that represent a logical plan
	// for applying the given schemas changes.
	PlanChanges(ctx context.Context, name string, changes []Change, opts ...PlanOption) (*Plan, error)
}

// Plan represents a database migration plan.
type Plan struct {
	// Name of the plan.
	Name string

	// Changes holds the list of changes to apply.
	Changes []Change
}

// PlanOption configures planning behavior.
type PlanOption func(*PlanOptions)

// PlanOptions holds the options for planning schema changes.
type PlanOptions struct {
	// SkipChanges defines a list of changes to skip when planning.
	SkipChanges []Change
}

// DiffMode defines the diffing mode.
type DiffMode int

// Inspector interface methods.
func (r *Realm) Schemas() []*Schema         { return r.Schemas }
func (r *Realm) Objects() []Object          { return r.Objects }
func (s *Schema) Tables() []*Table          { return s.Tables }
func (s *Schema) Views() []*View            { return s.Views }
func (s *Schema) Objects() []Object         { return s.Objects }
func (t *Table) Columns() []*Column         { return t.Columns }
func (t *Table) Indexes() []*Index          { return t.Indexes }
func (v *View) Columns() []*Column          { return v.Columns }
func (v *View) Indexes() []*Index           { return v.Indexes }
func (t *Table) PK() *Index                 { return t.PrimaryKey }
func (t *Table) Name() string               { return t.Name }
func (t *Table) Schema() *Schema            { return t.Schema }
func (v *View) Name() string                { return v.Name }
func (v *View) Schema() *Schema             { return v.Schema }
func (c *Column) Name() string              { return c.Name }
func (c *Column) Table() *Table             { return c.Table }
func (i *Index) Table() *Table              { return i.Table }
func (i *Index) Parts() []*IndexPart        { return i.Parts }
func (i *Index) Name() string               { return i.Name }
func (p *IndexPart) Column() *Column        { return p.C }
func (f *ForeignKey) Name() string          { return f.Symbol }
func (f *ForeignKey) Table() *Table         { return f.Table }
func (f *ForeignKey) RefTable() *Table      { return f.RefTable }
func (f *ForeignKey) RefColumns() []*Column { return f.RefColumns }

// IfExists and IfNotExists for migration.go
const (
	IfExists    = "IF EXISTS"
	IfNotExists = "IF NOT EXISTS"
)
