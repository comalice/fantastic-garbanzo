
# UBOM Workflow Engine

A comprehensive workflow engine implementation in Go following the UBOM (Universal Business Object Model) specification. This engine provides a robust, scalable, and maintainable solution for orchestrating complex business processes.

## Architecture Overview

The workflow engine follows a 3-layer architecture designed for orthogonality, immutability, and compensation logic:

```
┌─────────────────────────────────────────────────────────────┐
│                    Layer 2: Operational                    │
│  ┌─────────────────┐ ┌─────────────────┐ ┌──────────────┐  │
│  │ Runtime Engine  │ │ Persistence     │ │ Error        │  │
│  │                 │ │ Store           │ │ Handler      │  │
│  └─────────────────┘ └─────────────────┘ └──────────────┘  │
│  ┌─────────────────┐ ┌─────────────────┐ ┌──────────────┐  │
│  │ Transition      │ │ Lifecycle       │ │              │  │
│  │ Evaluator       │ │ Manager         │ │              │  │
│  └─────────────────┘ └─────────────────┘ └──────────────┘  │
├─────────────────────────────────────────────────────────────┤
│                   Layer 1: Compositional                   │
│  ┌─────────────────┐ ┌─────────────────┐ ┌──────────────┐  │
│  │ StateMachine    │ │ WorkExecution   │ │ Condition    │  │
│  │ Core            │ │ Core            │ │ Evaluation   │  │
│  └─────────────────┘ └─────────────────┘ └──────────────┘  │
│  ┌─────────────────┐ ┌─────────────────┐ ┌──────────────┐  │
│  │ Workflow        │ │ Orchestration   │ │              │  │
│  │ Definition      │ │ Services        │ │              │  │
│  └─────────────────┘ └─────────────────┘ └──────────────┘  │
├─────────────────────────────────────────────────────────────┤
│                  Layer 0: Atomic Foundation                │
│  ┌─────────────────┐ ┌─────────────────┐ ┌──────────────┐  │
│  │ State           │ │ Transition      │ │ Work         │  │
│  │                 │ │                 │ │              │  │
│  └─────────────────┘ └─────────────────┘ └──────────────┘  │
│  ┌─────────────────┐ ┌─────────────────┐ ┌──────────────┐  │
│  │ Condition       │ │ Context         │ │              │  │
│  │                 │ │                 │ │              │  │
│  └─────────────────┘ └─────────────────┘ └──────────────┘  │
└─────────────────────────────────────────────────────────────┘
```

## Key Features

- **Immutable Design**: All core components follow immutability principles
- **Thread-Safe**: Concurrent execution support for ≤100 users
- **Scalable**: Handles ≤250 million records efficiently
- **Extensible**: Plugin architecture for custom executors and evaluators
- **Persistent**: Configurable persistence with in-memory implementation
- **Observable**: Comprehensive lifecycle management and error handling
- **Testable**: Each component is independently testable

## Layer Descriptions

### Layer 0: Atomic Foundation
The foundational layer containing five core atomic components:

- **State**: Represents workflow states with metadata and validation
- **Transition**: Defines state transitions with conditions and actions
- **Work**: Encapsulates executable work units with retry policies
- **Condition**: Provides conditional logic evaluation
- **Context**: Thread-safe data storage across workflow execution

### Layer 1: Compositional
Core services that compose atomic components:

- **StateMachineCore**: Manages states and transitions
- **WorkExecutionCore**: Handles work execution with pluggable executors
- **ConditionEvaluationCore**: Evaluates conditions with custom evaluators
- **WorkflowDefinition**: Complete workflow specifications
- **Orchestration Services**: Workflow instance management

### Layer 2: Operational
Runtime framework components:

- **WorkflowRuntimeEngine**: Main execution engine
- **StatePersistenceStore**: Data persistence interface with in-memory implementation
- **TransitionEvaluator**: Evaluates transition conditions
- **ErrorHandler**: Comprehensive error management
- **WorkflowLifecycleManager**: Lifecycle event management

## Quick Start

### Installation

```bash
go mod init your-project
go get github.com/ubom/workflow
```

### Basic Usage

