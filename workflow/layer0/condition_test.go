package layer0

import (
	"testing"
	"time"
)

func TestNewCondition(t *testing.T) {
	id := ConditionID("test-condition")
	conditionType := ConditionTypeExpression
	name := "Test Condition"

	condition := NewCondition(id, conditionType, name)

	if condition.GetID() != id {
		t.Errorf("Expected ID %s, got %s", id, condition.GetID())
	}

	if condition.GetType() != conditionType {
		t.Errorf("Expected type %s, got %s", conditionType, condition.GetType())
	}

	if condition.GetStatus() != ConditionStatusPending {
		t.Errorf("Expected status %s, got %s", ConditionStatusPending, condition.GetStatus())
	}

	if condition.GetMetadata().Name != name {
		t.Errorf("Expected name %s, got %s", name, condition.GetMetadata().Name)
	}

	if condition.GetExpression().Language != "javascript" {
		t.Errorf("Expected default language javascript, got %s", condition.GetExpression().Language)
	}
}

func TestConditionSetStatus(t *testing.T) {
	condition := NewCondition("test", ConditionTypeExpression, "Test")
	originalTime := condition.Metadata.UpdatedAt

	time.Sleep(1 * time.Millisecond) // Ensure time difference

	newCondition := condition.SetStatus(ConditionStatusEvaluating)

	if newCondition.GetStatus() != ConditionStatusEvaluating {
		t.Errorf("Expected status %s, got %s", ConditionStatusEvaluating, newCondition.GetStatus())
	}

	if newCondition.Metadata.UpdatedAt.Equal(originalTime) {
		t.Error("UpdatedAt should be updated when status changes")
	}

	// Original condition should remain unchanged (immutability)
	if condition.GetStatus() != ConditionStatusPending {
		t.Error("Original condition should remain unchanged")
	}
}

func TestConditionSetResult(t *testing.T) {
	condition := NewCondition("test", ConditionTypeExpression, "Test")
	testResult := true

	newCondition := condition.SetResult(testResult)

	if newCondition.GetResult() != testResult {
		t.Errorf("Expected result %v, got %v", testResult, newCondition.GetResult())
	}

	// Original condition should remain unchanged (immutability)
	if condition.GetResult() != nil {
		t.Error("Original condition should remain unchanged")
	}
}

func TestConditionSetError(t *testing.T) {
	condition := NewCondition("test", ConditionTypeExpression, "Test")
	errorMsg := "test error"

	newCondition := condition.SetError(errorMsg)

	if newCondition.GetError() != errorMsg {
		t.Errorf("Expected error %s, got %s", errorMsg, newCondition.GetError())
	}

	// Original condition should remain unchanged (immutability)
	if condition.GetError() != "" {
		t.Error("Original condition should remain unchanged")
	}
}

func TestConditionAddDependency(t *testing.T) {
	condition := NewCondition("test", ConditionTypeExpression, "Test")
	dependencyID := ConditionID("dependency-1")

	newCondition := condition.AddDependency(dependencyID)

	dependencies := newCondition.GetDependencies()
	if len(dependencies) != 1 || dependencies[0] != dependencyID {
		t.Errorf("Expected dependency %s to be added", dependencyID)
	}

	// Original condition should remain unchanged (immutability)
	if len(condition.GetDependencies()) != 0 {
		t.Error("Original condition should remain unchanged")
	}
}

func TestConditionMarkEvaluated(t *testing.T) {
	condition := NewCondition("test", ConditionTypeExpression, "Test")

	// Test with boolean true result
	trueCondition := condition.MarkEvaluated(true)
	if !trueCondition.IsTrue() {
		t.Error("Condition with true result should be true")
	}
	if trueCondition.GetMetadata().EvaluatedAt == nil {
		t.Error("EvaluatedAt should be set")
	}

	// Test with boolean false result
	falseCondition := condition.MarkEvaluated(false)
	if !falseCondition.IsFalse() {
		t.Error("Condition with false result should be false")
	}

	// Test with nil result
	nilCondition := condition.MarkEvaluated(nil)
	if !nilCondition.IsFalse() {
		t.Error("Condition with nil result should be false")
	}

	// Test with non-boolean result
	stringCondition := condition.MarkEvaluated("test")
	if !stringCondition.IsTrue() {
		t.Error("Condition with non-nil non-boolean result should be true")
	}
}

func TestConditionMarkFailed(t *testing.T) {
	condition := NewCondition("test", ConditionTypeExpression, "Test")
	errorMsg := "evaluation failed"

	failedCondition := condition.MarkFailed(errorMsg)

	if !failedCondition.IsError() {
		t.Error("Failed condition should be in error state")
	}

	if failedCondition.GetError() != errorMsg {
		t.Errorf("Expected error %s, got %s", errorMsg, failedCondition.GetError())
	}

	if failedCondition.GetMetadata().EvaluatedAt == nil {
		t.Error("EvaluatedAt should be set for failed condition")
	}
}

