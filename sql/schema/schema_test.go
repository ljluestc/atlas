// Copyright 2021-present The Atlas Authors. All rights reserved.
// This source code is licensed under the Apache 2.0 license found
// in the LICENSE file in the root directory of this source tree.

package schema

import (
	"testing"
)

func TestReplaceOrAppend(t *testing.T) {
	attrs := []Attr{&Comment{Text: "original"}, &Pos{}}
	ReplaceOrAppend(&attrs, &Comment{Text: "replaced"})
	
	if len(attrs) != 2 {
		t.Errorf("Expected length to remain 2, got %d", len(attrs))
	}
	
	if c, ok := attrs[0].(*Comment); !ok || c.Text != "replaced" {
		t.Errorf("Expected first attribute to be replaced comment, got %v", attrs[0])
	}
	
	ReplaceOrAppend(&attrs, &Charset{V: "utf8"})
	
	if len(attrs) != 3 {
		t.Errorf("Expected length to be 3 after append, got %d", len(attrs))
	}
	
	if _, ok := attrs[2].(*Charset); !ok {
		t.Errorf("Expected third attribute to be charset, got %v", attrs[2])
	}
}

func TestRemoveAttr(t *testing.T) {
	attrs := []Attr{&Comment{Text: "comment"}, &Pos{}, &Comment{Text: "another"}}
	
	filtered := RemoveAttr[*Comment](attrs)
	
	if len(filtered) != 1 {
		t.Errorf("Expected length to be 1 after removal, got %d", len(filtered))
	}
	
	if _, ok := filtered[0].(*Pos); !ok {
		t.Errorf("Expected remaining attribute to be Pos, got %v", filtered[0])
	}
}

func TestTableHelpers(t *testing.T) {
	table := NewTable("users")
	if table.Name != "users" {
		t.Errorf("Expected table name to be 'users', got %s", table.Name)
	}
	
	schema := &Schema{Name: "test"}
	table = table.SetSchema(schema)
	if table.Schema != schema {
		t.Errorf("Expected table schema to be set")
	}
	
	col := &Column{Name: "id", Type: &IntegerType{T: "int"}}
	table = table.AddColumns(col)
	if len(table.Columns) != 1 || table.Columns[0] != col {
		t.Errorf("Expected column to be added")
	}
	
	check := &Check{Name: "positive_id", Expr: "id > 0"}
	table = table.AddConstraint(check)
	if len(table.Attrs) != 1 || table.Attrs[0] != check {
		t.Errorf("Expected constraint to be added")
	}
}
