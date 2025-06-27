
package layer1

import (
	"testing"

	"github.com/ubom/workflow/layer0"
)

func TestNewStateMachineCore(t *testing.T) {
	smc := NewStateMachineCore()
	
	if smc == nil {
		t.Error("NewStateMachineCore should return a non-nil instance")
	}
	
	states := smc.GetAllStates()
	if len(states) != 0 {
		t.Errorf("New state machine should have 0 states, got %d", len(states))
	}
}

func TestStateMachineCoreAddState(t *testing.T) {
	smc := NewStateMachineCore()
	state := layer0.NewState("test-state", layer0.StateTypeInitial, "Test State")
	
	err := smc.AddState(state)
	if err != nil {
		t.Errorf("AddState should not return error: %v", err)
	}
	
	// Try to add the same state again
	err = smc.AddState(state)
	if err == nil {
		t.Error("AddState should return error when adding duplicate state")
	}
	
	// Verify state was added
	retrievedState, err := smc.GetState(state.GetID())
	if err != nil {
		t.Errorf("GetState should not return error: %v", err)
	}
	
	if retrievedState.GetID() != state.GetID() {
		t.Error("Retrieved state should match added state")
	}
}

func TestStateMachineCoreRemoveState(t *testing.T) {
	smc := NewStateMachineCore()
	state := layer0.NewState("test-state", layer0.StateTypeInitial, "Test State")
	
	// Try to remove non-existent state
	err := smc.RemoveState(state.GetID())
	if err == nil {
		t.Error("RemoveState should return error for non-existent state")
	}
	
	// Add state and then remove it
	smc.AddState(state)
	err = smc.RemoveState(state.GetID())
	if err != nil {
		t.Errorf("RemoveState should not return error: %v", err)
	}
	
	// Verify state was removed
	_, err = smc.GetState(state.GetID())
	if err == nil {
		t.Error("GetState should return error for removed state")
	}
}

func TestStateMachineCoreAddTransition(t *testing.T) {
	smc := NewStateMachineCore()
	
	// Add states first
	fromState := layer0.NewState("from-state", layer0.StateTypeInitial, "From State")
	toState := layer0.NewState("to-state", layer0.StateTypeFinal, "To State")
	smc.AddState(fromState)
	smc.AddState(toState)
	
	// Add transition
	transition := layer0.NewTransition("test-transition", layer0.TransitionTypeAutomatic, fromState.GetID(), toState.GetID(), "Test Transition")
	err := smc.AddTransition(transition)
	if err != nil {
		t.Errorf("AddTransition should not return error: %v", err)
	}
	
	// Try to add the same transition again
	err = smc.AddTransition(transition)
	if err == nil {
		t.Error("AddTransition should return error when adding duplicate transition")
	}
	
	// Verify transition was added
	retrievedTransition, err := smc.GetTransition(transition.GetID())
	if err != nil {
		t.Errorf("GetTransition should not return error: %v", err)
	}
	
	if retrievedTransition.GetID() != transition.GetID() {
		t.Error("Retrieved transition should match added transition")
	}
}

func TestStateMachineCoreAddTransitionWithInvalidStates(t *testing.T) {
	smc := NewStateMachineCore()
	
	// Try to add transition without states
	transition := layer0.NewTransition("test-transition", layer0.TransitionTypeAutomatic, "non-existent-from", "non-existent-to", "Test Transition")
	err := smc.AddTransition(transition)
	if err == nil {
		t.Error("AddTransition should return error when states don't exist")
	}
}

func TestStateMachineCoreGetTransitionsFromState(t *testing.T) {
	smc := NewStateMachineCore()
	
	// Add states
	state1 := layer0.NewState("state1", layer0.StateTypeInitial, "State 1")
	state2 := layer0.NewState("state2", layer0.StateTypeIntermediate, "State 2")
	state3 := layer0.NewState("state3", layer0.StateTypeFinal, "State 3")
	smc.AddState(state1)
	smc.AddState(state2)
	smc.AddState(state3)
	
	// Add transitions
	transition1 := layer0.NewTransition("t1", layer0.TransitionTypeAutomatic, state1.GetID(), state2.GetID(), "Transition 1")
	transition2 := layer0.NewTransition("t2", layer0.TransitionTypeAutomatic, state1.GetID(), state3.GetID(), "Transition 2")
	transition3 := layer0.NewTransition("t3", layer0.TransitionTypeAutomatic, state2.GetID(), state3.GetID(), "Transition 3")
	smc.AddTransition(transition1)
	smc.AddTransition(transition2)
	smc.AddTransition(transition3)
	
	// Get transitions from state1
	transitions := smc.GetTransitionsFromState(state1.GetID())
	if len(transitions) != 2 {
		t.Errorf("Expected 2 transitions from state1, got %d", len(transitions))
	}
	
	// Get transitions from state2
	transitions = smc.GetTransitionsFromState(state2.GetID())
	if len(transitions) != 1 {
		t.Errorf("Expected 1 transition from state2, got %d", len(transitions))
	}
	
	// Get transitions from state3 (should be none)
	transitions = smc.GetTransitionsFromState(state3.GetID())
	if len(transitions) != 0 {
		t.Errorf("Expected 0 transitions from state3, got %d", len(transitions))
	}
}

