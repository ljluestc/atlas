// Copyright 2021-present The Atlas Authors. All rights reserved.
// This source code is licensed under the Apache 2.0 license found
// in the LICENSE file in the root directory of this source tree.

package schema

// ExcludeConstraint represents an EXCLUDE constraint.
type ExcludeConstraint struct {
	Name      string            // Constraint name
	Elements  []*ExcludeElement // Elements holds the exclude elements.
	Using     string            // Index method (e.g., "gist", "btree")
	Predicate string            // Optional WHERE predicate
	Attrs     []Attr            // Additional attributes
	table     *Table            // Table this constraint belongs to
}

// ExcludeElement represents an element in an EXCLUDE constraint.
type ExcludeElement struct {
	X        Expr    // Expression for the element
	C        *Column // Column reference
	Operator string  // Operator for the element (e.g., "=", "&&")
	Attrs    []Attr  // Additional attributes
}

// SetPos sets the position of the exclude constraint.
func (e *ExcludeConstraint) SetPos(p *Pos) {
	ReplaceOrAppend(&e.Attrs, p)
}

// Pos of the exclude constraint, if exists.
func (e *ExcludeConstraint) Pos() *Pos {
	for _, a := range e.Attrs {
		if p, ok := a.(*Pos); ok {
			return p
		}
	}
	return nil
}

// Table returns the table this constraint belongs to.
func (e *ExcludeConstraint) Table() *Table {
	return e.table
}

// SetTable sets the table for an ExcludeConstraint.
func (e *ExcludeConstraint) SetTable(t *Table) {
	e.table = t
}

// attr implements the Attr interface.
func (*ExcludeConstraint) attr() {}

// obj implements the Object interface.
func (*ExcludeConstraint) obj() {}

// ExcludeConstraint returns the ExcludeConstraint attribute if exists.
func (t *Table) ExcludeConstraint(name string) (*ExcludeConstraint, bool) {
	for _, attr := range t.Attrs {
		if c, ok := attr.(*ExcludeConstraint); ok && (name == "" || c.Name == name) {
			return c, true
		}
	}
	return nil, false
}
