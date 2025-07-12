// Copyright 2021-present The Atlas Authors. All rights reserved.
// This source code is licensed under the Apache 2.0 license found
// in the LICENSE file in the root directory of this source tree.

//go:build !ent

package postgres

import (
	"context"
	"io"
	"testing"
	"time"

	"ariga.io/atlas/sql/internal/sqltest"
	"ariga.io/atlas/sql/migrate"
	"ariga.io/atlas/sql/schema"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
)

func TestDriver_LockAcquired(t *testing.T) {
	db, m, err := sqlmock.New()
	require.NoError(t, err)
	name, hash := "name", 797654004

	t.Run("NoTimeout", func(t *testing.T) {
		m.ExpectQuery(sqltest.Escape("SELECT pg_try_advisory_lock($1)")).
			WithArgs(hash).
			WillReturnRows(sqlmock.NewRows([]string{"pg_advisory_lock"}).AddRow(1)).
			RowsWillBeClosed()
		m.ExpectQuery(sqltest.Escape("SELECT pg_advisory_unlock($1)")).
			WithArgs(hash).
			WillReturnRows(sqlmock.NewRows([]string{"pg_advisory_unlock"}).AddRow(1)).
			RowsWillBeClosed()

		d := &Driver{conn: &conn{ExecQuerier: db}}
		unlock, err := d.Lock(context.Background(), name, 0)
		require.NoError(t, err)
		require.NoError(t, unlock())
		require.NoError(t, m.ExpectationsWereMet())
	})

	t.Run("WithTimeout", func(t *testing.T) {
		m.ExpectQuery(sqltest.Escape("SELECT pg_try_advisory_lock($1)")).
			WithArgs(hash).
			WillReturnRows(sqlmock.NewRows([]string{"pg_advisory_lock"}).AddRow(0)).
			RowsWillBeClosed()
		m.ExpectQuery(sqltest.Escape("SELECT pg_try_advisory_lock($1)")).
			WithArgs(hash).
			WillReturnRows(sqlmock.NewRows([]string{"pg_advisory_lock"}).AddRow(1)).
			RowsWillBeClosed()
		m.ExpectQuery(sqltest.Escape("SELECT pg_advisory_unlock($1)")).
			WithArgs(hash).
			WillReturnRows(sqlmock.NewRows([]string{"pg_advisory_unlock"}).AddRow(1)).
			RowsWillBeClosed()

		d := &Driver{conn: &conn{ExecQuerier: db}}
		unlock, err := d.Lock(context.Background(), name, time.Second)
		require.NoError(t, err)
		require.NoError(t, unlock())
		require.NoError(t, m.ExpectationsWereMet())
	})
}

func TestDriver_LockError(t *testing.T) {
	db, m, err := sqlmock.New()
	require.NoError(t, err)
	d := &Driver{conn: &conn{ExecQuerier: db}}
	name, hash := "migrate", 979249972

	t.Run("Internal", func(t *testing.T) {
		m.ExpectQuery(sqltest.Escape("SELECT pg_try_advisory_lock($1)")).
			WithArgs(hash).
			WillReturnError(io.EOF).
			RowsWillBeClosed()
		unlock, err := d.Lock(context.Background(), name, time.Minute)
		require.Equal(t, io.EOF, err)
		require.Nil(t, unlock)
	})
}

func TestDriver_UnlockError(t *testing.T) {
	db, m, err := sqlmock.New()
	require.NoError(t, err)
	d := &Driver{conn: &conn{ExecQuerier: db}}
	name, hash := "up", 1551306158
	acquired := func() {
		m.ExpectQuery(sqltest.Escape("SELECT pg_try_advisory_lock($1)")).
			WithArgs(hash).
			WillReturnRows(sqlmock.NewRows([]string{"pg_try_advisory_lock"}).AddRow(1)).
			RowsWillBeClosed()
	}

	t.Run("NotHeld", func(t *testing.T) {
		acquired()
		unlock, err := d.Lock(context.Background(), name, 0)
		require.NoError(t, err)
		m.ExpectQuery(sqltest.Escape("SELECT pg_advisory_unlock($1)")).
			WithArgs(hash).
			WillReturnRows(sqlmock.NewRows([]string{"pg_advisory_unlock"}).AddRow(0)).
			RowsWillBeClosed()
		require.Error(t, unlock())
	})

	t.Run("Internal", func(t *testing.T) {
		acquired()
		unlock, err := d.Lock(context.Background(), name, 0)
		require.NoError(t, err)
		m.ExpectQuery(sqltest.Escape("SELECT pg_advisory_unlock($1)")).
			WithArgs(hash).
			WillReturnRows(sqlmock.NewRows([]string{"pg_advisory_unlock"}).AddRow(nil)).
			RowsWillBeClosed()
		require.Error(t, unlock())
	})
}

