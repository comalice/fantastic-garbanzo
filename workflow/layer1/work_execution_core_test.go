package layer1

import (
	"errors"
	"testing"
	"time"

	"github.com/ubom/workflow/layer0"
)

func TestNewWorkExecutionCore(t *testing.T) {
	wec := NewWorkExecutionCore()

	if wec == nil {
		t.Error("NewWorkExecutionCore should return a non-nil instance")
	}

	supportedTypes := wec.GetSupportedWorkTypes()
	if len(supportedTypes) != 0 {
		t.Errorf("New work execution core should have 0 supported types, got %d", len(supportedTypes))
	}
}

func TestWorkExecutionCoreRegisterExecutor(t *testing.T) {
	wec := NewWorkExecutionCore()
	executor := NewMockWorkExecutor([]layer0.WorkType{layer0.WorkTypeTask}, nil)

	err := wec.RegisterExecutor(layer0.WorkTypeTask, executor)
	if err != nil {
		t.Errorf("RegisterExecutor should not return error: %v", err)
	}

	// Try to register the same type again
	err = wec.RegisterExecutor(layer0.WorkTypeTask, executor)
	if err == nil {
		t.Error("RegisterExecutor should return error when registering duplicate type")
	}

	// Try to register nil executor
	err = wec.RegisterExecutor(layer0.WorkTypeService, nil)
	if err == nil {
		t.Error("RegisterExecutor should return error for nil executor")
	}

	// Verify executor was registered
	retrievedExecutor, err := wec.GetExecutor(layer0.WorkTypeTask)
	if err != nil {
		t.Errorf("GetExecutor should not return error: %v", err)
	}

	if retrievedExecutor != executor {
		t.Error("Retrieved executor should match registered executor")
	}
}

func TestWorkExecutionCoreUnregisterExecutor(t *testing.T) {
	wec := NewWorkExecutionCore()
	executor := NewMockWorkExecutor([]layer0.WorkType{layer0.WorkTypeTask}, nil)

	// Try to unregister non-existent executor
	err := wec.UnregisterExecutor(layer0.WorkTypeTask)
	if err == nil {
		t.Error("UnregisterExecutor should return error for non-existent executor")
	}

	// Register and then unregister
	wec.RegisterExecutor(layer0.WorkTypeTask, executor)
	err = wec.UnregisterExecutor(layer0.WorkTypeTask)
	if err != nil {
		t.Errorf("UnregisterExecutor should not return error: %v", err)
	}

	// Verify executor was unregistered
	_, err = wec.GetExecutor(layer0.WorkTypeTask)
	if err == nil {
		t.Error("GetExecutor should return error for unregistered executor")
	}
}

func TestWorkExecutionCoreGetSupportedWorkTypes(t *testing.T) {
	wec := NewWorkExecutionCore()

	// Initially no supported types
	types := wec.GetSupportedWorkTypes()
	if len(types) != 0 {
		t.Errorf("Expected 0 supported types, got %d", len(types))
	}

	// Register executors
	executor1 := NewMockWorkExecutor([]layer0.WorkType{layer0.WorkTypeTask}, nil)
	executor2 := NewMockWorkExecutor([]layer0.WorkType{layer0.WorkTypeService}, nil)
	wec.RegisterExecutor(layer0.WorkTypeTask, executor1)
	wec.RegisterExecutor(layer0.WorkTypeService, executor2)

	types = wec.GetSupportedWorkTypes()
	if len(types) != 2 {
		t.Errorf("Expected 2 supported types, got %d", len(types))
	}

	// Check that both types are present
	typeMap := make(map[layer0.WorkType]bool)
	for _, workType := range types {
		typeMap[workType] = true
	}

	if !typeMap[layer0.WorkTypeTask] || !typeMap[layer0.WorkTypeService] {
		t.Error("Expected both WorkTypeTask and WorkTypeService to be supported")
	}
}

