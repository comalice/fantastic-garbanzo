package layer1

import (
	"errors"
	"testing"
	"time"

	"github.com/ubom/workflow/layer0"
)

func TestNewConditionEvaluationCore(t *testing.T) {
	cec := NewConditionEvaluationCore()

	if cec == nil {
		t.Error("NewConditionEvaluationCore should return a non-nil instance")
	}

	supportedTypes := cec.GetSupportedConditionTypes()
	if len(supportedTypes) != 0 {
		t.Errorf("New condition evaluation core should have 0 supported types, got %d", len(supportedTypes))
	}
}

func TestConditionEvaluationCoreRegisterEvaluator(t *testing.T) {
	cec := NewConditionEvaluationCore()
	evaluator := NewMockConditionEvaluator([]layer0.ConditionType{layer0.ConditionTypeExpression}, nil)

	err := cec.RegisterEvaluator(layer0.ConditionTypeExpression, evaluator)
	if err != nil {
		t.Errorf("RegisterEvaluator should not return error: %v", err)
	}

	// Try to register the same type again
	err = cec.RegisterEvaluator(layer0.ConditionTypeExpression, evaluator)
	if err == nil {
		t.Error("RegisterEvaluator should return error when registering duplicate type")
	}

	// Try to register nil evaluator
	err = cec.RegisterEvaluator(layer0.ConditionTypeScript, nil)
	if err == nil {
		t.Error("RegisterEvaluator should return error for nil evaluator")
	}

	// Verify evaluator was registered
	retrievedEvaluator, err := cec.GetEvaluator(layer0.ConditionTypeExpression)
	if err != nil {
		t.Errorf("GetEvaluator should not return error: %v", err)
	}

	if retrievedEvaluator != evaluator {
		t.Error("Retrieved evaluator should match registered evaluator")
	}
}

func TestConditionEvaluationCoreUnregisterEvaluator(t *testing.T) {
	cec := NewConditionEvaluationCore()
	evaluator := NewMockConditionEvaluator([]layer0.ConditionType{layer0.ConditionTypeExpression}, nil)

	// Try to unregister non-existent evaluator
	err := cec.UnregisterEvaluator(layer0.ConditionTypeExpression)
	if err == nil {
		t.Error("UnregisterEvaluator should return error for non-existent evaluator")
	}

	// Register and then unregister
	cec.RegisterEvaluator(layer0.ConditionTypeExpression, evaluator)
	err = cec.UnregisterEvaluator(layer0.ConditionTypeExpression)
	if err != nil {
		t.Errorf("UnregisterEvaluator should not return error: %v", err)
	}

	// Verify evaluator was unregistered
	_, err = cec.GetEvaluator(layer0.ConditionTypeExpression)
	if err == nil {
		t.Error("GetEvaluator should return error for unregistered evaluator")
	}
}

func TestConditionEvaluationCoreGetSupportedConditionTypes(t *testing.T) {
	cec := NewConditionEvaluationCore()

	// Initially no supported types
	types := cec.GetSupportedConditionTypes()
	if len(types) != 0 {
		t.Errorf("Expected 0 supported types, got %d", len(types))
	}

	// Register evaluators
	evaluator1 := NewMockConditionEvaluator([]layer0.ConditionType{layer0.ConditionTypeExpression}, nil)
	evaluator2 := NewMockConditionEvaluator([]layer0.ConditionType{layer0.ConditionTypeScript}, nil)
	cec.RegisterEvaluator(layer0.ConditionTypeExpression, evaluator1)
	cec.RegisterEvaluator(layer0.ConditionTypeScript, evaluator2)

	types = cec.GetSupportedConditionTypes()
	if len(types) != 2 {
		t.Errorf("Expected 2 supported types, got %d", len(types))
	}

	// Check that both types are present
	typeMap := make(map[layer0.ConditionType]bool)
	for _, conditionType := range types {
		typeMap[conditionType] = true
	}

	if !typeMap[layer0.ConditionTypeExpression] || !typeMap[layer0.ConditionTypeScript] {
		t.Error("Expected both ConditionTypeExpression and ConditionTypeScript to be supported")
	}
}

