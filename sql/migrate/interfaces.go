// Copyright 2021-present The Atlas Authors. All rights reserved.
// This source code is licensed under the Apache 2.0 license found
// in the LICENSE file in the root directory of this source tree.

package migrate

import (
	"context"
	"time"

	"ariga.io/atlas/sql/schema"
)

// Migration represents a database migration.
type Migration struct {
	Name     string    // Name of the migration
	Version  string    // Version of the migration
	Created  time.Time // Creation time
	Applied  time.Time // When it was applied
	SQL      string    // SQL statement
	Metadata []byte    // Additional metadata
}

// RepeatableMigration is an interface that a migration can implement to indicate that it should
// be executed each time even if it has already been executed. This is useful for migrations that
// create or update views, stored procedures, or other database objects that are safe to recreate.
type RepeatableMigration interface {
	// Repeatable returns true if the migration can be applied multiple times.
	Repeatable() bool
}

// MigrationDriver is the interface that wraps the basic operations for executing migrations on a database.
type MigrationDriver interface {
	// ApplyMigration applies a single migration on the database.
	ApplyMigration(ctx context.Context, m *Migration) error

	// CheckRepeatable checks if a migration is repeatable.
	CheckRepeatable(ctx context.Context, m *Migration) (bool, error)
}

// TableSpec defines a table structure for migrations.
type TableSpec struct {
	Name    string
	Columns []*schema.Column
	Indexes []*schema.Index
	PK      *schema.Index
	FKs     []*schema.ForeignKey
}

// ExcludeSpec defines an EXCLUDE constraint for PostgreSQL.
type ExcludeSpec struct {
	Name    string
	Index   string
	Columns []string
	Ops     []string
}

// Migrate represents the main migration manager.
type Migrate struct {
	// Implementation details would go here
}
