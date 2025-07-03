package layer1

import (
	"fmt"
	"sync"
	"time"

	"github.com/ubom/workflow/layer0"
)

// WorkExecutor defines the interface for executing work
type WorkExecutor interface {
	Execute(work layer0.Work, context *layer0.Context) (interface{}, error)
	CanExecute(workType layer0.WorkType) bool
	GetSupportedTypes() []layer0.WorkType
}

// WorkExecutionResult represents the result of work execution
type WorkExecutionResult struct {
	WorkID      layer0.WorkID     `json:"work_id"`
	Status      layer0.WorkStatus `json:"status"`
	Output      interface{}       `json:"output"`
	Error       string            `json:"error,omitempty"`
	StartedAt   time.Time         `json:"started_at"`
	CompletedAt *time.Time        `json:"completed_at,omitempty"`
	Duration    time.Duration     `json:"duration"`
}

// WorkExecutionCore provides core work execution functionality
type WorkExecutionCore struct {
	executors        map[layer0.WorkType]WorkExecutor
	activeWork       map[layer0.WorkID]layer0.Work
	executionResults map[layer0.WorkID]WorkExecutionResult
	mutex            sync.RWMutex
}

// WorkExecutionCoreInterface defines the contract for work execution operations
type WorkExecutionCoreInterface interface {
	RegisterExecutor(workType layer0.WorkType, executor WorkExecutor) error
	UnregisterExecutor(workType layer0.WorkType) error
	GetExecutor(workType layer0.WorkType) (WorkExecutor, error)
	GetSupportedWorkTypes() []layer0.WorkType
	ExecuteWork(work layer0.Work, context *layer0.Context) (WorkExecutionResult, error)
	GetActiveWork() []layer0.Work
	GetExecutionResult(workID layer0.WorkID) (WorkExecutionResult, error)
	GetAllExecutionResults() []WorkExecutionResult
	CancelWork(workID layer0.WorkID) error
	IsWorkActive(workID layer0.WorkID) bool
}

// NewWorkExecutionCore creates a new work execution core
func NewWorkExecutionCore() *WorkExecutionCore {
	return &WorkExecutionCore{
		executors:        make(map[layer0.WorkType]WorkExecutor),
		activeWork:       make(map[layer0.WorkID]layer0.Work),
		executionResults: make(map[layer0.WorkID]WorkExecutionResult),
		mutex:            sync.RWMutex{},
	}
}

// RegisterExecutor registers a work executor for a specific work type
func (wec *WorkExecutionCore) RegisterExecutor(workType layer0.WorkType, executor WorkExecutor) error {
	if executor == nil {
		return fmt.Errorf("executor cannot be nil")
	}

	wec.mutex.Lock()
	defer wec.mutex.Unlock()

	if _, exists := wec.executors[workType]; exists {
		return fmt.Errorf("executor for work type %s already registered", workType)
	}

	wec.executors[workType] = executor
	return nil
}

// UnregisterExecutor unregisters a work executor for a specific work type
func (wec *WorkExecutionCore) UnregisterExecutor(workType layer0.WorkType) error {
	wec.mutex.Lock()
	defer wec.mutex.Unlock()

	if _, exists := wec.executors[workType]; !exists {
		return fmt.Errorf("no executor registered for work type %s", workType)
	}

	delete(wec.executors, workType)
	return nil
}

// GetExecutor retrieves the executor for a specific work type
func (wec *WorkExecutionCore) GetExecutor(workType layer0.WorkType) (WorkExecutor, error) {
	wec.mutex.RLock()
	defer wec.mutex.RUnlock()

	executor, exists := wec.executors[workType]
	if !exists {
		return nil, fmt.Errorf("no executor registered for work type %s", workType)
	}

	return executor, nil
}

// GetSupportedWorkTypes returns all supported work types
func (wec *WorkExecutionCore) GetSupportedWorkTypes() []layer0.WorkType {
	wec.mutex.RLock()
	defer wec.mutex.RUnlock()

	types := make([]layer0.WorkType, 0, len(wec.executors))
	for workType := range wec.executors {
		types = append(types, workType)
	}

	return types
}

