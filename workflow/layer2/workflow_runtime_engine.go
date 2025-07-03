package layer2

import (
	"fmt"
	"sync"
	"time"

	"github.com/ubom/workflow/layer0"
	"github.com/ubom/workflow/layer1"
)

// WorkflowRuntimeEngine provides the main runtime engine for executing workflows
type WorkflowRuntimeEngine struct {
	stateMachineCore        *layer1.StateMachineCore
	workExecutionCore       *layer1.WorkExecutionCore
	conditionEvaluationCore *layer1.ConditionEvaluationCore
	persistenceStore        StatePersistenceStore
	transitionEvaluator     TransitionEvaluator
	errorHandler            ErrorHandler
	lifecycleManager        WorkflowLifecycleManager
	activeInstances         map[WorkflowInstanceID]*WorkflowInstance
	mutex                   sync.RWMutex
}

// WorkflowRuntimeEngineInterface defines the contract for the workflow runtime engine
type WorkflowRuntimeEngineInterface interface {
	// Lifecycle operations
	StartWorkflow(definition layer1.WorkflowDefinition, initialContext layer0.Context) (WorkflowInstanceID, error)
	StopWorkflow(instanceID WorkflowInstanceID) error
	PauseWorkflow(instanceID WorkflowInstanceID) error
	ResumeWorkflow(instanceID WorkflowInstanceID) error
	CancelWorkflow(instanceID WorkflowInstanceID) error

	// Execution operations
	ExecuteStep(instanceID WorkflowInstanceID) error
	ExecuteWorkflow(instanceID WorkflowInstanceID) error

	// Query operations
	GetWorkflowInstance(instanceID WorkflowInstanceID) (*WorkflowInstance, error)
	GetWorkflowStatus(instanceID WorkflowInstanceID) (WorkflowInstanceStatus, error)
	ListActiveWorkflows() []WorkflowInstanceID

	// Configuration
	SetPersistenceStore(store StatePersistenceStore)
	SetTransitionEvaluator(evaluator TransitionEvaluator)
	SetErrorHandler(handler ErrorHandler)
	SetLifecycleManager(manager WorkflowLifecycleManager)

	// Cleanup
	Shutdown() error
}

// NewWorkflowRuntimeEngine creates a new workflow runtime engine
func NewWorkflowRuntimeEngine() *WorkflowRuntimeEngine {
	return &WorkflowRuntimeEngine{
		stateMachineCore:        layer1.NewStateMachineCore(),
		workExecutionCore:       layer1.NewWorkExecutionCore(),
		conditionEvaluationCore: layer1.NewConditionEvaluationCore(),
		persistenceStore:        NewInMemoryStatePersistenceStore(),
		transitionEvaluator:     NewDefaultTransitionEvaluator(),
		errorHandler:            NewDefaultErrorHandler(),
		lifecycleManager:        NewDefaultWorkflowLifecycleManager(),
		activeInstances:         make(map[WorkflowInstanceID]*WorkflowInstance),
		mutex:                   sync.RWMutex{},
	}
}

// StartWorkflow starts a new workflow instance
func (engine *WorkflowRuntimeEngine) StartWorkflow(definition layer1.WorkflowDefinition, initialContext *layer0.Context) (WorkflowInstanceID, error) {
	if !definition.CanExecute() {
		return "", fmt.Errorf("workflow definition cannot be executed")
	}

	// Generate instance ID
	instanceID := WorkflowInstanceID(fmt.Sprintf("%s-%d", definition.GetID(), time.Now().UnixNano()))

	// Create workflow instance
	now := time.Now()
	instance := WorkflowInstance{
		ID:                instanceID,
		DefinitionID:      definition.GetID(),
		DefinitionVersion: definition.GetVersion(),
		Status:            WorkflowInstanceStatusCreated,
		CurrentStateID:    definition.GetInitialStateID(),
		Context:           initialContext,
		CreatedAt:         now,
		UpdatedAt:         now,
		Metadata:          make(map[string]interface{}),
	}

	// Save to persistence store
	if err := engine.persistenceStore.SaveWorkflowInstance(instance); err != nil {
		return "", fmt.Errorf("failed to save workflow instance: %w", err)
	}

	// Add to active instances
	engine.mutex.Lock()
	engine.activeInstances[instanceID] = &instance
	engine.mutex.Unlock()

	// Initialize state machine with definition
	engine.stateMachineCore = definition.GetStateMachine()

	// Notify lifecycle manager
	if err := engine.lifecycleManager.OnWorkflowStarted(instanceID); err != nil {
		engine.errorHandler.HandleError(instanceID, fmt.Errorf("lifecycle manager error: %w", err))
	}

	// Start execution
	instance.Status = WorkflowInstanceStatusRunning
	startedAt := time.Now()
	instance.StartedAt = &startedAt
	instance.UpdatedAt = startedAt

	// Update persistence
	if err := engine.persistenceStore.UpdateWorkflowInstance(instance); err != nil {
		return instanceID, fmt.Errorf("failed to update workflow instance: %w", err)
	}

	// Update active instance
	engine.mutex.Lock()
	engine.activeInstances[instanceID] = &instance
	engine.mutex.Unlock()

	return instanceID, nil
}