```go
package main

import (
    "github.com/ubom/workflow/layer0"
    "github.com/ubom/workflow/layer1"
    "github.com/ubom/workflow/layer2"
)

func main() {
    // Create workflow runtime engine
    engine := layer2.NewWorkflowRuntimeEngine()
    defer engine.Shutdown()

    // Create workflow definition
    definition := createWorkflow()

    // Create initial context
    context := layer0.NewContext("demo-context", layer0.ContextScopeWorkflow, "Demo")
    context = context.Set("user_id", "user123")

    // Start workflow
    instanceID, err := engine.StartWorkflow(definition, context)
    if err != nil {
        panic(err)
    }

    // Execute workflow
    err = engine.ExecuteWorkflow(instanceID)
    if err != nil {
        panic(err)
    }

    // Check status
    status, _ := engine.GetWorkflowStatus(instanceID)
    fmt.Printf("Workflow completed with status: %s\n", status)
}

func createWorkflow() layer1.WorkflowDefinition {
    // Create workflow definition
    definition := layer1.NewWorkflowDefinition("sample", "1.0.0", "Sample Workflow")

    // Create state machine
    stateMachine := layer1.NewStateMachineCore()

    // Create states
    start := layer0.NewState("start", layer0.StateTypeInitial, "Start")
    end := layer0.NewState("end", layer0.StateTypeFinal, "End")

    // Add states
    stateMachine.AddState(start)
    stateMachine.AddState(end)

    // Create transition
    transition := layer0.NewTransition("start-to-end", layer0.TransitionTypeAutomatic, 
        start.GetID(), end.GetID(), "Start to End")
    stateMachine.AddTransition(transition)

    // Configure workflow
    return definition.SetStateMachine(stateMachine).
        SetInitialStateID(start.GetID()).
        AddFinalStateID(end.GetID()).
        SetStatus(layer1.WorkflowDefinitionStatusActive)
}
```

## Running the Demo

A complete demonstration is available in the `cmd` directory:

```bash
cd ~/ubom/workflow
go run ./cmd
```

This demo showcases:
- Workflow creation and execution
- State transitions
- Pause/resume functionality
- Error handling
- Lifecycle management

## Testing

The engine includes comprehensive tests for all layers:

```bash
# Test all layers
go test ./...

# Test specific layer
go test ./layer0/...
go test ./layer1/...
go test ./layer2/...

# Run with verbose output
go test -v ./...
```

## Configuration

### Persistence Store

The engine supports pluggable persistence stores. The default in-memory implementation is suitable for development and testing:

```go
// Use custom persistence store
customStore := NewCustomPersistenceStore()
engine.SetPersistenceStore(customStore)
```

### Work Executors

Register custom work executors for different work types:

```go
// Create custom executor
executor := &CustomWorkExecutor{}

// Register with work execution core
workCore := layer1.NewWorkExecutionCore()
workCore.RegisterExecutor(layer0.WorkTypeCustom, executor)
```

### Condition Evaluators

Add custom condition evaluators:

```go
// Create custom evaluator
evaluator := &CustomConditionEvaluator{}

// Register with condition evaluation core
conditionCore := layer1.NewConditionEvaluationCore()
conditionCore.RegisterEvaluator(layer0.ConditionTypeCustom, evaluator)
```

## Design Principles

### Orthogonality
Each component has a single, well-defined responsibility and can be used independently.

### Immutability
All state changes create new instances rather than modifying existing ones, ensuring thread safety and predictable behavior.

### Compensation Logic
Built-in support for compensating transactions and rollback scenarios.

### Unidirectional Dependencies
Clear dependency flow from Layer 2 → Layer 1 → Layer 0, preventing circular dependencies.

### Extension Points
Well-defined interfaces allow for custom implementations without modifying core components.

## Performance Characteristics

- **Concurrency**: Supports ≤100 concurrent users
- **Scale**: Handles ≤250 million records
- **Memory**: Efficient in-memory operations with optional persistence
- **Latency**: Low-latency state transitions and work execution

## Contributing

1. Follow the 3-layer architecture
2. Maintain immutability principles
3. Add comprehensive tests for new components
4. Update documentation for API changes
5. Ensure thread safety for concurrent operations

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Support

For questions, issues, or contributions, please refer to the project repository or contact the development team.
