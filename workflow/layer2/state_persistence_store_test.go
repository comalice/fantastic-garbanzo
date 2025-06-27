
package layer2

import (
        "testing"
        "time"

        "github.com/ubom/workflow/layer0"
)

func TestNewInMemoryStatePersistenceStore(t *testing.T) {
        store := NewInMemoryStatePersistenceStore()
        
        if store == nil {
                t.Error("NewInMemoryStatePersistenceStore should return a non-nil instance")
        }
        
        instances, err := store.ListAllWorkflowInstances()
        if err != nil {
                t.Errorf("ListAllWorkflowInstances should not return error: %v", err)
        }
        
        if len(instances) != 0 {
                t.Errorf("New store should have 0 instances, got %d", len(instances))
        }
}

func TestWorkflowInstanceOperations(t *testing.T) {
        store := NewInMemoryStatePersistenceStore()
        
        // Create test instance
        instance := WorkflowInstance{
                ID:                "test-instance",
                DefinitionID:      "test-definition",
                DefinitionVersion: "1.0.0",
                Status:            WorkflowInstanceStatusCreated,
                CurrentStateID:    "initial-state",
                Context:           layer0.NewContext("instance-context", layer0.ContextScopeWorkflow, "Instance Context"),
                CreatedAt:         time.Now(),
                UpdatedAt:         time.Now(),
                Metadata:          map[string]interface{}{"key": "value"},
        }
        
        // Test SaveWorkflowInstance
        err := store.SaveWorkflowInstance(instance)
        if err != nil {
                t.Errorf("SaveWorkflowInstance should not return error: %v", err)
        }
        
        // Test duplicate save
        err = store.SaveWorkflowInstance(instance)
        if err == nil {
                t.Error("SaveWorkflowInstance should return error for duplicate instance")
        }
        
        // Test GetWorkflowInstance
        retrievedInstance, err := store.GetWorkflowInstance(instance.ID)
        if err != nil {
                t.Errorf("GetWorkflowInstance should not return error: %v", err)
        }
        
        if retrievedInstance.ID != instance.ID {
                t.Error("Retrieved instance should match saved instance")
        }
        
        // Test UpdateWorkflowInstance
        instance.Status = WorkflowInstanceStatusRunning
        err = store.UpdateWorkflowInstance(instance)
        if err != nil {
                t.Errorf("UpdateWorkflowInstance should not return error: %v", err)
        }
        
        updatedInstance, _ := store.GetWorkflowInstance(instance.ID)
        if updatedInstance.Status != WorkflowInstanceStatusRunning {
                t.Error("Instance status should be updated")
        }
        
        // Test ListWorkflowInstances
        instances, err := store.ListWorkflowInstances(instance.DefinitionID)
        if err != nil {
                t.Errorf("ListWorkflowInstances should not return error: %v", err)
        }
        
        if len(instances) != 1 {
                t.Errorf("Expected 1 instance, got %d", len(instances))
        }
        
        // Test ListAllWorkflowInstances
        allInstances, err := store.ListAllWorkflowInstances()
        if err != nil {
                t.Errorf("ListAllWorkflowInstances should not return error: %v", err)
        }
        
        if len(allInstances) != 1 {
                t.Errorf("Expected 1 instance, got %d", len(allInstances))
        }
        
        // Test DeleteWorkflowInstance
        err = store.DeleteWorkflowInstance(instance.ID)
        if err != nil {
                t.Errorf("DeleteWorkflowInstance should not return error: %v", err)
        }
        
        _, err = store.GetWorkflowInstance(instance.ID)
        if err == nil {
                t.Error("GetWorkflowInstance should return error for deleted instance")
        }
}

