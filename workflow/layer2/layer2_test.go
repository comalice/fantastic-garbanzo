package layer2

import (
	"errors"
	"testing"

	"github.com/ubom/workflow/layer0"
	"github.com/ubom/workflow/layer1"
)

func TestWorkflowRuntimeEngineBasicOperations(t *testing.T) {
	engine := NewWorkflowRuntimeEngine()

	// Create a simple workflow definition
	definition := layer1.NewWorkflowDefinition("test-workflow", "1.0.0", "Test Workflow")

	// Create states
	initialState := layer0.NewState("initial", layer0.StateTypeInitial, "Initial State")
	finalState := layer0.NewState("final", layer0.StateTypeFinal, "Final State")

	// Add states to state machine
	stateMachine := layer1.NewStateMachineCore()
	stateMachine.AddState(initialState)
	stateMachine.AddState(finalState)

	// Create transition
	transition := layer0.NewTransition("t1", layer0.TransitionTypeAutomatic, initialState.GetID(), finalState.GetID(), "Transition")
	stateMachine.AddTransition(transition)

	// Update definition
	definition = definition.SetStateMachine(stateMachine).
		SetInitialStateID(initialState.GetID()).
		AddFinalStateID(finalState.GetID()).
		SetStatus(layer1.WorkflowDefinitionStatusActive)

	// Create initial context
	initialContext := layer0.NewContext("initial-context", layer0.ContextScopeWorkflow, "Initial Context")

	// Start workflow
	instanceID, err := engine.StartWorkflow(definition, initialContext)
	if err != nil {
		t.Errorf("StartWorkflow should not return error: %v", err)
	}

	if instanceID == "" {
		t.Error("StartWorkflow should return a valid instance ID")
	}

	// Check workflow status
	status, err := engine.GetWorkflowStatus(instanceID)
	if err != nil {
		t.Errorf("GetWorkflowStatus should not return error: %v", err)
	}

	if status != WorkflowInstanceStatusRunning {
		t.Errorf("Expected status %s, got %s", WorkflowInstanceStatusRunning, status)
	}

	// List active workflows
	activeWorkflows := engine.ListActiveWorkflows()
	if len(activeWorkflows) != 1 {
		t.Errorf("Expected 1 active workflow, got %d", len(activeWorkflows))
	}

	// Execute workflow
	err = engine.ExecuteWorkflow(instanceID)
	if err != nil {
		t.Errorf("ExecuteWorkflow should not return error: %v", err)
	}

	// Check final status
	finalStatus, err := engine.GetWorkflowStatus(instanceID)
	if err != nil {
		t.Errorf("GetWorkflowStatus should not return error: %v", err)
	}

	if finalStatus != WorkflowInstanceStatusCompleted {
		t.Errorf("Expected final status %s, got %s", WorkflowInstanceStatusCompleted, finalStatus)
	}
}

func TestWorkflowRuntimeEnginePauseResume(t *testing.T) {
	engine := NewWorkflowRuntimeEngine()

	// Create a simple workflow definition
	definition := layer1.NewWorkflowDefinition("test-workflow", "1.0.0", "Test Workflow")

	// Create states
	initialState := layer0.NewState("initial", layer0.StateTypeInitial, "Initial State")
	finalState := layer0.NewState("final", layer0.StateTypeFinal, "Final State")

	// Add states to state machine
	stateMachine := layer1.NewStateMachineCore()
	stateMachine.AddState(initialState)
	stateMachine.AddState(finalState)

	// Update definition
	definition = definition.SetStateMachine(stateMachine).
		SetInitialStateID(initialState.GetID()).
		AddFinalStateID(finalState.GetID()).
		SetStatus(layer1.WorkflowDefinitionStatusActive)

	// Create initial context
	initialContext := layer0.NewContext("initial-context", layer0.ContextScopeWorkflow, "Initial Context")

	// Start workflow
	instanceID, err := engine.StartWorkflow(definition, initialContext)
	if err != nil {
		t.Errorf("StartWorkflow should not return error: %v", err)
	}

	// Pause workflow
	err = engine.PauseWorkflow(instanceID)
	if err != nil {
		t.Errorf("PauseWorkflow should not return error: %v", err)
	}

	// Check paused status
	status, _ := engine.GetWorkflowStatus(instanceID)
	if status != WorkflowInstanceStatusPaused {
		t.Errorf("Expected status %s, got %s", WorkflowInstanceStatusPaused, status)
	}

	// Resume workflow
	err = engine.ResumeWorkflow(instanceID)
	if err != nil {
		t.Errorf("ResumeWorkflow should not return error: %v", err)
	}

	// Check resumed status
	status, _ = engine.GetWorkflowStatus(instanceID)
	if status != WorkflowInstanceStatusRunning {
		t.Errorf("Expected status %s, got %s", WorkflowInstanceStatusRunning, status)
	}
}