// StopWorkflow stops a running workflow instance
func (engine *WorkflowRuntimeEngine) StopWorkflow(instanceID WorkflowInstanceID) error {
	engine.mutex.Lock()
	defer engine.mutex.Unlock()

	instance, exists := engine.activeInstances[instanceID]
	if !exists {
		return fmt.Errorf("workflow instance %s not found", instanceID)
	}

	if instance.Status != WorkflowInstanceStatusRunning {
		return fmt.Errorf("workflow instance %s is not running", instanceID)
	}

	// Update status
	instance.Status = WorkflowInstanceStatusCompleted
	now := time.Now()
	instance.CompletedAt = &now
	instance.UpdatedAt = now

	// Update persistence
	if err := engine.persistenceStore.UpdateWorkflowInstance(*instance); err != nil {
		return fmt.Errorf("failed to update workflow instance: %w", err)
	}

	// Remove from active instances
	delete(engine.activeInstances, instanceID)

	// Notify lifecycle manager
	if err := engine.lifecycleManager.OnWorkflowCompleted(instanceID); err != nil {
		engine.errorHandler.HandleError(instanceID, fmt.Errorf("lifecycle manager error: %w", err))
	}

	return nil
}

// PauseWorkflow pauses a running workflow instance
func (engine *WorkflowRuntimeEngine) PauseWorkflow(instanceID WorkflowInstanceID) error {
	engine.mutex.Lock()
	defer engine.mutex.Unlock()

	instance, exists := engine.activeInstances[instanceID]
	if !exists {
		return fmt.Errorf("workflow instance %s not found", instanceID)
	}

	if instance.Status != WorkflowInstanceStatusRunning {
		return fmt.Errorf("workflow instance %s is not running", instanceID)
	}

	// Update status
	instance.Status = WorkflowInstanceStatusPaused
	instance.UpdatedAt = time.Now()

	// Update persistence
	if err := engine.persistenceStore.UpdateWorkflowInstance(*instance); err != nil {
		return fmt.Errorf("failed to update workflow instance: %w", err)
	}

	// Notify lifecycle manager
	if err := engine.lifecycleManager.OnWorkflowPaused(instanceID); err != nil {
		engine.errorHandler.HandleError(instanceID, fmt.Errorf("lifecycle manager error: %w", err))
	}

	return nil
}

// ResumeWorkflow resumes a paused workflow instance
func (engine *WorkflowRuntimeEngine) ResumeWorkflow(instanceID WorkflowInstanceID) error {
	engine.mutex.Lock()
	defer engine.mutex.Unlock()

	instance, exists := engine.activeInstances[instanceID]
	if !exists {
		return fmt.Errorf("workflow instance %s not found", instanceID)
	}

	if instance.Status != WorkflowInstanceStatusPaused {
		return fmt.Errorf("workflow instance %s is not paused", instanceID)
	}

	// Update status
	instance.Status = WorkflowInstanceStatusRunning
	instance.UpdatedAt = time.Now()

	// Update persistence
	if err := engine.persistenceStore.UpdateWorkflowInstance(*instance); err != nil {
		return fmt.Errorf("failed to update workflow instance: %w", err)
	}

	// Notify lifecycle manager
	if err := engine.lifecycleManager.OnWorkflowResumed(instanceID); err != nil {
		engine.errorHandler.HandleError(instanceID, fmt.Errorf("lifecycle manager error: %w", err))
	}

	return nil
}

