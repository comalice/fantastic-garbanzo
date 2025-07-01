# Pluggable Work Execution System Design
## Architectural Design for UBOM Workflow Engine Extensions

---

## Executive Summary

This document presents architectural approaches for extending the UBOM Workflow Engine with a pluggable work execution system that maintains core library integrity while enabling flexible work definitions. The design is informed by the established development philosophy emphasizing simplicity, orthogonality, and modular composability.

---

## Philosophy Analysis and Design Principles

### Core Philosophy Synthesis

Based on the development philosophy document, our pluggable work execution system must adhere to these foundational principles:

**1. Orthogonality and Conceptual Integrity**
- Work definitions must be distinct from the core workflow engine
- Each work type (docker, gRPC, serverless) represents a separate domain concept
- Plugin boundaries must reflect real-world distinctions

**2. Modular and Composable Design**
- Core engine remains unchanged and unopinionated about work types
- Work definitions compose through well-defined interfaces
- Each plugin has single responsibility and clear boundaries

**3. Simplicity and First Principles**
- Avoid complex plugin frameworks; use language features and minimal dependencies
- Build on established patterns rather than inventing new abstractions
- Keep the plugin interface minimal and focused

**4. Data-Centric Design**
- Define canonical work definition data structures
- Ensure consistency across storage, execution, and API layers
- Work metadata should be predictable and stable

**5. Clarity and Maintainability**
- Plugin system must be self-documenting through clear interfaces
- Extension points should be obvious and well-documented
- Avoid clever abstractions that obscure functionality

---

## Architectural Approaches

### Approach 1: Interface-Based Plugin Registry

**Core Concept**: Define a minimal work execution interface and maintain a registry of implementations.

```go
// Core interface in workflow engine
type WorkExecutor interface {
    Execute(ctx context.Context, definition WorkDefinition, inputs map[string]interface{}) (WorkResult, error)
    Validate(definition WorkDefinition) error
    GetSchema() WorkSchema
}

// Registry pattern
type WorkRegistry struct {
    executors map[string]WorkExecutor
}

func (r *WorkRegistry) Register(workType string, executor WorkExecutor) {
    r.executors[workType] = executor
}
```

**Plugin Structure**:
```
plugins/
├── docker/
│   ├── executor.go      // Implements WorkExecutor
│   ├── schema.go        // Work definition schema
│   └── plugin.go        // Registration logic
├── grpc/
│   ├── executor.go
│   ├── schema.go
│   └── plugin.go
└── serverless/
    ├── executor.go
    ├── schema.go
    └── plugin.go
```

**Pros**:
- Simple, Go-native approach using interfaces
- Clear separation between core and plugins
- Easy to test and mock
- Minimal dependencies

**Cons**:
- Requires compile-time linking of plugins
- Less dynamic than runtime plugin loading
- Plugin discovery is manual

### Approach 2: Factory Pattern with Dynamic Registration

**Core Concept**: Use factory functions and init() registration for automatic plugin discovery.

```go
// Core factory interface
type WorkExecutorFactory func(config map[string]interface{}) (WorkExecutor, error)

// Global registry with init() registration
var globalRegistry = NewWorkRegistry()

func RegisterWorkType(workType string, factory WorkExecutorFactory, schema WorkSchema) {
    globalRegistry.Register(workType, factory, schema)
}

// Plugin registration in init()
func init() {
    RegisterWorkType("docker", NewDockerExecutor, DockerWorkSchema)
}
```

**Plugin Organization**:
```
workdefs/
├── builtin/
│   ├── docker/
│   ├── grpc/
│   └── serverless/
├── external/
│   └── custom_plugin/
└── registry.go
```

**Pros**:
- Automatic plugin discovery through init()
- Clean separation of built-in vs external plugins
- Factory pattern allows configuration flexibility
- Maintains compile-time safety

**Cons**:
- Global state through registry
- Init() order dependencies possible
- Still requires compile-time inclusion

### Approach 3: Configuration-Driven Plugin System

**Core Concept**: Define work types through configuration files and use reflection/code generation for execution.

```yaml
# work-definitions.yaml
work_types:
  docker:
    executor: "builtin.docker"
    schema: "schemas/docker.json"
    config:
      default_runtime: "containerd"
  
  grpc:
    executor: "builtin.grpc"
    schema: "schemas/grpc.json"
    
  custom_api:
    executor: "plugin.custom_api"
    schema: "schemas/custom_api.json"
    plugin_path: "./plugins/custom_api.so"
```