func TestStateOperations(t *testing.T) {
        store := NewInMemoryStatePersistenceStore()
        
        // Create test instance
        instance := WorkflowInstance{
                ID:                "test-instance",
                DefinitionID:      "test-definition",
                DefinitionVersion: "1.0.0",
                Status:            WorkflowInstanceStatusCreated,
                CurrentStateID:    "initial-state",
                Context:           layer0.NewContext("instance-context", layer0.ContextScopeWorkflow, "Instance Context"),
                CreatedAt:         time.Now(),
                UpdatedAt:         time.Now(),
                Metadata:          map[string]interface{}{},
        }
        store.SaveWorkflowInstance(instance)
        
        // Create test state
        state := layer0.NewState("test-state", layer0.StateTypeInitial, "Test State")
        
        // Test SaveState
        err := store.SaveState(instance.ID, state)
        if err != nil {
                t.Errorf("SaveState should not return error: %v", err)
        }
        
        // Test duplicate save
        err = store.SaveState(instance.ID, state)
        if err == nil {
                t.Error("SaveState should return error for duplicate state")
        }
        
        // Test GetState
        retrievedState, err := store.GetState(instance.ID, state.GetID())
        if err != nil {
                t.Errorf("GetState should not return error: %v", err)
        }
        
        if retrievedState.GetID() != state.GetID() {
                t.Error("Retrieved state should match saved state")
        }
        
        // Test UpdateState
        updatedState := state.SetStatus(layer0.StateStatusActive)
        err = store.UpdateState(instance.ID, updatedState)
        if err != nil {
                t.Errorf("UpdateState should not return error: %v", err)
        }
        
        retrievedUpdatedState, _ := store.GetState(instance.ID, state.GetID())
        if retrievedUpdatedState.GetStatus() != layer0.StateStatusActive {
                t.Error("State status should be updated")
        }
        
        // Test ListStates
        states, err := store.ListStates(instance.ID)
        if err != nil {
                t.Errorf("ListStates should not return error: %v", err)
        }
        
        if len(states) != 1 {
                t.Errorf("Expected 1 state, got %d", len(states))
        }
        
        // Test operations on non-existent instance
        _, err = store.GetState("non-existent", state.GetID())
        if err == nil {
                t.Error("GetState should return error for non-existent instance")
        }
}

func TestTransitionOperations(t *testing.T) {
        store := NewInMemoryStatePersistenceStore()
        
        // Create test instance
        instance := WorkflowInstance{
                ID:                "test-instance",
                DefinitionID:      "test-definition",
                DefinitionVersion: "1.0.0",
                Status:            WorkflowInstanceStatusCreated,
                CurrentStateID:    "initial-state",
                Context:           layer0.NewContext("instance-context", layer0.ContextScopeWorkflow, "Instance Context"),
                CreatedAt:         time.Now(),
                UpdatedAt:         time.Now(),
                Metadata:          map[string]interface{}{},
        }
        store.SaveWorkflowInstance(instance)
        
        // Create test transition
        transition := layer0.NewTransition("test-transition", layer0.TransitionTypeAutomatic, "from-state", "to-state", "Test Transition")
        
        // Test SaveTransition
        err := store.SaveTransition(instance.ID, transition)
        if err != nil {
                t.Errorf("SaveTransition should not return error: %v", err)
        }
        
        // Test GetTransition
        retrievedTransition, err := store.GetTransition(instance.ID, transition.GetID())
        if err != nil {
                t.Errorf("GetTransition should not return error: %v", err)
        }
        
        if retrievedTransition.GetID() != transition.GetID() {
                t.Error("Retrieved transition should match saved transition")
        }
        
        // Test UpdateTransition
        updatedTransition := transition.SetStatus(layer0.TransitionStatusReady)
        err = store.UpdateTransition(instance.ID, updatedTransition)
        if err != nil {
                t.Errorf("UpdateTransition should not return error: %v", err)
        }
        
        // Test ListTransitions
        transitions, err := store.ListTransitions(instance.ID)
        if err != nil {
                t.Errorf("ListTransitions should not return error: %v", err)
        }
        
        if len(transitions) != 1 {
                t.Errorf("Expected 1 transition, got %d", len(transitions))
        }
}

