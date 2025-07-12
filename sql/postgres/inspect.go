// Copyright 2021-present The Atlas Authors. All rights reserved.
// This source code is licensed under the Apache 2.0 license found
// in the LICENSE file in the root directory of this source tree.

package postgres

package postgres

import (
	"context"

	"ariga.io/atlas/sql/schema"
)

// BitType represents a bit string type.
// https://postgresql.org/docs/current/datatype-bit.html
type BitType struct {
	schema.Type
	T   string
	Len int
}

// Size returns the length of the bit string.
func (b *BitType) Size() int {
	return b.Len
}

// inspect implements the schema inspection for PostgreSQL.
type inspect struct {
	conn   schema.ExecQuerier
	schema string
}

// parseConstraints parses the table constraints.
func (i *inspect) parseConstraints(ctx context.Context, s *schema.Schema, tables []*schema.Table) error {
	return i.parseTableConstraints(ctx, s, tables)
}

// parseTableConstraints parses the table constraints.
func (i *inspect) parseTableConstraints(ctx context.Context, s *schema.Schema, tables []*schema.Table) error {
	// Implementation to be filled in
	return nil
}