func TestWorkExecutionCoreExecuteWork(t *testing.T) {
	wec := NewWorkExecutionCore()

	// Create work and context
	work := layer0.NewWork("test-work", layer0.WorkTypeTask, "Test Work")
	context := layer0.NewContext("test-context", layer0.ContextScopeWork, "Test Context")

	// Try to execute without registered executor
	_, err := wec.ExecuteWork(work, context)
	if err == nil {
		t.Error("ExecuteWork should return error when no executor is registered")
	}

	// Register executor
	executor := NewMockWorkExecutor([]layer0.WorkType{layer0.WorkTypeTask}, func(w layer0.Work, c layer0.Context) (interface{}, error) {
		return "test result", nil
	})
	wec.RegisterExecutor(layer0.WorkTypeTask, executor)

	// Execute work
	result, err := wec.ExecuteWork(work, context)
	if err != nil {
		t.Errorf("ExecuteWork should not return error: %v", err)
	}

	if result.WorkID != work.GetID() {
		t.Error("Result should have correct work ID")
	}

	if result.Status != layer0.WorkStatusCompleted {
		t.Errorf("Expected status %s, got %s", layer0.WorkStatusCompleted, result.Status)
	}

	if result.Output != "test result" {
		t.Errorf("Expected output 'test result', got %v", result.Output)
	}

	if result.CompletedAt == nil {
		t.Error("CompletedAt should be set")
	}

	if result.Duration <= 0 {
		t.Error("Duration should be positive")
	}
}

func TestWorkExecutionCoreExecuteWorkWithError(t *testing.T) {
	wec := NewWorkExecutionCore()

	// Create work and context
	work := layer0.NewWork("test-work", layer0.WorkTypeTask, "Test Work")
	context := layer0.NewContext("test-context", layer0.ContextScopeWork, "Test Context")

	// Register executor that returns error
	executor := NewMockWorkExecutor([]layer0.WorkType{layer0.WorkTypeTask}, func(w layer0.Work, c layer0.Context) (interface{}, error) {
		return nil, errors.New("execution failed")
	})
	wec.RegisterExecutor(layer0.WorkTypeTask, executor)

	// Execute work
	result, err := wec.ExecuteWork(work, context)
	if err != nil {
		t.Errorf("ExecuteWork should not return error even when execution fails: %v", err)
	}

	if result.Status != layer0.WorkStatusFailed {
		t.Errorf("Expected status %s, got %s", layer0.WorkStatusFailed, result.Status)
	}

	if result.Error != "execution failed" {
		t.Errorf("Expected error 'execution failed', got %s", result.Error)
	}

	if result.Output != nil {
		t.Error("Output should be nil for failed execution")
	}
}

func TestWorkExecutionCoreExecuteNonExecutableWork(t *testing.T) {
	wec := NewWorkExecutionCore()

	// Create completed work (not executable)
	work := layer0.NewWork("test-work", layer0.WorkTypeTask, "Test Work").SetStatus(layer0.WorkStatusCompleted)
	context := layer0.NewContext("test-context", layer0.ContextScopeWork, "Test Context")

	// Register executor
	executor := NewMockWorkExecutor([]layer0.WorkType{layer0.WorkTypeTask}, nil)
	wec.RegisterExecutor(layer0.WorkTypeTask, executor)

	// Try to execute non-executable work
	_, err := wec.ExecuteWork(work, context)
	if err == nil {
		t.Error("ExecuteWork should return error for non-executable work")
	}
}

func TestWorkExecutionCoreExecuteAlreadyActiveWork(t *testing.T) {
	wec := NewWorkExecutionCore()

	// Create work and context
	work := layer0.NewWork("test-work", layer0.WorkTypeTask, "Test Work")
	context := layer0.NewContext("test-context", layer0.ContextScopeWork, "Test Context")

	// Register executor that takes time
	executor := NewMockWorkExecutor([]layer0.WorkType{layer0.WorkTypeTask}, func(w layer0.Work, c layer0.Context) (interface{}, error) {
		time.Sleep(100 * time.Millisecond)
		return "result", nil
	})
	wec.RegisterExecutor(layer0.WorkTypeTask, executor)

	// Start execution in goroutine
	go wec.ExecuteWork(work, context)

	// Wait a bit to ensure execution starts
	time.Sleep(10 * time.Millisecond)

	// Try to execute the same work again
	_, err := wec.ExecuteWork(work, context)
	if err == nil {
		t.Error("ExecuteWork should return error when work is already active")
	}
}