func TestWorkOperations(t *testing.T) {
        store := NewInMemoryStatePersistenceStore()
        
        // Create test instance
        instance := WorkflowInstance{
                ID:                "test-instance",
                DefinitionID:      "test-definition",
                DefinitionVersion: "1.0.0",
                Status:            WorkflowInstanceStatusCreated,
                CurrentStateID:    "initial-state",
                Context:           layer0.NewContext("instance-context", layer0.ContextScopeWorkflow, "Instance Context"),
                CreatedAt:         time.Now(),
                UpdatedAt:         time.Now(),
                Metadata:          map[string]interface{}{},
        }
        store.SaveWorkflowInstance(instance)
        
        // Create test work
        work := layer0.NewWork("test-work", layer0.WorkTypeTask, "Test Work")
        
        // Test SaveWork
        err := store.SaveWork(instance.ID, work)
        if err != nil {
                t.Errorf("SaveWork should not return error: %v", err)
        }
        
        // Test GetWork
        retrievedWork, err := store.GetWork(instance.ID, work.GetID())
        if err != nil {
                t.Errorf("GetWork should not return error: %v", err)
        }
        
        if retrievedWork.GetID() != work.GetID() {
                t.Error("Retrieved work should match saved work")
        }
        
        // Test UpdateWork
        updatedWork := work.SetStatus(layer0.WorkStatusExecuting)
        err = store.UpdateWork(instance.ID, updatedWork)
        if err != nil {
                t.Errorf("UpdateWork should not return error: %v", err)
        }
        
        // Test ListWork
        workItems, err := store.ListWork(instance.ID)
        if err != nil {
                t.Errorf("ListWork should not return error: %v", err)
        }
        
        if len(workItems) != 1 {
                t.Errorf("Expected 1 work item, got %d", len(workItems))
        }
}

func TestContextOperations(t *testing.T) {
        store := NewInMemoryStatePersistenceStore()
        
        // Create test instance
        instance := WorkflowInstance{
                ID:                "test-instance",
                DefinitionID:      "test-definition",
                DefinitionVersion: "1.0.0",
                Status:            WorkflowInstanceStatusCreated,
                CurrentStateID:    "initial-state",
                Context:           layer0.NewContext("instance-context", layer0.ContextScopeWorkflow, "Instance Context"),
                CreatedAt:         time.Now(),
                UpdatedAt:         time.Now(),
                Metadata:          map[string]interface{}{},
        }
        store.SaveWorkflowInstance(instance)
        
        // Create test context
        context := layer0.NewContext("test-context", layer0.ContextScopeState, "Test Context")
        context = context.Set("key", "value")
        
        // Test SaveContext
        err := store.SaveContext(instance.ID, context)
        if err != nil {
                t.Errorf("SaveContext should not return error: %v", err)
        }
        
        // Test GetContext
        retrievedContext, err := store.GetContext(instance.ID, context.GetID())
        if err != nil {
                t.Errorf("GetContext should not return error: %v", err)
        }
        
        if retrievedContext.GetID() != context.GetID() {
                t.Error("Retrieved context should match saved context")
        }
        
        // Test UpdateContext
        updatedContext := context.Set("new-key", "new-value")
        err = store.UpdateContext(instance.ID, updatedContext)
        if err != nil {
                t.Errorf("UpdateContext should not return error: %v", err)
        }
        
        // Test ListContexts
        contexts, err := store.ListContexts(instance.ID)
        if err != nil {
                t.Errorf("ListContexts should not return error: %v", err)
        }
        
        if len(contexts) != 1 {
                t.Errorf("Expected 1 context, got %d", len(contexts))
        }
}