func TestWorkflowRuntimeEngineCancel(t *testing.T) {
	engine := NewWorkflowRuntimeEngine()

	// Create a simple workflow definition
	definition := layer1.NewWorkflowDefinition("test-workflow", "1.0.0", "Test Workflow")

	// Create states
	initialState := layer0.NewState("initial", layer0.StateTypeInitial, "Initial State")
	finalState := layer0.NewState("final", layer0.StateTypeFinal, "Final State")

	// Add states to state machine
	stateMachine := layer1.NewStateMachineCore()
	stateMachine.AddState(initialState)
	stateMachine.AddState(finalState)

	// Update definition
	definition = definition.SetStateMachine(stateMachine).
		SetInitialStateID(initialState.GetID()).
		AddFinalStateID(finalState.GetID()).
		SetStatus(layer1.WorkflowDefinitionStatusActive)

	// Create initial context
	initialContext := layer0.NewContext("initial-context", layer0.ContextScopeWorkflow, "Initial Context")

	// Start workflow
	instanceID, err := engine.StartWorkflow(definition, initialContext)
	if err != nil {
		t.Errorf("StartWorkflow should not return error: %v", err)
	}

	// Cancel workflow
	err = engine.CancelWorkflow(instanceID)
	if err != nil {
		t.Errorf("CancelWorkflow should not return error: %v", err)
	}

	// Check cancelled status
	status, _ := engine.GetWorkflowStatus(instanceID)
	if status != WorkflowInstanceStatusCancelled {
		t.Errorf("Expected status %s, got %s", WorkflowInstanceStatusCancelled, status)
	}

	// Check that workflow is no longer active
	activeWorkflows := engine.ListActiveWorkflows()
	if len(activeWorkflows) != 0 {
		t.Errorf("Expected 0 active workflows after cancellation, got %d", len(activeWorkflows))
	}
}

func TestDefaultTransitionEvaluator(t *testing.T) {
	evaluator := NewDefaultTransitionEvaluator()

	// Create test transition
	transition := layer0.NewTransition("test-transition", layer0.TransitionTypeAutomatic, "from", "to", "Test Transition")
	context := layer0.NewContext("test-context", layer0.ContextScopeWorkflow, "Test Context")

	// Test transition without conditions
	canTransition, err := evaluator.CanTransition(transition, context)
	if err != nil {
		t.Errorf("CanTransition should not return error: %v", err)
	}

	if !canTransition {
		t.Error("Transition without conditions should be allowed")
	}

	// Test transition with conditions
	transitionWithConditions := transition.AddCondition("test-condition")

	// Context without the condition should allow transition (default behavior)
	canTransition, err = evaluator.CanTransition(transitionWithConditions, context)
	if err != nil {
		t.Errorf("CanTransition should not return error: %v", err)
	}

	if !canTransition {
		t.Error("Transition with default condition should be allowed")
	}

	// Test with context that has the condition set to false
	contextWithFalseCondition := context.Set("test-condition", false)
	canTransition, err = evaluator.CanTransition(transitionWithConditions, contextWithFalseCondition)
	if err != nil {
		t.Errorf("CanTransition should not return error: %v", err)
	}

	if canTransition {
		t.Error("Transition with false condition should not be allowed")
	}

	// Test with context that has the condition set to true
	contextWithTrueCondition := context.Set("test-condition", true)
	canTransition, err = evaluator.CanTransition(transitionWithConditions, contextWithTrueCondition)
	if err != nil {
		t.Errorf("CanTransition should not return error: %v", err)
	}

	if !canTransition {
		t.Error("Transition with true condition should be allowed")
	}
}

func TestDefaultErrorHandler(t *testing.T) {
	handler := NewDefaultErrorHandler()
	instanceID := WorkflowInstanceID("test-instance")
	testError := errors.New("test error")

	// Test HandleError
	err := handler.HandleError(instanceID, testError)
	if err != nil {
		t.Errorf("HandleError should not return error: %v", err)
	}

	// Test GetErrors
	workflowErrors := handler.GetErrors(instanceID)
	if len(workflowErrors) != 1 {
		t.Errorf("Expected 1 error, got %d", len(workflowErrors))
	}

	if workflowErrors[0].Error.Error() != testError.Error() {
		t.Error("Error should match the original error")
	}

	if workflowErrors[0].Severity != ErrorSeverityMedium {
		t.Errorf("Expected severity %s, got %s", ErrorSeverityMedium, workflowErrors[0].Severity)
	}

	// Test HandleErrorWithSeverity
	criticalError := errors.New("critical error")
	err = handler.HandleErrorWithSeverity(instanceID, criticalError, ErrorSeverityCritical)
	if err != nil {
		t.Errorf("HandleErrorWithSeverity should not return error: %v", err)
	}

	// Test GetAllErrors
	allErrors := handler.GetAllErrors()
	if len(allErrors) != 2 {
		t.Errorf("Expected 2 errors, got %d", len(allErrors))
	}

	// Test IsRecoverable
	if !handler.IsRecoverable(errors.New("timeout error")) {
		t.Error("Timeout error should be recoverable")
	}

	if handler.IsRecoverable(errors.New("validation error")) {
		t.Error("Validation error should not be recoverable")
	}

	// Test ClearErrors
	err = handler.ClearErrors(instanceID)
	if err != nil {
		t.Errorf("ClearErrors should not return error: %v", err)
	}

	clearedErrors := handler.GetErrors(instanceID)
	if len(clearedErrors) != 0 {
		t.Errorf("Expected 0 errors after clearing, got %d", len(clearedErrors))
	}
}

