package layer2

import (
	"fmt"
	"sync"
	"time"

	"github.com/ubom/workflow/layer0"
	"github.com/ubom/workflow/layer1"
)

// WorkflowInstanceID represents a unique identifier for a workflow instance
type WorkflowInstanceID string

// WorkflowInstanceStatus represents the status of a workflow instance
type WorkflowInstanceStatus string

const (
	WorkflowInstanceStatusCreated   WorkflowInstanceStatus = "created"
	WorkflowInstanceStatusRunning   WorkflowInstanceStatus = "running"
	WorkflowInstanceStatusPaused    WorkflowInstanceStatus = "paused"
	WorkflowInstanceStatusCompleted WorkflowInstanceStatus = "completed"
	WorkflowInstanceStatusFailed    WorkflowInstanceStatus = "failed"
	WorkflowInstanceStatusCancelled WorkflowInstanceStatus = "cancelled"
)

// WorkflowInstance represents a running instance of a workflow
type WorkflowInstance struct {
	ID                WorkflowInstanceID               `json:"id"`
	DefinitionID      layer1.WorkflowDefinitionID      `json:"definition_id"`
	DefinitionVersion layer1.WorkflowDefinitionVersion `json:"definition_version"`
	Status            WorkflowInstanceStatus           `json:"status"`
	CurrentStateID    layer0.StateID                   `json:"current_state_id"`
	Context           *layer0.Context                  `json:"context"`
	CreatedAt         time.Time                        `json:"created_at"`
	UpdatedAt         time.Time                        `json:"updated_at"`
	StartedAt         *time.Time                       `json:"started_at,omitempty"`
	CompletedAt       *time.Time                       `json:"completed_at,omitempty"`
	Error             string                           `json:"error,omitempty"`
	Metadata          map[string]interface{}           `json:"metadata"`
}

// StatePersistenceStore defines the interface for persisting workflow state
type StatePersistenceStore interface {
	// Workflow Instance operations
	SaveWorkflowInstance(instance WorkflowInstance) error
	GetWorkflowInstance(instanceID WorkflowInstanceID) (WorkflowInstance, error)
	UpdateWorkflowInstance(instance WorkflowInstance) error
	DeleteWorkflowInstance(instanceID WorkflowInstanceID) error
	ListWorkflowInstances(definitionID layer1.WorkflowDefinitionID) ([]WorkflowInstance, error)
	ListAllWorkflowInstances() ([]WorkflowInstance, error)

	// State operations
	SaveState(instanceID WorkflowInstanceID, state layer0.State) error
	GetState(instanceID WorkflowInstanceID, stateID layer0.StateID) (layer0.State, error)
	UpdateState(instanceID WorkflowInstanceID, state layer0.State) error
	ListStates(instanceID WorkflowInstanceID) ([]layer0.State, error)

	// Transition operations
	SaveTransition(instanceID WorkflowInstanceID, transition layer0.Transition) error
	GetTransition(instanceID WorkflowInstanceID, transitionID layer0.TransitionID) (layer0.Transition, error)
	UpdateTransition(instanceID WorkflowInstanceID, transition layer0.Transition) error
	ListTransitions(instanceID WorkflowInstanceID) ([]layer0.Transition, error)

	// Work operations
	SaveWork(instanceID WorkflowInstanceID, work layer0.Work) error
	GetWork(instanceID WorkflowInstanceID, workID layer0.WorkID) (layer0.Work, error)
	UpdateWork(instanceID WorkflowInstanceID, work layer0.Work) error
	ListWork(instanceID WorkflowInstanceID) ([]layer0.Work, error)

	// Context operations
	SaveContext(instanceID WorkflowInstanceID, context *layer0.Context) error
	GetContext(instanceID WorkflowInstanceID, contextID layer0.ContextID) (*layer0.Context, error)
	UpdateContext(instanceID WorkflowInstanceID, context *layer0.Context) error
	ListContexts(instanceID WorkflowInstanceID) ([]*layer0.Context, error)

	// Cleanup operations
	Cleanup() error
	GetStats() (map[string]interface{}, error)
}

