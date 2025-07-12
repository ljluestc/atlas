// Copyright 2021-present The Atlas Authors. All rights reserved.
// This source code is licensed under the Apache 2.0 license found
// in the LICENSE file in the root directory of this source tree.

package schema

import (
	"context"
	"database/sql"
	"errors"
)

// A NotExistError wraps another error to retain its original text
// but makes it possible to the migrator to catch it.
type NotExistError struct {
	Err error
}

func (e NotExistError) Error() string { return e.Err.Error() }

// IsNotExistError reports if an error is a NotExistError.
func IsNotExistError(err error) bool {
	if err == nil {
		return false
	}
	var e *NotExistError
	return errors.As(err, &e)
}

// ExecQuerier wraps the two standard sql.DB methods.
type ExecQuerier interface {
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

// InspectMode and related types are now declared in interface.go
// Comment out these declarations here

/*
	// InspectMode defines inspection behavior.
	type InspectMode uint

	// Inspection mode options.
	const (
		InspectSchemas InspectMode = 1 << iota
		InspectTables
		InspectTableColumns
		InspectTableIndexes
		InspectTableForeignKeys
	)

	// InspectOptions holds schema inspection options.
	type InspectOptions struct {
		// Mode defines the inspection behavior.
		Mode InspectMode
		// Tables filters inspection by table names.
		Tables []string
	}

	// InspectRealmOption allows configuring RealmInspectOptions.
	type InspectRealmOption func(*RealmInspectOptions)
*/

// Normalizer is the interface implemented by the different database drivers for
// "normalizing" schema objects. i.e. converting schema objects defined in natural
// form to their representation in the database. Thus, two schema objects are equal
// if their normal forms are equal.
type Normalizer interface {
	// NormalizeSchema returns the normal representation of a schema.
	NormalizeSchema(context.Context, *Schema) (*Schema, error)

	// NormalizeRealm returns the normal representation of a database.
	NormalizeRealm(context.Context, *Realm) (*Realm, error)
}
