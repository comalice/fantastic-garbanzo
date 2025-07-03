package layer1

import (
	"testing"
	"time"

	"github.com/ubom/workflow/layer0"
)

func TestNewWorkflowDefinition(t *testing.T) {
	id := WorkflowDefinitionID("test-workflow")
	version := WorkflowDefinitionVersion("1.0.0")
	name := "Test Workflow"

	wd := NewWorkflowDefinition(id, version, name)

	if wd.GetID() != id {
		t.Errorf("Expected ID %s, got %s", id, wd.GetID())
	}

	if wd.GetVersion() != version {
		t.Errorf("Expected version %s, got %s", version, wd.GetVersion())
	}

	if wd.GetStatus() != WorkflowDefinitionStatusDraft {
		t.Errorf("Expected status %s, got %s", WorkflowDefinitionStatusDraft, wd.GetStatus())
	}

	if wd.GetMetadata().Name != name {
		t.Errorf("Expected name %s, got %s", name, wd.GetMetadata().Name)
	}

	if wd.GetStateMachine() == nil {
		t.Error("State machine should not be nil")
	}

	if wd.GetConfiguration().MaxConcurrentInstances != 100 {
		t.Errorf("Expected max concurrent instances 100, got %d", wd.GetConfiguration().MaxConcurrentInstances)
	}
}

func TestWorkflowDefinitionSetStatus(t *testing.T) {
	wd := NewWorkflowDefinition("test", "1.0.0", "Test")
	originalTime := wd.Metadata.UpdatedAt

	time.Sleep(1 * time.Millisecond) // Ensure time difference

	newWd := wd.SetStatus(WorkflowDefinitionStatusActive)

	if newWd.GetStatus() != WorkflowDefinitionStatusActive {
		t.Errorf("Expected status %s, got %s", WorkflowDefinitionStatusActive, newWd.GetStatus())
	}

	if newWd.Metadata.UpdatedAt.Equal(originalTime) {
		t.Error("UpdatedAt should be updated when status changes")
	}

	// Original workflow definition should remain unchanged (immutability)
	if wd.GetStatus() != WorkflowDefinitionStatusDraft {
		t.Error("Original workflow definition should remain unchanged")
	}
}

func TestWorkflowDefinitionSetStateMachine(t *testing.T) {
	wd := NewWorkflowDefinition("test", "1.0.0", "Test")
	newStateMachine := NewStateMachineCore()

	newWd := wd.SetStateMachine(newStateMachine)

	if newWd.GetStateMachine() != newStateMachine {
		t.Error("State machine should be updated")
	}

	// Original workflow definition should remain unchanged (immutability)
	if wd.GetStateMachine() == newStateMachine {
		t.Error("Original workflow definition should remain unchanged")
	}
}

func TestWorkflowDefinitionSetInitialStateID(t *testing.T) {
	wd := NewWorkflowDefinition("test", "1.0.0", "Test")
	stateID := layer0.StateID("initial-state")

	newWd := wd.SetInitialStateID(stateID)

	if newWd.GetInitialStateID() != stateID {
		t.Errorf("Expected initial state ID %s, got %s", stateID, newWd.GetInitialStateID())
	}

	// Original workflow definition should remain unchanged (immutability)
	if wd.GetInitialStateID() == stateID {
		t.Error("Original workflow definition should remain unchanged")
	}
}

func TestWorkflowDefinitionAddFinalStateID(t *testing.T) {
	wd := NewWorkflowDefinition("test", "1.0.0", "Test")
	stateID := layer0.StateID("final-state")

	newWd := wd.AddFinalStateID(stateID)

	finalStates := newWd.GetFinalStateIDs()
	if len(finalStates) != 1 || finalStates[0] != stateID {
		t.Errorf("Expected final state %s to be added", stateID)
	}

	// Original workflow definition should remain unchanged (immutability)
	if len(wd.GetFinalStateIDs()) != 0 {
		t.Error("Original workflow definition should remain unchanged")
	}
}