// InMemoryStatePersistenceStore provides an in-memory implementation of StatePersistenceStore
type InMemoryStatePersistenceStore struct {
	workflowInstances map[WorkflowInstanceID]WorkflowInstance
	states            map[WorkflowInstanceID]map[layer0.StateID]layer0.State
	transitions       map[WorkflowInstanceID]map[layer0.TransitionID]layer0.Transition
	work              map[WorkflowInstanceID]map[layer0.WorkID]layer0.Work
	contexts          map[WorkflowInstanceID]map[layer0.ContextID]*layer0.Context
	mutex             sync.RWMutex
}

// NewInMemoryStatePersistenceStore creates a new in-memory state persistence store
func NewInMemoryStatePersistenceStore() *InMemoryStatePersistenceStore {
	return &InMemoryStatePersistenceStore{
		workflowInstances: make(map[WorkflowInstanceID]WorkflowInstance),
		states:            make(map[WorkflowInstanceID]map[layer0.StateID]layer0.State),
		transitions:       make(map[WorkflowInstanceID]map[layer0.TransitionID]layer0.Transition),
		work:              make(map[WorkflowInstanceID]map[layer0.WorkID]layer0.Work),
		contexts:          make(map[WorkflowInstanceID]map[layer0.ContextID]*layer0.Context),
		mutex:             sync.RWMutex{},
	}
}

// SaveWorkflowInstance saves a workflow instance
func (store *InMemoryStatePersistenceStore) SaveWorkflowInstance(instance WorkflowInstance) error {
	store.mutex.Lock()
	defer store.mutex.Unlock()

	if _, exists := store.workflowInstances[instance.ID]; exists {
		return fmt.Errorf("workflow instance %s already exists", instance.ID)
	}

	store.workflowInstances[instance.ID] = instance

	// Initialize maps for this instance
	store.states[instance.ID] = make(map[layer0.StateID]layer0.State)
	store.transitions[instance.ID] = make(map[layer0.TransitionID]layer0.Transition)
	store.work[instance.ID] = make(map[layer0.WorkID]layer0.Work)
	store.contexts[instance.ID] = make(map[layer0.ContextID]*layer0.Context)

	return nil
}

// GetWorkflowInstance retrieves a workflow instance
func (store *InMemoryStatePersistenceStore) GetWorkflowInstance(instanceID WorkflowInstanceID) (WorkflowInstance, error) {
	store.mutex.RLock()
	defer store.mutex.RUnlock()

	instance, exists := store.workflowInstances[instanceID]
	if !exists {
		return WorkflowInstance{}, fmt.Errorf("workflow instance %s not found", instanceID)
	}

	return instance, nil
}

// UpdateWorkflowInstance updates a workflow instance
func (store *InMemoryStatePersistenceStore) UpdateWorkflowInstance(instance WorkflowInstance) error {
	store.mutex.Lock()
	defer store.mutex.Unlock()

	if _, exists := store.workflowInstances[instance.ID]; !exists {
		return fmt.Errorf("workflow instance %s not found", instance.ID)
	}

	instance.UpdatedAt = time.Now()
	store.workflowInstances[instance.ID] = instance
	return nil
}

// DeleteWorkflowInstance deletes a workflow instance and all its associated data
func (store *InMemoryStatePersistenceStore) DeleteWorkflowInstance(instanceID WorkflowInstanceID) error {
	store.mutex.Lock()
	defer store.mutex.Unlock()

	if _, exists := store.workflowInstances[instanceID]; !exists {
		return fmt.Errorf("workflow instance %s not found", instanceID)
	}

	delete(store.workflowInstances, instanceID)
	delete(store.states, instanceID)
	delete(store.transitions, instanceID)
	delete(store.work, instanceID)
	delete(store.contexts, instanceID)

	return nil
}

// ListWorkflowInstances lists all workflow instances for a specific definition
func (store *InMemoryStatePersistenceStore) ListWorkflowInstances(definitionID layer1.WorkflowDefinitionID) ([]WorkflowInstance, error) {
	store.mutex.RLock()
	defer store.mutex.RUnlock()

	var instances []WorkflowInstance
	for _, instance := range store.workflowInstances {
		if instance.DefinitionID == definitionID {
			instances = append(instances, instance)
		}
	}

	return instances, nil
}