func TestDefaultWorkflowLifecycleManager(t *testing.T) {
	manager := NewDefaultWorkflowLifecycleManager()
	instanceID := WorkflowInstanceID("test-instance")

	// Test OnWorkflowStarted
	err := manager.OnWorkflowStarted(instanceID)
	if err != nil {
		t.Errorf("OnWorkflowStarted should not return error: %v", err)
	}

	// Test OnWorkflowCompleted
	err = manager.OnWorkflowCompleted(instanceID)
	if err != nil {
		t.Errorf("OnWorkflowCompleted should not return error: %v", err)
	}

	// Test OnStateChanged
	err = manager.OnStateChanged(instanceID, "state1", "state2")
	if err != nil {
		t.Errorf("OnStateChanged should not return error: %v", err)
	}

	// Test GetEvents
	events := manager.GetEvents(instanceID)
	if len(events) != 3 {
		t.Errorf("Expected 3 events, got %d", len(events))
	}

	// Verify event types
	expectedEventTypes := []string{"workflow_started", "workflow_completed", "state_changed"}
	for i, event := range events {
		if event.EventType != expectedEventTypes[i] {
			t.Errorf("Expected event type %s, got %s", expectedEventTypes[i], event.EventType)
		}
	}

	// Test GetAllEvents
	allEvents := manager.GetAllEvents()
	if len(allEvents) != 3 {
		t.Errorf("Expected 3 events, got %d", len(allEvents))
	}

	// Test ClearEvents
	err = manager.ClearEvents(instanceID)
	if err != nil {
		t.Errorf("ClearEvents should not return error: %v", err)
	}

	clearedEvents := manager.GetEvents(instanceID)
	if len(clearedEvents) != 0 {
		t.Errorf("Expected 0 events after clearing, got %d", len(clearedEvents))
	}
}

func TestWorkflowRuntimeEngineShutdown(t *testing.T) {
	engine := NewWorkflowRuntimeEngine()

	// Create a simple workflow definition
	definition := layer1.NewWorkflowDefinition("test-workflow", "1.0.0", "Test Workflow")

	// Create states
	initialState := layer0.NewState("initial", layer0.StateTypeInitial, "Initial State")
	finalState := layer0.NewState("final", layer0.StateTypeFinal, "Final State")

	// Add states to state machine
	stateMachine := layer1.NewStateMachineCore()
	stateMachine.AddState(initialState)
	stateMachine.AddState(finalState)

	// Update definition
	definition = definition.SetStateMachine(stateMachine).
		SetInitialStateID(initialState.GetID()).
		AddFinalStateID(finalState.GetID()).
		SetStatus(layer1.WorkflowDefinitionStatusActive)

	// Create initial context
	initialContext := layer0.NewContext("initial-context", layer0.ContextScopeWorkflow, "Initial Context")

	// Start multiple workflows
	instanceID1, _ := engine.StartWorkflow(definition, initialContext)
	instanceID2, _ := engine.StartWorkflow(definition, initialContext)

	// Verify workflows are active
	activeWorkflows := engine.ListActiveWorkflows()
	if len(activeWorkflows) != 2 {
		t.Errorf("Expected 2 active workflows, got %d", len(activeWorkflows))
	}

	// Shutdown engine
	err := engine.Shutdown()
	if err != nil {
		t.Errorf("Shutdown should not return error: %v", err)
	}

	// Verify all workflows are cancelled
	status1, _ := engine.GetWorkflowStatus(instanceID1)
	status2, _ := engine.GetWorkflowStatus(instanceID2)

	if status1 != WorkflowInstanceStatusCancelled {
		t.Errorf("Expected workflow 1 to be cancelled, got %s", status1)
	}

	if status2 != WorkflowInstanceStatusCancelled {
		t.Errorf("Expected workflow 2 to be cancelled, got %s", status2)
	}

	// Verify no active workflows
	activeWorkflowsAfterShutdown := engine.ListActiveWorkflows()
	if len(activeWorkflowsAfterShutdown) != 0 {
		t.Errorf("Expected 0 active workflows after shutdown, got %d", len(activeWorkflowsAfterShutdown))
	}
}
