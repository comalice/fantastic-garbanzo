package layer1

import (
	"fmt"
	"sync"
	"time"

	"github.com/ubom/workflow/layer0"
)

// ConditionEvaluator defines the interface for evaluating conditions
type ConditionEvaluator interface {
	Evaluate(condition layer0.Condition, context *layer0.Context) (interface{}, error)
	CanEvaluate(conditionType layer0.ConditionType) bool
	GetSupportedTypes() []layer0.ConditionType
}

// ConditionEvaluationResult represents the result of condition evaluation
type ConditionEvaluationResult struct {
	ConditionID layer0.ConditionID     `json:"condition_id"`
	Status      layer0.ConditionStatus `json:"status"`
	Result      interface{}            `json:"result"`
	Error       string                 `json:"error,omitempty"`
	EvaluatedAt time.Time              `json:"evaluated_at"`
	Duration    time.Duration          `json:"duration"`
}

// ConditionEvaluationCore provides core condition evaluation functionality
type ConditionEvaluationCore struct {
	evaluators        map[layer0.ConditionType]ConditionEvaluator
	evaluationResults map[layer0.ConditionID]ConditionEvaluationResult
	activeEvaluations map[layer0.ConditionID]layer0.Condition
	mutex             sync.RWMutex
}

// ConditionEvaluationCoreInterface defines the contract for condition evaluation operations
type ConditionEvaluationCoreInterface interface {
	RegisterEvaluator(conditionType layer0.ConditionType, evaluator ConditionEvaluator) error
	UnregisterEvaluator(conditionType layer0.ConditionType) error
	GetEvaluator(conditionType layer0.ConditionType) (ConditionEvaluator, error)
	GetSupportedConditionTypes() []layer0.ConditionType
	EvaluateCondition(condition layer0.Condition, context *layer0.Context) (ConditionEvaluationResult, error)
	EvaluateConditions(conditions []layer0.Condition, context *layer0.Context, operator layer0.ConditionOperator) (bool, error)
	GetEvaluationResult(conditionID layer0.ConditionID) (ConditionEvaluationResult, error)
	GetAllEvaluationResults() []ConditionEvaluationResult
	IsConditionEvaluating(conditionID layer0.ConditionID) bool
	GetActiveEvaluations() []layer0.Condition
}

// NewConditionEvaluationCore creates a new condition evaluation core
func NewConditionEvaluationCore() *ConditionEvaluationCore {
	return &ConditionEvaluationCore{
		evaluators:        make(map[layer0.ConditionType]ConditionEvaluator),
		evaluationResults: make(map[layer0.ConditionID]ConditionEvaluationResult),
		activeEvaluations: make(map[layer0.ConditionID]layer0.Condition),
		mutex:             sync.RWMutex{},
	}
}

// RegisterEvaluator registers a condition evaluator for a specific condition type
func (cec *ConditionEvaluationCore) RegisterEvaluator(conditionType layer0.ConditionType, evaluator ConditionEvaluator) error {
	if evaluator == nil {
		return fmt.Errorf("evaluator cannot be nil")
	}

	cec.mutex.Lock()
	defer cec.mutex.Unlock()

	if _, exists := cec.evaluators[conditionType]; exists {
		return fmt.Errorf("evaluator for condition type %s already registered", conditionType)
	}

	cec.evaluators[conditionType] = evaluator
	return nil
}

// UnregisterEvaluator unregisters a condition evaluator for a specific condition type
func (cec *ConditionEvaluationCore) UnregisterEvaluator(conditionType layer0.ConditionType) error {
	cec.mutex.Lock()
	defer cec.mutex.Unlock()

	if _, exists := cec.evaluators[conditionType]; !exists {
		return fmt.Errorf("no evaluator registered for condition type %s", conditionType)
	}

	delete(cec.evaluators, conditionType)
	return nil
}