// ListAllWorkflowInstances lists all workflow instances
func (store *InMemoryStatePersistenceStore) ListAllWorkflowInstances() ([]WorkflowInstance, error) {
	store.mutex.RLock()
	defer store.mutex.RUnlock()

	instances := make([]WorkflowInstance, 0, len(store.workflowInstances))
	for _, instance := range store.workflowInstances {
		instances = append(instances, instance)
	}

	return instances, nil
}

// SaveState saves a state for a workflow instance
func (store *InMemoryStatePersistenceStore) SaveState(instanceID WorkflowInstanceID, state layer0.State) error {
	store.mutex.Lock()
	defer store.mutex.Unlock()

	if _, exists := store.workflowInstances[instanceID]; !exists {
		return fmt.Errorf("workflow instance %s not found", instanceID)
	}

	if _, exists := store.states[instanceID][state.GetID()]; exists {
		return fmt.Errorf("state %s already exists for instance %s", state.GetID(), instanceID)
	}

	store.states[instanceID][state.GetID()] = state
	return nil
}

// GetState retrieves a state for a workflow instance
func (store *InMemoryStatePersistenceStore) GetState(instanceID WorkflowInstanceID, stateID layer0.StateID) (layer0.State, error) {
	store.mutex.RLock()
	defer store.mutex.RUnlock()

	if _, exists := store.workflowInstances[instanceID]; !exists {
		return layer0.State{}, fmt.Errorf("workflow instance %s not found", instanceID)
	}

	state, exists := store.states[instanceID][stateID]
	if !exists {
		return layer0.State{}, fmt.Errorf("state %s not found for instance %s", stateID, instanceID)
	}

	return state, nil
}

// UpdateState updates a state for a workflow instance
func (store *InMemoryStatePersistenceStore) UpdateState(instanceID WorkflowInstanceID, state layer0.State) error {
	store.mutex.Lock()
	defer store.mutex.Unlock()

	if _, exists := store.workflowInstances[instanceID]; !exists {
		return fmt.Errorf("workflow instance %s not found", instanceID)
	}

	if _, exists := store.states[instanceID][state.GetID()]; !exists {
		return fmt.Errorf("state %s not found for instance %s", state.GetID(), instanceID)
	}

	store.states[instanceID][state.GetID()] = state
	return nil
}

// ListStates lists all states for a workflow instance
func (store *InMemoryStatePersistenceStore) ListStates(instanceID WorkflowInstanceID) ([]layer0.State, error) {
	store.mutex.RLock()
	defer store.mutex.RUnlock()

	if _, exists := store.workflowInstances[instanceID]; !exists {
		return nil, fmt.Errorf("workflow instance %s not found", instanceID)
	}

	states := make([]layer0.State, 0, len(store.states[instanceID]))
	for _, state := range store.states[instanceID] {
		states = append(states, state)
	}

	return states, nil
}

// SaveTransition saves a transition for a workflow instance
func (store *InMemoryStatePersistenceStore) SaveTransition(instanceID WorkflowInstanceID, transition layer0.Transition) error {
	store.mutex.Lock()
	defer store.mutex.Unlock()

	if _, exists := store.workflowInstances[instanceID]; !exists {
		return fmt.Errorf("workflow instance %s not found", instanceID)
	}

	if _, exists := store.transitions[instanceID][transition.GetID()]; exists {
		return fmt.Errorf("transition %s already exists for instance %s", transition.GetID(), instanceID)
	}

	store.transitions[instanceID][transition.GetID()] = transition
	return nil
}

