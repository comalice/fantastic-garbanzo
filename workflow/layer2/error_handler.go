package layer2

import (
	"fmt"
	"log"
	"sync"
	"time"
)

// ErrorSeverity defines the severity level of an error
type ErrorSeverity string

const (
	ErrorSeverityLow      ErrorSeverity = "low"
	ErrorSeverityMedium   ErrorSeverity = "medium"
	ErrorSeverityHigh     ErrorSeverity = "high"
	ErrorSeverityCritical ErrorSeverity = "critical"
)

// WorkflowError represents an error that occurred during workflow execution
type WorkflowError struct {
	ID          string                 `json:"id"`
	InstanceID  WorkflowInstanceID     `json:"instance_id"`
	Error       error                  `json:"error"`
	Severity    ErrorSeverity          `json:"severity"`
	Timestamp   time.Time              `json:"timestamp"`
	Context     map[string]interface{} `json:"context"`
	Recoverable bool                   `json:"recoverable"`
	Handled     bool                   `json:"handled"`
}

// ErrorHandler defines the interface for handling workflow errors
type ErrorHandler interface {
	HandleError(instanceID WorkflowInstanceID, err error) error
	HandleErrorWithSeverity(instanceID WorkflowInstanceID, err error, severity ErrorSeverity) error
	HandleErrorWithContext(instanceID WorkflowInstanceID, err error, severity ErrorSeverity, context map[string]interface{}) error
	GetErrors(instanceID WorkflowInstanceID) []WorkflowError
	GetAllErrors() []WorkflowError
	ClearErrors(instanceID WorkflowInstanceID) error
	IsRecoverable(err error) bool
}

// DefaultErrorHandler provides a default implementation of ErrorHandler
type DefaultErrorHandler struct {
	errors map[WorkflowInstanceID][]WorkflowError
	mutex  sync.RWMutex
}

// NewDefaultErrorHandler creates a new default error handler
func NewDefaultErrorHandler() *DefaultErrorHandler {
	return &DefaultErrorHandler{
		errors: make(map[WorkflowInstanceID][]WorkflowError),
		mutex:  sync.RWMutex{},
	}
}

// HandleError handles an error with default severity
func (handler *DefaultErrorHandler) HandleError(instanceID WorkflowInstanceID, err error) error {
	return handler.HandleErrorWithSeverity(instanceID, err, ErrorSeverityMedium)
}

// HandleErrorWithSeverity handles an error with specified severity
func (handler *DefaultErrorHandler) HandleErrorWithSeverity(instanceID WorkflowInstanceID, err error, severity ErrorSeverity) error {
	return handler.HandleErrorWithContext(instanceID, err, severity, nil)
}

// HandleErrorWithContext handles an error with severity and context
func (handler *DefaultErrorHandler) HandleErrorWithContext(instanceID WorkflowInstanceID, err error, severity ErrorSeverity, context map[string]interface{}) error {
	handler.mutex.Lock()
	defer handler.mutex.Unlock()

	// Create workflow error
	workflowError := WorkflowError{
		ID:          fmt.Sprintf("%s-%d", instanceID, time.Now().UnixNano()),
		InstanceID:  instanceID,
		Error:       err,
		Severity:    severity,
		Timestamp:   time.Now(),
		Context:     context,
		Recoverable: handler.IsRecoverable(err),
		Handled:     false,
	}

	// Add to errors list
	if _, exists := handler.errors[instanceID]; !exists {
		handler.errors[instanceID] = []WorkflowError{}
	}
	handler.errors[instanceID] = append(handler.errors[instanceID], workflowError)

	// Log the error
	log.Printf("[%s] Workflow %s error: %v", severity, instanceID, err)

	// Handle based on severity
	switch severity {
	case ErrorSeverityCritical:
		log.Printf("CRITICAL ERROR in workflow %s: %v", instanceID, err)
		// In a real implementation, this might trigger alerts, notifications, etc.
	case ErrorSeverityHigh:
		log.Printf("HIGH SEVERITY ERROR in workflow %s: %v", instanceID, err)
	case ErrorSeverityMedium:
		log.Printf("MEDIUM SEVERITY ERROR in workflow %s: %v", instanceID, err)
	case ErrorSeverityLow:
		log.Printf("LOW SEVERITY ERROR in workflow %s: %v", instanceID, err)
	}

	// Mark as handled
	if len(handler.errors[instanceID]) > 0 {
		lastIndex := len(handler.errors[instanceID]) - 1
		handler.errors[instanceID][lastIndex].Handled = true
	}

	return nil
}

// GetErrors retrieves all errors for a specific workflow instance
func (handler *DefaultErrorHandler) GetErrors(instanceID WorkflowInstanceID) []WorkflowError {
	handler.mutex.RLock()
	defer handler.mutex.RUnlock()

	errors, exists := handler.errors[instanceID]
	if !exists {
		return []WorkflowError{}
	}

	// Return a copy to prevent external modification
	result := make([]WorkflowError, len(errors))
	copy(result, errors)
	return result
}

// GetAllErrors retrieves all errors across all workflow instances
func (handler *DefaultErrorHandler) GetAllErrors() []WorkflowError {
	handler.mutex.RLock()
	defer handler.mutex.RUnlock()

	var allErrors []WorkflowError
	for _, errors := range handler.errors {
		allErrors = append(allErrors, errors...)
	}

	return allErrors
}

// ClearErrors clears all errors for a specific workflow instance
func (handler *DefaultErrorHandler) ClearErrors(instanceID WorkflowInstanceID) error {
	handler.mutex.Lock()
	defer handler.mutex.Unlock()

	delete(handler.errors, instanceID)
	return nil
}

// IsRecoverable determines if an error is recoverable
func (handler *DefaultErrorHandler) IsRecoverable(err error) bool {
	if err == nil {
		return true
	}

	// Simple heuristics for determining recoverability
	errorMsg := err.Error()

	// Network-related errors are often recoverable
	if contains(errorMsg, "timeout") || contains(errorMsg, "connection") || contains(errorMsg, "network") {
		return true
	}

	// Resource-related errors might be recoverable
	if contains(errorMsg, "resource") || contains(errorMsg, "memory") || contains(errorMsg, "disk") {
		return true
	}

	// Validation errors are typically not recoverable without intervention
	if contains(errorMsg, "validation") || contains(errorMsg, "invalid") || contains(errorMsg, "malformed") {
		return false
	}

	// Permission errors are typically not recoverable
	if contains(errorMsg, "permission") || contains(errorMsg, "unauthorized") || contains(errorMsg, "forbidden") {
		return false
	}

	// Default to recoverable for unknown errors
	return true
}

// contains checks if a string contains a substring (case-insensitive)
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		(len(s) > len(substr) &&
			(s[:len(substr)] == substr ||
				s[len(s)-len(substr):] == substr ||
				containsSubstring(s, substr))))
}

// containsSubstring checks if a string contains a substring
func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
