// Copyright 2021-present The Atlas Authors. All rights reserved.
// This source code is licensed under the Apache 2.0 license found
// in the LICENSE file in the root directory of this source tree.

package schema

import (
	"context"
	"errors"
	"reflect"
	"time"
)

// List of diff modes.
const (
	DiffModeUnset         DiffMode = 1 << iota // Default, backwards compatability.
	DiffModeNotNormalized                      // Diff objects are considered to be in not normalized state.
	DiffModeNormalized                         // Diff objects are considered to be in normalized state.
	DiffModeSkipInvalid                        // Invalid changes are skipped, instead of returning an error.
)

// ChangeKind is a bit mask for different types of changes.
const (
	// NoChange is already defined in interface.go, so we don't redefine it here
	ChangeUnique ChangeKind = 1 << iota
	ChangeParts
	ChangeAttr
	ChangeComment
	ChangeRefTable
	ChangeRefColumn
	ChangeColumn
	ChangeUpdateAction
	ChangeDeleteAction
)

// Is reports whether c is match the given change kind.
func (k ChangeKind) Is(c ChangeKind) bool {
	return k == c || k&c != 0
}

type (
	// Differ is the interface implemented by the different
	// drivers for comparing and diffing schema top elements.
	Differ interface {
		// RealmDiff returns a diff report for migrating a realm
		// (or a database) from state "from" to state "to". An error
		// is returned if such step is not possible.
		RealmDiff(from, to *Realm, opts ...DiffOption) ([]Change, error)

		// SchemaDiff returns a diff report for migrating a schema
		// from state "from" to state "to". An error is returned
		// if such step is not possible.
		SchemaDiff(from, to *Schema, opts ...DiffOption) ([]Change, error)

		// TableDiff returns a diff report for migrating a table
		// from state "from" to state "to". An error is returned
		// if such step is not possible.
		TableDiff(from, to *Table, opts ...DiffOption) ([]Change, error)
	}

	// DiffMode defines the diffing mode, e.g. if objects are normalized or not.
	DiffMode uint8

	// DiffOptions and DiffOption are already defined in interface.go, so we don't redefine them here
)

// Is reports whether m is match the given mode.
func (m DiffMode) Is(m1 DiffMode) bool {
	return m == m1 || m&m1 != 0
}

// NewDiffOptions creates a new DiffOptions from the given configuration.
func NewDiffOptions(opts ...DiffOption) *DiffOptions {
	o := &DiffOptions{}
	for _, opt := range opts {
		opt(o)
	}
	return o
}

// DiffSkipChanges returns a DiffOption that skips the given change types.
// For example, in order to skip all destructive changes, use:
//
//	DiffSkipChanges(&DropSchema{}, &DropTable{}, &DropColumn{}, &DropIndex{}, &DropForeignKey{})
func DiffSkipChanges(changes ...Change) DiffOption {
	return func(o *DiffOptions) {
		o.SkipChanges = append(o.SkipChanges, changes...)
	}
}

// DiffNormalized returns a DiffOption that sets DiffMode to DiffModeNormalized,
// indicating the Differ should consider input objects as normalized, For example:
//
//	DiffNormalized()
func DiffNormalized() DiffOption {
	return func(o *DiffOptions) {
		o.Mode = DiffModeNormalized
	}
}

// Skipped reports whether the given change should be skipped.
func (o *DiffOptions) Skipped(c Change) bool {
	for _, s := range o.SkipChanges {
		if reflect.TypeOf(c) == reflect.TypeOf(s) {
			return true
		}
	}
	return false
}

// AddOrSkip adds the given change to the list of changes if it is not skipped.
func (o *DiffOptions) AddOrSkip(changes Changes, cs ...Change) Changes {
	for _, c := range cs {
		if !o.Skipped(c) {
			changes = append(changes, c)
		}
	}
	return changes

}

// ErrLocked is returned on Lock calls which have failed to obtain the lock.
var ErrLocked = errors.New("sql/schema: lock is held by other session")

type (
	// UnlockFunc is returned by the Locker to explicitly
	// release the named "advisory lock".
	UnlockFunc func() error

	// Locker is an interface that is optionally implemented by the different drivers
	// for obtaining an "advisory lock" with the given name.
	Locker interface {
		// Lock acquires a named "advisory lock", using the given timeout. Negative value means no timeout,
		// and the zero value means a "try lock" mode. i.e. return immediately if the lock is already taken.
		// The returned unlock function is used to release the advisory lock acquired by the session.
		//
		// An ErrLocked is returned if the operation failed to obtain the lock in all different timeout modes.
		Lock(ctx context.Context, name string, timeout time.Duration) (UnlockFunc, error)
	}
)

