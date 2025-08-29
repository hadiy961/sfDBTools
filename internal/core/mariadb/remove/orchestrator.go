package remove

import (
	"context"
	"fmt"
	"sfDBTools/internal/logger"
	"sfDBTools/utils/system"
	"sfDBTools/utils/terminal"
)

// Dependencies holds all external dependencies for the orchestrator
type Dependencies struct {
	PackageManager system.PackageManager
	ServiceManager system.ServiceManager
	FileSystem     system.FileSystem
}

// State holds the current state of the removal process
type State struct {
	Config       *RemovalConfig
	Installation *DetectedInstallation
	BackupPath   string
	RollbackData map[string]interface{} // For storing rollback information
}

// Step interface defines a step in the removal pipeline
type Step interface {
	Name() string
	Execute(ctx context.Context, state *State) error
	Rollback(ctx context.Context, state *State) error
	Validate(state *State) error
}

// Pipeline orchestrates the execution of removal steps
type Pipeline struct {
	steps        []Step
	dependencies Dependencies
}

// Orchestrator coordinates the entire removal process
type Orchestrator struct {
	dependencies Dependencies
	pipeline     *Pipeline
}

// NewOrchestrator creates a new removal orchestrator
func NewOrchestrator(deps Dependencies) *Orchestrator {
	pipeline := &Pipeline{
		dependencies: deps,
		steps: []Step{
			&DetectionStep{deps},
			&ValidationStep{deps},
			&BackupStep{deps},
			&ServiceStopStep{deps},
			&PackageRemovalStep{deps},
			&DataCleanupStep{deps},
		},
	}

	return &Orchestrator{
		dependencies: deps,
		pipeline:     pipeline,
	}
}

// Execute runs the complete removal process
func (o *Orchestrator) Execute(ctx context.Context, config *RemovalConfig) error {
	lg, _ := logger.Get()
	lg.Info("Starting MariaDB removal orchestration")

	state := &State{
		Config:       config,
		RollbackData: make(map[string]interface{}),
	}

	// Execute pipeline
	if err := o.pipeline.Execute(ctx, state); err != nil {
		return fmt.Errorf("removal pipeline failed: %w", err)
	}

	lg.Info("MariaDB removal orchestration completed successfully")
	return nil
}

// Execute runs all steps in the pipeline
func (p *Pipeline) Execute(ctx context.Context, state *State) error {
	lg, _ := logger.Get()

	for i, step := range p.steps {
		lg.Info("Executing step", logger.String("step", step.Name()), logger.Int("index", i+1), logger.Int("total", len(p.steps)))

		// Show progress
		terminal.PrintInfo(fmt.Sprintf("Step %d/%d: %s", i+1, len(p.steps), step.Name()))

		// Validate step preconditions
		if err := step.Validate(state); err != nil {
			return fmt.Errorf("step %s validation failed: %w", step.Name(), err)
		}

		// Execute step
		if err := step.Execute(ctx, state); err != nil {
			lg.Error("Step execution failed", logger.String("step", step.Name()), logger.Error(err))

			// Attempt rollback
			rollbackErr := p.rollback(ctx, state, i-1)
			if rollbackErr != nil {
				lg.Error("Rollback failed", logger.Error(rollbackErr))
				return fmt.Errorf("step %s failed: %w (rollback also failed: %v)", step.Name(), err, rollbackErr)
			}

			return fmt.Errorf("step %s failed: %w", step.Name(), err)
		}

		lg.Info("Step completed successfully", logger.String("step", step.Name()))
	}

	return nil
}

// rollback attempts to rollback executed steps in reverse order
func (p *Pipeline) rollback(ctx context.Context, state *State, lastExecutedIndex int) error {
	lg, _ := logger.Get()
	lg.Info("Starting rollback process", logger.Int("lastExecutedIndex", lastExecutedIndex))

	for i := lastExecutedIndex; i >= 0; i-- {
		step := p.steps[i]
		lg.Info("Rolling back step", logger.String("step", step.Name()))

		if err := step.Rollback(ctx, state); err != nil {
			lg.Error("Step rollback failed", logger.String("step", step.Name()), logger.Error(err))
			// Continue with other rollbacks even if one fails
		}
	}

	return nil
}
