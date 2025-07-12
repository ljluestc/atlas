// Copyright 2021-present The Atlas Authors. All rights reserved.
// This source code is licensed under the Apache 2.0 license found
// in the LICENSE file in the root directory of this source tree.

package schema

// Constraint represents a database constraint.
type Constraint interface {
	Attr
}

// TableConstraint represents a constraint that belongs to a table.
type TableConstraint interface {
	Constraint
	// Table returns the table this constraint belongs to.
	Table() *Table
}