**Implementation Structure**:
```go
type WorkDefinitionConfig struct {
    WorkTypes map[string]WorkTypeConfig `yaml:"work_types"`
}

type WorkTypeConfig struct {
    Executor   string                 `yaml:"executor"`
    Schema     string                 `yaml:"schema"`
    Config     map[string]interface{} `yaml:"config"`
    PluginPath string                 `yaml:"plugin_path,omitempty"`
}
```

**Pros**:
- Highly configurable without code changes
- Clear separation of built-in vs external
- Schema validation through JSON Schema
- Runtime plugin loading possible

**Cons**:
- More complex configuration management
- Potential runtime errors from misconfiguration
- Requires additional tooling for validation

### Approach 4: Layered Extension Architecture

**Core Concept**: Create distinct layers for core execution, built-in work types, and external plugins.

```
┌─────────────────────────────────────┐
│           External Plugins          │
├─────────────────────────────────────┤
│         Built-in Work Types         │
│    (docker, grpc, serverless)       │
├─────────────────────────────────────┤
│        Work Execution Layer         │
│     (common execution logic)        │
├─────────────────────────────────────┤
│         UBOM Core Engine            │
│        (unchanged)                  │
└─────────────────────────────────────┘
```

**Layer Definitions**:

1. **Core Engine**: Unchanged UBOM workflow engine
2. **Work Execution Layer**: Common execution patterns, result handling, error management
3. **Built-in Work Types**: Standard implementations (docker, gRPC, serverless)
4. **External Plugins**: User-defined work types

**Implementation**:
```go
// Work Execution Layer
type BaseWorkExecutor struct {
    logger Logger
    metrics MetricsCollector
}

func (b *BaseWorkExecutor) ExecuteWithCommonLogic(ctx context.Context, fn ExecutionFunc) (WorkResult, error) {
    // Common pre-execution logic
    // Error handling, logging, metrics
    // Post-execution cleanup
}

// Built-in work types extend base
type DockerExecutor struct {
    BaseWorkExecutor
    dockerClient DockerClient
}

// External plugins implement interface
type ExternalWorkExecutor interface {
    WorkExecutor
    PluginInfo() PluginMetadata
}
```

**Pros**:
- Clear architectural boundaries
- Shared common functionality
- Easy to extend with new layers
- Follows orthogonality principle

**Cons**:
- More complex initial setup
- Potential for layer coupling
- Requires careful interface design

---

## Plugin System Design Details

### Core Interface Definition

```go
// Minimal, focused interface following simplicity principle
type WorkExecutor interface {
    // Execute work with given definition and inputs
    Execute(ctx context.Context, definition WorkDefinition, inputs map[string]interface{}) (WorkResult, error)
    
    // Validate work definition before execution
    Validate(definition WorkDefinition) error
    
    // Return schema for work definition validation
    GetSchema() WorkSchema
    
    // Return metadata about this work type
    GetMetadata() WorkMetadata
}

// Canonical data structures (data-centric design)
type WorkDefinition struct {
    Type       string                 `json:"type"`
    Version    string                 `json:"version"`
    Config     map[string]interface{} `json:"config"`
    Metadata   map[string]string      `json:"metadata"`
}

type WorkResult struct {
    Success    bool                   `json:"success"`
    Outputs    map[string]interface{} `json:"outputs"`
    Logs       []LogEntry            `json:"logs"`
    Metrics    ExecutionMetrics      `json:"metrics"`
    Error      string                `json:"error,omitempty"`
}

type WorkSchema struct {
    JSONSchema string            `json:"json_schema"`
    Examples   []WorkDefinition  `json:"examples"`
    Documentation string         `json:"documentation"`
}
```

### Built-in Work Definitions

#### Docker Work Definition
```go
type DockerExecutor struct {
    client DockerClient
    logger Logger
}

type DockerWorkConfig struct {
    Image       string            `json:"image"`
    Command     []string          `json:"command,omitempty"`
    Environment map[string]string `json:"environment,omitempty"`
    Volumes     []VolumeMount     `json:"volumes,omitempty"`
    Resources   ResourceLimits    `json:"resources,omitempty"`
}

func (d *DockerExecutor) Execute(ctx context.Context, definition WorkDefinition, inputs map[string]interface{}) (WorkResult, error) {
    var config DockerWorkConfig
    if err := mapstructure.Decode(definition.Config, &config); err != nil {
        return WorkResult{}, fmt.Errorf("invalid docker config: %w", err)
    }
    
    // Docker execution logic
    container, err := d.client.CreateContainer(ctx, config, inputs)
    if err != nil {
        return WorkResult{Success: false, Error: err.Error()}, nil
    }
    
    result, err := d.client.RunContainer(ctx, container.ID)
    return d.convertToWorkResult(result), err
}
```

