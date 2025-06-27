# NOTE: Vibecoded using abacus.ai !!!

# UBOM - Universal Business Object Model

A comprehensive Product Lifecycle Management (PLM) system implementation in Go following the UBOM specification. This system provides a robust, scalable, and maintainable solution for managing complex business processes across the entire product lifecycle.

## Project Structure

```
ubom/
├── go.mod                    # Root module definition
├── go.sum                    # Dependency checksums
├── README.md                 # This file
└── workflow/                 # Workflow Engine Component
    ├── README.md            # Workflow-specific documentation
    ├── cmd/                 # Command-line applications
    ├── layer0/              # Atomic Foundation Layer
    ├── layer1/              # Compositional Layer
    └── layer2/              # Operational Layer
```

## Architecture Overview

UBOM is designed as a modular PLM system with multiple components:

- **Workflow Engine** - Core workflow orchestration and execution
- **Quality Management** - Quality control and assurance processes (planned)
- **Change Management** - Change control and approval workflows (planned)
- **CRUD Interfaces** - Data management and API endpoints (planned)
- **Document Management** - Document lifecycle and version control (planned)
- **Supplier Management** - Supplier relationship and qualification (planned)

## Current Implementation Status

### ✅ Workflow Engine (Complete)
The workflow engine is fully implemented with a 3-layer architecture:

- **Layer 0**: Atomic components (State, Transition, Work, Condition, Context)
- **Layer 1**: Compositional services (StateMachine, WorkExecution, ConditionEvaluation)
- **Layer 2**: Operational framework (RuntimeEngine, Persistence, ErrorHandling)

### 🚧 Planned Components
- Quality Management System
- Change Management System
- CRUD Interface Layer
- Document Management System
- Supplier Management System

## Quick Start

### Installation

```bash
git clone <repository-url>
cd ubom
go mod download
```

### Using the Workflow Engine

```go
package main

import (
    "fmt"
    "github.com/ubom/workflow/layer0"
    "github.com/ubom/workflow/layer1"
    "github.com/ubom/workflow/layer2"
)

func main() {
    // Create workflow runtime engine
    engine := layer2.NewWorkflowRuntimeEngine()
    defer engine.Shutdown()

    // Create workflow definition
    definition := createSampleWorkflow()

    // Create initial context
    context := layer0.NewContext("demo-context", layer0.ContextScopeWorkflow, "Demo")
    context = context.Set("user_id", "user123")

    // Start and execute workflow
    instanceID, err := engine.StartWorkflow(definition, context)
    if err != nil {
        panic(err)
    }

    err = engine.ExecuteWorkflow(instanceID)
    if err != nil {
        panic(err)
    }

    // Check final status
    status, _ := engine.GetWorkflowStatus(instanceID)
    fmt.Printf("Workflow completed with status: %s\n", status)
}
```

### Running the Demo

```bash
# Run the workflow engine demo
go run ./workflow/cmd

# Run tests for all components
go test ./...

# Run tests for specific component
go test ./workflow/...
```

## Module Structure

The project uses Go modules with the root module at `github.com/ubom`. Each component is organized as a subpackage:

- `github.com/ubom/workflow` - Workflow engine implementation
- `github.com/ubom/quality` - Quality management (planned)
- `github.com/ubom/change` - Change management (planned)
- `github.com/ubom/crud` - CRUD interfaces (planned)

## Development Guidelines

### Adding New Components

1. Create a new directory under the root (e.g., `quality/`, `change/`)
2. Follow the established 3-layer architecture pattern where applicable
3. Import the workflow engine using: `import "github.com/ubom/workflow/layerX"`
4. Add comprehensive tests and documentation
5. Update this README with the new component status

### Code Organization

- **Root Level**: Module definition, shared utilities, and system-wide configuration
- **Component Level**: Individual PLM system components
- **Layer Level**: (For workflow engine) Architectural layers with clear separation

### Testing

```bash
# Test entire system
go test ./...

# Test specific component
go test ./workflow/...

# Test with coverage
go test -cover ./...

# Verbose test output
go test -v ./...
```

## Workflow Engine Features

The workflow engine provides:

- **Immutable Design**: Thread-safe operations with immutable state
- **Scalable Architecture**: Supports ≤100 concurrent users, ≤250M records
- **Extensible Framework**: Plugin architecture for custom executors and evaluators
- **Persistent Storage**: Configurable persistence with in-memory implementation
- **Comprehensive Testing**: Full test coverage across all layers
- **Observable Operations**: Lifecycle management and error handling

For detailed workflow engine documentation, see [workflow/README.md](workflow/README.md).

## Configuration

### Environment Setup

```bash
# Set Go environment
export GO111MODULE=on
export GOPROXY=https://proxy.golang.org

# Build all components
go build ./...

# Install dependencies
go mod tidy
```

### Module Dependencies

The project uses minimal external dependencies to maintain simplicity and reduce security surface area. Core dependencies are limited to:

- Go standard library
- Testing frameworks (for development)

## Performance Characteristics

### Workflow Engine
- **Concurrency**: ≤100 concurrent users
- **Scale**: ≤250 million records
- **Memory**: Efficient in-memory operations
- **Latency**: Low-latency state transitions

### System-wide (Planned)
- **Multi-tenant**: Support for multiple organizations
- **Distributed**: Microservices architecture capability
- **High Availability**: Fault-tolerant design patterns

## Contributing

1. **Architecture**: Follow the established patterns and layer separation
2. **Testing**: Add comprehensive tests for all new functionality
3. **Documentation**: Update relevant README files and code comments
4. **Dependencies**: Minimize external dependencies, justify additions
5. **Performance**: Consider scalability and concurrency requirements

### Pull Request Process

1. Create feature branch from main
2. Implement changes following project conventions
3. Add/update tests with good coverage
4. Update documentation as needed
5. Submit PR with clear description of changes

## Roadmap

### Phase 1: Foundation (Complete)
- ✅ Workflow Engine implementation
- ✅ 3-layer architecture establishment
- ✅ Core testing framework
- ✅ Module restructuring

### Phase 2: Core PLM Components (Planned)
- 🚧 Quality Management System
- 🚧 Change Management System
- 🚧 Basic CRUD interfaces

### Phase 3: Advanced Features (Future)
- 📋 Document Management
- 📋 Supplier Management
- 📋 Integration APIs
- 📋 Web UI components

### Phase 4: Enterprise Features (Future)
- 📋 Multi-tenant architecture
- 📋 Advanced analytics
- 📋 Compliance reporting
- 📋 Third-party integrations

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Support

For questions, issues, or contributions:

1. Check existing documentation in component README files
2. Review test files for usage examples
3. Create issues for bugs or feature requests
4. Submit pull requests for contributions

---

**Note**: This is an active development project. The workflow engine is production-ready, while other PLM components are in planning/development phases. Check individual component README files for specific status and usage instructions.
