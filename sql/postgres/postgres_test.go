// Copyright 2021-present The Atlas Authors. All rights reserved.
// This source code is licensed under the Apache 2.0 license found
// in the LICENSE file in the root directory of this source tree.

package postgres

import (
	"context"
	"database/sql"
	"testing"

	"ariga.io/atlas/sql/migrate"
	"ariga.io/atlas/sql/schema"
	"github.com/stretchr/testify/require"
)

type mockExecQuerier struct {
	schema.ExecQuerier
}

func (m *mockExecQuerier) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return nil, nil
}

func (m *mockExecQuerier) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return nil, nil
}

type dialectOpener func(string) (schema.ExecQuerier, error)

func (d dialectOpener) Open(name string) (schema.ExecQuerier, error) {
	return d(name)
}

func TestDriver_Table(t *testing.T) {
	tbl := &schema.Table{
		Name: "users",
		Columns: []*schema.Column{
			{Name: "id", Type: &schema.ColumnType{Type: &schema.IntegerType{T: "int"}}},
			{Name: "age", Type: &schema.ColumnType{Type: &schema.IntegerType{T: "int"}}, Default: &schema.RawExpr{X: "10"}},
			{Name: "c1", Type: &schema.ColumnType{Type: &schema.IntegerType{T: "int"}, Null: true}},
			{Name: "c2", Type: &schema.ColumnType{Type: &schema.IntegerType{T: "int"}}, Default: &schema.Literal{V: "10"}},
			{Name: "c3", Type: &schema.ColumnType{Raw: "serial", Type: &schema.IntegerType{T: "serial"}}},
			{Name: "c4", Type: &schema.ColumnType{Type: &schema.TimeType{T: "date"}}, Default: &schema.RawExpr{X: "CURRENT_DATE"}},
			{Name: "c5", Type: &schema.ColumnType{Type: &schema.TimeType{T: "time"}}},
			{Name: "c6", Type: &schema.ColumnType{Type: &schema.TimeType{T: "time with time zone"}}},
			{Name: "c7", Type: &schema.ColumnType{Type: &schema.JSONType{T: "json"}}},
			{Name: "c8", Type: &schema.ColumnType{Type: &schema.JSONType{T: "jsonb"}}},
			{Name: "c9", Type: &schema.ColumnType{Raw: "hstore", Type: &schema.UnsupportedType{T: "hstore"}}},
			{Name: "c10", Type: &schema.ColumnType{Type: &schema.EnumType{T: "enum_int", Values: []string{"a", "b"}}}},
			{Name: "c11", Type: &schema.ColumnType{Type: &schema.StringType{T: "character", Size: 10}}},
			{Name: "c12", Type: &schema.ColumnType{Type: &schema.SpatialType{T: "geography"}}, Default: &schema.RawExpr{X: "ST_GeomFromText('POINT(0 0)')"}},
			{Name: "c13", Type: &schema.ColumnType{Type: &schema.SpatialType{T: "geometry"}}, Default: &schema.RawExpr{X: "ST_MakePoint(0, 0)"}},
			{Name: "c14", Type: &schema.ColumnType{Type: &schema.StringType{T: "char", Size: 10}}},
			{Name: "c15", Type: &schema.ColumnType{Type: &schema.StringType{T: "character varying", Size: 10}}},
			{Name: "c16", Type: &schema.ColumnType{Type: &schema.StringType{T: "varchar", Size: 10}}},
			{Name: "c17", Type: &schema.ColumnType{Type: &schema.BinaryType{T: "bytea"}}},
			{Name: "c18", Type: &schema.ColumnType{Type: &schema.BinaryType{T: "bytea"}}, Default: &schema.RawExpr{X: "decode('0102', 'hex')"}},
			{Name: "c19", Type: &schema.ColumnType{Type: &schema.BoolType{T: "boolean"}}},
			{Name: "c20", Type: &schema.ColumnType{Type: &schema.BoolType{T: "boolean"}}, Default: &schema.Literal{V: "true"}},
			{Name: "c21", Type: &schema.ColumnType{Type: &schema.IntegerType{T: "smallint"}}},
			{Name: "c22", Type: &schema.ColumnType{Type: &schema.DecimalType{T: "decimal", Precision: 10, Scale: 2}}},
			{Name: "c23", Type: &schema.ColumnType{Type: &schema.IntegerType{T: "int", Unsigned: true}}},
			{Name: "c24", Type: &schema.ColumnType{Type: &schema.FloatType{T: "float", Precision: 10}}},
			{Name: "c25", Type: &schema.ColumnType{Type: &schema.FloatType{T: "real"}}},
			{Name: "c26", Type: &schema.ColumnType{Type: &schema.UUIDType{T: "uuid"}}},
			{Name: "c27", Type: &schema.ColumnType{Type: &schema.TimeType{T: "timestamp"}}},
			{Name: "c28", Type: &schema.ColumnType{Type: &schema.TimeType{T: "timestamp with time zone"}}},
			{Name: "c29", Type: &schema.ColumnType{Type: &schema.TimeType{T: "timestamp with time zone"}}, Default: &schema.RawExpr{X: "CURRENT_TIMESTAMP"}},
			{Name: "c30", Type: &schema.ColumnType{Type: &schema.StringType{T: "citext"}}},
			{Name: "c31", Type: &schema.ColumnType{Type: &schema.StringType{T: "citext"}}, Default: &schema.Literal{V: "hello"}},
			{Name: "c32", Type: &schema.ColumnType{Type: &schema.BitType{T: "bit", Len: 10}}},
			{Name: "c33", Type: &schema.ColumnType{Type: &schema.BitType{T: "bit varying", Len: 10}}},
			{Name: "c34", Type: &schema.ColumnType{Type: &schema.BitType{T: "varbit", Len: 10}}},
			{Name: "c35", Type: &schema.ColumnType{Type: &schema.BitType{T: "bit", Len: 10}}, Default: &schema.RawExpr{X: "B'0'"}},
			{Name: "c36", Type: &schema.ColumnType{Type: &schema.StringType{T: "text"}}, Attrs: []schema.Attr{&schema.GeneratedExpr{Expr: "'gen_col'", Type: "STORED"}}},
			{Name: "c37", Type: &schema.ColumnType{Type: &schema.StringType{T: "text"}}, Attrs: []schema.Attr{&schema.GeneratedExpr{Expr: "'gen_col'", Type: "VIRTUAL"}}},
			{Name: "c38", Type: &schema.ColumnType{Type: &schema.IntegerType{T: "integer"}, Null: true}},
		},
		Attrs: []schema.Attr{
			&schema.Check{
				Name:  "users_c1",
				Expr:  "c1 <> 0",
				Attrs: []schema.Attr{&schema.NoInherit{}},
			},
			&schema.Check{
				Name: "users_c38_check",
				Expr: "c38 > 0",
			},
		},
	}
	s := `CREATE TABLE "users" (
  "id" integer NOT NULL,
  "age" integer NOT NULL DEFAULT 10,
  "c1" integer NULL,
  "c2" integer NOT NULL DEFAULT '10',
  "c3" serial NOT NULL,
  "c4" date NOT NULL DEFAULT CURRENT_DATE,
  "c5" time NOT NULL,
  "c6" time with time zone NOT NULL,
  "c7" json NOT NULL,
  "c8" jsonb NOT NULL,
  "c9" hstore NOT NULL,
  "c10" "enum_int" NOT NULL,
  "c11" character(10) NOT NULL,
  "c12" geography NOT NULL DEFAULT ST_GeomFromText('POINT(0 0)'),
  "c13" geometry NOT NULL DEFAULT ST_MakePoint(0, 0),
  "c14" char(10) NOT NULL,
  "c15" character varying(10) NOT NULL,
  "c16" varchar(10) NOT NULL,
  "c17" bytea NOT NULL,
  "c18" bytea NOT NULL DEFAULT decode('0102', 'hex'),
  "c19" boolean NOT NULL,
  "c20" boolean NOT NULL DEFAULT 'true',
  "c21" smallint NOT NULL,
  "c22" decimal(10,2) NOT NULL,
  "c23" integer NOT NULL,
  "c24" float(10) NOT NULL,
  "c25" real NOT NULL,
  "c26" uuid NOT NULL,
  "c27" timestamp NOT NULL,
  "c28" timestamp with time zone NOT NULL,
  "c29" timestamp with time zone NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "c30" citext NOT NULL,
  "c31" citext NOT NULL DEFAULT 'hello',
  "c32" bit(10) NOT NULL,
  "c33" bit varying(10) NOT NULL,
  "c34" varbit(10) NOT NULL,
  "c35" bit(10) NOT NULL DEFAULT B'0',
  "c36" text GENERATED ALWAYS AS ('gen_col') STORED NOT NULL,
  "c37" text GENERATED ALWAYS AS ('gen_col') STORED NOT NULL,
  "c38" integer NULL,
  CONSTRAINT "users_c1" CHECK (c1 <> 0) NO INHERIT,
  CONSTRAINT "users_c38_check" CHECK (c38 > 0)
)`
	t.Run("Create", func(t *testing.T) {
		drv := &Driver{dialectOpener: dialectOpener(func(string) (schema.ExecQuerier, error) {
			return &mockExecQuerier{}, nil
		})}
		require.Equal(t, s, drv.tableSpec(tbl).String())
	})
}

