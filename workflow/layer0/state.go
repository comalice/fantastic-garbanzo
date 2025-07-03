package layer0

import (
	"fmt"
	"time"
)

// StateID represents a unique identifier for a state
type StateID string

// StateType defines the type of state
type StateType string

const (
	// StateTypeInitial represents the starting state of a workflow
	StateTypeInitial StateType = "initial"
	// StateTypeIntermediate represents a state that is neither initial nor final
	StateTypeIntermediate StateType = "intermediate"
	// StateTypeFinal represents a terminal state that ends the workflow successfully
	StateTypeFinal StateType = "final"
	// StateTypeError represents a terminal state that ends the workflow with an error
	StateTypeError StateType = "error"
)

// StateStatus represents the current status of a state
type StateStatus string

const (
	// StateStatusActive indicates the state is currently active and processing
	StateStatusActive StateStatus = "active"
	// StateStatusInactive indicates the state is not currently active
	StateStatusInactive StateStatus = "inactive"
	// StateStatusPending indicates the state is waiting to become active
	StateStatusPending StateStatus = "pending"
	// StateStatusComplete indicates the state has finished processing successfully
	StateStatusComplete StateStatus = "complete"
	// StateStatusFailed indicates the state has failed during processing
	StateStatusFailed StateStatus = "failed"
)

// StateMetadata contains metadata about a state
type StateMetadata struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Tags        []string          `json:"tags"`
	Properties  map[string]string `json:"properties"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

// State represents an atomic state in the workflow system
type State struct {
	ID       StateID       `json:"id"`
	Type     StateType     `json:"type"`
	Status   StateStatus   `json:"status"`
	Metadata StateMetadata `json:"metadata"`
	Data     interface{}   `json:"data"`
}

// StateInterface defines the contract for state operations
type StateInterface interface {
	GetID() StateID
	GetType() StateType
	GetStatus() StateStatus
	GetMetadata() StateMetadata
	GetData() interface{}
	SetStatus(status StateStatus) State
	SetData(data interface{}) State
	IsActive() bool
	IsFinal() bool
	IsError() bool
	Clone() State
	Validate() error
}

// NewState creates a new state with the given parameters
func NewState(id StateID, stateType StateType, name string) State {
	now := time.Now()
	return State{
		ID:     id,
		Type:   stateType,
		Status: StateStatusInactive,
		Metadata: StateMetadata{
			Name:        name,
			Description: "",
			Tags:        []string{},
			Properties:  make(map[string]string),
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		Data: nil,
	}
}

// GetID returns the state ID
func (s State) GetID() StateID {
	return s.ID
}

// GetType returns the state type
func (s State) GetType() StateType {
	return s.Type
}

// GetStatus returns the state status
func (s State) GetStatus() StateStatus {
	return s.Status
}

// GetMetadata returns the state metadata
func (s State) GetMetadata() StateMetadata {
	return s.Metadata
}

// GetData returns the state data
func (s State) GetData() interface{} {
	return s.Data
}

// SetStatus creates a new state with updated status (immutable)
func (s State) SetStatus(status StateStatus) State {
	newState := s.Clone()
	newState.Status = status
	newState.Metadata.UpdatedAt = time.Now()
	return newState
}

// SetData creates a new state with updated data (immutable)
func (s State) SetData(data interface{}) State {
	newState := s.Clone()
	newState.Data = data
	newState.Metadata.UpdatedAt = time.Now()
	return newState
}

// IsActive checks if the state is active
func (s State) IsActive() bool {
	return s.Status == StateStatusActive
}

// IsFinal checks if the state is a final state
func (s State) IsFinal() bool {
	return s.Type == StateTypeFinal
}

// IsError checks if the state is an error state
func (s State) IsError() bool {
	return s.Type == StateTypeError || s.Status == StateStatusFailed
}

// Clone creates a deep copy of the state
func (s State) Clone() State {
	metadata := StateMetadata{
		Name:        s.Metadata.Name,
		Description: s.Metadata.Description,
		Tags:        make([]string, len(s.Metadata.Tags)),
		Properties:  make(map[string]string),
		CreatedAt:   s.Metadata.CreatedAt,
		UpdatedAt:   s.Metadata.UpdatedAt,
	}

	copy(metadata.Tags, s.Metadata.Tags)
	for k, v := range s.Metadata.Properties {
		metadata.Properties[k] = v
	}

	return State{
		ID:       s.ID,
		Type:     s.Type,
		Status:   s.Status,
		Metadata: metadata,
		Data:     s.Data, // Shallow copy for data - caller responsible for deep copy if needed
	}
}

// Validate checks if the state is valid
func (s State) Validate() error {
	if s.ID == "" {
		return fmt.Errorf("state ID cannot be empty")
	}

	if s.Type == "" {
		return fmt.Errorf("state type cannot be empty")
	}

	if s.Status == "" {
		return fmt.Errorf("state status cannot be empty")
	}

	if s.Metadata.Name == "" {
		return fmt.Errorf("state name cannot be empty")
	}

	return nil
}
