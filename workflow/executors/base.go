
package executors

import (
	"context"
	"time"

	"github.com/ubom/workflow/layer0"
)

// WorkExecutor is the unified interface for all work executors
// This replaces both the original WorkExecutor and EnhancedWorkExecutor interfaces
type WorkExecutor interface {
	// Core execution methods
	Execute(ctx context.Context, work layer0.Work, workContext *layer0.Context) (WorkResult, error)
	Validate(work layer0.Work) error
	
	// Capability methods
	CanExecute(workType layer0.WorkType) bool
	GetSupportedTypes() []layer0.WorkType
	
	// Schema and metadata support
	GetSchema() WorkSchema
	GetMetadata() WorkMetadata
}

// WorkResult represents execution result with comprehensive information
type WorkResult struct {
	Success bool                   `json:"success"`
	Outputs map[string]interface{} `json:"outputs"`
	Logs    []LogEntry            `json:"logs"`
	Metrics ExecutionMetrics      `json:"metrics"`
	Error   string                `json:"error,omitempty"`
}

// LogEntry represents a log entry during execution
type LogEntry struct {
	Timestamp time.Time `json:"timestamp"`
	Level     string    `json:"level"`
	Message   string    `json:"message"`
	Source    string    `json:"source"`
}

// ExecutionMetrics contains execution performance metrics
type ExecutionMetrics struct {
	StartTime    time.Time     `json:"start_time"`
	EndTime      time.Time     `json:"end_time"`
	Duration     time.Duration `json:"duration"`
	CPUUsage     float64       `json:"cpu_usage"`
	MemoryUsage  int64         `json:"memory_usage"`
	NetworkIO    NetworkIO     `json:"network_io"`
}

// NetworkIO represents network I/O metrics
type NetworkIO struct {
	BytesSent     int64 `json:"bytes_sent"`
	BytesReceived int64 `json:"bytes_received"`
}

// WorkSchema defines the JSON schema for work validation
type WorkSchema struct {
	JSONSchema    string           `json:"json_schema"`
	Examples      []WorkDefinition `json:"examples"`
	Documentation string           `json:"documentation"`
}

// WorkDefinition represents a work definition example
type WorkDefinition struct {
	Name          string                 `json:"name"`
	Description   string                 `json:"description"`
	Configuration map[string]interface{} `json:"configuration"`
	Input         interface{}            `json:"input"`
	ExpectedOutput interface{}           `json:"expected_output"`
}

// WorkMetadata contains executor metadata
type WorkMetadata struct {
	Name        string   `json:"name"`
	Version     string   `json:"version"`
	Author      string   `json:"author"`
	Description string   `json:"description"`
	WorkTypes   []string `json:"work_types"`
}

// BaseExecutor provides common functionality for all executors
type BaseExecutor struct {
	name         string
	version      string
	author       string
	description  string
	supportedTypes []layer0.WorkType
}

// NewBaseExecutor creates a new base executor
func NewBaseExecutor(name, version, author, description string, supportedTypes []layer0.WorkType) *BaseExecutor {
	return &BaseExecutor{
		name:           name,
		version:        version,
		author:         author,
		description:    description,
		supportedTypes: supportedTypes,
	}
}

// GetMetadata returns metadata about the executor
func (e *BaseExecutor) GetMetadata() WorkMetadata {
	workTypes := make([]string, len(e.supportedTypes))
	for i, wt := range e.supportedTypes {
		workTypes[i] = string(wt)
	}
	
	return WorkMetadata{
		Name:        e.name,
		Version:     e.version,
		Author:      e.author,
		Description: e.description,
		WorkTypes:   workTypes,
	}
}

// CanExecute checks if the executor can handle the work type
func (e *BaseExecutor) CanExecute(workType layer0.WorkType) bool {
	for _, supportedType := range e.supportedTypes {
		if supportedType == workType {
			return true
		}
	}
	return false
}

// GetSupportedTypes returns the supported work types
func (e *BaseExecutor) GetSupportedTypes() []layer0.WorkType {
	return e.supportedTypes
}