func TestDriver_CheckClean(t *testing.T) {
	s := schema.New("test")
	drv := &Driver{Inspector: &mockInspector{schema: s}, conn: &conn{schema: "test"}}
	// Empty schema.
	err := drv.CheckClean(context.Background(), nil)
	require.NoError(t, err)
	// Revisions table found.
	s.AddTables(schema.NewTable("revisions"))
	err = drv.CheckClean(context.Background(), &migrate.TableIdent{Name: "revisions", Schema: "test"})
	require.NoError(t, err)
	// Multiple tables.
	s.Tables = []*schema.Table{schema.NewTable("a"), schema.NewTable("revisions")}
	err = drv.CheckClean(context.Background(), &migrate.TableIdent{Name: "revisions", Schema: "test"})
	require.EqualError(t, err, `sql/migrate: connected database is not clean: found table "a" in schema "test"`)

	r := schema.NewRealm()
	drv.schema = ""
	drv.Inspector = &mockInspector{realm: r}
	// Empty realm.
	err = drv.CheckClean(context.Background(), nil)
	require.NoError(t, err)
	// Revisions table found.
	s.Tables = []*schema.Table{schema.NewTable("revisions").SetSchema(s)}
	r.AddSchemas(s)
	err = drv.CheckClean(context.Background(), &migrate.TableIdent{Name: "revisions", Schema: "test"})
	require.NoError(t, err)
	// Unknown table.
	s.Tables[0].Name = "unknown"
	err = drv.CheckClean(context.Background(), &migrate.TableIdent{Schema: "test", Name: "revisions"})
	require.EqualError(t, err, `sql/migrate: connected database is not clean: found table "unknown" in schema "test"`)
	// Multiple tables.
	s.Tables = []*schema.Table{schema.NewTable("a"), schema.NewTable("revisions")}
	err = drv.CheckClean(context.Background(), &migrate.TableIdent{Schema: "test", Name: "revisions"})
	require.EqualError(t, err, `sql/migrate: connected database is not clean: found 2 tables in schema "test"`)
	// With auto created public schema.
	s.Tables = []*schema.Table{schema.NewTable("revisions")}
	r.AddSchemas(schema.New("public"))
	err = drv.CheckClean(context.Background(), &migrate.TableIdent{Schema: "test", Name: "revisions"})
	require.NoError(t, err)
}

func TestDriver_ExcludeConstraint(t *testing.T) {
	drv, err := Open(pgURL(t, migrate.Snapshot{}))
	require.NoError(t, err)

	// Enable btree_gist extension if needed for test
	_, err = drv.ExecContext(context.Background(), "CREATE EXTENSION IF NOT EXISTS btree_gist")
	require.NoError(t, err)

	undo := func() {
		_, err := drv.ExecContext(context.Background(), "drop table if exists reservations")
		require.NoError(t, err)
	}
	t.Cleanup(undo)
	undo()

	// Create a table with a tsrange column and an EXCLUDE constraint
	reservations := &schema.Table{
		Name: "reservations",
		Columns: []*schema.Column{
			{Name: "id", Type: &schema.ColumnType{Type: &postgres.IntegerType{T: "int"}, Raw: "int"}},
			{Name: "room", Type: &schema.ColumnType{Type: &schema.StringType{T: "text"}, Raw: "text"}},
			{Name: "during", Type: &schema.ColumnType{Type: &postgres.OtherType{T: "tsrange"}, Raw: "tsrange"}},
		},
		Constraints: []schema.Constraint{
			&schema.ExcludeConstraint{
				Name:    "reservations_room_during_excl",
				Index:   "GIST",
				Columns: []string{"room", "during"},
				Ops:     []string{"=", "&&"},
			},
		},
	}

	// Create the table with the EXCLUDE constraint
	err = drv.Create(context.Background(), reservations)
	require.NoError(t, err)

	// Verify the constraint exists
	rows, err := drv.QueryContext(context.Background(),
		"SELECT constraint_name FROM information_schema.table_constraints WHERE table_name = 'reservations' AND constraint_name = 'reservations_room_during_excl'")
	require.NoError(t, err)
	defer rows.Close()

	var found bool
	for rows.Next() {
		var name string
		require.NoError(t, rows.Scan(&name))
		if name == "reservations_room_during_excl" {
			found = true
			break
		}
	}
	require.True(t, found, "EXCLUDE constraint was not created")

	// Test that the constraint works by inserting overlapping records
	_, err = drv.ExecContext(context.Background(),
		"INSERT INTO reservations (id, room, during) VALUES (1, '123A', '[2023-01-01 14:00, 2023-01-01 15:00)')")
	require.NoError(t, err)

	// This should fail due to the EXCLUDE constraint
	_, err = drv.ExecContext(context.Background(),
		"INSERT INTO reservations (id, room, during) VALUES (2, '123A', '[2023-01-01 14:30, 2023-01-01 15:30)')")
	require.Error(t, err)
	require.Contains(t, err.Error(), "violates exclusion constraint")

	// This should succeed (different room)
	_, err = drv.ExecContext(context.Background(),
		"INSERT INTO reservations (id, room, during) VALUES (3, '123B', '[2023-01-01 14:30, 2023-01-01 15:30)')")
	require.NoError(t, err)
}