type (
	// Changes is a list of changes allow for searching and mutating changes.
	Changes []Change

	// ChangeDepender wraps the ChangeDeps method, which returns a dependency map
	// from each change to its dependent changes. This interface can optionally
	// be implemented by drivers.
	ChangeDepender interface {
		ChangeDeps([]Change) map[Change][]Change
	}
)

// IndexAddTable returns the index of the first AddTable in the changes
// with the given name, or -1 if there is no such change in the Changes.
func (c Changes) IndexAddTable(name string) int {
	return c.search(func(c Change) bool {
		a, ok := c.(*AddTable)
		return ok && a.T.Name == name
	})
}

// IndexDropTable returns the index of the first DropTable in the changes
// with the given name, or -1 if there is no such change in the Changes.
func (c Changes) IndexDropTable(name string) int {
	return c.search(func(c Change) bool {
		a, ok := c.(*DropTable)
		return ok && a.T.Name == name
	})
}

// LastIndexAddTable returns the index of the last AddTable in the changes
// with the given name, or -1 if there is no such change in the Changes.
func (c Changes) LastIndexAddTable(name string) int {
	return c.rsearch(func(c Change) bool {
		a, ok := c.(*AddTable)
		return ok && a.T.Name == name
	})
}

// LastIndexDropTable returns the index of the last DropTable in the changes
// with the given name, or -1 if there is no such change in the Changes.
func (c Changes) LastIndexDropTable(name string) int {
	return c.rsearch(func(c Change) bool {
		a, ok := c.(*DropTable)
		return ok && a.T.Name == name
	})
}

// IndexAddColumn returns the index of the first AddColumn in the changes
// with the given name, or -1 if there is no such change in the Changes.
func (c Changes) IndexAddColumn(name string) int {
	return c.search(func(c Change) bool {
		a, ok := c.(*AddColumn)
		return ok && a.C.Name == name
	})
}

// IndexDropColumn returns the index of the first DropColumn in the changes
// with the given name, or -1 if there is no such change in the Changes.
func (c Changes) IndexDropColumn(name string) int {
	return c.search(func(c Change) bool {
		d, ok := c.(*DropColumn)
		return ok && d.C.Name == name
	})
}

// IndexModifyColumn returns the index of the first ModifyColumn in the changes
// with the given name, or -1 if there is no such change in the Changes.
func (c Changes) IndexModifyColumn(name string) int {
	return c.search(func(c Change) bool {
		a, ok := c.(*ModifyColumn)
		return ok && a.From.Name == name
	})
}

// IndexAddIndex returns the index of the first AddIndex in the changes
// with the given name, or -1 if there is no such change in the Changes.
func (c Changes) IndexAddIndex(name string) int {
	return c.search(func(c Change) bool {
		a, ok := c.(*AddIndex)
		return ok && a.I.Name == name
	})
}

// IndexDropIndex returns the index of the first DropIndex in the changes
// with the given name, or -1 if there is no such change in the Changes.
func (c Changes) IndexDropIndex(name string) int {
	return c.search(func(c Change) bool {
		a, ok := c.(*DropIndex)
		return ok && a.I.Name == name
	})
}

// RemoveIndex removes elements in the given indexes from the Changes.
func (c *Changes) RemoveIndex(indexes ...int) {
	changes := make([]Change, 0, len(*c)-len(indexes))
Loop:
	for i := range *c {
		for _, idx := range indexes {
			if i == idx {
				continue Loop
			}
		}
		changes = append(changes, (*c)[i])
	}
	*c = changes
}

// search returns the index of the first call to f that returns true, or -1.
func (c Changes) search(f func(Change) bool) int {
	for i := range c {
		if f(c[i]) {
			return i
		}
	}
	return -1
}

// rsearch is the reversed version of search. It returns the
// index of the last call to f that returns true, or -1.
func (c Changes) rsearch(f func(Change) bool) int {
	for i := len(c) - 1; i >= 0; i-- {
		if f(c[i]) {
			return i
		}
	}
	return -1
}

// Define the required constants
const (
	IfExists    = "IF EXISTS"
	IfNotExists = "IF NOT EXISTS"
)

// Define the proper types for IF EXISTS and IF NOT EXISTS clauses
type IfExistsType string
type IfNotExistsType string

func (IfExistsType) clause()    {}
func (IfNotExistsType) clause() {}
