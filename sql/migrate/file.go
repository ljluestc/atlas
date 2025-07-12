// Copyright 2021-present The Atlas Authors. All rights reserved.
// This source code is licensed under the Apache 2.0 license found
// in the LICENSE file in the root directory of this source tree.

package migrate

// File represents a migration file.
type File interface {
	// Name returns the name of the migration file.
	Name() string
	// Version returns the version of the migration.
	Version() string
	// Desc returns the description of the migration.
	Desc() string
}

// CheckpointFile is a File that also indicates if it's a checkpoint.
type CheckpointFile interface {
	File
	// IsCheckpoint returns if the file represents a checkpoint migration.
	IsCheckpoint() bool
}

// Dir is an interface for interacting with a migration directory.
type Dir interface {
	// Files returns all migration files in the directory.
	Files() ([]File, error)
	// WriteFile writes a file to the directory.
	WriteFile(name string, data []byte) error
	// ReadFile reads a file from the directory.
	ReadFile(name string) ([]byte, error)
	// Checksum returns a Checksummer for the directory.
	Checksum() (Checksummer, error)
}

// CheckpointDir is a Dir that also supports checkpoint operations.
type CheckpointDir interface {
	Dir
	// WriteCheckpoint writes a checkpoint file to the directory.
	WriteCheckpoint(name, tag string, data []byte) error
}

// Checksummer computes checksums for migration files.
type Checksummer interface {
	// SumByName returns the checksum for a file by name.
	SumByName(name string) (string, error)
}

// WriteSumFile writes a checksum file to the directory.
func WriteSumFile(dir Dir, sum Checksummer) error {
	// This is a stub implementation
	return nil
}

// Formatter formats a Plan into migration files.
type Formatter interface {
	// Format formats a Plan into migration files.
	Format(plan *Plan) ([]File, error)
}

// DefaultFormatter is the default Formatter.
var DefaultFormatter Formatter = &defaultFormatter{}

// defaultFormatter is the default implementation of Formatter.
type defaultFormatter struct{}

// Format implements Formatter.
func (d *defaultFormatter) Format(plan *Plan) ([]File, error) {
	// This is a stub implementation
	return nil, nil
}