func TestDriver_ExcludeConstraint(t *testing.T) {
	tbl := &schema.Table{
		Name: "meetings",
		Columns: []*schema.Column{
			{Name: "id", Type: &schema.ColumnType{Type: &schema.IntegerType{T: "int"}}},
			{Name: "room", Type: &schema.ColumnType{Type: &schema.IntegerType{T: "int"}}},
			{Name: "during", Type: &schema.ColumnType{Raw: "tsrange", Type: &schema.UnsupportedType{T: "tsrange"}}},
		},
		Attrs: []schema.Attr{
			&schema.ExcludeConstraint{
				Name:    "exclude_overlapping_meetings",
				Columns: []string{"room", "during"},
				Ops:     []string{"=", "&&"},
				Using:   "gist",
			},
		},
	}
	s := `CREATE TABLE "meetings" (
  "id" integer NOT NULL,
  "room" integer NOT NULL,
  "during" tsrange NOT NULL,
  CONSTRAINT "exclude_overlapping_meetings" EXCLUDE USING gist ("room" WITH =, "during" WITH &&)
)`
	t.Run("Create with Exclude", func(t *testing.T) {
		drv := &Driver{dialectOpener: dialectOpener(func(string) (schema.ExecQuerier, error) {
			return &mockExecQuerier{}, nil
		})}
		require.Equal(t, s, drv.tableSpec(tbl).String())
	})

	// Test with a predicate
	tbl2 := &schema.Table{
		Name: "meetings",
		Columns: []*schema.Column{
			{Name: "id", Type: &schema.ColumnType{Type: &schema.IntegerType{T: "int"}}},
			{Name: "room", Type: &schema.ColumnType{Type: &schema.IntegerType{T: "int"}}},
			{Name: "during", Type: &schema.ColumnType{Raw: "tsrange", Type: &schema.UnsupportedType{T: "tsrange"}}},
			{Name: "active", Type: &schema.ColumnType{Type: &schema.BoolType{T: "boolean"}}},
		},
		Attrs: []schema.Attr{
			&schema.ExcludeConstraint{
				Name:      "exclude_overlapping_active_meetings",
				Columns:   []string{"room", "during"},
				Ops:       []string{"=", "&&"},
				Using:     "gist",
				Predicate: "active = true",
			},
		},
	}
	s2 := `CREATE TABLE "meetings" (
  "id" integer NOT NULL,
  "room" integer NOT NULL,
  "during" tsrange NOT NULL,
  "active" boolean NOT NULL,
  CONSTRAINT "exclude_overlapping_active_meetings" EXCLUDE USING gist ("room" WITH =, "during" WITH &&) WHERE active = true
)`
	t.Run("Create with Exclude and Predicate", func(t *testing.T) {
		drv := &Driver{dialectOpener: dialectOpener(func(string) (schema.ExecQuerier, error) {
			return &mockExecQuerier{}, nil
		})}
		require.Equal(t, s2, drv.tableSpec(tbl2).String())
	})

	// Test with expressions
	tbl3 := &schema.Table{
		Name: "points",
		Columns: []*schema.Column{
			{Name: "id", Type: &schema.ColumnType{Type: &schema.IntegerType{T: "int"}}},
			{Name: "p", Type: &schema.ColumnType{Type: &schema.SpatialType{T: "point"}}},
		},
		Attrs: []schema.Attr{
			&schema.ExcludeConstraint{
				Name:  "exclude_overlapping_points",
				Exprs: []string{"p <-> p", "p"},
				Ops:   []string{"<", "&&"},
				Using: "gist",
			},
		},
	}
	s3 := `CREATE TABLE "points" (
  "id" integer NOT NULL,
  "p" point NOT NULL,
  CONSTRAINT "exclude_overlapping_points" EXCLUDE USING gist (p <-> p WITH <, p WITH &&)
)`
	t.Run("Create with Exclude using Expressions", func(t *testing.T) {
		drv := &Driver{dialectOpener: dialectOpener(func(string) (schema.ExecQuerier, error) {
			return &mockExecQuerier{}, nil
		})}
		require.Equal(t, s3, drv.tableSpec(tbl3).String())
	})
}