// CancelWorkflow cancels a workflow instance
func (engine *WorkflowRuntimeEngine) CancelWorkflow(instanceID WorkflowInstanceID) error {
	engine.mutex.Lock()
	defer engine.mutex.Unlock()

	return engine.cancelWorkflowUnsafe(instanceID)
}

// cancelWorkflowUnsafe cancels a workflow instance without acquiring the mutex
// This method assumes the caller already holds the mutex lock
func (engine *WorkflowRuntimeEngine) cancelWorkflowUnsafe(instanceID WorkflowInstanceID) error {
	instance, exists := engine.activeInstances[instanceID]
	if !exists {
		return fmt.Errorf("workflow instance %s not found", instanceID)
	}

	// Update status
	instance.Status = WorkflowInstanceStatusCancelled
	now := time.Now()
	instance.CompletedAt = &now
	instance.UpdatedAt = now

	// Update persistence
	if err := engine.persistenceStore.UpdateWorkflowInstance(*instance); err != nil {
		return fmt.Errorf("failed to update workflow instance: %w", err)
	}

	// Remove from active instances
	delete(engine.activeInstances, instanceID)

	// Notify lifecycle manager
	if err := engine.lifecycleManager.OnWorkflowCancelled(instanceID); err != nil {
		engine.errorHandler.HandleError(instanceID, fmt.Errorf("lifecycle manager error: %w", err))
	}

	return nil
}

// ExecuteStep executes a single step of the workflow
func (engine *WorkflowRuntimeEngine) ExecuteStep(instanceID WorkflowInstanceID) error {
	engine.mutex.RLock()
	instance, exists := engine.activeInstances[instanceID]
	engine.mutex.RUnlock()

	if !exists {
		return fmt.Errorf("workflow instance %s not found", instanceID)
	}

	if instance.Status != WorkflowInstanceStatusRunning {
		return fmt.Errorf("workflow instance %s is not running", instanceID)
	}

	// Get current state
	currentState, err := engine.stateMachineCore.GetState(instance.CurrentStateID)
	if err != nil {
		return fmt.Errorf("failed to get current state: %w", err)
	}

	// Check if current state is final
	if currentState.IsFinal() {
		return engine.StopWorkflow(instanceID)
	}

	// Get available transitions
	transitions := engine.stateMachineCore.GetTransitionsFromState(instance.CurrentStateID)
	if len(transitions) == 0 {
		return fmt.Errorf("no transitions available from state %s", instance.CurrentStateID)
	}

	// Evaluate transitions
	for _, transition := range transitions {
		canTransition, err := engine.transitionEvaluator.CanTransition(transition, instance.Context)
		if err != nil {
			engine.errorHandler.HandleError(instanceID, fmt.Errorf("transition evaluation error: %w", err))
			continue
		}

		if canTransition {
			// Execute transition
			if err := engine.executeTransition(instanceID, transition); err != nil {
				engine.errorHandler.HandleError(instanceID, fmt.Errorf("transition execution error: %w", err))
				continue
			}
			return nil
		}
	}

	return fmt.Errorf("no valid transitions found from state %s", instance.CurrentStateID)
}

// executeTransition executes a specific transition
func (engine *WorkflowRuntimeEngine) executeTransition(instanceID WorkflowInstanceID, transition layer0.Transition) error {
	engine.mutex.Lock()
	instance := engine.activeInstances[instanceID]
	engine.mutex.Unlock()

	// Execute transition actions (work items)
	for _, actionID := range transition.GetActions() {
		// Create work item
		work := layer0.NewWork(layer0.WorkID(actionID), layer0.WorkTypeTask, fmt.Sprintf("Action %s", actionID))

		// Execute work
		result, err := engine.workExecutionCore.ExecuteWork(work, instance.Context)
		if err != nil {
			return fmt.Errorf("failed to execute work %s: %w", actionID, err)
		}

		if result.Status == layer0.WorkStatusFailed {
			return fmt.Errorf("work %s failed: %s", actionID, result.Error)
		}

		// Update context with work output if available
		if result.Output != nil {
			instance.Context = instance.Context.Set(fmt.Sprintf("work_%s_output", actionID), result.Output)
		}
	}

	// Update current state
	instance.CurrentStateID = transition.GetToStateID()
	instance.UpdatedAt = time.Now()

	// Update persistence
	if err := engine.persistenceStore.UpdateWorkflowInstance(*instance); err != nil {
		return fmt.Errorf("failed to update workflow instance: %w", err)
	}

	// Update active instance
	engine.mutex.Lock()
	engine.activeInstances[instanceID] = instance
	engine.mutex.Unlock()

	return nil
}

