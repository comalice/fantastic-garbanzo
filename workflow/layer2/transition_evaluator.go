
package layer2

import (
        "github.com/ubom/workflow/layer0"
        "github.com/ubom/workflow/layer1"
)

// TransitionEvaluator defines the interface for evaluating whether a transition can be taken
type TransitionEvaluator interface {
        CanTransition(transition layer0.Transition, context *layer0.Context) (bool, error)
        EvaluateConditions(conditionIDs []string, context *layer0.Context) (bool, error)
}

// DefaultTransitionEvaluator provides a default implementation of TransitionEvaluator
type DefaultTransitionEvaluator struct {
        conditionEvaluationCore *layer1.ConditionEvaluationCore
}

// NewDefaultTransitionEvaluator creates a new default transition evaluator
func NewDefaultTransitionEvaluator() *DefaultTransitionEvaluator {
        return &DefaultTransitionEvaluator{
                conditionEvaluationCore: layer1.NewConditionEvaluationCore(),
        }
}

// CanTransition evaluates whether a transition can be taken
func (evaluator *DefaultTransitionEvaluator) CanTransition(transition layer0.Transition, context *layer0.Context) (bool, error) {
        // Check transition status
        if !transition.IsReady() && transition.GetStatus() != layer0.TransitionStatusPending {
                return false, nil
        }

        // If no conditions, transition is always possible
        conditions := transition.GetConditions()
        if len(conditions) == 0 {
                return true, nil
        }

        // Evaluate all conditions
        return evaluator.EvaluateConditions(conditions, context)
}

// EvaluateConditions evaluates a list of condition IDs
func (evaluator *DefaultTransitionEvaluator) EvaluateConditions(conditionIDs []string, context *layer0.Context) (bool, error) {
        if len(conditionIDs) == 0 {
                return true, nil
        }

        // For simplicity, we'll assume all conditions must be true (AND logic)
        // In a real implementation, this would be more sophisticated
        for _, conditionID := range conditionIDs {
                // Create a simple condition for evaluation
                condition := layer0.NewCondition(layer0.ConditionID(conditionID), layer0.ConditionTypeExpression, conditionID)
                condition.Expression.Expression = "true" // Default to true for now
                
                // Register a simple evaluator if not already registered
                if len(evaluator.conditionEvaluationCore.GetSupportedConditionTypes()) == 0 {
                        simpleEvaluator := layer1.NewMockConditionEvaluator(
                                []layer0.ConditionType{layer0.ConditionTypeExpression},
                                func(c layer0.Condition, ctx *layer0.Context) (interface{}, error) {
                                        // Simple evaluation: check if context has a key matching the condition ID
                                        if value, exists := ctx.Get(string(c.GetID())); exists {
                                                if boolValue, ok := value.(bool); ok {
                                                        return boolValue, nil
                                                }
                                                return value != nil, nil
                                        }
                                        return true, nil // Default to true if no specific condition
                                },
                        )
                        evaluator.conditionEvaluationCore.RegisterEvaluator(layer0.ConditionTypeExpression, simpleEvaluator)
                }

                result, err := evaluator.conditionEvaluationCore.EvaluateCondition(condition, context)
                if err != nil {
                        return false, err
                }

                if result.Status != layer0.ConditionStatusTrue {
                        return false, nil
                }
        }

        return true, nil
}