func TestWorkExecutionCoreGetActiveWork(t *testing.T) {
	wec := NewWorkExecutionCore()

	// Initially no active work
	activeWork := wec.GetActiveWork()
	if len(activeWork) != 0 {
		t.Errorf("Expected 0 active work items, got %d", len(activeWork))
	}

	// Create work and context
	work := layer0.NewWork("test-work", layer0.WorkTypeTask, "Test Work")
	context := layer0.NewContext("test-context", layer0.ContextScopeWork, "Test Context")

	// Register executor that takes time
	executor := NewMockWorkExecutor([]layer0.WorkType{layer0.WorkTypeTask}, func(w layer0.Work, c layer0.Context) (interface{}, error) {
		time.Sleep(100 * time.Millisecond)
		return "result", nil
	})
	wec.RegisterExecutor(layer0.WorkTypeTask, executor)

	// Start execution in goroutine
	go wec.ExecuteWork(work, context)

	// Wait a bit to ensure execution starts
	time.Sleep(10 * time.Millisecond)

	// Check active work
	activeWork = wec.GetActiveWork()
	if len(activeWork) != 1 {
		t.Errorf("Expected 1 active work item, got %d", len(activeWork))
	}

	if activeWork[0].GetID() != work.GetID() {
		t.Error("Active work should match the executing work")
	}

	// Wait for execution to complete
	time.Sleep(200 * time.Millisecond)

	// Should be no active work now
	activeWork = wec.GetActiveWork()
	if len(activeWork) != 0 {
		t.Errorf("Expected 0 active work items after completion, got %d", len(activeWork))
	}
}

func TestWorkExecutionCoreGetExecutionResult(t *testing.T) {
	wec := NewWorkExecutionCore()

	// Try to get non-existent result
	_, err := wec.GetExecutionResult("non-existent")
	if err == nil {
		t.Error("GetExecutionResult should return error for non-existent work")
	}

	// Create work and context
	work := layer0.NewWork("test-work", layer0.WorkTypeTask, "Test Work")
	context := layer0.NewContext("test-context", layer0.ContextScopeWork, "Test Context")

	// Register executor
	executor := NewMockWorkExecutor([]layer0.WorkType{layer0.WorkTypeTask}, func(w layer0.Work, c layer0.Context) (interface{}, error) {
		return "test result", nil
	})
	wec.RegisterExecutor(layer0.WorkTypeTask, executor)

	// Execute work
	wec.ExecuteWork(work, context)

	// Get execution result
	result, err := wec.GetExecutionResult(work.GetID())
	if err != nil {
		t.Errorf("GetExecutionResult should not return error: %v", err)
	}

	if result.WorkID != work.GetID() {
		t.Error("Result should have correct work ID")
	}

	if result.Status != layer0.WorkStatusCompleted {
		t.Error("Result should have completed status")
	}
}

func TestWorkExecutionCoreGetAllExecutionResults(t *testing.T) {
	wec := NewWorkExecutionCore()

	// Initially no results
	results := wec.GetAllExecutionResults()
	if len(results) != 0 {
		t.Errorf("Expected 0 execution results, got %d", len(results))
	}

	// Execute multiple work items
	work1 := layer0.NewWork("work1", layer0.WorkTypeTask, "Work 1")
	work2 := layer0.NewWork("work2", layer0.WorkTypeTask, "Work 2")
	context := layer0.NewContext("test-context", layer0.ContextScopeWork, "Test Context")

	executor := NewMockWorkExecutor([]layer0.WorkType{layer0.WorkTypeTask}, func(w layer0.Work, c layer0.Context) (interface{}, error) {
		return "result", nil
	})
	wec.RegisterExecutor(layer0.WorkTypeTask, executor)

	wec.ExecuteWork(work1, context)
	wec.ExecuteWork(work2, context)

	// Get all results
	results = wec.GetAllExecutionResults()
	if len(results) != 2 {
		t.Errorf("Expected 2 execution results, got %d", len(results))
	}
}

