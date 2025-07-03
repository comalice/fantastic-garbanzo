package main

import (
	"fmt"
	"log"

	"github.com/ubom/workflow/layer0"
	"github.com/ubom/workflow/layer1"
	"github.com/ubom/workflow/layer2"
)

func main() {
	fmt.Println("UBOM Workflow Engine Demo")
	fmt.Println("========================")

	// Create workflow runtime engine
	engine := layer2.NewWorkflowRuntimeEngine()
	defer engine.Shutdown()

	// Create a simple workflow definition
	definition := createSampleWorkflow()

	// Create initial context
	initialContext := layer0.NewContext("demo-context", layer0.ContextScopeWorkflow, "Demo Context")
	initialContext = initialContext.Set("user_id", "demo-user")
	initialContext = initialContext.Set("process_data", true)

	fmt.Println("\n1. Starting workflow...")
	instanceID, err := engine.StartWorkflow(definition, initialContext)
	if err != nil {
		log.Fatalf("Failed to start workflow: %v", err)
	}
	fmt.Printf("   Workflow started with instance ID: %s\n", instanceID)

	// Check initial status
	status, err := engine.GetWorkflowStatus(instanceID)
	if err != nil {
		log.Fatalf("Failed to get workflow status: %v", err)
	}
	fmt.Printf("   Initial status: %s\n", status)

	fmt.Println("\n2. Executing workflow...")
	err = engine.ExecuteWorkflow(instanceID)
	if err != nil {
		log.Fatalf("Failed to execute workflow: %v", err)
	}

	// Check final status
	finalStatus, err := engine.GetWorkflowStatus(instanceID)
	if err != nil {
		log.Fatalf("Failed to get final workflow status: %v", err)
	}
	fmt.Printf("   Final status: %s\n", finalStatus)

	// Get workflow instance details
	instance, err := engine.GetWorkflowInstance(instanceID)
	if err != nil {
		log.Fatalf("Failed to get workflow instance: %v", err)
	}

	fmt.Println("\n3. Workflow execution summary:")
	fmt.Printf("   Instance ID: %s\n", instance.ID)
	fmt.Printf("   Definition ID: %s\n", instance.DefinitionID)
	fmt.Printf("   Status: %s\n", instance.Status)
	fmt.Printf("   Current State: %s\n", instance.CurrentStateID)
	fmt.Printf("   Created At: %s\n", instance.CreatedAt.Format("2006-01-02 15:04:05"))
	if instance.StartedAt != nil {
		fmt.Printf("   Started At: %s\n", instance.StartedAt.Format("2006-01-02 15:04:05"))
	}
	if instance.CompletedAt != nil {
		fmt.Printf("   Completed At: %s\n", instance.CompletedAt.Format("2006-01-02 15:04:05"))
		duration := instance.CompletedAt.Sub(*instance.StartedAt)
		fmt.Printf("   Duration: %s\n", duration)
	}

	fmt.Println("\n4. Demonstrating pause/resume functionality...")

	// Start another workflow for pause/resume demo
	pauseResumeInstanceID, err := engine.StartWorkflow(definition, initialContext)
	if err != nil {
		log.Fatalf("Failed to start second workflow: %v", err)
	}
	fmt.Printf("   Started workflow: %s\n", pauseResumeInstanceID)

	// Pause the workflow
	err = engine.PauseWorkflow(pauseResumeInstanceID)
	if err != nil {
		log.Fatalf("Failed to pause workflow: %v", err)
	}
	fmt.Printf("   Paused workflow: %s\n", pauseResumeInstanceID)

	// Check paused status
	pausedStatus, _ := engine.GetWorkflowStatus(pauseResumeInstanceID)
	fmt.Printf("   Status after pause: %s\n", pausedStatus)

	// Resume the workflow
	err = engine.ResumeWorkflow(pauseResumeInstanceID)
	if err != nil {
		log.Fatalf("Failed to resume workflow: %v", err)
	}
	fmt.Printf("   Resumed workflow: %s\n", pauseResumeInstanceID)

	// Execute the resumed workflow
	err = engine.ExecuteWorkflow(pauseResumeInstanceID)
	if err != nil {
		log.Fatalf("Failed to execute resumed workflow: %v", err)
	}

	resumedFinalStatus, _ := engine.GetWorkflowStatus(pauseResumeInstanceID)
	fmt.Printf("   Final status after resume: %s\n", resumedFinalStatus)

	fmt.Println("\n5. Workflow engine statistics:")

	// Get persistence store statistics
	store := layer2.NewInMemoryStatePersistenceStore()
	engine.SetPersistenceStore(store)

	stats, err := store.GetStats()
	if err == nil {
		fmt.Printf("   Workflow instances: %v\n", stats["workflow_instances"])
		fmt.Printf("   Total states: %v\n", stats["total_states"])
		fmt.Printf("   Total transitions: %v\n", stats["total_transitions"])
		fmt.Printf("   Total work items: %v\n", stats["total_work"])
		fmt.Printf("   Total contexts: %v\n", stats["total_contexts"])
	}

	fmt.Println("\nDemo completed successfully!")
}