func TestConditionEvaluationCoreEvaluateCondition(t *testing.T) {
	cec := NewConditionEvaluationCore()

	// Create condition and context
	condition := layer0.NewCondition("test-condition", layer0.ConditionTypeExpression, "Test Condition")
	condition.Expression.Expression = "true"
	context := layer0.NewContext("test-context", layer0.ContextScopeWork, "Test Context")

	// Try to evaluate without registered evaluator
	_, err := cec.EvaluateCondition(condition, context)
	if err == nil {
		t.Error("EvaluateCondition should return error when no evaluator is registered")
	}

	// Register evaluator
	evaluator := NewMockConditionEvaluator([]layer0.ConditionType{layer0.ConditionTypeExpression}, func(c layer0.Condition, ctx layer0.Context) (interface{}, error) {
		return true, nil
	})
	cec.RegisterEvaluator(layer0.ConditionTypeExpression, evaluator)

	// Evaluate condition
	result, err := cec.EvaluateCondition(condition, context)
	if err != nil {
		t.Errorf("EvaluateCondition should not return error: %v", err)
	}

	if result.ConditionID != condition.GetID() {
		t.Error("Result should have correct condition ID")
	}

	if result.Status != layer0.ConditionStatusTrue {
		t.Errorf("Expected status %s, got %s", layer0.ConditionStatusTrue, result.Status)
	}

	if result.Result != true {
		t.Errorf("Expected result true, got %v", result.Result)
	}

	if result.Duration <= 0 {
		t.Error("Duration should be positive")
	}
}

func TestConditionEvaluationCoreEvaluateConditionWithError(t *testing.T) {
	cec := NewConditionEvaluationCore()

	// Create condition and context
	condition := layer0.NewCondition("test-condition", layer0.ConditionTypeExpression, "Test Condition")
	condition.Expression.Expression = "invalid"
	context := layer0.NewContext("test-context", layer0.ContextScopeWork, "Test Context")

	// Register evaluator that returns error
	evaluator := NewMockConditionEvaluator([]layer0.ConditionType{layer0.ConditionTypeExpression}, func(c layer0.Condition, ctx layer0.Context) (interface{}, error) {
		return nil, errors.New("evaluation failed")
	})
	cec.RegisterEvaluator(layer0.ConditionTypeExpression, evaluator)

	// Evaluate condition
	result, err := cec.EvaluateCondition(condition, context)
	if err != nil {
		t.Errorf("EvaluateCondition should not return error even when evaluation fails: %v", err)
	}

	if result.Status != layer0.ConditionStatusError {
		t.Errorf("Expected status %s, got %s", layer0.ConditionStatusError, result.Status)
	}

	if result.Error != "evaluation failed" {
		t.Errorf("Expected error 'evaluation failed', got %s", result.Error)
	}

	if result.Result != nil {
		t.Error("Result should be nil for failed evaluation")
	}
}

func TestConditionEvaluationCoreEvaluateConditionWithDifferentResults(t *testing.T) {
	cec := NewConditionEvaluationCore()
	context := layer0.NewContext("test-context", layer0.ContextScopeWork, "Test Context")

	// Test boolean false result
	evaluator := NewMockConditionEvaluator([]layer0.ConditionType{layer0.ConditionTypeExpression}, func(c layer0.Condition, ctx layer0.Context) (interface{}, error) {
		return false, nil
	})
	cec.RegisterEvaluator(layer0.ConditionTypeExpression, evaluator)

	condition := layer0.NewCondition("test-false", layer0.ConditionTypeExpression, "Test False")
	condition.Expression.Expression = "false"
	result, _ := cec.EvaluateCondition(condition, context)

	if result.Status != layer0.ConditionStatusFalse {
		t.Errorf("Expected status %s for false result, got %s", layer0.ConditionStatusFalse, result.Status)
	}

	// Test nil result
	cec.UnregisterEvaluator(layer0.ConditionTypeExpression)
	evaluator = NewMockConditionEvaluator([]layer0.ConditionType{layer0.ConditionTypeExpression}, func(c layer0.Condition, ctx layer0.Context) (interface{}, error) {
		return nil, nil
	})
	cec.RegisterEvaluator(layer0.ConditionTypeExpression, evaluator)

	condition = layer0.NewCondition("test-nil", layer0.ConditionTypeExpression, "Test Nil")
	condition.Expression.Expression = "nil"
	result, _ = cec.EvaluateCondition(condition, context)

	if result.Status != layer0.ConditionStatusFalse {
		t.Errorf("Expected status %s for nil result, got %s", layer0.ConditionStatusFalse, result.Status)
	}

	// Test non-boolean result
	cec.UnregisterEvaluator(layer0.ConditionTypeExpression)
	evaluator = NewMockConditionEvaluator([]layer0.ConditionType{layer0.ConditionTypeExpression}, func(c layer0.Condition, ctx layer0.Context) (interface{}, error) {
		return "non-empty string", nil
	})
	cec.RegisterEvaluator(layer0.ConditionTypeExpression, evaluator)

	condition = layer0.NewCondition("test-string", layer0.ConditionTypeExpression, "Test String")
	condition.Expression.Expression = "string"
	result, _ = cec.EvaluateCondition(condition, context)

	if result.Status != layer0.ConditionStatusTrue {
		t.Errorf("Expected status %s for non-boolean result, got %s", layer0.ConditionStatusTrue, result.Status)
	}
}

