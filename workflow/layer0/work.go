package layer0

import (
	"fmt"
	"time"
)

// WorkID represents a unique identifier for work
type WorkID string

// WorkType defines the type of work
type WorkType string

const (
	WorkTypeTask         WorkType = "task"
	WorkTypeService      WorkType = "service"
	WorkTypeScript       WorkType = "script"
	WorkTypeHuman        WorkType = "human"
	WorkTypeCompensation WorkType = "compensation"
)

// WorkStatus represents the current status of work
type WorkStatus string

const (
	WorkStatusPending   WorkStatus = "pending"
	WorkStatusScheduled WorkStatus = "scheduled"
	WorkStatusExecuting WorkStatus = "executing"
	WorkStatusCompleted WorkStatus = "completed"
	WorkStatusFailed    WorkStatus = "failed"
	WorkStatusCancelled WorkStatus = "cancelled"
	WorkStatusRetrying  WorkStatus = "retrying"
)

// WorkPriority defines the priority level of work
type WorkPriority int

const (
	WorkPriorityLow      WorkPriority = 1
	WorkPriorityNormal   WorkPriority = 5
	WorkPriorityHigh     WorkPriority = 10
	WorkPriorityCritical WorkPriority = 15
)

// WorkMetadata contains metadata about work
type WorkMetadata struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Tags        []string          `json:"tags"`
	Properties  map[string]string `json:"properties"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
	ScheduledAt *time.Time        `json:"scheduled_at,omitempty"`
	StartedAt   *time.Time        `json:"started_at,omitempty"`
	CompletedAt *time.Time        `json:"completed_at,omitempty"`
}

// WorkConfiguration contains configuration for work execution
type WorkConfiguration struct {
	TimeoutSeconds    int                    `json:"timeout_seconds"`
	RetryCount        int                    `json:"retry_count"`
	RetryDelaySeconds int                    `json:"retry_delay_seconds"`
	Parameters        map[string]interface{} `json:"parameters"`
	Environment       map[string]string      `json:"environment"`
}

// Work represents an atomic unit of work in the workflow system
type Work struct {
	ID                 WorkID            `json:"id"`
	Type               WorkType          `json:"type"`
	Status             WorkStatus        `json:"status"`
	Priority           WorkPriority      `json:"priority"`
	Metadata           WorkMetadata      `json:"metadata"`
	Configuration      WorkConfiguration `json:"configuration"`
	Input              interface{}       `json:"input"`
	Output             interface{}       `json:"output"`
	Error              string            `json:"error,omitempty"`
	CompensationWorkID *WorkID           `json:"compensation_work_id,omitempty"`
}

// WorkInterface defines the contract for work operations
type WorkInterface interface {
	GetID() WorkID
	GetType() WorkType
	GetStatus() WorkStatus
	GetPriority() WorkPriority
	GetMetadata() WorkMetadata
	GetConfiguration() WorkConfiguration
	GetInput() interface{}
	GetOutput() interface{}
	GetError() string
	GetCompensationWorkID() *WorkID
	SetStatus(status WorkStatus) Work
	SetInput(input interface{}) Work
	SetOutput(output interface{}) Work
	SetError(error string) Work
	SetCompensationWorkID(workID WorkID) Work
	MarkStarted() Work
	MarkCompleted(output interface{}) Work
	MarkFailed(error string) Work
	IsExecutable() bool
	IsCompleted() bool
	IsFailed() bool
	RequiresCompensation() bool
	Clone() Work
	Validate() error
}

// NewWork creates a new work with the given parameters
func NewWork(id WorkID, workType WorkType, name string) Work {
	now := time.Now()
	return Work{
		ID:       id,
		Type:     workType,
		Status:   WorkStatusPending,
		Priority: WorkPriorityNormal,
		Metadata: WorkMetadata{
			Name:        name,
			Description: "",
			Tags:        []string{},
			Properties:  make(map[string]string),
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		Configuration: WorkConfiguration{
			TimeoutSeconds:    300, // 5 minutes default
			RetryCount:        3,
			RetryDelaySeconds: 5,
			Parameters:        make(map[string]interface{}),
			Environment:       make(map[string]string),
		},
		Input:  nil,
		Output: nil,
		Error:  "",
	}
}

// GetID returns the work ID
func (w Work) GetID() WorkID {
	return w.ID
}

// GetType returns the work type
func (w Work) GetType() WorkType {
	return w.Type
}

// GetStatus returns the work status
func (w Work) GetStatus() WorkStatus {
	return w.Status
}

// GetPriority returns the work priority
func (w Work) GetPriority() WorkPriority {
	return w.Priority
}

// GetMetadata returns the work metadata
func (w Work) GetMetadata() WorkMetadata {
	return w.Metadata
}

// GetConfiguration returns the work configuration
func (w Work) GetConfiguration() WorkConfiguration {
	return w.Configuration
}

// GetInput returns the work input
func (w Work) GetInput() interface{} {
	return w.Input
}

// GetOutput returns the work output
func (w Work) GetOutput() interface{} {
	return w.Output
}

// GetError returns the work error
func (w Work) GetError() string {
	return w.Error
}

// GetCompensationWorkID returns the compensation work ID
func (w Work) GetCompensationWorkID() *WorkID {
	return w.CompensationWorkID
}

// SetStatus creates a new work with updated status (immutable)
func (w Work) SetStatus(status WorkStatus) Work {
	newWork := w.Clone()
	newWork.Status = status
	newWork.Metadata.UpdatedAt = time.Now()
	return newWork
}

// SetInput creates a new work with updated input (immutable)
func (w Work) SetInput(input interface{}) Work {
	newWork := w.Clone()
	newWork.Input = input
	newWork.Metadata.UpdatedAt = time.Now()
	return newWork
}

// SetOutput creates a new work with updated output (immutable)
func (w Work) SetOutput(output interface{}) Work {
	newWork := w.Clone()
	newWork.Output = output
	newWork.Metadata.UpdatedAt = time.Now()
	return newWork
}

// SetError creates a new work with updated error (immutable)
func (w Work) SetError(error string) Work {
	newWork := w.Clone()
	newWork.Error = error
	newWork.Metadata.UpdatedAt = time.Now()
	return newWork
}

// SetCompensationWorkID creates a new work with updated compensation work ID (immutable)
func (w Work) SetCompensationWorkID(workID WorkID) Work {
	newWork := w.Clone()
	newWork.CompensationWorkID = &workID
	newWork.Metadata.UpdatedAt = time.Now()
	return newWork
}

// MarkStarted marks the work as started
func (w Work) MarkStarted() Work {
	newWork := w.SetStatus(WorkStatusExecuting)
	now := time.Now()
	newWork.Metadata.StartedAt = &now
	return newWork
}

// MarkCompleted marks the work as completed with output
func (w Work) MarkCompleted(output interface{}) Work {
	newWork := w.SetStatus(WorkStatusCompleted).SetOutput(output)
	now := time.Now()
	newWork.Metadata.CompletedAt = &now
	return newWork
}

// MarkFailed marks the work as failed with error
func (w Work) MarkFailed(error string) Work {
	newWork := w.SetStatus(WorkStatusFailed).SetError(error)
	now := time.Now()
	newWork.Metadata.CompletedAt = &now
	return newWork
}

// IsExecutable checks if the work can be executed
func (w Work) IsExecutable() bool {
	return w.Status == WorkStatusPending || w.Status == WorkStatusScheduled || w.Status == WorkStatusRetrying
}

// IsCompleted checks if the work has completed successfully
func (w Work) IsCompleted() bool {
	return w.Status == WorkStatusCompleted
}

// IsFailed checks if the work has failed
func (w Work) IsFailed() bool {
	return w.Status == WorkStatusFailed
}

// RequiresCompensation checks if the work requires compensation
func (w Work) RequiresCompensation() bool {
	return w.CompensationWorkID != nil && w.IsCompleted()
}

// Clone creates a deep copy of the work
func (w Work) Clone() Work {
	metadata := WorkMetadata{
		Name:        w.Metadata.Name,
		Description: w.Metadata.Description,
		Tags:        make([]string, len(w.Metadata.Tags)),
		Properties:  make(map[string]string),
		CreatedAt:   w.Metadata.CreatedAt,
		UpdatedAt:   w.Metadata.UpdatedAt,
	}

	copy(metadata.Tags, w.Metadata.Tags)
	for k, v := range w.Metadata.Properties {
		metadata.Properties[k] = v
	}

	if w.Metadata.ScheduledAt != nil {
		scheduledAt := *w.Metadata.ScheduledAt
		metadata.ScheduledAt = &scheduledAt
	}

	if w.Metadata.StartedAt != nil {
		startedAt := *w.Metadata.StartedAt
		metadata.StartedAt = &startedAt
	}

	if w.Metadata.CompletedAt != nil {
		completedAt := *w.Metadata.CompletedAt
		metadata.CompletedAt = &completedAt
	}

	configuration := WorkConfiguration{
		TimeoutSeconds:    w.Configuration.TimeoutSeconds,
		RetryCount:        w.Configuration.RetryCount,
		RetryDelaySeconds: w.Configuration.RetryDelaySeconds,
		Parameters:        make(map[string]interface{}),
		Environment:       make(map[string]string),
	}

	for k, v := range w.Configuration.Parameters {
		configuration.Parameters[k] = v
	}

	for k, v := range w.Configuration.Environment {
		configuration.Environment[k] = v
	}

	var compensationWorkID *WorkID
	if w.CompensationWorkID != nil {
		id := *w.CompensationWorkID
		compensationWorkID = &id
	}

	return Work{
		ID:                 w.ID,
		Type:               w.Type,
		Status:             w.Status,
		Priority:           w.Priority,
		Metadata:           metadata,
		Configuration:      configuration,
		Input:              w.Input,  // Shallow copy
		Output:             w.Output, // Shallow copy
		Error:              w.Error,
		CompensationWorkID: compensationWorkID,
	}
}

// Validate checks if the work is valid
func (w Work) Validate() error {
	if w.ID == "" {
		return fmt.Errorf("work ID cannot be empty")
	}

	if w.Type == "" {
		return fmt.Errorf("work type cannot be empty")
	}

	if w.Status == "" {
		return fmt.Errorf("work status cannot be empty")
	}

	if w.Metadata.Name == "" {
		return fmt.Errorf("work name cannot be empty")
	}

	if w.Configuration.TimeoutSeconds <= 0 {
		return fmt.Errorf("timeout seconds must be positive")
	}

	if w.Configuration.RetryCount < 0 {
		return fmt.Errorf("retry count cannot be negative")
	}

	if w.Configuration.RetryDelaySeconds < 0 {
		return fmt.Errorf("retry delay seconds cannot be negative")
	}

	return nil
}