func TestConditionIsTrue(t *testing.T) {
	condition := NewCondition("test", ConditionTypeExpression, "Test")

	if condition.IsTrue() {
		t.Error("New condition should not be true")
	}

	trueCondition := condition.SetStatus(ConditionStatusTrue)
	if !trueCondition.IsTrue() {
		t.Error("Condition with true status should be true")
	}
}

func TestConditionIsFalse(t *testing.T) {
	condition := NewCondition("test", ConditionTypeExpression, "Test")

	if condition.IsFalse() {
		t.Error("New condition should not be false")
	}

	falseCondition := condition.SetStatus(ConditionStatusFalse)
	if !falseCondition.IsFalse() {
		t.Error("Condition with false status should be false")
	}
}

func TestConditionIsError(t *testing.T) {
	condition := NewCondition("test", ConditionTypeExpression, "Test")

	if condition.IsError() {
		t.Error("New condition should not be in error state")
	}

	errorCondition := condition.SetStatus(ConditionStatusError)
	if !errorCondition.IsError() {
		t.Error("Condition with error status should be in error state")
	}
}

func TestConditionIsPending(t *testing.T) {
	condition := NewCondition("test", ConditionTypeExpression, "Test")

	if !condition.IsPending() {
		t.Error("New condition should be pending")
	}

	evaluatingCondition := condition.SetStatus(ConditionStatusEvaluating)
	if evaluatingCondition.IsPending() {
		t.Error("Evaluating condition should not be pending")
	}
}

func TestConditionIsEvaluated(t *testing.T) {
	condition := NewCondition("test", ConditionTypeExpression, "Test")

	if condition.IsEvaluated() {
		t.Error("New condition should not be evaluated")
	}

	// Test true status
	trueCondition := condition.SetStatus(ConditionStatusTrue)
	if !trueCondition.IsEvaluated() {
		t.Error("True condition should be evaluated")
	}

	// Test false status
	falseCondition := condition.SetStatus(ConditionStatusFalse)
	if !falseCondition.IsEvaluated() {
		t.Error("False condition should be evaluated")
	}

	// Test error status
	errorCondition := condition.SetStatus(ConditionStatusError)
	if !errorCondition.IsEvaluated() {
		t.Error("Error condition should be evaluated")
	}

	// Test evaluating status
	evaluatingCondition := condition.SetStatus(ConditionStatusEvaluating)
	if evaluatingCondition.IsEvaluated() {
		t.Error("Evaluating condition should not be evaluated")
	}
}

func TestConditionClone(t *testing.T) {
	original := NewCondition("test", ConditionTypeExpression, "Test")
	original.Metadata.Tags = []string{"tag1", "tag2"}
	original.Metadata.Properties["key"] = "value"
	original.Expression.Variables["var"] = "value"
	original = original.AddDependency("dep1")

	cloned := original.Clone()

	// Verify clone has same values
	if cloned.GetID() != original.GetID() {
		t.Error("Cloned condition should have same ID")
	}

	if len(cloned.GetDependencies()) != len(original.GetDependencies()) {
		t.Error("Cloned condition should have same dependencies")
	}

	// Verify independence (modify clone)
	cloned.Metadata.Tags[0] = "modified"
	cloned.Metadata.Properties["key"] = "modified"
	cloned.Expression.Variables["var"] = "modified"
	cloned.Dependencies[0] = "modified"

	if original.Metadata.Tags[0] == "modified" {
		t.Error("Original condition tags should not be affected by clone modification")
	}

	if original.Metadata.Properties["key"] == "modified" {
		t.Error("Original condition properties should not be affected by clone modification")
	}

	if original.Expression.Variables["var"] == "modified" {
		t.Error("Original condition variables should not be affected by clone modification")
	}

	if original.Dependencies[0] == "modified" {
		t.Error("Original condition dependencies should not be affected by clone modification")
	}
}

func TestConditionValidate(t *testing.T) {
	// Valid condition
	validCondition := NewCondition("test", ConditionTypeExpression, "Test")
	validCondition.Expression.Expression = "true"
	if err := validCondition.Validate(); err != nil {
		t.Errorf("Valid condition should not return error: %v", err)
	}

	// Invalid conditions
	invalidConditions := []Condition{
		{ID: "", Type: ConditionTypeExpression, Status: ConditionStatusPending, Metadata: ConditionMetadata{Name: "Test"}},
		{ID: "test", Type: "", Status: ConditionStatusPending, Metadata: ConditionMetadata{Name: "Test"}},
		{ID: "test", Type: ConditionTypeExpression, Status: "", Metadata: ConditionMetadata{Name: "Test"}},
		{ID: "test", Type: ConditionTypeExpression, Status: ConditionStatusPending, Metadata: ConditionMetadata{Name: ""}},
		{ID: "test", Type: ConditionTypeExpression, Status: ConditionStatusPending, Metadata: ConditionMetadata{Name: "Test"}, Expression: ConditionExpression{Expression: ""}},
	}

	for i, condition := range invalidConditions {
		if err := condition.Validate(); err == nil {
			t.Errorf("Invalid condition %d should return error", i)
		}
	}
}