func TestStateMachineCoreGetTransitionsToState(t *testing.T) {
	smc := NewStateMachineCore()
	
	// Add states
	state1 := layer0.NewState("state1", layer0.StateTypeInitial, "State 1")
	state2 := layer0.NewState("state2", layer0.StateTypeIntermediate, "State 2")
	state3 := layer0.NewState("state3", layer0.StateTypeFinal, "State 3")
	smc.AddState(state1)
	smc.AddState(state2)
	smc.AddState(state3)
	
	// Add transitions
	transition1 := layer0.NewTransition("t1", layer0.TransitionTypeAutomatic, state1.GetID(), state2.GetID(), "Transition 1")
	transition2 := layer0.NewTransition("t2", layer0.TransitionTypeAutomatic, state1.GetID(), state3.GetID(), "Transition 2")
	transition3 := layer0.NewTransition("t3", layer0.TransitionTypeAutomatic, state2.GetID(), state3.GetID(), "Transition 3")
	smc.AddTransition(transition1)
	smc.AddTransition(transition2)
	smc.AddTransition(transition3)
	
	// Get transitions to state3
	transitions := smc.GetTransitionsToState(state3.GetID())
	if len(transitions) != 2 {
		t.Errorf("Expected 2 transitions to state3, got %d", len(transitions))
	}
	
	// Get transitions to state2
	transitions = smc.GetTransitionsToState(state2.GetID())
	if len(transitions) != 1 {
		t.Errorf("Expected 1 transition to state2, got %d", len(transitions))
	}
	
	// Get transitions to state1 (should be none)
	transitions = smc.GetTransitionsToState(state1.GetID())
	if len(transitions) != 0 {
		t.Errorf("Expected 0 transitions to state1, got %d", len(transitions))
	}
}

func TestStateMachineCoreCurrentState(t *testing.T) {
	smc := NewStateMachineCore()
	
	// Initially no current state
	_, err := smc.GetCurrentState()
	if err == nil {
		t.Error("GetCurrentState should return error when no current state is set")
	}
	
	// Add a state and set it as current
	state := layer0.NewState("test-state", layer0.StateTypeInitial, "Test State")
	smc.AddState(state)
	
	err = smc.SetCurrentState(state.GetID())
	if err != nil {
		t.Errorf("SetCurrentState should not return error: %v", err)
	}
	
	currentStateID, err := smc.GetCurrentState()
	if err != nil {
		t.Errorf("GetCurrentState should not return error: %v", err)
	}
	
	if *currentStateID != state.GetID() {
		t.Error("Current state should match the set state")
	}
	
	// Try to set non-existent state as current
	err = smc.SetCurrentState("non-existent")
	if err == nil {
		t.Error("SetCurrentState should return error for non-existent state")
	}
}

func TestStateMachineCoreCanTransition(t *testing.T) {
	smc := NewStateMachineCore()
	
	// Add states
	state1 := layer0.NewState("state1", layer0.StateTypeInitial, "State 1")
	state2 := layer0.NewState("state2", layer0.StateTypeFinal, "State 2")
	state3 := layer0.NewState("state3", layer0.StateTypeFinal, "State 3")
	smc.AddState(state1)
	smc.AddState(state2)
	smc.AddState(state3)
	
	// Add transition from state1 to state2
	transition := layer0.NewTransition("t1", layer0.TransitionTypeAutomatic, state1.GetID(), state2.GetID(), "Transition 1")
	smc.AddTransition(transition)
	
	// Test valid transition
	if !smc.CanTransition(state1.GetID(), state2.GetID()) {
		t.Error("Should be able to transition from state1 to state2")
	}
	
	// Test invalid transition
	if smc.CanTransition(state1.GetID(), state3.GetID()) {
		t.Error("Should not be able to transition from state1 to state3")
	}
	
	// Test reverse transition
	if smc.CanTransition(state2.GetID(), state1.GetID()) {
		t.Error("Should not be able to transition from state2 to state1")
	}
}

