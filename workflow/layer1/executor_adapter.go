package layer1

import (
        "context"
        "fmt"
        "time"

        "github.com/ubom/workflow/layer0"
        "github.com/ubom/workflow/executors"
)

// ExecutorAdapter adapts the new unified WorkExecutor interface to the legacy layer1 interface
type ExecutorAdapter struct {
        unified executors.WorkExecutor
}

// NewExecutorAdapter creates a new adapter for backward compatibility
func NewExecutorAdapter(unified executors.WorkExecutor) WorkExecutor {
        return &ExecutorAdapter{unified: unified}
}

// Execute adapts the unified interface to the legacy interface
func (a *ExecutorAdapter) Execute(work layer0.Work, workContext *layer0.Context) (interface{}, error) {
        result, err := a.unified.Execute(context.Background(), work, workContext)
        if err != nil {
                return nil, err
        }
        
        if !result.Success {
                return nil, fmt.Errorf(result.Error)
        }
        
        // Return the main result or all outputs
        if mainResult, exists := result.Outputs["result"]; exists {
                return mainResult, nil
        }
        
        return result.Outputs, nil
}

// CanExecute delegates to the unified executor
func (a *ExecutorAdapter) CanExecute(workType layer0.WorkType) bool {
        return a.unified.CanExecute(workType)
}

// GetSupportedTypes delegates to the unified executor
func (a *ExecutorAdapter) GetSupportedTypes() []layer0.WorkType {
        return a.unified.GetSupportedTypes()
}

// UnifiedExecutorAdapter adapts the legacy layer1 interface to the new unified interface
type UnifiedExecutorAdapter struct {
        *executors.BaseExecutor
        legacy WorkExecutor
}

// NewUnifiedExecutorAdapter creates an adapter from legacy to unified interface
func NewUnifiedExecutorAdapter(legacy WorkExecutor, name, version, author, description string) executors.WorkExecutor {
        baseExecutor := executors.NewBaseExecutor(
                name,
                version,
                author,
                description,
                legacy.GetSupportedTypes(),
        )
        
        return &UnifiedExecutorAdapter{
                BaseExecutor: baseExecutor,
                legacy:       legacy,
        }
}

// Execute adapts the legacy interface to the unified interface
func (a *UnifiedExecutorAdapter) Execute(ctx context.Context, work layer0.Work, workContext *layer0.Context) (executors.WorkResult, error) {
        startTime := time.Now()
        
        output, err := a.legacy.Execute(work, workContext)
        endTime := time.Now()
        
        result := executors.WorkResult{
                Logs: []executors.LogEntry{
                        {
                                Timestamp: startTime,
                                Level:     "INFO",
                                Message:   "Legacy executor execution started",
                                Source:    "legacy-adapter",
                        },
                },
                Metrics: executors.ExecutionMetrics{
                        StartTime: startTime,
                        EndTime:   endTime,
                        Duration:  endTime.Sub(startTime),
                },
        }
        
        if err != nil {
                result.Success = false
                result.Error = err.Error()
                result.Logs = append(result.Logs, executors.LogEntry{
                        Timestamp: endTime,
                        Level:     "ERROR",
                        Message:   "Legacy executor execution failed: " + err.Error(),
                        Source:    "legacy-adapter",
                })
                return result, err
        }
        
        result.Success = true
        result.Outputs = map[string]interface{}{
                "result": output,
        }
        result.Logs = append(result.Logs, executors.LogEntry{
                Timestamp: endTime,
                Level:     "INFO",
                Message:   "Legacy executor execution completed successfully",
                Source:    "legacy-adapter",
        })
        
        return result, nil
}

// Validate provides basic validation (legacy executors don't have this method)
func (a *UnifiedExecutorAdapter) Validate(work layer0.Work) error {
        // Basic validation - check if executor can handle the work type
        if !a.legacy.CanExecute(work.GetType()) {
                return fmt.Errorf("executor cannot handle work type %s", work.GetType())
        }
        return nil
}

// GetSchema provides a basic schema (legacy executors don't have this method)
func (a *UnifiedExecutorAdapter) GetSchema() executors.WorkSchema {
        return executors.WorkSchema{
                JSONSchema:    "{}",
                Examples:      []executors.WorkDefinition{},
                Documentation: "Legacy executor - no schema available",
        }
}