// ExecuteWorkflow executes a workflow until completion or error
func (engine *WorkflowRuntimeEngine) ExecuteWorkflow(instanceID WorkflowInstanceID) error {
	maxSteps := 1000 // Prevent infinite loops

	for i := 0; i < maxSteps; i++ {
		err := engine.ExecuteStep(instanceID)
		if err != nil {
			// Check if workflow completed normally
			if err.Error() == fmt.Sprintf("workflow instance %s is not running", instanceID) {
				return nil // Workflow completed
			}
			return err
		}

		// Check if workflow is still running
		engine.mutex.RLock()
		instance, exists := engine.activeInstances[instanceID]
		engine.mutex.RUnlock()

		if !exists || instance.Status != WorkflowInstanceStatusRunning {
			return nil // Workflow completed or stopped
		}
	}

	return fmt.Errorf("workflow execution exceeded maximum steps (%d)", maxSteps)
}

// GetWorkflowInstance retrieves a workflow instance
func (engine *WorkflowRuntimeEngine) GetWorkflowInstance(instanceID WorkflowInstanceID) (*WorkflowInstance, error) {
	engine.mutex.RLock()
	instance, exists := engine.activeInstances[instanceID]
	engine.mutex.RUnlock()

	if exists {
		return instance, nil
	}

	// Try to get from persistence store
	persistedInstance, err := engine.persistenceStore.GetWorkflowInstance(instanceID)
	if err != nil {
		return nil, fmt.Errorf("workflow instance %s not found", instanceID)
	}

	return &persistedInstance, nil
}

// GetWorkflowStatus retrieves the status of a workflow instance
func (engine *WorkflowRuntimeEngine) GetWorkflowStatus(instanceID WorkflowInstanceID) (WorkflowInstanceStatus, error) {
	instance, err := engine.GetWorkflowInstance(instanceID)
	if err != nil {
		return "", err
	}

	return instance.Status, nil
}

// ListActiveWorkflows returns a list of active workflow instance IDs
func (engine *WorkflowRuntimeEngine) ListActiveWorkflows() []WorkflowInstanceID {
	engine.mutex.RLock()
	defer engine.mutex.RUnlock()

	instanceIDs := make([]WorkflowInstanceID, 0, len(engine.activeInstances))
	for instanceID := range engine.activeInstances {
		instanceIDs = append(instanceIDs, instanceID)
	}

	return instanceIDs
}

// SetPersistenceStore sets the persistence store
func (engine *WorkflowRuntimeEngine) SetPersistenceStore(store StatePersistenceStore) {
	engine.persistenceStore = store
}

// SetTransitionEvaluator sets the transition evaluator
func (engine *WorkflowRuntimeEngine) SetTransitionEvaluator(evaluator TransitionEvaluator) {
	engine.transitionEvaluator = evaluator
}

// SetErrorHandler sets the error handler
func (engine *WorkflowRuntimeEngine) SetErrorHandler(handler ErrorHandler) {
	engine.errorHandler = handler
}

// SetLifecycleManager sets the lifecycle manager
func (engine *WorkflowRuntimeEngine) SetLifecycleManager(manager WorkflowLifecycleManager) {
	engine.lifecycleManager = manager
}

// Shutdown shuts down the workflow runtime engine
func (engine *WorkflowRuntimeEngine) Shutdown() error {
	engine.mutex.Lock()
	defer engine.mutex.Unlock()

	// Stop all active workflows
	for instanceID := range engine.activeInstances {
		if err := engine.cancelWorkflowUnsafe(instanceID); err != nil {
			// Log error but continue shutdown
			engine.errorHandler.HandleError(instanceID, fmt.Errorf("shutdown error: %w", err))
		}
	}

	// Clear active instances
	engine.activeInstances = make(map[WorkflowInstanceID]*WorkflowInstance)

	return nil
}
