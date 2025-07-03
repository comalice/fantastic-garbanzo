package layer0

import (
	"testing"
	"time"
)

func TestNewTransition(t *testing.T) {
	id := TransitionID("test-transition")
	transitionType := TransitionTypeAutomatic
	fromStateID := StateID("from-state")
	toStateID := StateID("to-state")
	name := "Test Transition"

	transition := NewTransition(id, transitionType, fromStateID, toStateID, name)

	if transition.GetID() != id {
		t.Errorf("Expected ID %s, got %s", id, transition.GetID())
	}

	if transition.GetType() != transitionType {
		t.Errorf("Expected type %s, got %s", transitionType, transition.GetType())
	}

	if transition.GetFromStateID() != fromStateID {
		t.Errorf("Expected from state ID %s, got %s", fromStateID, transition.GetFromStateID())
	}

	if transition.GetToStateID() != toStateID {
		t.Errorf("Expected to state ID %s, got %s", toStateID, transition.GetToStateID())
	}

	if transition.GetStatus() != TransitionStatusPending {
		t.Errorf("Expected status %s, got %s", TransitionStatusPending, transition.GetStatus())
	}

	if transition.GetMetadata().Name != name {
		t.Errorf("Expected name %s, got %s", name, transition.GetMetadata().Name)
	}
}

func TestTransitionSetStatus(t *testing.T) {
	transition := NewTransition("test", TransitionTypeAutomatic, "from", "to", "Test")
	originalTime := transition.Metadata.UpdatedAt

	time.Sleep(1 * time.Millisecond) // Ensure time difference

	newTransition := transition.SetStatus(TransitionStatusReady)

	if newTransition.GetStatus() != TransitionStatusReady {
		t.Errorf("Expected status %s, got %s", TransitionStatusReady, newTransition.GetStatus())
	}

	if newTransition.Metadata.UpdatedAt.Equal(originalTime) {
		t.Error("UpdatedAt should be updated when status changes")
	}

	// Original transition should remain unchanged (immutability)
	if transition.GetStatus() != TransitionStatusPending {
		t.Error("Original transition should remain unchanged")
	}
}

func TestTransitionSetData(t *testing.T) {
	transition := NewTransition("test", TransitionTypeAutomatic, "from", "to", "Test")
	testData := map[string]interface{}{"key": "value"}

	newTransition := transition.SetData(testData)

	if newTransition.GetData() == nil {
		t.Error("Data should be set")
	}

	// Original transition should remain unchanged (immutability)
	if transition.GetData() != nil {
		t.Error("Original transition should remain unchanged")
	}
}

func TestTransitionAddCondition(t *testing.T) {
	transition := NewTransition("test", TransitionTypeAutomatic, "from", "to", "Test")
	conditionID := "condition-1"

	newTransition := transition.AddCondition(conditionID)

	conditions := newTransition.GetConditions()
	if len(conditions) != 1 || conditions[0] != conditionID {
		t.Errorf("Expected condition %s to be added", conditionID)
	}

	// Original transition should remain unchanged (immutability)
	if len(transition.GetConditions()) != 0 {
		t.Error("Original transition should remain unchanged")
	}
}

func TestTransitionAddAction(t *testing.T) {
	transition := NewTransition("test", TransitionTypeAutomatic, "from", "to", "Test")
	actionID := "action-1"

	newTransition := transition.AddAction(actionID)

	actions := newTransition.GetActions()
	if len(actions) != 1 || actions[0] != actionID {
		t.Errorf("Expected action %s to be added", actionID)
	}

	// Original transition should remain unchanged (immutability)
	if len(transition.GetActions()) != 0 {
		t.Error("Original transition should remain unchanged")
	}
}

func TestTransitionIsReady(t *testing.T) {
	transition := NewTransition("test", TransitionTypeAutomatic, "from", "to", "Test")

	if transition.IsReady() {
		t.Error("New transition should not be ready")
	}

	readyTransition := transition.SetStatus(TransitionStatusReady)
	if !readyTransition.IsReady() {
		t.Error("Transition with ready status should be ready")
	}
}