func TestWorkExecutionCoreCancelWork(t *testing.T) {
	wec := NewWorkExecutionCore()

	// Try to cancel non-active work
	err := wec.CancelWork("non-existent")
	if err == nil {
		t.Error("CancelWork should return error for non-active work")
	}

	// Create work and context
	work := layer0.NewWork("test-work", layer0.WorkTypeTask, "Test Work")
	context := layer0.NewContext("test-context", layer0.ContextScopeWork, "Test Context")

	// Register executor that takes time
	executor := NewMockWorkExecutor([]layer0.WorkType{layer0.WorkTypeTask}, func(w layer0.Work, c layer0.Context) (interface{}, error) {
		time.Sleep(200 * time.Millisecond)
		return "result", nil
	})
	wec.RegisterExecutor(layer0.WorkTypeTask, executor)

	// Start execution in goroutine
	go wec.ExecuteWork(work, context)

	// Wait a bit to ensure execution starts
	time.Sleep(10 * time.Millisecond)

	// Cancel work
	err = wec.CancelWork(work.GetID())
	if err != nil {
		t.Errorf("CancelWork should not return error: %v", err)
	}

	// Check that work is no longer active
	if wec.IsWorkActive(work.GetID()) {
		t.Error("Work should not be active after cancellation")
	}

	// Check cancellation result
	result, err := wec.GetExecutionResult(work.GetID())
	if err != nil {
		t.Errorf("GetExecutionResult should not return error: %v", err)
	}

	if result.Status != layer0.WorkStatusCancelled {
		t.Errorf("Expected status %s, got %s", layer0.WorkStatusCancelled, result.Status)
	}

	if result.Error != "work was cancelled" {
		t.Errorf("Expected error 'work was cancelled', got %s", result.Error)
	}
}

func TestWorkExecutionCoreIsWorkActive(t *testing.T) {
	wec := NewWorkExecutionCore()

	// Initially no work is active
	if wec.IsWorkActive("test-work") {
		t.Error("IsWorkActive should return false for non-existent work")
	}

	// Create work and context
	work := layer0.NewWork("test-work", layer0.WorkTypeTask, "Test Work")
	context := layer0.NewContext("test-context", layer0.ContextScopeWork, "Test Context")

	// Register executor that takes time
	executor := NewMockWorkExecutor([]layer0.WorkType{layer0.WorkTypeTask}, func(w layer0.Work, c layer0.Context) (interface{}, error) {
		time.Sleep(100 * time.Millisecond)
		return "result", nil
	})
	wec.RegisterExecutor(layer0.WorkTypeTask, executor)

	// Start execution in goroutine
	go wec.ExecuteWork(work, context)

	// Wait a bit to ensure execution starts
	time.Sleep(10 * time.Millisecond)

	// Check that work is active
	if !wec.IsWorkActive(work.GetID()) {
		t.Error("IsWorkActive should return true for active work")
	}

	// Wait for execution to complete
	time.Sleep(200 * time.Millisecond)

	// Check that work is no longer active
	if wec.IsWorkActive(work.GetID()) {
		t.Error("IsWorkActive should return false after work completion")
	}
}

func TestMockWorkExecutor(t *testing.T) {
	supportedTypes := []layer0.WorkType{layer0.WorkTypeTask, layer0.WorkTypeService}
	executeFunc := func(work layer0.Work, context layer0.Context) (interface{}, error) {
		return "custom result", nil
	}

	executor := NewMockWorkExecutor(supportedTypes, executeFunc)

	// Test supported types
	if !executor.CanExecute(layer0.WorkTypeTask) {
		t.Error("Executor should support WorkTypeTask")
	}

	if !executor.CanExecute(layer0.WorkTypeService) {
		t.Error("Executor should support WorkTypeService")
	}

	if executor.CanExecute(layer0.WorkTypeScript) {
		t.Error("Executor should not support WorkTypeScript")
	}

	// Test get supported types
	types := executor.GetSupportedTypes()
	if len(types) != 2 {
		t.Errorf("Expected 2 supported types, got %d", len(types))
	}

	// Test execution
	work := layer0.NewWork("test", layer0.WorkTypeTask, "Test")
	context := layer0.NewContext("test", layer0.ContextScopeWork, "Test")

	result, err := executor.Execute(work, context)
	if err != nil {
		t.Errorf("Execute should not return error: %v", err)
	}

	if result != "custom result" {
		t.Errorf("Expected 'custom result', got %v", result)
	}

	// Test default execution (no custom function)
	defaultExecutor := NewMockWorkExecutor(supportedTypes, nil)
	result, err = defaultExecutor.Execute(work, context)
	if err != nil {
		t.Errorf("Execute should not return error: %v", err)
	}

	if result != "mock result" {
		t.Errorf("Expected 'mock result', got %v", result)
	}
}
