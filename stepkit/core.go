package main

import (
	"context"
	"fmt"
)

// Context sharing between steps lives here
type PipelineContext struct {
}

type Step interface {
	Name() string
	Run(ctx context.Context, pc *PipelineContext) error
}

type Pipeline struct {
	steps []Step
}

func (p *Pipeline) Run(ctx context.Context, pc *PipelineContext) error {
	for _, step := range p.steps {
		if err := step.Run(ctx, pc); err != nil {
			return fmt.Errorf("step %s failed: %w", step.Name(), err)
		}
	}
	return nil
}