func TestCleanupAndStats(t *testing.T) {
        store := NewInMemoryStatePersistenceStore()
        
        // Create test data
        instance := WorkflowInstance{
                ID:                "test-instance",
                DefinitionID:      "test-definition",
                DefinitionVersion: "1.0.0",
                Status:            WorkflowInstanceStatusCreated,
                CurrentStateID:    "initial-state",
                Context:           layer0.NewContext("instance-context", layer0.ContextScopeWorkflow, "Instance Context"),
                CreatedAt:         time.Now(),
                UpdatedAt:         time.Now(),
                Metadata:          map[string]interface{}{},
        }
        store.SaveWorkflowInstance(instance)
        
        state := layer0.NewState("test-state", layer0.StateTypeInitial, "Test State")
        store.SaveState(instance.ID, state)
        
        transition := layer0.NewTransition("test-transition", layer0.TransitionTypeAutomatic, "from", "to", "Test Transition")
        store.SaveTransition(instance.ID, transition)
        
        work := layer0.NewWork("test-work", layer0.WorkTypeTask, "Test Work")
        store.SaveWork(instance.ID, work)
        
        context := layer0.NewContext("test-context", layer0.ContextScopeState, "Test Context")
        store.SaveContext(instance.ID, context)
        
        // Test GetStats
        stats, err := store.GetStats()
        if err != nil {
                t.Errorf("GetStats should not return error: %v", err)
        }
        
        if stats["workflow_instances"] != 1 {
                t.Errorf("Expected 1 workflow instance, got %v", stats["workflow_instances"])
        }
        
        if stats["total_states"] != 1 {
                t.Errorf("Expected 1 state, got %v", stats["total_states"])
        }
        
        if stats["total_transitions"] != 1 {
                t.Errorf("Expected 1 transition, got %v", stats["total_transitions"])
        }
        
        if stats["total_work"] != 1 {
                t.Errorf("Expected 1 work item, got %v", stats["total_work"])
        }
        
        if stats["total_contexts"] != 1 {
                t.Errorf("Expected 1 context, got %v", stats["total_contexts"])
        }
        
        // Test Cleanup
        err = store.Cleanup()
        if err != nil {
                t.Errorf("Cleanup should not return error: %v", err)
        }
        
        // Verify cleanup
        instances, _ := store.ListAllWorkflowInstances()
        if len(instances) != 0 {
                t.Errorf("Expected 0 instances after cleanup, got %d", len(instances))
        }
        
        stats, _ = store.GetStats()
        if stats["workflow_instances"] != 0 {
                t.Errorf("Expected 0 workflow instances after cleanup, got %v", stats["workflow_instances"])
        }
}

func TestErrorCases(t *testing.T) {
        store := NewInMemoryStatePersistenceStore()
        
        // Test operations on non-existent instance
        _, err := store.GetWorkflowInstance("non-existent")
        if err == nil {
                t.Error("GetWorkflowInstance should return error for non-existent instance")
        }
        
        err = store.UpdateWorkflowInstance(WorkflowInstance{ID: "non-existent"})
        if err == nil {
                t.Error("UpdateWorkflowInstance should return error for non-existent instance")
        }
        
        err = store.DeleteWorkflowInstance("non-existent")
        if err == nil {
                t.Error("DeleteWorkflowInstance should return error for non-existent instance")
        }
        
        // Test state operations on non-existent instance
        state := layer0.NewState("test-state", layer0.StateTypeInitial, "Test State")
        err = store.SaveState("non-existent", state)
        if err == nil {
                t.Error("SaveState should return error for non-existent instance")
        }
        
        // Test transition operations on non-existent instance
        transition := layer0.NewTransition("test-transition", layer0.TransitionTypeAutomatic, "from", "to", "Test Transition")
        err = store.SaveTransition("non-existent", transition)
        if err == nil {
                t.Error("SaveTransition should return error for non-existent instance")
        }
        
        // Test work operations on non-existent instance
        work := layer0.NewWork("test-work", layer0.WorkTypeTask, "Test Work")
        err = store.SaveWork("non-existent", work)
        if err == nil {
                t.Error("SaveWork should return error for non-existent instance")
        }
        
        // Test context operations on non-existent instance
        context := layer0.NewContext("test-context", layer0.ContextScopeState, "Test Context")
        err = store.SaveContext("non-existent", context)
        if err == nil {
                t.Error("SaveContext should return error for non-existent instance")
        }
}