// GetTransition retrieves a transition for a workflow instance
func (store *InMemoryStatePersistenceStore) GetTransition(instanceID WorkflowInstanceID, transitionID layer0.TransitionID) (layer0.Transition, error) {
	store.mutex.RLock()
	defer store.mutex.RUnlock()

	if _, exists := store.workflowInstances[instanceID]; !exists {
		return layer0.Transition{}, fmt.Errorf("workflow instance %s not found", instanceID)
	}

	transition, exists := store.transitions[instanceID][transitionID]
	if !exists {
		return layer0.Transition{}, fmt.Errorf("transition %s not found for instance %s", transitionID, instanceID)
	}

	return transition, nil
}

// UpdateTransition updates a transition for a workflow instance
func (store *InMemoryStatePersistenceStore) UpdateTransition(instanceID WorkflowInstanceID, transition layer0.Transition) error {
	store.mutex.Lock()
	defer store.mutex.Unlock()

	if _, exists := store.workflowInstances[instanceID]; !exists {
		return fmt.Errorf("workflow instance %s not found", instanceID)
	}

	if _, exists := store.transitions[instanceID][transition.GetID()]; !exists {
		return fmt.Errorf("transition %s not found for instance %s", transition.GetID(), instanceID)
	}

	store.transitions[instanceID][transition.GetID()] = transition
	return nil
}

// ListTransitions lists all transitions for a workflow instance
func (store *InMemoryStatePersistenceStore) ListTransitions(instanceID WorkflowInstanceID) ([]layer0.Transition, error) {
	store.mutex.RLock()
	defer store.mutex.RUnlock()

	if _, exists := store.workflowInstances[instanceID]; !exists {
		return nil, fmt.Errorf("workflow instance %s not found", instanceID)
	}

	transitions := make([]layer0.Transition, 0, len(store.transitions[instanceID]))
	for _, transition := range store.transitions[instanceID] {
		transitions = append(transitions, transition)
	}

	return transitions, nil
}

// SaveWork saves work for a workflow instance
func (store *InMemoryStatePersistenceStore) SaveWork(instanceID WorkflowInstanceID, work layer0.Work) error {
	store.mutex.Lock()
	defer store.mutex.Unlock()

	if _, exists := store.workflowInstances[instanceID]; !exists {
		return fmt.Errorf("workflow instance %s not found", instanceID)
	}

	if _, exists := store.work[instanceID][work.GetID()]; exists {
		return fmt.Errorf("work %s already exists for instance %s", work.GetID(), instanceID)
	}

	store.work[instanceID][work.GetID()] = work
	return nil
}

// GetWork retrieves work for a workflow instance
func (store *InMemoryStatePersistenceStore) GetWork(instanceID WorkflowInstanceID, workID layer0.WorkID) (layer0.Work, error) {
	store.mutex.RLock()
	defer store.mutex.RUnlock()

	if _, exists := store.workflowInstances[instanceID]; !exists {
		return layer0.Work{}, fmt.Errorf("workflow instance %s not found", instanceID)
	}

	work, exists := store.work[instanceID][workID]
	if !exists {
		return layer0.Work{}, fmt.Errorf("work %s not found for instance %s", workID, instanceID)
	}

	return work, nil
}

// UpdateWork updates work for a workflow instance
func (store *InMemoryStatePersistenceStore) UpdateWork(instanceID WorkflowInstanceID, work layer0.Work) error {
	store.mutex.Lock()
	defer store.mutex.Unlock()

	if _, exists := store.workflowInstances[instanceID]; !exists {
		return fmt.Errorf("workflow instance %s not found", instanceID)
	}

	if _, exists := store.work[instanceID][work.GetID()]; !exists {
		return fmt.Errorf("work %s not found for instance %s", work.GetID(), instanceID)
	}

	store.work[instanceID][work.GetID()] = work
	return nil
}

// ListWork lists all work for a workflow instance
func (store *InMemoryStatePersistenceStore) ListWork(instanceID WorkflowInstanceID) ([]layer0.Work, error) {
	store.mutex.RLock()
	defer store.mutex.RUnlock()

	if _, exists := store.workflowInstances[instanceID]; !exists {
		return nil, fmt.Errorf("workflow instance %s not found", instanceID)
	}

	workItems := make([]layer0.Work, 0, len(store.work[instanceID]))
	for _, work := range store.work[instanceID] {
		workItems = append(workItems, work)
	}

	return workItems, nil
}