func TestConditionEvaluationCoreEvaluateAlreadyActiveCondition(t *testing.T) {
	cec := NewConditionEvaluationCore()

	// Create condition and context
	condition := layer0.NewCondition("test-condition", layer0.ConditionTypeExpression, "Test Condition")
	condition.Expression.Expression = "true"
	context := layer0.NewContext("test-context", layer0.ContextScopeWork, "Test Context")

	// Register evaluator that takes time
	evaluator := NewMockConditionEvaluator([]layer0.ConditionType{layer0.ConditionTypeExpression}, func(c layer0.Condition, ctx layer0.Context) (interface{}, error) {
		time.Sleep(100 * time.Millisecond)
		return true, nil
	})
	cec.RegisterEvaluator(layer0.ConditionTypeExpression, evaluator)

	// Start evaluation in goroutine
	go cec.EvaluateCondition(condition, context)

	// Wait a bit to ensure evaluation starts
	time.Sleep(10 * time.Millisecond)

	// Try to evaluate the same condition again
	_, err := cec.EvaluateCondition(condition, context)
	if err == nil {
		t.Error("EvaluateCondition should return error when condition is already being evaluated")
	}
}

func TestConditionEvaluationCoreEvaluateConditions(t *testing.T) {
	cec := NewConditionEvaluationCore()
	context := layer0.NewContext("test-context", layer0.ContextScopeWork, "Test Context")

	// Register evaluator
	evaluator := NewMockConditionEvaluator([]layer0.ConditionType{layer0.ConditionTypeExpression}, func(c layer0.Condition, ctx layer0.Context) (interface{}, error) {
		// Return true for conditions with "true" in expression, false otherwise
		return c.GetExpression().Expression == "true", nil
	})
	cec.RegisterEvaluator(layer0.ConditionTypeExpression, evaluator)

	// Create conditions
	condition1 := layer0.NewCondition("cond1", layer0.ConditionTypeExpression, "Condition 1")
	condition1.Expression.Expression = "true"
	condition2 := layer0.NewCondition("cond2", layer0.ConditionTypeExpression, "Condition 2")
	condition2.Expression.Expression = "true"
	condition3 := layer0.NewCondition("cond3", layer0.ConditionTypeExpression, "Condition 3")
	condition3.Expression.Expression = "false"

	// Test AND operator with all true conditions
	result, err := cec.EvaluateConditions([]layer0.Condition{condition1, condition2}, context, layer0.ConditionOperatorAnd)
	if err != nil {
		t.Errorf("EvaluateConditions should not return error: %v", err)
	}
	if !result {
		t.Error("AND with all true conditions should return true")
	}

	// Test AND operator with one false condition
	result, err = cec.EvaluateConditions([]layer0.Condition{condition1, condition3}, context, layer0.ConditionOperatorAnd)
	if err != nil {
		t.Errorf("EvaluateConditions should not return error: %v", err)
	}
	if result {
		t.Error("AND with one false condition should return false")
	}

	// Test OR operator with one true condition
	result, err = cec.EvaluateConditions([]layer0.Condition{condition1, condition3}, context, layer0.ConditionOperatorOr)
	if err != nil {
		t.Errorf("EvaluateConditions should not return error: %v", err)
	}
	if !result {
		t.Error("OR with one true condition should return true")
	}

	// Test OR operator with all false conditions
	result, err = cec.EvaluateConditions([]layer0.Condition{condition3}, context, layer0.ConditionOperatorOr)
	if err != nil {
		t.Errorf("EvaluateConditions should not return error: %v", err)
	}
	if result {
		t.Error("OR with all false conditions should return false")
	}

	// Test NOT operator with true condition
	result, err = cec.EvaluateConditions([]layer0.Condition{condition1}, context, layer0.ConditionOperatorNot)
	if err != nil {
		t.Errorf("EvaluateConditions should not return error: %v", err)
	}
	if result {
		t.Error("NOT with true condition should return false")
	}

	// Test NOT operator with false condition
	result, err = cec.EvaluateConditions([]layer0.Condition{condition3}, context, layer0.ConditionOperatorNot)
	if err != nil {
		t.Errorf("EvaluateConditions should not return error: %v", err)
	}
	if !result {
		t.Error("NOT with false condition should return true")
	}

	// Test NOT operator with multiple conditions (should fail)
	_, err = cec.EvaluateConditions([]layer0.Condition{condition1, condition2}, context, layer0.ConditionOperatorNot)
	if err == nil {
		t.Error("NOT operator with multiple conditions should return error")
	}

	// Test empty conditions (should return true)
	result, err = cec.EvaluateConditions([]layer0.Condition{}, context, layer0.ConditionOperatorAnd)
	if err != nil {
		t.Errorf("EvaluateConditions should not return error for empty conditions: %v", err)
	}
	if !result {
		t.Error("Empty conditions should return true")
	}
}

