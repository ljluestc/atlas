// Copyright 2021-present The Atlas Authors. All rights reserved.
// This source code is licensed under the Apache 2.0 license found
// in the LICENSE file in the root directory of this source tree.

package migrate

// PlanOption allows configuring Plan using functional options.
type PlanOption interface {
	apply(*planOptions)
}

// planOptions holds the options for Plan.
type planOptions struct {
	mode    PlanMode
	schemas []string
}

// planOption implements PlanOption.
type planOption struct {
	applyFunc func(*planOptions)
}

// apply implements PlanOption.
func (p planOption) apply(opts *planOptions) {
	p.applyFunc(opts)
}

// WithPlanMode returns a PlanOption that sets the plan mode.
func WithPlanMode(m PlanMode) PlanOption {
	return planOption{
		applyFunc: func(opts *planOptions) {
			opts.mode = m
		},
	}
}

// PlanApplier applies a migration plan to a database.
type PlanApplier struct {
	stateReader StateReader
	conn        ExecQuerier
	dir         Dir
}

// NewPlanApplier creates a new PlanApplier.
func NewPlanApplier(sr StateReader, conn ExecQuerier, dir Dir) *PlanApplier {
	return &PlanApplier{
		stateReader: sr,
		conn:        conn,
		dir:         dir,
	}
}