func TestWorkflowDefinitionAddErrorStateID(t *testing.T) {
	wd := NewWorkflowDefinition("test", "1.0.0", "Test")
	stateID := layer0.StateID("error-state")

	newWd := wd.AddErrorStateID(stateID)

	errorStates := newWd.GetErrorStateIDs()
	if len(errorStates) != 1 || errorStates[0] != stateID {
		t.Errorf("Expected error state %s to be added", stateID)
	}

	// Original workflow definition should remain unchanged (immutability)
	if len(wd.GetErrorStateIDs()) != 0 {
		t.Error("Original workflow definition should remain unchanged")
	}
}

func TestWorkflowDefinitionUpdateGlobalContext(t *testing.T) {
	wd := NewWorkflowDefinition("test", "1.0.0", "Test")
	newContext := layer0.NewContext("new-context", layer0.ContextScopeGlobal, "New Context")
	newContext = newContext.Set("key", "value")

	newWd := wd.UpdateGlobalContext(newContext)

	if newWd.GetGlobalContext().GetID() != newContext.GetID() {
		t.Error("Global context should be updated")
	}

	// Check that the context has the new data
	value, exists := newWd.GetGlobalContext().Get("key")
	if !exists || value != "value" {
		t.Error("Global context should contain the new data")
	}

	// Original workflow definition should remain unchanged (immutability)
	_, exists = wd.GetGlobalContext().Get("key")
	if exists {
		t.Error("Original workflow definition should remain unchanged")
	}
}

func TestWorkflowDefinitionUpdateConfiguration(t *testing.T) {
	wd := NewWorkflowDefinition("test", "1.0.0", "Test")
	newConfig := WorkflowConfiguration{
		MaxConcurrentInstances: 50,
		DefaultTimeoutSeconds:  1800,
		RetryPolicy: RetryPolicy{
			MaxRetries:        5,
			InitialDelay:      2 * time.Second,
			MaxDelay:          2 * time.Minute,
			BackoffMultiplier: 3.0,
			RetryableErrors:   []string{"timeout"},
		},
		CompensationEnabled: false,
		PersistenceEnabled:  false,
		LoggingLevel:        "DEBUG",
		Environment:         map[string]string{"ENV": "test"},
	}

	newWd := wd.UpdateConfiguration(newConfig)

	if newWd.GetConfiguration().MaxConcurrentInstances != 50 {
		t.Error("Configuration should be updated")
	}

	if newWd.GetConfiguration().RetryPolicy.MaxRetries != 5 {
		t.Error("Retry policy should be updated")
	}

	// Original workflow definition should remain unchanged (immutability)
	if wd.GetConfiguration().MaxConcurrentInstances == 50 {
		t.Error("Original workflow definition should remain unchanged")
	}
}

func TestWorkflowDefinitionIsActive(t *testing.T) {
	wd := NewWorkflowDefinition("test", "1.0.0", "Test")

	if wd.IsActive() {
		t.Error("New workflow definition should not be active")
	}

	activeWd := wd.SetStatus(WorkflowDefinitionStatusActive)
	if !activeWd.IsActive() {
		t.Error("Workflow definition with active status should be active")
	}
}

func TestWorkflowDefinitionCanExecute(t *testing.T) {
	wd := NewWorkflowDefinition("test", "1.0.0", "Test")

	// Draft workflow should not be executable
	if wd.CanExecute() {
		t.Error("Draft workflow definition should not be executable")
	}

	// Set up a valid workflow
	stateMachine := NewStateMachineCore()
	initialState := layer0.NewState("initial", layer0.StateTypeInitial, "Initial State")
	finalState := layer0.NewState("final", layer0.StateTypeFinal, "Final State")

	stateMachine.AddState(initialState)
	stateMachine.AddState(finalState)

	transition := layer0.NewTransition("t1", layer0.TransitionTypeAutomatic, initialState.GetID(), finalState.GetID(), "Transition")
	stateMachine.AddTransition(transition)

	validWd := wd.SetStateMachine(stateMachine).
		SetInitialStateID(initialState.GetID()).
		AddFinalStateID(finalState.GetID()).
		SetStatus(WorkflowDefinitionStatusActive)

	if !validWd.CanExecute() {
		t.Error("Valid active workflow definition should be executable")
	}
}

