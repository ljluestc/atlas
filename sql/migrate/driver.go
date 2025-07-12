// Copyright 2021-present The Atlas Authors. All rights reserved.
// This source code is licensed under the Apache 2.0 license found
// in the LICENSE file in the root directory of this source tree.

package migrate

import (
	"context"
	
	"ariga.io/atlas/sql/schema"
)

// Driver wraps the functionality needed to execute migrations on a database.
type Driver interface {
	// ExecContext executes a SQL statement on the database.
	ExecContext(context.Context, string) error
	
	// InspectSchema inspects the database schema with the given name.
	InspectSchema(ctx context.Context, name string, opts *schema.InspectOptions) (*schema.Schema, error)
	
	// InspectRealm inspects the entire database realm.
	InspectRealm(ctx context.Context, opts *schema.InspectRealmOption) (*schema.Realm, error)
	
	// RealmDiff computes the difference between two database realms.
	RealmDiff(from, to *schema.Realm, opts ...schema.DiffOption) ([]schema.Change, error)
	
	// SchemaDiff computes the difference between two database schemas.
	SchemaDiff(from, to *schema.Schema, opts ...schema.DiffOption) ([]schema.Change, error)
	
	// PlanChanges creates a migration plan from a list of changes.
	PlanChanges(ctx context.Context, name string, changes []schema.Change, opts ...PlanOption) (*Plan, error)
	
	// CheckClean checks if the database is in a clean state.
	CheckClean(ctx context.Context, ident *TableIdent) error
}

// PlanOptions are options for planning changes.
type PlanOptions struct {
	// SchemaQualifier is the schema to prefix tables and other resources with.
	SchemaQualifier *string
	
	// Indent is the indentation to use for generated SQL statements.
	Indent string
	
	// Mode is the planning mode to use.
	Mode PlanMode
}
