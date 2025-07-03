package layer0

import (
	"fmt"
	"time"
)

// TransitionID represents a unique identifier for a transition
type TransitionID string

// TransitionType defines the type of transition
type TransitionType string

const (
	TransitionTypeAutomatic    TransitionType = "automatic"
	TransitionTypeManual       TransitionType = "manual"
	TransitionTypeConditional  TransitionType = "conditional"
	TransitionTypeCompensation TransitionType = "compensation"
)

// TransitionStatus represents the current status of a transition
type TransitionStatus string

const (
	TransitionStatusPending    TransitionStatus = "pending"
	TransitionStatusEvaluating TransitionStatus = "evaluating"
	TransitionStatusReady      TransitionStatus = "ready"
	TransitionStatusExecuting  TransitionStatus = "executing"
	TransitionStatusCompleted  TransitionStatus = "completed"
	TransitionStatusFailed     TransitionStatus = "failed"
	TransitionStatusSkipped    TransitionStatus = "skipped"
)

// TransitionMetadata contains metadata about a transition
type TransitionMetadata struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Tags        []string          `json:"tags"`
	Properties  map[string]string `json:"properties"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

// Transition represents an atomic transition in the workflow system
type Transition struct {
	ID          TransitionID       `json:"id"`
	Type        TransitionType     `json:"type"`
	Status      TransitionStatus   `json:"status"`
	FromStateID StateID            `json:"from_state_id"`
	ToStateID   StateID            `json:"to_state_id"`
	Metadata    TransitionMetadata `json:"metadata"`
	Conditions  []string           `json:"conditions"` // References to condition IDs
	Actions     []string           `json:"actions"`    // References to work IDs
	Priority    int                `json:"priority"`
	Data        interface{}        `json:"data"`
}

// TransitionInterface defines the contract for transition operations
type TransitionInterface interface {
	GetID() TransitionID
	GetType() TransitionType
	GetStatus() TransitionStatus
	GetFromStateID() StateID
	GetToStateID() StateID
	GetMetadata() TransitionMetadata
	GetConditions() []string
	GetActions() []string
	GetPriority() int
	GetData() interface{}
	SetStatus(status TransitionStatus) Transition
	SetData(data interface{}) Transition
	AddCondition(conditionID string) Transition
	AddAction(actionID string) Transition
	IsReady() bool
	IsCompleted() bool
	IsFailed() bool
	Clone() Transition
	Validate() error
}

// NewTransition creates a new transition with the given parameters
func NewTransition(id TransitionID, transitionType TransitionType, fromStateID, toStateID StateID, name string) Transition {
	now := time.Now()
	return Transition{
		ID:          id,
		Type:        transitionType,
		Status:      TransitionStatusPending,
		FromStateID: fromStateID,
		ToStateID:   toStateID,
		Metadata: TransitionMetadata{
			Name:        name,
			Description: "",
			Tags:        []string{},
			Properties:  make(map[string]string),
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		Conditions: []string{},
		Actions:    []string{},
		Priority:   0,
		Data:       nil,
	}
}

// GetID returns the transition ID
func (t Transition) GetID() TransitionID {
	return t.ID
}

// GetType returns the transition type
func (t Transition) GetType() TransitionType {
	return t.Type
}

// GetStatus returns the transition status
func (t Transition) GetStatus() TransitionStatus {
	return t.Status
}

// GetFromStateID returns the source state ID
func (t Transition) GetFromStateID() StateID {
	return t.FromStateID
}

// GetToStateID returns the target state ID
func (t Transition) GetToStateID() StateID {
	return t.ToStateID
}

// GetMetadata returns the transition metadata
func (t Transition) GetMetadata() TransitionMetadata {
	return t.Metadata
}

// GetConditions returns the condition IDs
func (t Transition) GetConditions() []string {
	return t.Conditions
}

// GetActions returns the action IDs
func (t Transition) GetActions() []string {
	return t.Actions
}

// GetPriority returns the transition priority
func (t Transition) GetPriority() int {
	return t.Priority
}

// GetData returns the transition data
func (t Transition) GetData() interface{} {
	return t.Data
}

// SetStatus creates a new transition with updated status (immutable)
func (t Transition) SetStatus(status TransitionStatus) Transition {
	newTransition := t.Clone()
	newTransition.Status = status
	newTransition.Metadata.UpdatedAt = time.Now()
	return newTransition
}

// SetData creates a new transition with updated data (immutable)
func (t Transition) SetData(data interface{}) Transition {
	newTransition := t.Clone()
	newTransition.Data = data
	newTransition.Metadata.UpdatedAt = time.Now()
	return newTransition
}

// AddCondition creates a new transition with an additional condition (immutable)
func (t Transition) AddCondition(conditionID string) Transition {
	newTransition := t.Clone()
	newTransition.Conditions = append(newTransition.Conditions, conditionID)
	newTransition.Metadata.UpdatedAt = time.Now()
	return newTransition
}

// AddAction creates a new transition with an additional action (immutable)
func (t Transition) AddAction(actionID string) Transition {
	newTransition := t.Clone()
	newTransition.Actions = append(newTransition.Actions, actionID)
	newTransition.Metadata.UpdatedAt = time.Now()
	return newTransition
}

// IsReady checks if the transition is ready for execution
func (t Transition) IsReady() bool {
	return t.Status == TransitionStatusReady
}

// IsCompleted checks if the transition has completed
func (t Transition) IsCompleted() bool {
	return t.Status == TransitionStatusCompleted
}

// IsFailed checks if the transition has failed
func (t Transition) IsFailed() bool {
	return t.Status == TransitionStatusFailed
}

// Clone creates a deep copy of the transition
func (t Transition) Clone() Transition {
	metadata := TransitionMetadata{
		Name:        t.Metadata.Name,
		Description: t.Metadata.Description,
		Tags:        make([]string, len(t.Metadata.Tags)),
		Properties:  make(map[string]string),
		CreatedAt:   t.Metadata.CreatedAt,
		UpdatedAt:   t.Metadata.UpdatedAt,
	}

	copy(metadata.Tags, t.Metadata.Tags)
	for k, v := range t.Metadata.Properties {
		metadata.Properties[k] = v
	}

	conditions := make([]string, len(t.Conditions))
	copy(conditions, t.Conditions)

	actions := make([]string, len(t.Actions))
	copy(actions, t.Actions)

	return Transition{
		ID:          t.ID,
		Type:        t.Type,
		Status:      t.Status,
		FromStateID: t.FromStateID,
		ToStateID:   t.ToStateID,
		Metadata:    metadata,
		Conditions:  conditions,
		Actions:     actions,
		Priority:    t.Priority,
		Data:        t.Data, // Shallow copy for data
	}
}

// Validate checks if the transition is valid
func (t Transition) Validate() error {
	if t.ID == "" {
		return fmt.Errorf("transition ID cannot be empty")
	}

	if t.Type == "" {
		return fmt.Errorf("transition type cannot be empty")
	}

	if t.Status == "" {
		return fmt.Errorf("transition status cannot be empty")
	}

	if t.FromStateID == "" {
		return fmt.Errorf("from state ID cannot be empty")
	}

	if t.ToStateID == "" {
		return fmt.Errorf("to state ID cannot be empty")
	}

	if t.Metadata.Name == "" {
		return fmt.Errorf("transition name cannot be empty")
	}

	return nil
}
