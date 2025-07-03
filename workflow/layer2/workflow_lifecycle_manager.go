package layer2

import (
	"log"
	"sync"
	"time"
)

// WorkflowLifecycleEvent represents an event in the workflow lifecycle
type WorkflowLifecycleEvent struct {
	InstanceID WorkflowInstanceID     `json:"instance_id"`
	EventType  string                 `json:"event_type"`
	Timestamp  time.Time              `json:"timestamp"`
	Data       map[string]interface{} `json:"data"`
}

// WorkflowLifecycleManager defines the interface for managing workflow lifecycle events
type WorkflowLifecycleManager interface {
	OnWorkflowStarted(instanceID WorkflowInstanceID) error
	OnWorkflowCompleted(instanceID WorkflowInstanceID) error
	OnWorkflowFailed(instanceID WorkflowInstanceID, err error) error
	OnWorkflowPaused(instanceID WorkflowInstanceID) error
	OnWorkflowResumed(instanceID WorkflowInstanceID) error
	OnWorkflowCancelled(instanceID WorkflowInstanceID) error
	OnStateChanged(instanceID WorkflowInstanceID, fromState, toState string) error
	GetEvents(instanceID WorkflowInstanceID) []WorkflowLifecycleEvent
	GetAllEvents() []WorkflowLifecycleEvent
	ClearEvents(instanceID WorkflowInstanceID) error
}

// DefaultWorkflowLifecycleManager provides a default implementation of WorkflowLifecycleManager
type DefaultWorkflowLifecycleManager struct {
	events map[WorkflowInstanceID][]WorkflowLifecycleEvent
	mutex  sync.RWMutex
}

// NewDefaultWorkflowLifecycleManager creates a new default workflow lifecycle manager
func NewDefaultWorkflowLifecycleManager() *DefaultWorkflowLifecycleManager {
	return &DefaultWorkflowLifecycleManager{
		events: make(map[WorkflowInstanceID][]WorkflowLifecycleEvent),
		mutex:  sync.RWMutex{},
	}
}

// OnWorkflowStarted handles workflow started event
func (manager *DefaultWorkflowLifecycleManager) OnWorkflowStarted(instanceID WorkflowInstanceID) error {
	event := WorkflowLifecycleEvent{
		InstanceID: instanceID,
		EventType:  "workflow_started",
		Timestamp:  time.Now(),
		Data:       map[string]interface{}{},
	}

	manager.addEvent(instanceID, event)
	log.Printf("Workflow %s started", instanceID)
	return nil
}

// OnWorkflowCompleted handles workflow completed event
func (manager *DefaultWorkflowLifecycleManager) OnWorkflowCompleted(instanceID WorkflowInstanceID) error {
	event := WorkflowLifecycleEvent{
		InstanceID: instanceID,
		EventType:  "workflow_completed",
		Timestamp:  time.Now(),
		Data:       map[string]interface{}{},
	}

	manager.addEvent(instanceID, event)
	log.Printf("Workflow %s completed", instanceID)
	return nil
}

// OnWorkflowFailed handles workflow failed event
func (manager *DefaultWorkflowLifecycleManager) OnWorkflowFailed(instanceID WorkflowInstanceID, err error) error {
	event := WorkflowLifecycleEvent{
		InstanceID: instanceID,
		EventType:  "workflow_failed",
		Timestamp:  time.Now(),
		Data: map[string]interface{}{
			"error": err.Error(),
		},
	}

	manager.addEvent(instanceID, event)
	log.Printf("Workflow %s failed: %v", instanceID, err)
	return nil
}

// OnWorkflowPaused handles workflow paused event
func (manager *DefaultWorkflowLifecycleManager) OnWorkflowPaused(instanceID WorkflowInstanceID) error {
	event := WorkflowLifecycleEvent{
		InstanceID: instanceID,
		EventType:  "workflow_paused",
		Timestamp:  time.Now(),
		Data:       map[string]interface{}{},
	}

	manager.addEvent(instanceID, event)
	log.Printf("Workflow %s paused", instanceID)
	return nil
}

// OnWorkflowResumed handles workflow resumed event
func (manager *DefaultWorkflowLifecycleManager) OnWorkflowResumed(instanceID WorkflowInstanceID) error {
	event := WorkflowLifecycleEvent{
		InstanceID: instanceID,
		EventType:  "workflow_resumed",
		Timestamp:  time.Now(),
		Data:       map[string]interface{}{},
	}

	manager.addEvent(instanceID, event)
	log.Printf("Workflow %s resumed", instanceID)
	return nil
}

// OnWorkflowCancelled handles workflow cancelled event
func (manager *DefaultWorkflowLifecycleManager) OnWorkflowCancelled(instanceID WorkflowInstanceID) error {
	event := WorkflowLifecycleEvent{
		InstanceID: instanceID,
		EventType:  "workflow_cancelled",
		Timestamp:  time.Now(),
		Data:       map[string]interface{}{},
	}

	manager.addEvent(instanceID, event)
	log.Printf("Workflow %s cancelled", instanceID)
	return nil
}

// OnStateChanged handles state changed event
func (manager *DefaultWorkflowLifecycleManager) OnStateChanged(instanceID WorkflowInstanceID, fromState, toState string) error {
	event := WorkflowLifecycleEvent{
		InstanceID: instanceID,
		EventType:  "state_changed",
		Timestamp:  time.Now(),
		Data: map[string]interface{}{
			"from_state": fromState,
			"to_state":   toState,
		},
	}

	manager.addEvent(instanceID, event)
	log.Printf("Workflow %s state changed from %s to %s", instanceID, fromState, toState)
	return nil
}

// GetEvents retrieves all events for a specific workflow instance
func (manager *DefaultWorkflowLifecycleManager) GetEvents(instanceID WorkflowInstanceID) []WorkflowLifecycleEvent {
	manager.mutex.RLock()
	defer manager.mutex.RUnlock()

	events, exists := manager.events[instanceID]
	if !exists {
		return []WorkflowLifecycleEvent{}
	}

	// Return a copy to prevent external modification
	result := make([]WorkflowLifecycleEvent, len(events))
	copy(result, events)
	return result
}

// GetAllEvents retrieves all events across all workflow instances
func (manager *DefaultWorkflowLifecycleManager) GetAllEvents() []WorkflowLifecycleEvent {
	manager.mutex.RLock()
	defer manager.mutex.RUnlock()

	var allEvents []WorkflowLifecycleEvent
	for _, events := range manager.events {
		allEvents = append(allEvents, events...)
	}

	return allEvents
}

// ClearEvents clears all events for a specific workflow instance
func (manager *DefaultWorkflowLifecycleManager) ClearEvents(instanceID WorkflowInstanceID) error {
	manager.mutex.Lock()
	defer manager.mutex.Unlock()

	delete(manager.events, instanceID)
	return nil
}

// addEvent adds an event to the manager
func (manager *DefaultWorkflowLifecycleManager) addEvent(instanceID WorkflowInstanceID, event WorkflowLifecycleEvent) {
	manager.mutex.Lock()
	defer manager.mutex.Unlock()

	if _, exists := manager.events[instanceID]; !exists {
		manager.events[instanceID] = []WorkflowLifecycleEvent{}
	}

	manager.events[instanceID] = append(manager.events[instanceID], event)
}
