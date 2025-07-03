package layer0

import (
	"fmt"
	"time"
)

// ConditionID represents a unique identifier for a condition
type ConditionID string

// ConditionType defines the type of condition
type ConditionType string

const (
	ConditionTypeExpression ConditionType = "expression"
	ConditionTypeScript     ConditionType = "script"
	ConditionTypeService    ConditionType = "service"
	ConditionTypeTime       ConditionType = "time"
	ConditionTypeEvent      ConditionType = "event"
)

// ConditionStatus represents the current status of a condition
type ConditionStatus string

const (
	ConditionStatusPending    ConditionStatus = "pending"
	ConditionStatusEvaluating ConditionStatus = "evaluating"
	ConditionStatusTrue       ConditionStatus = "true"
	ConditionStatusFalse      ConditionStatus = "false"
	ConditionStatusError      ConditionStatus = "error"
)

// ConditionOperator defines logical operators for combining conditions
type ConditionOperator string

const (
	ConditionOperatorAnd ConditionOperator = "and"
	ConditionOperatorOr  ConditionOperator = "or"
	ConditionOperatorNot ConditionOperator = "not"
)

// ConditionMetadata contains metadata about a condition
type ConditionMetadata struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Tags        []string          `json:"tags"`
	Properties  map[string]string `json:"properties"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
	EvaluatedAt *time.Time        `json:"evaluated_at,omitempty"`
}

// ConditionExpression represents a condition expression
type ConditionExpression struct {
	Expression string                 `json:"expression"`
	Variables  map[string]interface{} `json:"variables"`
	Language   string                 `json:"language"` // e.g., "javascript", "python", "go"
}

// Condition represents an atomic condition in the workflow system
type Condition struct {
	ID           ConditionID         `json:"id"`
	Type         ConditionType       `json:"type"`
	Status       ConditionStatus     `json:"status"`
	Metadata     ConditionMetadata   `json:"metadata"`
	Expression   ConditionExpression `json:"expression"`
	Result       interface{}         `json:"result"`
	Error        string              `json:"error,omitempty"`
	Dependencies []ConditionID       `json:"dependencies"` // Other conditions this depends on
}

// ConditionInterface defines the contract for condition operations
type ConditionInterface interface {
	GetID() ConditionID
	GetType() ConditionType
	GetStatus() ConditionStatus
	GetMetadata() ConditionMetadata
	GetExpression() ConditionExpression
	GetResult() interface{}
	GetError() string
	GetDependencies() []ConditionID
	SetStatus(status ConditionStatus) Condition
	SetResult(result interface{}) Condition
	SetError(error string) Condition
	AddDependency(conditionID ConditionID) Condition
	MarkEvaluated(result interface{}) Condition
	MarkFailed(error string) Condition
	IsTrue() bool
	IsFalse() bool
	IsError() bool
	IsPending() bool
	IsEvaluated() bool
	Clone() Condition
	Validate() error
}

// NewCondition creates a new condition with the given parameters
func NewCondition(id ConditionID, conditionType ConditionType, name string) Condition {
	now := time.Now()
	return Condition{
		ID:     id,
		Type:   conditionType,
		Status: ConditionStatusPending,
		Metadata: ConditionMetadata{
			Name:        name,
			Description: "",
			Tags:        []string{},
			Properties:  make(map[string]string),
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		Expression: ConditionExpression{
			Expression: "",
			Variables:  make(map[string]interface{}),
			Language:   "javascript", // Default language
		},
		Result:       nil,
		Error:        "",
		Dependencies: []ConditionID{},
	}
}

// GetID returns the condition ID
func (c Condition) GetID() ConditionID {
	return c.ID
}

// GetType returns the condition type
func (c Condition) GetType() ConditionType {
	return c.Type
}

// GetStatus returns the condition status
func (c Condition) GetStatus() ConditionStatus {
	return c.Status
}

// GetMetadata returns the condition metadata
func (c Condition) GetMetadata() ConditionMetadata {
	return c.Metadata
}

// GetExpression returns the condition expression
func (c Condition) GetExpression() ConditionExpression {
	return c.Expression
}

// GetResult returns the condition result
func (c Condition) GetResult() interface{} {
	return c.Result
}

// GetError returns the condition error
func (c Condition) GetError() string {
	return c.Error
}

// GetDependencies returns the condition dependencies
func (c Condition) GetDependencies() []ConditionID {
	return c.Dependencies
}

// SetStatus creates a new condition with updated status (immutable)
func (c Condition) SetStatus(status ConditionStatus) Condition {
	newCondition := c.Clone()
	newCondition.Status = status
	newCondition.Metadata.UpdatedAt = time.Now()
	return newCondition
}

// SetResult creates a new condition with updated result (immutable)
func (c Condition) SetResult(result interface{}) Condition {
	newCondition := c.Clone()
	newCondition.Result = result
	newCondition.Metadata.UpdatedAt = time.Now()
	return newCondition
}

// SetError creates a new condition with updated error (immutable)
func (c Condition) SetError(error string) Condition {
	newCondition := c.Clone()
	newCondition.Error = error
	newCondition.Metadata.UpdatedAt = time.Now()
	return newCondition
}

// AddDependency creates a new condition with an additional dependency (immutable)
func (c Condition) AddDependency(conditionID ConditionID) Condition {
	newCondition := c.Clone()
	newCondition.Dependencies = append(newCondition.Dependencies, conditionID)
	newCondition.Metadata.UpdatedAt = time.Now()
	return newCondition
}

// MarkEvaluated marks the condition as evaluated with a result
func (c Condition) MarkEvaluated(result interface{}) Condition {
	newCondition := c.SetResult(result)
	now := time.Now()
	newCondition.Metadata.EvaluatedAt = &now

	// Determine status based on result
	if result == nil {
		newCondition.Status = ConditionStatusFalse
	} else if boolResult, ok := result.(bool); ok {
		if boolResult {
			newCondition.Status = ConditionStatusTrue
		} else {
			newCondition.Status = ConditionStatusFalse
		}
	} else {
		// Non-boolean results are considered true if not nil/empty
		newCondition.Status = ConditionStatusTrue
	}

	return newCondition
}

// MarkFailed marks the condition as failed with an error
func (c Condition) MarkFailed(error string) Condition {
	newCondition := c.SetStatus(ConditionStatusError).SetError(error)
	now := time.Now()
	newCondition.Metadata.EvaluatedAt = &now
	return newCondition
}

// IsTrue checks if the condition evaluated to true
func (c Condition) IsTrue() bool {
	return c.Status == ConditionStatusTrue
}

// IsFalse checks if the condition evaluated to false
func (c Condition) IsFalse() bool {
	return c.Status == ConditionStatusFalse
}

// IsError checks if the condition evaluation resulted in an error
func (c Condition) IsError() bool {
	return c.Status == ConditionStatusError
}

// IsPending checks if the condition is pending evaluation
func (c Condition) IsPending() bool {
	return c.Status == ConditionStatusPending
}

// IsEvaluated checks if the condition has been evaluated
func (c Condition) IsEvaluated() bool {
	return c.Status == ConditionStatusTrue || c.Status == ConditionStatusFalse || c.Status == ConditionStatusError
}

// Clone creates a deep copy of the condition
func (c Condition) Clone() Condition {
	metadata := ConditionMetadata{
		Name:        c.Metadata.Name,
		Description: c.Metadata.Description,
		Tags:        make([]string, len(c.Metadata.Tags)),
		Properties:  make(map[string]string),
		CreatedAt:   c.Metadata.CreatedAt,
		UpdatedAt:   c.Metadata.UpdatedAt,
	}

	copy(metadata.Tags, c.Metadata.Tags)
	for k, v := range c.Metadata.Properties {
		metadata.Properties[k] = v
	}

	if c.Metadata.EvaluatedAt != nil {
		evaluatedAt := *c.Metadata.EvaluatedAt
		metadata.EvaluatedAt = &evaluatedAt
	}

	expression := ConditionExpression{
		Expression: c.Expression.Expression,
		Variables:  make(map[string]interface{}),
		Language:   c.Expression.Language,
	}

	for k, v := range c.Expression.Variables {
		expression.Variables[k] = v
	}

	dependencies := make([]ConditionID, len(c.Dependencies))
	copy(dependencies, c.Dependencies)

	return Condition{
		ID:           c.ID,
		Type:         c.Type,
		Status:       c.Status,
		Metadata:     metadata,
		Expression:   expression,
		Result:       c.Result, // Shallow copy
		Error:        c.Error,
		Dependencies: dependencies,
	}
}

// Validate checks if the condition is valid
func (c Condition) Validate() error {
	if c.ID == "" {
		return fmt.Errorf("condition ID cannot be empty")
	}

	if c.Type == "" {
		return fmt.Errorf("condition type cannot be empty")
	}

	if c.Status == "" {
		return fmt.Errorf("condition status cannot be empty")
	}

	if c.Metadata.Name == "" {
		return fmt.Errorf("condition name cannot be empty")
	}

	if c.Type == ConditionTypeExpression && c.Expression.Expression == "" {
		return fmt.Errorf("expression condition must have an expression")
	}

	return nil
}