func TestTransitionIsCompleted(t *testing.T) {
	transition := NewTransition("test", TransitionTypeAutomatic, "from", "to", "Test")

	if transition.IsCompleted() {
		t.Error("New transition should not be completed")
	}

	completedTransition := transition.SetStatus(TransitionStatusCompleted)
	if !completedTransition.IsCompleted() {
		t.Error("Transition with completed status should be completed")
	}
}

func TestTransitionIsFailed(t *testing.T) {
	transition := NewTransition("test", TransitionTypeAutomatic, "from", "to", "Test")

	if transition.IsFailed() {
		t.Error("New transition should not be failed")
	}

	failedTransition := transition.SetStatus(TransitionStatusFailed)
	if !failedTransition.IsFailed() {
		t.Error("Transition with failed status should be failed")
	}
}

func TestTransitionClone(t *testing.T) {
	original := NewTransition("test", TransitionTypeAutomatic, "from", "to", "Test")
	original.Metadata.Tags = []string{"tag1", "tag2"}
	original.Metadata.Properties["key"] = "value"
	original = original.AddCondition("condition1").AddAction("action1")

	cloned := original.Clone()

	// Verify clone has same values
	if cloned.GetID() != original.GetID() {
		t.Error("Cloned transition should have same ID")
	}

	if len(cloned.GetConditions()) != len(original.GetConditions()) {
		t.Error("Cloned transition should have same conditions")
	}

	if len(cloned.GetActions()) != len(original.GetActions()) {
		t.Error("Cloned transition should have same actions")
	}

	// Verify independence (modify clone)
	cloned.Metadata.Tags[0] = "modified"
	cloned.Metadata.Properties["key"] = "modified"
	cloned.Conditions[0] = "modified"
	cloned.Actions[0] = "modified"

	if original.Metadata.Tags[0] == "modified" {
		t.Error("Original transition tags should not be affected by clone modification")
	}

	if original.Metadata.Properties["key"] == "modified" {
		t.Error("Original transition properties should not be affected by clone modification")
	}

	if original.Conditions[0] == "modified" {
		t.Error("Original transition conditions should not be affected by clone modification")
	}

	if original.Actions[0] == "modified" {
		t.Error("Original transition actions should not be affected by clone modification")
	}
}

func TestTransitionValidate(t *testing.T) {
	// Valid transition
	validTransition := NewTransition("test", TransitionTypeAutomatic, "from", "to", "Test")
	if err := validTransition.Validate(); err != nil {
		t.Errorf("Valid transition should not return error: %v", err)
	}

	// Invalid transitions
	invalidTransitions := []Transition{
		{ID: "", Type: TransitionTypeAutomatic, Status: TransitionStatusPending, FromStateID: "from", ToStateID: "to", Metadata: TransitionMetadata{Name: "Test"}},
		{ID: "test", Type: "", Status: TransitionStatusPending, FromStateID: "from", ToStateID: "to", Metadata: TransitionMetadata{Name: "Test"}},
		{ID: "test", Type: TransitionTypeAutomatic, Status: "", FromStateID: "from", ToStateID: "to", Metadata: TransitionMetadata{Name: "Test"}},
		{ID: "test", Type: TransitionTypeAutomatic, Status: TransitionStatusPending, FromStateID: "", ToStateID: "to", Metadata: TransitionMetadata{Name: "Test"}},
		{ID: "test", Type: TransitionTypeAutomatic, Status: TransitionStatusPending, FromStateID: "from", ToStateID: "", Metadata: TransitionMetadata{Name: "Test"}},
		{ID: "test", Type: TransitionTypeAutomatic, Status: TransitionStatusPending, FromStateID: "from", ToStateID: "to", Metadata: TransitionMetadata{Name: ""}},
	}

	for i, transition := range invalidTransitions {
		if err := transition.Validate(); err == nil {
			t.Errorf("Invalid transition %d should return error", i)
		}
	}
}
