// Copyright 2021-present The Atlas Authors. All rights reserved.
// This source code is licensed under the Apache 2.0 license found
// in the LICENSE file in the root directory of this source tree.

package migrate

import (
	"context"
	"io/fs"

	"ariga.io/atlas/sql/schema"
)

// mockDriver is a mock implementation of the Driver interface.
type mockDriver struct {
	MigrateFunc func(context.Context, []*Statement) error
	MigrateErr  error
	InspectFunc func(context.Context) (*schema.Realm, error)
	InspectErr  error
}

func (m *mockDriver) Migrate(ctx context.Context, stmts []*Statement) error {
	if m.MigrateFunc != nil {
		return m.MigrateFunc(ctx, stmts)
	}
	return m.MigrateErr
}

func (m *mockDriver) Inspect(ctx context.Context) (*schema.Realm, error) {
	if m.InspectFunc != nil {
		return m.InspectFunc(ctx)
	}
	return nil, m.InspectErr
}

// MockMigrateDriver is a mock implementation of the migrate driver.
type MockMigrateDriver struct {
	mockDriver
}

func (m *MockMigrateDriver) Lock(context.Context, string) (schema.ReleaseLockFunc, error) {
	return func() error { return nil }, nil
}

// mockRevisionReadWriter is a mock implementation of the RevisionReadWriter interface.
type mockRevisionReadWriter struct {
	GetFunc       func(context.Context) (*Revision, error)
	GetErr        error
	WriteFunc     func(context.Context, *Revision) error
	WriteErr      error
	ReadVersions  []string
	WriteVersions []string
}

func (m *mockRevisionReadWriter) Revision(ctx context.Context) (*Revision, error) {
	if m.GetFunc != nil {
		return m.GetFunc(ctx)
	}
	return nil, m.GetErr
}

func (m *mockRevisionReadWriter) WriteRevision(ctx context.Context, rev *Revision) error {
	if m.WriteFunc != nil {
		return m.WriteFunc(ctx, rev)
	}
	return m.WriteErr
}

func (m *mockRevisionReadWriter) ReadDir() ([]interface{}, error) {
	entries := make([]interface{}, len(m.ReadVersions))
	for i, v := range m.ReadVersions {
		entries[i] = dirEntry(v)
	}
	return entries, nil
}

type dirEntry string

func (d dirEntry) Name() string               { return string(d) }
func (d dirEntry) IsDir() bool                { return false }
func (d dirEntry) Type() fs.FileMode          { return 0 }
func (d dirEntry) Info() (fs.FileInfo, error) { return nil, nil }