func createSampleWorkflow() layer1.WorkflowDefinition {
	// Create workflow definition
	definition := layer1.NewWorkflowDefinition("sample-workflow", "1.0.0", "Sample Workflow")

	// Create state machine
	stateMachine := layer1.NewStateMachineCore()

	// Create states
	initialState := layer0.NewState("start", layer0.StateTypeInitial, "Start State")
	processingState := layer0.NewState("processing", layer0.StateTypeIntermediate, "Processing State")
	finalState := layer0.NewState("end", layer0.StateTypeFinal, "End State")
	errorState := layer0.NewState("error", layer0.StateTypeError, "Error State")

	// Add states to state machine
	if err := stateMachine.AddState(initialState); err != nil {
		log.Fatalf("Failed to add initial state: %v", err)
	}
	if err := stateMachine.AddState(processingState); err != nil {
		log.Fatalf("Failed to add processing state: %v", err)
	}
	if err := stateMachine.AddState(finalState); err != nil {
		log.Fatalf("Failed to add final state: %v", err)
	}
	if err := stateMachine.AddState(errorState); err != nil {
		log.Fatalf("Failed to add error state: %v", err)
	}

	// Create transitions
	startToProcessing := layer0.NewTransition("start-to-processing", layer0.TransitionTypeAutomatic,
		initialState.GetID(), processingState.GetID(), "Start to Processing")
	startToProcessing = startToProcessing.AddCondition("process_data")

	processingToEnd := layer0.NewTransition("processing-to-end", layer0.TransitionTypeAutomatic,
		processingState.GetID(), finalState.GetID(), "Processing to End")

	processingToError := layer0.NewTransition("processing-to-error", layer0.TransitionTypeAutomatic,
		processingState.GetID(), errorState.GetID(), "Processing to Error")

	// Add transitions to state machine
	if err := stateMachine.AddTransition(startToProcessing); err != nil {
		log.Fatalf("Failed to add start-to-processing transition: %v", err)
	}
	if err := stateMachine.AddTransition(processingToEnd); err != nil {
		log.Fatalf("Failed to add processing-to-end transition: %v", err)
	}
	if err := stateMachine.AddTransition(processingToError); err != nil {
		log.Fatalf("Failed to add processing-to-error transition: %v", err)
	}

	// Update workflow definition
	definition = definition.SetStateMachine(stateMachine).
		SetInitialStateID(initialState.GetID()).
		AddFinalStateID(finalState.GetID()).
		AddErrorStateID(errorState.GetID()).
		SetStatus(layer1.WorkflowDefinitionStatusActive)

	// Validate the workflow definition
	if err := definition.Validate(); err != nil {
		log.Fatalf("Invalid workflow definition: %v", err)
	}

	return definition
}
