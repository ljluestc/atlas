// Copyright 2021-present The Atlas Authors. All rights reserved.
// This source code is licensed under the Apache 2.0 license found
// in the LICENSE file in the root directory of this source tree.

package postgres

import (
	"ariga.io/atlas/sql/migrate"
	"ariga.io/atlas/sql/schema"
	"context"
	"fmt"
	"strings"
)

// Migrate is a PostgreSQL schema migration driver.
type Migrate struct {
	// Add fields as needed
}

// tableSpec builds the table creation spec.
func (m *Migrate) tableSpec(t *schema.Table) (*migrate.TableSpec, error) {
	// TODO: Replace with driver logic
	return &migrate.TableSpec{}, nil
}

// formatExclude formats an EXCLUDE constraint from a migration spec.
func formatExclude(e *migrate.ExcludeSpec) string {
	var b strings.Builder
	b.WriteString("CONSTRAINT ")
	if e.Name != "" {
		b.WriteString(quote(e.Name))
		b.WriteString(" ")
	}
	b.WriteString("EXCLUDE")

	// Add USING clause if specified
	if e.Using != "" {
		b.WriteString(" USING ")
		b.WriteString(e.Using)
	}

	// Format the columns/expressions with operators
	b.WriteString(" (")
	for i := 0; i < max(len(e.Columns), len(e.Exprs)); i++ {
		if i > 0 {
			b.WriteString(", ")
		}

		// If we have an expression, use it, otherwise use the column
		if i < len(e.Exprs) && e.Exprs[i] != "" {
			b.WriteString(e.Exprs[i])
		} else if i < len(e.Columns) {
			b.WriteString(quote(e.Columns[i]))
		}

		// Add the operator if available
		if i < len(e.Ops) {
			b.WriteString(" WITH ")
			b.WriteString(e.Ops[i])
		}
	}
	b.WriteString(")")

	// Add predicate if specified
	if e.Predicate != "" {
		b.WriteString(" WHERE ")
		b.WriteString(e.Predicate)
	}
	// Variable 'create' is undefined - commented out until proper implementation
	/*
		for _, exclude := range create.Excludes {
			b.WriteString(",\n  ")
			b.WriteString(formatExclude(exclude))
		}
		for _, exclude := range create.Excludes {
			b.WriteString(",\n  ")
			b.WriteString(formatExclude(exclude))
		}
	*/

	return b.String()
}

// max returns the maximum of two integers.
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// ExcludeSpec represents a PostgreSQL EXCLUDE constraint for adding constraints to existing tables.
type ExcludeSpec struct {
	Name      string
	Using     string   // Index method (e.g., "gist", "btree")
	Index     string   // Alias for Using for backward compatibility
	Columns   []string // Column names
	Ops       []string // Operators for each column (e.g., "=", "&&")
	Predicate string   // Optional WHERE predicate
	Exprs     []string // Optional expressions instead of columns
}

// AddExcludeConstraint adds an EXCLUDE constraint to a table.
func AddExcludeConstraint(ctx context.Context, conn schema.ExecQuerier, t *schema.Table, c *ExcludeSpec) error {
	var b strings.Builder
	b.WriteString("ALTER TABLE ")
	b.WriteString(fmt.Sprintf("%q", t.Name))
	b.WriteString(" ADD CONSTRAINT ")
	b.WriteString(fmt.Sprintf("%q", c.Name))
	b.WriteString(" EXCLUDE USING ")

	// Use Using field if available, otherwise use Index
	indexMethod := c.Using
	if indexMethod == "" {
		indexMethod = c.Index
	}
	b.WriteString(indexMethod)

	b.WriteString(" (")
	for i := 0; i < max(len(c.Columns), len(c.Exprs)); i++ {
		if i > 0 {
			b.WriteString(", ")
		}

		// If we have an expression, use it, otherwise use the column
		if i < len(c.Exprs) && c.Exprs[i] != "" {
			b.WriteString(c.Exprs[i])
		} else if i < len(c.Columns) {
			b.WriteString(fmt.Sprintf("%q", c.Columns[i]))
		}

		// Add the operator if available
		if i < len(c.Ops) {
			b.WriteString(" WITH ")
			b.WriteString(c.Ops[i])
		}
	}
	b.WriteString(")")

	// Add predicate if specified
	if c.Predicate != "" {
		b.WriteString(" WHERE ")
		b.WriteString(c.Predicate)
	}

	_, err := conn.ExecContext(ctx, b.String())
	return err
}
