
package layer0

import (
	"testing"
	"time"
)

func TestNewState(t *testing.T) {
	id := StateID("test-state")
	stateType := StateTypeInitial
	name := "Test State"
	
	state := NewState(id, stateType, name)
	
	if state.GetID() != id {
		t.Errorf("Expected ID %s, got %s", id, state.GetID())
	}
	
	if state.GetType() != stateType {
		t.Errorf("Expected type %s, got %s", stateType, state.GetType())
	}
	
	if state.GetStatus() != StateStatusInactive {
		t.Errorf("Expected status %s, got %s", StateStatusInactive, state.GetStatus())
	}
	
	if state.GetMetadata().Name != name {
		t.Errorf("Expected name %s, got %s", name, state.GetMetadata().Name)
	}
}

func TestStateSetStatus(t *testing.T) {
	state := NewState("test", StateTypeInitial, "Test")
	originalTime := state.Metadata.UpdatedAt
	
	time.Sleep(1 * time.Millisecond) // Ensure time difference
	
	newState := state.SetStatus(StateStatusActive)
	
	if newState.GetStatus() != StateStatusActive {
		t.Errorf("Expected status %s, got %s", StateStatusActive, newState.GetStatus())
	}
	
	if newState.Metadata.UpdatedAt.Equal(originalTime) {
		t.Error("UpdatedAt should be updated when status changes")
	}
	
	// Original state should remain unchanged (immutability)
	if state.GetStatus() != StateStatusInactive {
		t.Error("Original state should remain unchanged")
	}
}

func TestStateSetData(t *testing.T) {
	state := NewState("test", StateTypeInitial, "Test")
	testData := map[string]interface{}{"key": "value"}
	
	newState := state.SetData(testData)
	
	if newState.GetData() == nil {
		t.Error("Data should be set")
	}
	
	// Original state should remain unchanged (immutability)
	if state.GetData() != nil {
		t.Error("Original state should remain unchanged")
	}
}

func TestStateIsActive(t *testing.T) {
	state := NewState("test", StateTypeInitial, "Test")
	
	if state.IsActive() {
		t.Error("New state should not be active")
	}
	
	activeState := state.SetStatus(StateStatusActive)
	if !activeState.IsActive() {
		t.Error("State with active status should be active")
	}
}

func TestStateIsFinal(t *testing.T) {
	initialState := NewState("test", StateTypeInitial, "Test")
	finalState := NewState("test", StateTypeFinal, "Test")
	
	if initialState.IsFinal() {
		t.Error("Initial state should not be final")
	}
	
	if !finalState.IsFinal() {
		t.Error("Final state should be final")
	}
}

func TestStateIsError(t *testing.T) {
	normalState := NewState("test", StateTypeInitial, "Test")
	errorState := NewState("test", StateTypeError, "Test")
	failedState := NewState("test", StateTypeInitial, "Test").SetStatus(StateStatusFailed)
	
	if normalState.IsError() {
		t.Error("Normal state should not be error")
	}
	
	if !errorState.IsError() {
		t.Error("Error type state should be error")
	}
	
	if !failedState.IsError() {
		t.Error("Failed status state should be error")
	}
}

func TestStateClone(t *testing.T) {
	original := NewState("test", StateTypeInitial, "Test")
	original.Metadata.Tags = []string{"tag1", "tag2"}
	original.Metadata.Properties["key"] = "value"
	
	cloned := original.Clone()
	
	// Verify clone has same values
	if cloned.GetID() != original.GetID() {
		t.Error("Cloned state should have same ID")
	}
	
	// Verify independence (modify clone)
	cloned.Metadata.Tags[0] = "modified"
	cloned.Metadata.Properties["key"] = "modified"
	
	if original.Metadata.Tags[0] == "modified" {
		t.Error("Original state tags should not be affected by clone modification")
	}
	
	if original.Metadata.Properties["key"] == "modified" {
		t.Error("Original state properties should not be affected by clone modification")
	}
}

func TestStateValidate(t *testing.T) {
	// Valid state
	validState := NewState("test", StateTypeInitial, "Test")
	if err := validState.Validate(); err != nil {
		t.Errorf("Valid state should not return error: %v", err)
	}
	
	// Invalid states
	invalidStates := []State{
		{ID: "", Type: StateTypeInitial, Status: StateStatusInactive, Metadata: StateMetadata{Name: "Test"}},
		{ID: "test", Type: "", Status: StateStatusInactive, Metadata: StateMetadata{Name: "Test"}},
		{ID: "test", Type: StateTypeInitial, Status: "", Metadata: StateMetadata{Name: "Test"}},
		{ID: "test", Type: StateTypeInitial, Status: StateStatusInactive, Metadata: StateMetadata{Name: ""}},
	}
	
	for i, state := range invalidStates {
		if err := state.Validate(); err == nil {
			t.Errorf("Invalid state %d should return error", i)
		}
	}
}