#### gRPC/API Work Definition
```go
type GRPCExecutor struct {
    connectionPool ConnectionPool
    logger         Logger
}

type GRPCWorkConfig struct {
    Endpoint    string            `json:"endpoint"`
    Method      string            `json:"method"`
    Headers     map[string]string `json:"headers,omitempty"`
    Timeout     time.Duration     `json:"timeout,omitempty"`
    Retry       RetryConfig       `json:"retry,omitempty"`
    TLS         TLSConfig         `json:"tls,omitempty"`
}

func (g *GRPCExecutor) Execute(ctx context.Context, definition WorkDefinition, inputs map[string]interface{}) (WorkResult, error) {
    var config GRPCWorkConfig
    if err := mapstructure.Decode(definition.Config, &config); err != nil {
        return WorkResult{}, fmt.Errorf("invalid grpc config: %w", err)
    }
    
    // gRPC call execution logic
    conn, err := g.connectionPool.GetConnection(config.Endpoint, config.TLS)
    if err != nil {
        return WorkResult{Success: false, Error: err.Error()}, nil
    }
    
    response, err := g.makeGRPCCall(ctx, conn, config, inputs)
    return g.convertToWorkResult(response), err
}
```

#### Serverless Work Definition
```go
type ServerlessExecutor struct {
    providers map[string]ServerlessProvider
    logger    Logger
}

type ServerlessWorkConfig struct {
    Provider     string            `json:"provider"` // aws, gcp, azure
    Function     string            `json:"function"`
    Region       string            `json:"region,omitempty"`
    Credentials  CredentialConfig  `json:"credentials,omitempty"`
    Timeout      time.Duration     `json:"timeout,omitempty"`
    Environment  map[string]string `json:"environment,omitempty"`
}

func (s *ServerlessExecutor) Execute(ctx context.Context, definition WorkDefinition, inputs map[string]interface{}) (WorkResult, error) {
    var config ServerlessWorkConfig
    if err := mapstructure.Decode(definition.Config, &config); err != nil {
        return WorkResult{}, fmt.Errorf("invalid serverless config: %w", err)
    }
    
    provider, exists := s.providers[config.Provider]
    if !exists {
        return WorkResult{Success: false, Error: fmt.Sprintf("unsupported provider: %s", config.Provider)}, nil
    }
    
    result, err := provider.InvokeFunction(ctx, config, inputs)
    return s.convertToWorkResult(result), err
}
```

### External Plugin Implementation

#### Plugin Interface
```go
// External plugins must implement this extended interface
type ExternalWorkPlugin interface {
    WorkExecutor
    
    // Plugin lifecycle
    Initialize(config map[string]interface{}) error
    Shutdown() error
    
    // Plugin metadata
    GetPluginInfo() PluginInfo
    
    // Health check
    HealthCheck() error
}

type PluginInfo struct {
    Name        string   `json:"name"`
    Version     string   `json:"version"`
    Author      string   `json:"author"`
    Description string   `json:"description"`
    WorkTypes   []string `json:"work_types"`
}
```

#### Example External Plugin
```go
// Custom database work plugin
type DatabaseWorkPlugin struct {
    connections map[string]*sql.DB
    logger      Logger
}

func (d *DatabaseWorkPlugin) Initialize(config map[string]interface{}) error {
    // Initialize database connections
    d.connections = make(map[string]*sql.DB)
    return nil
}

func (d *DatabaseWorkPlugin) Execute(ctx context.Context, definition WorkDefinition, inputs map[string]interface{}) (WorkResult, error) {
    var config DatabaseWorkConfig
    if err := mapstructure.Decode(definition.Config, &config); err != nil {
        return WorkResult{}, fmt.Errorf("invalid database config: %w", err)
    }
    
    db, err := d.getConnection(config.ConnectionString)
    if err != nil {
        return WorkResult{Success: false, Error: err.Error()}, nil
    }
    
    result, err := d.executeQuery(ctx, db, config.Query, inputs)
    return d.convertToWorkResult(result), err
}

func (d *DatabaseWorkPlugin) GetPluginInfo() PluginInfo {
    return PluginInfo{
        Name:        "database-work",
        Version:     "1.0.0",
        Author:      "Custom Plugin Team",
        Description: "Execute database queries as workflow work",
        WorkTypes:   []string{"database", "sql"},
    }
}
```

---

## Implementation Strategy

### Phase 1: Core Interface and Registry
1. Define minimal WorkExecutor interface
2. Implement basic registry pattern
3. Create work definition data structures
4. Add validation framework

### Phase 2: Built-in Work Types
1. Implement DockerExecutor with comprehensive Docker support
2. Implement GRPCExecutor with connection pooling and retry logic
3. Implement ServerlessExecutor with multi-provider support
4. Add comprehensive testing for each work type