func TestConditionEvaluationCoreGetEvaluationResult(t *testing.T) {
	cec := NewConditionEvaluationCore()

	// Try to get non-existent result
	_, err := cec.GetEvaluationResult("non-existent")
	if err == nil {
		t.Error("GetEvaluationResult should return error for non-existent condition")
	}

	// Create condition and context
	condition := layer0.NewCondition("test-condition", layer0.ConditionTypeExpression, "Test Condition")
	condition.Expression.Expression = "true"
	context := layer0.NewContext("test-context", layer0.ContextScopeWork, "Test Context")

	// Register evaluator
	evaluator := NewMockConditionEvaluator([]layer0.ConditionType{layer0.ConditionTypeExpression}, func(c layer0.Condition, ctx layer0.Context) (interface{}, error) {
		return true, nil
	})
	cec.RegisterEvaluator(layer0.ConditionTypeExpression, evaluator)

	// Evaluate condition
	cec.EvaluateCondition(condition, context)

	// Get evaluation result
	result, err := cec.GetEvaluationResult(condition.GetID())
	if err != nil {
		t.Errorf("GetEvaluationResult should not return error: %v", err)
	}

	if result.ConditionID != condition.GetID() {
		t.Error("Result should have correct condition ID")
	}

	if result.Status != layer0.ConditionStatusTrue {
		t.Error("Result should have true status")
	}
}

func TestConditionEvaluationCoreGetAllEvaluationResults(t *testing.T) {
	cec := NewConditionEvaluationCore()

	// Initially no results
	results := cec.GetAllEvaluationResults()
	if len(results) != 0 {
		t.Errorf("Expected 0 evaluation results, got %d", len(results))
	}

	// Evaluate multiple conditions
	condition1 := layer0.NewCondition("cond1", layer0.ConditionTypeExpression, "Condition 1")
	condition1.Expression.Expression = "true"
	condition2 := layer0.NewCondition("cond2", layer0.ConditionTypeExpression, "Condition 2")
	condition2.Expression.Expression = "true"
	context := layer0.NewContext("test-context", layer0.ContextScopeWork, "Test Context")

	evaluator := NewMockConditionEvaluator([]layer0.ConditionType{layer0.ConditionTypeExpression}, func(c layer0.Condition, ctx layer0.Context) (interface{}, error) {
		return true, nil
	})
	cec.RegisterEvaluator(layer0.ConditionTypeExpression, evaluator)

	cec.EvaluateCondition(condition1, context)
	cec.EvaluateCondition(condition2, context)

	// Get all results
	results = cec.GetAllEvaluationResults()
	if len(results) != 2 {
		t.Errorf("Expected 2 evaluation results, got %d", len(results))
	}
}

func TestConditionEvaluationCoreIsConditionEvaluating(t *testing.T) {
	cec := NewConditionEvaluationCore()

	// Initially no condition is evaluating
	if cec.IsConditionEvaluating("test-condition") {
		t.Error("IsConditionEvaluating should return false for non-existent condition")
	}

	// Create condition and context
	condition := layer0.NewCondition("test-condition", layer0.ConditionTypeExpression, "Test Condition")
	condition.Expression.Expression = "true"
	context := layer0.NewContext("test-context", layer0.ContextScopeWork, "Test Context")

	// Register evaluator that takes time
	evaluator := NewMockConditionEvaluator([]layer0.ConditionType{layer0.ConditionTypeExpression}, func(c layer0.Condition, ctx layer0.Context) (interface{}, error) {
		time.Sleep(100 * time.Millisecond)
		return true, nil
	})
	cec.RegisterEvaluator(layer0.ConditionTypeExpression, evaluator)

	// Start evaluation in goroutine
	go cec.EvaluateCondition(condition, context)

	// Wait a bit to ensure evaluation starts
	time.Sleep(10 * time.Millisecond)

	// Check that condition is evaluating
	if !cec.IsConditionEvaluating(condition.GetID()) {
		t.Error("IsConditionEvaluating should return true for active condition")
	}

	// Wait for evaluation to complete
	time.Sleep(200 * time.Millisecond)

	// Check that condition is no longer evaluating
	if cec.IsConditionEvaluating(condition.GetID()) {
		t.Error("IsConditionEvaluating should return false after condition evaluation")
	}
}