// ExecuteWork executes a work item using the appropriate executor
func (wec *WorkExecutionCore) ExecuteWork(work layer0.Work, context *layer0.Context) (WorkExecutionResult, error) {
	if err := work.Validate(); err != nil {
		return WorkExecutionResult{}, fmt.Errorf("invalid work: %w", err)
	}

	if !work.IsExecutable() {
		return WorkExecutionResult{}, fmt.Errorf("work %s is not executable (status: %s)", work.GetID(), work.GetStatus())
	}

	wec.mutex.Lock()

	// Check if work is already active
	if _, isActive := wec.activeWork[work.GetID()]; isActive {
		wec.mutex.Unlock()
		return WorkExecutionResult{}, fmt.Errorf("work %s is already being executed", work.GetID())
	}

	// Get executor
	executor, exists := wec.executors[work.GetType()]
	if !exists {
		wec.mutex.Unlock()
		return WorkExecutionResult{}, fmt.Errorf("no executor registered for work type %s", work.GetType())
	}

	// Mark work as active
	startedWork := work.MarkStarted()
	wec.activeWork[work.GetID()] = startedWork
	wec.mutex.Unlock()

	// Create initial result
	startTime := time.Now()
	result := WorkExecutionResult{
		WorkID:    work.GetID(),
		Status:    layer0.WorkStatusExecuting,
		StartedAt: startTime,
	}

	// Execute work
	output, err := executor.Execute(work, context)
	endTime := time.Now()
	duration := endTime.Sub(startTime)

	wec.mutex.Lock()
	defer wec.mutex.Unlock()

	// Remove from active work
	delete(wec.activeWork, work.GetID())

	// Update result
	result.Duration = duration
	result.CompletedAt = &endTime

	if err != nil {
		result.Status = layer0.WorkStatusFailed
		result.Error = err.Error()
	} else {
		result.Status = layer0.WorkStatusCompleted
		result.Output = output
	}

	// Store result
	wec.executionResults[work.GetID()] = result

	return result, nil
}

// GetActiveWork returns all currently active work items
func (wec *WorkExecutionCore) GetActiveWork() []layer0.Work {
	wec.mutex.RLock()
	defer wec.mutex.RUnlock()

	activeWork := make([]layer0.Work, 0, len(wec.activeWork))
	for _, work := range wec.activeWork {
		activeWork = append(activeWork, work)
	}

	return activeWork
}

// GetExecutionResult retrieves the execution result for a specific work ID
func (wec *WorkExecutionCore) GetExecutionResult(workID layer0.WorkID) (WorkExecutionResult, error) {
	wec.mutex.RLock()
	defer wec.mutex.RUnlock()

	result, exists := wec.executionResults[workID]
	if !exists {
		return WorkExecutionResult{}, fmt.Errorf("no execution result found for work ID %s", workID)
	}

	return result, nil
}

// GetAllExecutionResults returns all execution results
func (wec *WorkExecutionCore) GetAllExecutionResults() []WorkExecutionResult {
	wec.mutex.RLock()
	defer wec.mutex.RUnlock()

	results := make([]WorkExecutionResult, 0, len(wec.executionResults))
	for _, result := range wec.executionResults {
		results = append(results, result)
	}

	return results
}

// CancelWork cancels an active work item
func (wec *WorkExecutionCore) CancelWork(workID layer0.WorkID) error {
	wec.mutex.Lock()
	defer wec.mutex.Unlock()

	work, isActive := wec.activeWork[workID]
	if !isActive {
		return fmt.Errorf("work %s is not currently active", workID)
	}

	// Remove from active work
	delete(wec.activeWork, workID)

	// Create cancellation result
	now := time.Now()
	var startedAt time.Time
	if work.GetMetadata().StartedAt != nil {
		startedAt = *work.GetMetadata().StartedAt
	} else {
		startedAt = now
	}

	result := WorkExecutionResult{
		WorkID:      workID,
		Status:      layer0.WorkStatusCancelled,
		Error:       "work was cancelled",
		StartedAt:   startedAt,
		CompletedAt: &now,
		Duration:    now.Sub(startedAt),
	}

	wec.executionResults[workID] = result
	return nil
}

// IsWorkActive checks if a work item is currently being executed
func (wec *WorkExecutionCore) IsWorkActive(workID layer0.WorkID) bool {
	wec.mutex.RLock()
	defer wec.mutex.RUnlock()

	_, isActive := wec.activeWork[workID]
	return isActive
}

// MockWorkExecutor is a simple mock executor for testing
type MockWorkExecutor struct {
	supportedTypes []layer0.WorkType
	executeFunc    func(work layer0.Work, context layer0.Context) (interface{}, error)
}

// NewMockWorkExecutor creates a new mock work executor
func NewMockWorkExecutor(supportedTypes []layer0.WorkType, executeFunc func(layer0.Work, layer0.Context) (interface{}, error)) *MockWorkExecutor {
	return &MockWorkExecutor{
		supportedTypes: supportedTypes,
		executeFunc:    executeFunc,
	}
}

// Execute executes the work using the mock function
func (mwe *MockWorkExecutor) Execute(work layer0.Work, context layer0.Context) (interface{}, error) {
	if mwe.executeFunc != nil {
		return mwe.executeFunc(work, context)
	}
	return "mock result", nil
}

// CanExecute checks if the executor can execute the given work type
func (mwe *MockWorkExecutor) CanExecute(workType layer0.WorkType) bool {
	for _, supportedType := range mwe.supportedTypes {
		if supportedType == workType {
			return true
		}
	}
	return false
}

// GetSupportedTypes returns the supported work types
func (mwe *MockWorkExecutor) GetSupportedTypes() []layer0.WorkType {
	return mwe.supportedTypes
}