// GetEvaluator retrieves the evaluator for a specific condition type
func (cec *ConditionEvaluationCore) GetEvaluator(conditionType layer0.ConditionType) (ConditionEvaluator, error) {
	cec.mutex.RLock()
	defer cec.mutex.RUnlock()

	evaluator, exists := cec.evaluators[conditionType]
	if !exists {
		return nil, fmt.Errorf("no evaluator registered for condition type %s", conditionType)
	}

	return evaluator, nil
}

// GetSupportedConditionTypes returns all supported condition types
func (cec *ConditionEvaluationCore) GetSupportedConditionTypes() []layer0.ConditionType {
	cec.mutex.RLock()
	defer cec.mutex.RUnlock()

	types := make([]layer0.ConditionType, 0, len(cec.evaluators))
	for conditionType := range cec.evaluators {
		types = append(types, conditionType)
	}

	return types
}

// EvaluateCondition evaluates a single condition using the appropriate evaluator
func (cec *ConditionEvaluationCore) EvaluateCondition(condition layer0.Condition, context *layer0.Context) (ConditionEvaluationResult, error) {
	if err := condition.Validate(); err != nil {
		return ConditionEvaluationResult{}, fmt.Errorf("invalid condition: %w", err)
	}

	cec.mutex.Lock()

	// Check if condition is already being evaluated
	if _, isActive := cec.activeEvaluations[condition.GetID()]; isActive {
		cec.mutex.Unlock()
		return ConditionEvaluationResult{}, fmt.Errorf("condition %s is already being evaluated", condition.GetID())
	}

	// Get evaluator
	evaluator, exists := cec.evaluators[condition.GetType()]
	if !exists {
		cec.mutex.Unlock()
		return ConditionEvaluationResult{}, fmt.Errorf("no evaluator registered for condition type %s", condition.GetType())
	}

	// Mark condition as being evaluated
	evaluatingCondition := condition.SetStatus(layer0.ConditionStatusEvaluating)
	cec.activeEvaluations[condition.GetID()] = evaluatingCondition
	cec.mutex.Unlock()

	// Evaluate condition
	startTime := time.Now()
	result, err := evaluator.Evaluate(condition, context)
	endTime := time.Now()
	duration := endTime.Sub(startTime)

	cec.mutex.Lock()
	defer cec.mutex.Unlock()

	// Remove from active evaluations
	delete(cec.activeEvaluations, condition.GetID())

	// Create evaluation result
	evalResult := ConditionEvaluationResult{
		ConditionID: condition.GetID(),
		EvaluatedAt: endTime,
		Duration:    duration,
	}

	if err != nil {
		evalResult.Status = layer0.ConditionStatusError
		evalResult.Error = err.Error()
	} else {
		evalResult.Result = result
		// Determine status based on result
		if result == nil {
			evalResult.Status = layer0.ConditionStatusFalse
		} else if boolResult, ok := result.(bool); ok {
			if boolResult {
				evalResult.Status = layer0.ConditionStatusTrue
			} else {
				evalResult.Status = layer0.ConditionStatusFalse
			}
		} else {
			// Non-boolean results are considered true if not nil/empty
			evalResult.Status = layer0.ConditionStatusTrue
		}
	}

	// Store result
	cec.evaluationResults[condition.GetID()] = evalResult

	return evalResult, nil
}