func TestDriver_Version(t *testing.T) {
	db, m, err := sqlmock.New()
	require.NoError(t, err)
	mock{m}.version("130000")
	drv, err := Open(db)
	require.NoError(t, err)

	type vr interface{ Version() string }
	require.Implements(t, (*vr)(nil), drv)
	require.Equal(t, "130000", drv.(vr).Version())
}

func TestDriver_RealmRestoreFunc(t *testing.T) {
	var (
		apply   = &mockPlanApplier{}
		inspect = &mockInspector{}
		drv     = &Driver{
			Inspector:   inspect,
			Differ:      DefaultDiff,
			conn:        &conn{schema: "test"},
			PlanApplier: apply,
		}
	)
	f := drv.RealmRestoreFunc(schema.NewRealm().AddSchemas(schema.New("public")))

	// No changes.
	inspect.realm = schema.NewRealm().AddSchemas(schema.New("public"))
	err := f(context.Background())
	require.NoError(t, err)
	require.Empty(t, apply.applied)

	// Schema changes.
	inspect.realm = schema.NewRealm().AddSchemas(schema.New("public").AddTables(schema.NewTable("t1")))
	err = f(context.Background())
	require.NoError(t, err)
	require.Len(t, apply.applied, 2)
	drop, ok := apply.applied[0].(*schema.DropSchema)
	require.True(t, ok)
	require.Equal(t, "public", drop.S.Name)
	create, ok := apply.applied[1].(*schema.AddSchema)
	require.True(t, ok)
	require.Equal(t, "public", create.S.Name)

	// Recreate the public schema.
	apply.applied = nil
	inspect.realm = schema.NewRealm().AddSchemas(schema.New("test").AddTables(schema.NewTable("t1")))
	err = f(context.Background())
	require.NoError(t, err)
	require.Len(t, apply.applied, 2)
	drop, ok = apply.applied[0].(*schema.DropSchema)
	require.True(t, ok)
	require.Equal(t, "test", drop.S.Name)
	create, ok = apply.applied[1].(*schema.AddSchema)
	require.True(t, ok)
	require.Equal(t, "public", create.S.Name)

	// Non-public changes.
	apply.applied = nil
	f = drv.RealmRestoreFunc(schema.NewRealm().AddSchemas(schema.New("test")))
	inspect.realm = schema.NewRealm().AddSchemas(schema.New("test").AddTables(schema.NewTable("t1")))
	err = f(context.Background())
	require.NoError(t, err)
	require.Len(t, apply.applied, 1)
	dropT, ok := apply.applied[0].(*schema.DropTable)
	require.True(t, ok)
	require.Equal(t, "t1", dropT.T.Name)
}

type mockInspector struct {
	schema.Inspector
	realm  *schema.Realm
	schema *schema.Schema
}

func (m *mockInspector) InspectSchema(context.Context, string, *schema.InspectOptions) (*schema.Schema, error) {
	if m.schema == nil {
		return nil, &schema.NotExistError{}
	}
	return m.schema, nil
}

func (m *mockInspector) InspectRealm(context.Context, *schema.InspectRealmOption) (*schema.Realm, error) {
	return m.realm, nil
}

type mockPlanApplier struct {
	planned []schema.Change
	applied []schema.Change
}

func (m *mockPlanApplier) PlanChanges(_ context.Context, _ string, planned []schema.Change, _ ...migrate.PlanOption) (*migrate.Plan, error) {
	m.planned = append(m.planned, planned...)
	return nil, nil
}

func (m *mockPlanApplier) ApplyChanges(_ context.Context, applied []schema.Change, _ ...migrate.PlanOption) error {
	m.applied = append(m.applied, applied...)
	return nil
}

type mockExecQuerier struct {
	lastQuery string
}

func (m *mockExecQuerier) ExecContext(ctx context.Context, query string, args ...any) (any, error) {
	m.lastQuery = query
	return nil, nil
}

func (m *mockExecQuerier) QueryContext(ctx context.Context, query string, args ...any) (any, error) {
	return nil, nil
}

func TestAddExcludeConstraint(t *testing.T) {
	mock := &mockExecQuerier{}
	drv := &driver{conn: &conn{ExecQuerier: mock}}
	table := &schema.Table{Name: "events"}
	constraint := &schema.ExcludeConstraint{
		Name: "exclude_events",
		Elements: []*schema.ExcludeElement{
			{Column: "room", Op: "="},
			{Column: "during", Op: "&&"},
		},
		Attrs: []schema.Attr{
			&schema.IndexType{T: "GIST"},
		},
	}
	err := drv.AddExcludeConstraint(context.Background(), table, constraint)
	if err != nil {
		t.Fatalf("failed to add exclude constraint: %v", err)
	}
	expected := `ALTER TABLE "events" ADD CONSTRAINT "exclude_events" EXCLUDE USING GIST ("room" WITH =, "during" WITH &&)`
	if mock.lastQuery != expected {
		t.Errorf("unexpected query:\ngot:  %s\nwant: %s", mock.lastQuery, expected)
	}
}