func TestStateMachineCoreGetAvailableTransitions(t *testing.T) {
	smc := NewStateMachineCore()
	
	// Initially no current state, should return empty
	transitions := smc.GetAvailableTransitions()
	if len(transitions) != 0 {
		t.Errorf("Expected 0 available transitions with no current state, got %d", len(transitions))
	}
	
	// Add states
	state1 := layer0.NewState("state1", layer0.StateTypeInitial, "State 1")
	state2 := layer0.NewState("state2", layer0.StateTypeIntermediate, "State 2")
	state3 := layer0.NewState("state3", layer0.StateTypeFinal, "State 3")
	smc.AddState(state1)
	smc.AddState(state2)
	smc.AddState(state3)
	
	// Add transitions
	transition1 := layer0.NewTransition("t1", layer0.TransitionTypeAutomatic, state1.GetID(), state2.GetID(), "Transition 1")
	transition2 := layer0.NewTransition("t2", layer0.TransitionTypeAutomatic, state1.GetID(), state3.GetID(), "Transition 2")
	smc.AddTransition(transition1)
	smc.AddTransition(transition2)
	
	// Set current state to state1
	smc.SetCurrentState(state1.GetID())
	
	// Get available transitions
	transitions = smc.GetAvailableTransitions()
	if len(transitions) != 2 {
		t.Errorf("Expected 2 available transitions from state1, got %d", len(transitions))
	}
	
	// Set current state to state3 (final state with no outgoing transitions)
	smc.SetCurrentState(state3.GetID())
	transitions = smc.GetAvailableTransitions()
	if len(transitions) != 0 {
		t.Errorf("Expected 0 available transitions from state3, got %d", len(transitions))
	}
}

func TestStateMachineCoreValidateStateMachine(t *testing.T) {
	smc := NewStateMachineCore()
	
	// Empty state machine should be invalid
	err := smc.ValidateStateMachine()
	if err == nil {
		t.Error("Empty state machine should be invalid")
	}
	
	// Add a non-initial state
	state := layer0.NewState("test-state", layer0.StateTypeIntermediate, "Test State")
	smc.AddState(state)
	
	// Should be invalid without initial state
	err = smc.ValidateStateMachine()
	if err == nil {
		t.Error("State machine without initial state should be invalid")
	}
	
	// Add initial state
	initialState := layer0.NewState("initial-state", layer0.StateTypeInitial, "Initial State")
	smc.AddState(initialState)
	
	// Should be valid now
	err = smc.ValidateStateMachine()
	if err != nil {
		t.Errorf("Valid state machine should not return error: %v", err)
	}
	
	// Add transition with valid states
	transition := layer0.NewTransition("test-transition", layer0.TransitionTypeAutomatic, initialState.GetID(), state.GetID(), "Test Transition")
	smc.AddTransition(transition)
	
	// Should still be valid
	err = smc.ValidateStateMachine()
	if err != nil {
		t.Errorf("Valid state machine with transitions should not return error: %v", err)
	}
}

func TestStateMachineCoreRemoveStateWithTransitions(t *testing.T) {
	smc := NewStateMachineCore()
	
	// Add states
	state1 := layer0.NewState("state1", layer0.StateTypeInitial, "State 1")
	state2 := layer0.NewState("state2", layer0.StateTypeFinal, "State 2")
	smc.AddState(state1)
	smc.AddState(state2)
	
	// Add transition
	transition := layer0.NewTransition("t1", layer0.TransitionTypeAutomatic, state1.GetID(), state2.GetID(), "Transition 1")
	smc.AddTransition(transition)
	
	// Try to remove state1 (should fail because it's referenced by transition)
	err := smc.RemoveState(state1.GetID())
	if err == nil {
		t.Error("Should not be able to remove state referenced by transition")
	}
	
	// Try to remove state2 (should fail because it's referenced by transition)
	err = smc.RemoveState(state2.GetID())
	if err == nil {
		t.Error("Should not be able to remove state referenced by transition")
	}
	
	// Remove transition first
	smc.RemoveTransition(transition.GetID())
	
	// Now should be able to remove states
	err = smc.RemoveState(state1.GetID())
	if err != nil {
		t.Errorf("Should be able to remove state after removing referencing transition: %v", err)
	}
}

func TestStateMachineCoreRemoveCurrentState(t *testing.T) {
	smc := NewStateMachineCore()
	
	// Add state and set as current
	state := layer0.NewState("test-state", layer0.StateTypeInitial, "Test State")
	smc.AddState(state)
	smc.SetCurrentState(state.GetID())
	
	// Try to remove current state (should fail)
	err := smc.RemoveState(state.GetID())
	if err == nil {
		t.Error("Should not be able to remove current state")
	}
}
