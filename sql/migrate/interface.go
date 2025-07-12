// Copyright 2021-present The Atlas Authors. All rights reserved.
// This source code is licensed under the Apache 2.0 license found
// in the LICENSE file in the root directory of this source tree.

package migrate

import (
	"context"

	"ariga.io/atlas/sql/schema"
)

// Inspector is an interface for database schema inspection operations.
type Inspector interface {
	// InspectSchema returns the schema as it is inspected from the database.
	InspectSchema(ctx context.Context, name string, opts *schema.InspectOptions) (*schema.Schema, error)
}

// Statement represents a migration statement.
type Statement struct {
	Text   string
	Schema string
}

// Driver is the interface that wraps the Migrate and Inspect methods.
type Driver interface {
	Migrate(context.Context, []*Statement) error
	Inspect(context.Context) (*schema.Realm, error)
	InspectRealm(ctx context.Context, opts *schema.InspectRealmOption) (*schema.Realm, error)
	InspectSchema(ctx context.Context, name string, opts *schema.InspectOptions) (*schema.Schema, error)
	RealmDiff(current, desired *schema.Realm, opts ...schema.DiffOption) ([]schema.Change, error)
	SchemaDiff(current, desired *schema.Schema, opts ...schema.DiffOption) ([]schema.Change, error)
	PlanChanges(ctx context.Context, name string, changes []schema.Change, opts ...PlanOption) (*Plan, error)
	ExecContext(ctx context.Context, query string, args ...interface{}) (interface{}, error)
	CheckClean(ctx context.Context, ident *TableIdent) error
}

// RevisionReadWriter is the interface for reading and writing revisions.
type RevisionReadWriter interface {
	Ident() *TableIdent
	ReadRevisions(context.Context) ([]*Revision, error)
	ReadRevision(context.Context, string) (*Revision, error)
	WriteRevision(context.Context, *Revision) error
	DeleteRevision(context.Context, string) error
}

// Execer is an interface for executing SQL statements.
type Execer interface {
	// Execute executes a query.
	Execute(ctx context.Context, query string, skip func(string) bool) error
}

// StateReader reads a database schema state.
type StateReader interface {
	ReadState(ctx context.Context) (*schema.Realm, error)
}

// StateReaderFunc is a function that implements StateReader.
type StateReaderFunc func(ctx context.Context) (*schema.Realm, error)

// ReadState implements StateReader.
func (f StateReaderFunc) ReadState(ctx context.Context) (*schema.Realm, error) {
	return f(ctx)
}

// PlanMode defines the mode of operation for the plan.
type PlanMode int

// PlanMode enumeration.
const (
	PlanModeUnset        PlanMode = iota // Driver default.
	PlanModeDump                         // Multiple files in one transaction.
	PlanModeInPlace                      // Multiple SQL statements in one file.
	PlanModeUnsortedDump                 // Multiple files in one transaction, unordered.
)

// Is reports whether m is equal to the given mode.
func (m PlanMode) Is(mode PlanMode) bool {
	return m == mode
}