func TestWorkflowDefinitionClone(t *testing.T) {
	original := NewWorkflowDefinition("test", "1.0.0", "Test")
	original.Metadata.Tags = []string{"tag1", "tag2"}
	original.Metadata.Properties["key"] = "value"
	original = original.AddFinalStateID("final1").AddErrorStateID("error1")
	original.Configuration.Environment["ENV"] = "test"
	original.Configuration.RetryPolicy.RetryableErrors = []string{"error1"}

	cloned := original.Clone()

	// Verify clone has same values
	if cloned.GetID() != original.GetID() {
		t.Error("Cloned workflow definition should have same ID")
	}

	if len(cloned.GetFinalStateIDs()) != len(original.GetFinalStateIDs()) {
		t.Error("Cloned workflow definition should have same final states")
	}

	if len(cloned.GetErrorStateIDs()) != len(original.GetErrorStateIDs()) {
		t.Error("Cloned workflow definition should have same error states")
	}

	// Verify independence (modify clone)
	cloned.Metadata.Tags[0] = "modified"
	cloned.Metadata.Properties["key"] = "modified"
	cloned.FinalStateIDs[0] = "modified"
	cloned.ErrorStateIDs[0] = "modified"
	cloned.Configuration.Environment["ENV"] = "modified"
	cloned.Configuration.RetryPolicy.RetryableErrors[0] = "modified"

	if original.Metadata.Tags[0] == "modified" {
		t.Error("Original workflow definition tags should not be affected by clone modification")
	}

	if original.Metadata.Properties["key"] == "modified" {
		t.Error("Original workflow definition properties should not be affected by clone modification")
	}

	if original.FinalStateIDs[0] == "modified" {
		t.Error("Original workflow definition final states should not be affected by clone modification")
	}

	if original.ErrorStateIDs[0] == "modified" {
		t.Error("Original workflow definition error states should not be affected by clone modification")
	}

	if original.Configuration.Environment["ENV"] == "modified" {
		t.Error("Original workflow definition environment should not be affected by clone modification")
	}

	if original.Configuration.RetryPolicy.RetryableErrors[0] == "modified" {
		t.Error("Original workflow definition retryable errors should not be affected by clone modification")
	}
}