func TestProcessExcludeConstraints(t *testing.T) {
	// Create a mock inspector and schema
	i := &inspect{conn: nil} // In a real test you'd mock the connection
	s := &schema.Schema{Name: "test"}

	// Create test tables
	tables := []*schema.Table{
		{Name: "events"},
		{Name: "other_table"},
	}

	// Create mock exclude constraints
	excludes := map[string][]*pgExcludeConstraint{
		"events": {
			{
				name:      "no_overlapping_events",
				columns:   []string{"room", "time_period"},
				ops:       []string{"=", "&&"},
				using:     "gist",
				predicate: "WHERE active = true",
				exprs:     []string{"room", "tsrange(start_time, end_time) AS time_period"},
			},
		},
	}

	// Process the constraints
	err := i.processExcludeConstraints(context.Background(), s, tables, excludes)
	if err != nil {
		t.Fatalf("Failed to process exclude constraints: %v", err)
	}

	// Verify the constraint was added to the table
	eventTable := tables[0]
	found := false
	for _, attr := range eventTable.Attrs {
		if ex, ok := attr.(*schema.ExcludeConstraint); ok {
			if ex.Name == "no_overlapping_events" {
				found = true
				if ex.Using != "gist" {
					t.Errorf("Expected using method 'gist', got '%s'", ex.Using)
				}
				if len(ex.Columns) != 2 {
					t.Errorf("Expected 2 columns, got %d", len(ex.Columns))
				}
				if len(ex.Ops) != 2 {
					t.Errorf("Expected 2 operators, got %d", len(ex.Ops))
				}
				if ex.Predicate != "WHERE active = true" {
					t.Errorf("Expected predicate 'WHERE active = true', got '%s'", ex.Predicate)
				}
			}
		}
	}

	if !found {
		t.Error("Exclude constraint was not added to the table")
	}
}