func TestConditionEvaluationCoreGetActiveEvaluations(t *testing.T) {
	cec := NewConditionEvaluationCore()

	// Initially no active evaluations
	activeEvaluations := cec.GetActiveEvaluations()
	if len(activeEvaluations) != 0 {
		t.Errorf("Expected 0 active evaluations, got %d", len(activeEvaluations))
	}

	// Create condition and context
	condition := layer0.NewCondition("test-condition", layer0.ConditionTypeExpression, "Test Condition")
	condition.Expression.Expression = "true"
	context := layer0.NewContext("test-context", layer0.ContextScopeWork, "Test Context")

	// Register evaluator that takes time
	evaluator := NewMockConditionEvaluator([]layer0.ConditionType{layer0.ConditionTypeExpression}, func(c layer0.Condition, ctx layer0.Context) (interface{}, error) {
		time.Sleep(100 * time.Millisecond)
		return true, nil
	})
	cec.RegisterEvaluator(layer0.ConditionTypeExpression, evaluator)

	// Start evaluation in goroutine
	go cec.EvaluateCondition(condition, context)

	// Wait a bit to ensure evaluation starts
	time.Sleep(10 * time.Millisecond)

	// Check active evaluations
	activeEvaluations = cec.GetActiveEvaluations()
	if len(activeEvaluations) != 1 {
		t.Errorf("Expected 1 active evaluation, got %d", len(activeEvaluations))
	}

	if activeEvaluations[0].GetID() != condition.GetID() {
		t.Error("Active evaluation should match the evaluating condition")
	}

	// Wait for evaluation to complete
	time.Sleep(200 * time.Millisecond)

	// Should be no active evaluations now
	activeEvaluations = cec.GetActiveEvaluations()
	if len(activeEvaluations) != 0 {
		t.Errorf("Expected 0 active evaluations after completion, got %d", len(activeEvaluations))
	}
}

func TestMockConditionEvaluator(t *testing.T) {
	supportedTypes := []layer0.ConditionType{layer0.ConditionTypeExpression, layer0.ConditionTypeScript}
	evaluateFunc := func(condition layer0.Condition, context layer0.Context) (interface{}, error) {
		return "custom result", nil
	}

	evaluator := NewMockConditionEvaluator(supportedTypes, evaluateFunc)

	// Test supported types
	if !evaluator.CanEvaluate(layer0.ConditionTypeExpression) {
		t.Error("Evaluator should support ConditionTypeExpression")
	}

	if !evaluator.CanEvaluate(layer0.ConditionTypeScript) {
		t.Error("Evaluator should support ConditionTypeScript")
	}

	if evaluator.CanEvaluate(layer0.ConditionTypeService) {
		t.Error("Evaluator should not support ConditionTypeService")
	}

	// Test get supported types
	types := evaluator.GetSupportedTypes()
	if len(types) != 2 {
		t.Errorf("Expected 2 supported types, got %d", len(types))
	}

	// Test evaluation
	condition := layer0.NewCondition("test", layer0.ConditionTypeExpression, "Test")
	context := layer0.NewContext("test", layer0.ContextScopeWork, "Test")

	result, err := evaluator.Evaluate(condition, context)
	if err != nil {
		t.Errorf("Evaluate should not return error: %v", err)
	}

	if result != "custom result" {
		t.Errorf("Expected 'custom result', got %v", result)
	}

	// Test default evaluation (no custom function)
	defaultEvaluator := NewMockConditionEvaluator(supportedTypes, nil)
	result, err = defaultEvaluator.Evaluate(condition, context)
	if err != nil {
		t.Errorf("Evaluate should not return error: %v", err)
	}

	if result != true {
		t.Errorf("Expected true, got %v", result)
	}
}