func TestWorkflowDefinitionValidate(t *testing.T) {
	// Create a valid workflow definition
	wd := NewWorkflowDefinition("test", "1.0.0", "Test")
	stateMachine := NewStateMachineCore()
	initialState := layer0.NewState("initial", layer0.StateTypeInitial, "Initial State")
	finalState := layer0.NewState("final", layer0.StateTypeFinal, "Final State")

	stateMachine.AddState(initialState)
	stateMachine.AddState(finalState)

	transition := layer0.NewTransition("t1", layer0.TransitionTypeAutomatic, initialState.GetID(), finalState.GetID(), "Transition")
	stateMachine.AddTransition(transition)

	validWd := wd.SetStateMachine(stateMachine).
		SetInitialStateID(initialState.GetID()).
		AddFinalStateID(finalState.GetID())

	if err := validWd.Validate(); err != nil {
		t.Errorf("Valid workflow definition should not return error: %v", err)
	}

	// Test invalid workflow definitions
	invalidWorkflows := []struct {
		name string
		wd   WorkflowDefinition
	}{
		{
			name: "empty ID",
			wd:   WorkflowDefinition{ID: "", Version: "1.0.0", Status: WorkflowDefinitionStatusDraft, Metadata: WorkflowDefinitionMetadata{Name: "Test"}},
		},
		{
			name: "empty version",
			wd:   WorkflowDefinition{ID: "test", Version: "", Status: WorkflowDefinitionStatusDraft, Metadata: WorkflowDefinitionMetadata{Name: "Test"}},
		},
		{
			name: "empty status",
			wd:   WorkflowDefinition{ID: "test", Version: "1.0.0", Status: "", Metadata: WorkflowDefinitionMetadata{Name: "Test"}},
		},
		{
			name: "empty name",
			wd:   WorkflowDefinition{ID: "test", Version: "1.0.0", Status: WorkflowDefinitionStatusDraft, Metadata: WorkflowDefinitionMetadata{Name: ""}},
		},
		{
			name: "nil state machine",
			wd:   WorkflowDefinition{ID: "test", Version: "1.0.0", Status: WorkflowDefinitionStatusDraft, Metadata: WorkflowDefinitionMetadata{Name: "Test"}, StateMachine: nil},
		},
		{
			name: "empty initial state",
			wd:   validWd.SetInitialStateID(""),
		},
		{
			name: "non-existent initial state",
			wd:   validWd.SetInitialStateID("non-existent"),
		},
		{
			name: "invalid configuration - zero max concurrent instances",
			wd: func() WorkflowDefinition {
				config := validWd.GetConfiguration()
				config.MaxConcurrentInstances = 0
				return validWd.UpdateConfiguration(config)
			}(),
		},
		{
			name: "invalid configuration - zero timeout",
			wd: func() WorkflowDefinition {
				config := validWd.GetConfiguration()
				config.DefaultTimeoutSeconds = 0
				return validWd.UpdateConfiguration(config)
			}(),
		},
		{
			name: "invalid configuration - negative max retries",
			wd: func() WorkflowDefinition {
				config := validWd.GetConfiguration()
				config.RetryPolicy.MaxRetries = -1
				return validWd.UpdateConfiguration(config)
			}(),
		},
		{
			name: "invalid configuration - negative initial delay",
			wd: func() WorkflowDefinition {
				config := validWd.GetConfiguration()
				config.RetryPolicy.InitialDelay = -time.Second
				return validWd.UpdateConfiguration(config)
			}(),
		},
		{
			name: "invalid configuration - max delay less than initial delay",
			wd: func() WorkflowDefinition {
				config := validWd.GetConfiguration()
				config.RetryPolicy.InitialDelay = time.Minute
				config.RetryPolicy.MaxDelay = time.Second
				return validWd.UpdateConfiguration(config)
			}(),
		},
		{
			name: "invalid configuration - zero backoff multiplier",
			wd: func() WorkflowDefinition {
				config := validWd.GetConfiguration()
				config.RetryPolicy.BackoffMultiplier = 0
				return validWd.UpdateConfiguration(config)
			}(),
		},
	}

	for _, test := range invalidWorkflows {
		t.Run(test.name, func(t *testing.T) {
			if err := test.wd.Validate(); err == nil {
				t.Errorf("Invalid workflow definition '%s' should return error", test.name)
			}
		})
	}
}

func TestWorkflowDefinitionValidateWithNonExistentFinalState(t *testing.T) {
	wd := NewWorkflowDefinition("test", "1.0.0", "Test")
	stateMachine := NewStateMachineCore()
	initialState := layer0.NewState("initial", layer0.StateTypeInitial, "Initial State")

	stateMachine.AddState(initialState)

	invalidWd := wd.SetStateMachine(stateMachine).
		SetInitialStateID(initialState.GetID()).
		AddFinalStateID("non-existent")

	if err := invalidWd.Validate(); err == nil {
		t.Error("Workflow definition with non-existent final state should return error")
	}
}

func TestWorkflowDefinitionValidateWithNonExistentErrorState(t *testing.T) {
	wd := NewWorkflowDefinition("test", "1.0.0", "Test")
	stateMachine := NewStateMachineCore()
	initialState := layer0.NewState("initial", layer0.StateTypeInitial, "Initial State")

	stateMachine.AddState(initialState)

	invalidWd := wd.SetStateMachine(stateMachine).
		SetInitialStateID(initialState.GetID()).
		AddErrorStateID("non-existent")

	if err := invalidWd.Validate(); err == nil {
		t.Error("Workflow definition with non-existent error state should return error")
	}
}
