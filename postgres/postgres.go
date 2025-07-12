package postgres

import (
	"ariga.io/atlas/sql/migrate"
	"ariga.io/atlas/sql/schema"
	"context"
	"fmt"
	"strings"
)

// pgIndex represents a PostgreSQL index.
type pgIndex struct {
	name   string // Index name
	table  string // Table name
	unique bool   // Unique index
}

// pgConstraint represents a PostgreSQL constraint.
type pgConstraint struct {
	name string // Constraint name
	typ  string // Constraint type
}

// pgExcludeConstraint represents a PostgreSQL EXCLUDE constraint.
type pgExcludeConstraint struct {
	name      string   // Constraint name
	table     string   // Table name
	columns   []string // Column names
	ops       []string // Operators
	predicate string   // WHERE clause
	using     string   // USING method
	exprs     []string // Expressions
}

// max returns the maximum of two integers.
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// formatExclude formats an EXCLUDE constraint.
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

	return b.String()
}

// quote returns the quoted identifier.
func quote(s string) string {
	return fmt.Sprintf("%q", s)
}
