// Copyright 2021-present The Atlas Authors. All rights reserved.
// This source code is licensed under the Apache 2.0 license found
// in the LICENSE file in the root directory of this source tree.

package migrate

import (
	"context"
	"fmt"
)

// Stmt represents a SQL statement in a migration file
type Stmt struct {
	Text string
	// Additional statement information like comments and directives
	Comments []string
	// Optional: Line number in the file
	Line int
}

// fileStmts reads and parses the migration file into SQL statements
func (e *Executor) fileStmts(m File) ([]*Stmt, error) {
	// This is a stub implementation that would normally parse the file
	// and return the statements within
	content, err := e.dir.ReadFile(m.Name())
	if err != nil {
		return nil, fmt.Errorf("reading file: %w", err)
	}
	
	// A real implementation would parse the content into separate statements
	// For now, we just return a single statement with the whole content
	return []*Stmt{
		{
			Text: string(content),
			Line: 1,
		},
	}, nil
}

// fileChecks executes any pre-migration checks defined in the file
func (e *Executor) fileChecks(ctx context.Context, m File, r *Revision) error {
	// This is a stub implementation
	// In a real implementation, it would look for and execute check statements
	return nil
}

// ValidateDir validates that the migration directory is in a consistent state
func (e *Executor) ValidateDir(ctx context.Context) error {
	// This is a stub implementation
	// In a real implementation, it would check for file integrity, etc.
	return nil
}

// SkipCheckpointFiles filters out checkpoint files from a list of files
func SkipCheckpointFiles(files []File) []File {
	result := make([]File, 0, len(files))
	for _, f := range files {
		if cf, ok := f.(CheckpointFile); !ok || !cf.IsCheckpoint() {
			result = append(result, f)
		}
	}
	return result
}

// FilesLastIndex returns the index of the last file in the slice that matches the predicate
func FilesLastIndex(files []File, predicate func(File) bool) int {
	for i := len(files) - 1; i >= 0; i-- {
		if predicate(files[i]) {
			return i
		}
	}
	return -1
}

// FilesFromLastCheckpoint returns all files after the last checkpoint file
func FilesFromLastCheckpoint(dir Dir) ([]File, error) {
	files, err := dir.Files()
	if err != nil {
		return nil, fmt.Errorf("reading migration directory files: %w", err)
	}
	
	// Find the last checkpoint file
	lastCheckpoint := -1
	for i, f := range files {
		if cf, ok := f.(CheckpointFile); ok && cf.IsCheckpoint() {
			lastCheckpoint = i
		}
	}
	
	// If no checkpoint was found, return all files
	if lastCheckpoint == -1 {
		return SkipCheckpointFiles(files), nil
	}
	
	// Return all files after the last checkpoint
	return SkipCheckpointFiles(files[lastCheckpoint+1:]), nil
}