// EvaluateConditions evaluates multiple conditions with a logical operator
func (cec *ConditionEvaluationCore) EvaluateConditions(conditions []layer0.Condition, context *layer0.Context, operator layer0.ConditionOperator) (bool, error) {
	if len(conditions) == 0 {
		return true, nil // Empty condition list is considered true
	}

	results := make([]bool, len(conditions))

	// Evaluate all conditions
	for i, condition := range conditions {
		evalResult, err := cec.EvaluateCondition(condition, context)
		if err != nil {
			return false, fmt.Errorf("failed to evaluate condition %s: %w", condition.GetID(), err)
		}

		if evalResult.Status == layer0.ConditionStatusError {
			return false, fmt.Errorf("condition %s evaluation failed: %s", condition.GetID(), evalResult.Error)
		}

		results[i] = evalResult.Status == layer0.ConditionStatusTrue
	}

	// Apply logical operator
	switch operator {
	case layer0.ConditionOperatorAnd:
		for _, result := range results {
			if !result {
				return false, nil
			}
		}
		return true, nil

	case layer0.ConditionOperatorOr:
		for _, result := range results {
			if result {
				return true, nil
			}
		}
		return false, nil

	case layer0.ConditionOperatorNot:
		if len(results) != 1 {
			return false, fmt.Errorf("NOT operator requires exactly one condition, got %d", len(results))
		}
		return !results[0], nil

	default:
		return false, fmt.Errorf("unsupported condition operator: %s", operator)
	}
}

// GetEvaluationResult retrieves the evaluation result for a specific condition ID
func (cec *ConditionEvaluationCore) GetEvaluationResult(conditionID layer0.ConditionID) (ConditionEvaluationResult, error) {
	cec.mutex.RLock()
	defer cec.mutex.RUnlock()

	result, exists := cec.evaluationResults[conditionID]
	if !exists {
		return ConditionEvaluationResult{}, fmt.Errorf("no evaluation result found for condition ID %s", conditionID)
	}

	return result, nil
}

// GetAllEvaluationResults returns all evaluation results
func (cec *ConditionEvaluationCore) GetAllEvaluationResults() []ConditionEvaluationResult {
	cec.mutex.RLock()
	defer cec.mutex.RUnlock()

	results := make([]ConditionEvaluationResult, 0, len(cec.evaluationResults))
	for _, result := range cec.evaluationResults {
		results = append(results, result)
	}

	return results
}

// IsConditionEvaluating checks if a condition is currently being evaluated
func (cec *ConditionEvaluationCore) IsConditionEvaluating(conditionID layer0.ConditionID) bool {
	cec.mutex.RLock()
	defer cec.mutex.RUnlock()

	_, isActive := cec.activeEvaluations[conditionID]
	return isActive
}

// GetActiveEvaluations returns all currently active condition evaluations
func (cec *ConditionEvaluationCore) GetActiveEvaluations() []layer0.Condition {
	cec.mutex.RLock()
	defer cec.mutex.RUnlock()

	activeConditions := make([]layer0.Condition, 0, len(cec.activeEvaluations))
	for _, condition := range cec.activeEvaluations {
		activeConditions = append(activeConditions, condition)
	}

	return activeConditions
}

// MockConditionEvaluator is a simple mock evaluator for testing
type MockConditionEvaluator struct {
	supportedTypes []layer0.ConditionType
	evaluateFunc   func(condition layer0.Condition, context *layer0.Context) (interface{}, error)
}

// NewMockConditionEvaluator creates a new mock condition evaluator
func NewMockConditionEvaluator(supportedTypes []layer0.ConditionType, evaluateFunc func(layer0.Condition, *layer0.Context) (interface{}, error)) *MockConditionEvaluator {
	return &MockConditionEvaluator{
		supportedTypes: supportedTypes,
		evaluateFunc:   evaluateFunc,
	}
}

// Evaluate evaluates the condition using the mock function
func (mce *MockConditionEvaluator) Evaluate(condition layer0.Condition, context *layer0.Context) (interface{}, error) {
	if mce.evaluateFunc != nil {
		return mce.evaluateFunc(condition, context)
	}
	return true, nil // Default to true
}

// CanEvaluate checks if the evaluator can evaluate the given condition type
func (mce *MockConditionEvaluator) CanEvaluate(conditionType layer0.ConditionType) bool {
	for _, supportedType := range mce.supportedTypes {
		if supportedType == conditionType {
			return true
		}
	}
	return false
}

// GetSupportedTypes returns the supported condition types
func (mce *MockConditionEvaluator) GetSupportedTypes() []layer0.ConditionType {
	return mce.supportedTypes
}