// SaveContext saves a context for a workflow instance
func (store *InMemoryStatePersistenceStore) SaveContext(instanceID WorkflowInstanceID, context *layer0.Context) error {
	store.mutex.Lock()
	defer store.mutex.Unlock()

	if _, exists := store.workflowInstances[instanceID]; !exists {
		return fmt.Errorf("workflow instance %s not found", instanceID)
	}

	if _, exists := store.contexts[instanceID][context.GetID()]; exists {
		return fmt.Errorf("context %s already exists for instance %s", context.GetID(), instanceID)
	}

	store.contexts[instanceID][context.GetID()] = context
	return nil
}

// GetContext retrieves a context for a workflow instance
func (store *InMemoryStatePersistenceStore) GetContext(instanceID WorkflowInstanceID, contextID layer0.ContextID) (*layer0.Context, error) {
	store.mutex.RLock()
	defer store.mutex.RUnlock()

	if _, exists := store.workflowInstances[instanceID]; !exists {
		return nil, fmt.Errorf("workflow instance %s not found", instanceID)
	}

	context, exists := store.contexts[instanceID][contextID]
	if !exists {
		return nil, fmt.Errorf("context %s not found for instance %s", contextID, instanceID)
	}

	return context, nil
}

// UpdateContext updates a context for a workflow instance
func (store *InMemoryStatePersistenceStore) UpdateContext(instanceID WorkflowInstanceID, context *layer0.Context) error {
	store.mutex.Lock()
	defer store.mutex.Unlock()

	if _, exists := store.workflowInstances[instanceID]; !exists {
		return fmt.Errorf("workflow instance %s not found", instanceID)
	}

	if _, exists := store.contexts[instanceID][context.GetID()]; !exists {
		return fmt.Errorf("context %s not found for instance %s", context.GetID(), instanceID)
	}

	store.contexts[instanceID][context.GetID()] = context
	return nil
}

// ListContexts lists all contexts for a workflow instance
func (store *InMemoryStatePersistenceStore) ListContexts(instanceID WorkflowInstanceID) ([]*layer0.Context, error) {
	store.mutex.RLock()
	defer store.mutex.RUnlock()

	if _, exists := store.workflowInstances[instanceID]; !exists {
		return nil, fmt.Errorf("workflow instance %s not found", instanceID)
	}

	contexts := make([]*layer0.Context, 0, len(store.contexts[instanceID]))
	for _, context := range store.contexts[instanceID] {
		contexts = append(contexts, context)
	}

	return contexts, nil
}

// Cleanup clears all data from the store
func (store *InMemoryStatePersistenceStore) Cleanup() error {
	store.mutex.Lock()
	defer store.mutex.Unlock()

	store.workflowInstances = make(map[WorkflowInstanceID]WorkflowInstance)
	store.states = make(map[WorkflowInstanceID]map[layer0.StateID]layer0.State)
	store.transitions = make(map[WorkflowInstanceID]map[layer0.TransitionID]layer0.Transition)
	store.work = make(map[WorkflowInstanceID]map[layer0.WorkID]layer0.Work)
	store.contexts = make(map[WorkflowInstanceID]map[layer0.ContextID]*layer0.Context)

	return nil
}

// GetStats returns statistics about the store
func (store *InMemoryStatePersistenceStore) GetStats() (map[string]interface{}, error) {
	store.mutex.RLock()
	defer store.mutex.RUnlock()

	stats := map[string]interface{}{
		"workflow_instances": len(store.workflowInstances),
		"total_states":       0,
		"total_transitions":  0,
		"total_work":         0,
		"total_contexts":     0,
	}

	for instanceID := range store.workflowInstances {
		stats["total_states"] = stats["total_states"].(int) + len(store.states[instanceID])
		stats["total_transitions"] = stats["total_transitions"].(int) + len(store.transitions[instanceID])
		stats["total_work"] = stats["total_work"].(int) + len(store.work[instanceID])
		stats["total_contexts"] = stats["total_contexts"].(int) + len(store.contexts[instanceID])
	}

	return stats, nil
}