### Phase 3: Plugin System
1. Define external plugin interface
2. Implement plugin discovery and loading
3. Add plugin lifecycle management
4. Create plugin development documentation and examples

### Phase 4: Integration and Tooling
1. Integrate with UBOM workflow engine
2. Create CLI tools for plugin management
3. Add monitoring and observability
4. Performance optimization and profiling

### Boundary Rules and Core Protection

**Core Library Boundaries**:
- UBOM workflow engine remains completely unchanged
- No direct dependencies from core to work execution system
- Work execution system depends only on stable core interfaces

**Plugin Boundaries**:
- Plugins cannot access core engine internals
- All communication through defined interfaces
- Plugin failures cannot crash core system
- Resource isolation between plugins

**Data Boundaries**:
- Canonical work definition format
- Standardized result structures
- Schema validation at boundaries
- No plugin-specific data in core storage

---

## Recommended Approach

### Primary Recommendation: Layered Extension Architecture (Approach 4)

**Rationale**:
1. **Orthogonality**: Clear separation between core, execution layer, built-ins, and plugins
2. **Simplicity**: Each layer has focused responsibility
3. **Composability**: Layers can be developed and tested independently
4. **Maintainability**: Clear boundaries make system easier to understand and modify
5. **Extensibility**: New work types can be added without affecting existing code

**Implementation Priority**:
1. Start with Interface-Based Registry (Approach 1) for immediate needs
2. Evolve to Layered Architecture as system matures
3. Add Configuration-Driven aspects (Approach 3) for operational flexibility

### Secondary Recommendation: Hybrid Approach

Combine the best aspects of multiple approaches:
- **Core**: Interface-based registry for simplicity
- **Built-ins**: Factory pattern with init() registration
- **External**: Configuration-driven plugin loading
- **Operations**: Layered architecture for clear boundaries

---

## Concrete Implementation Examples

### Workflow Definition with Pluggable Work
```yaml
# workflow.yaml
name: "data-processing-pipeline"
version: "1.0"

work:
  - name: "extract-data"
    type: "docker"
    config:
      image: "data-extractor:latest"
      environment:
        SOURCE_URL: "${inputs.source_url}"
      resources:
        memory: "512Mi"
        cpu: "0.5"
    
  - name: "transform-data"
    type: "grpc"
    config:
      endpoint: "transform-service:9090"
      method: "TransformData"
      timeout: "30s"
      retry:
        max_attempts: 3
        backoff: "exponential"
    
  - name: "load-data"
    type: "serverless"
    config:
      provider: "aws"
      function: "data-loader"
      region: "us-west-2"
      timeout: "5m"
    
  - name: "notify-completion"
    type: "custom-notification"  # External plugin
    config:
      webhook_url: "${secrets.notification_webhook}"
      message_template: "Pipeline completed successfully"

dependencies:
  - from: "extract-data"
    to: "transform-data"
  - from: "transform-data"
    to: "load-data"
  - from: "load-data"
    to: "notify-completion"
```

### Plugin Registration Example
```go
// main.go - Application startup
func main() {
    // Initialize work registry
    registry := workexec.NewRegistry()
    
    // Register built-in work types
    registry.Register("docker", docker.NewExecutor())
    registry.Register("grpc", grpc.NewExecutor())
    registry.Register("serverless", serverless.NewExecutor())
    
    // Load external plugins
    pluginLoader := workexec.NewPluginLoader()
    plugins, err := pluginLoader.LoadPlugins("./plugins")
    if err != nil {
        log.Fatalf("Failed to load plugins: %v", err)
    }
    
    for _, plugin := range plugins {
        for _, workType := range plugin.GetPluginInfo().WorkTypes {
            registry.Register(workType, plugin)
        }
    }
    
    // Initialize workflow engine with work registry
    engine := ubom.NewEngine(ubom.Config{
        WorkRegistry: registry,
    })
    
    // Start engine
    engine.Start()
}
```

---

## Conclusion

The proposed pluggable work execution system maintains strict adherence to the established development philosophy while providing the flexibility needed for diverse work types. The layered extension architecture provides the best balance of simplicity, orthogonality, and extensibility.

Key success factors:
- **Minimal Core Changes**: Zero modifications to UBOM workflow engine
- **Clear Boundaries**: Well-defined interfaces between all components
- **Incremental Development**: Can be implemented in phases
- **Plugin Safety**: Robust isolation and error handling
- **Developer Experience**: Clear documentation and examples

This design enables the UBOM Workflow Engine to support diverse work types while maintaining its core integrity and following established architectural principles.
