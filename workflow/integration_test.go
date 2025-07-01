
package workflow

import (
        "context"
        "fmt"
        "testing"
        "time"

        "github.com/ubom/workflow/layer0"
        "github.com/ubom/workflow/layer1"
        "github.com/ubom/workflow/executors"
        "github.com/ubom/workflow/executors/docker"
        "github.com/ubom/workflow/executors/grpc"
        "github.com/ubom/workflow/executors/serverless"
        "github.com/ubom/workflow/plugins"
        "github.com/ubom/workflow/schemas"
)

// Integration tests for the pluggable work execution system

func TestPluggableExecutionIntegration(t *testing.T) {
        // Create enhanced registry
        registry := executors.NewEnhancedWorkRegistry()
        
        // Create schema validator
        validator := schemas.NewSchemaValidator()
        
        // Register built-in executors
        dockerExecutor := docker.NewDockerExecutor(&MockDockerClient{}, &MockLogger{})
        grpcExecutor := grpc.NewGRPCExecutor(&MockConnectionPool{}, &MockLogger{})
        serverlessExecutor := serverless.NewServerlessExecutor(map[string]serverless.ServerlessProvider{
                "aws": &MockServerlessProvider{name: "aws"},
        }, &MockLogger{})
        
        // Register executors
        err := registry.RegisterExecutor(docker.WorkTypeDocker, dockerExecutor)
        if err != nil {
                t.Fatalf("Failed to register Docker executor: %v", err)
        }
        
        err = registry.RegisterExecutor(grpc.WorkTypeGRPC, grpcExecutor)
        if err != nil {
                t.Fatalf("Failed to register gRPC executor: %v", err)
        }
        
        err = registry.RegisterExecutor(serverless.WorkTypeServerless, serverlessExecutor)
        if err != nil {
                t.Fatalf("Failed to register Serverless executor: %v", err)
        }
        
        // Register schemas
        validator.RegisterSchema(docker.WorkTypeDocker, dockerExecutor.GetSchema())
        validator.RegisterSchema(grpc.WorkTypeGRPC, grpcExecutor.GetSchema())
        validator.RegisterSchema(serverless.WorkTypeServerless, serverlessExecutor.GetSchema())
        
        // Test Docker work execution
        t.Run("DockerExecution", func(t *testing.T) {
                work := createDockerWork()
                testWorkExecution(t, registry, validator, work)
        })
        
        // Test gRPC work execution
        t.Run("GRPCExecution", func(t *testing.T) {
                work := createGRPCWork()
                testWorkExecution(t, registry, validator, work)
        })
        
        // Test Serverless work execution
        t.Run("ServerlessExecution", func(t *testing.T) {
                work := createServerlessWork()
                testWorkExecution(t, registry, validator, work)
        })
}

func TestPluginSystemIntegration(t *testing.T) {
        // Create plugin loader
        loader := plugins.NewDefaultPluginLoader(&MockLogger{})
        
        // Register a mock plugin
        factory := func() plugins.ExternalWorkPlugin {
                return NewMockPlugin()
        }
        
        err := loader.RegisterPlugin("mock-plugin", factory)
        if err != nil {
                t.Fatalf("Failed to register plugin: %v", err)
        }
        
        // Initialize plugin
        config := map[string]interface{}{"test": "config"}
        err = loader.InitializePlugin("mock-plugin", config)
        if err != nil {
                t.Fatalf("Failed to initialize plugin: %v", err)
        }
        
        // Test plugin execution
        plugin, err := loader.GetPlugin("mock-plugin")
        if err != nil {
                t.Fatalf("Failed to get plugin: %v", err)
        }
        
        work := layer0.NewWork("test-work", "mock", "Test Mock Work")
        ctx := context.Background()
        workContext := &layer0.Context{}
        
        result, err := plugin.Execute(ctx, work, workContext)
        if err != nil {
                t.Fatalf("Plugin execution failed: %v", err)
        }
        
        if !result.Success {
                t.Errorf("Expected successful execution, got failure: %s", result.Error)
        }
        
        // Test health check
        err = plugin.HealthCheck()
        if err != nil {
                t.Errorf("Plugin health check failed: %v", err)
        }
        
        // Test shutdown
        err = loader.ShutdownPlugin("mock-plugin")
        if err != nil {
                t.Errorf("Failed to shutdown plugin: %v", err)
        }
}

func TestBackwardCompatibility(t *testing.T) {
        // Test that enhanced executors work with the original interface
        dockerExecutor := docker.NewDockerExecutor(&MockDockerClient{}, &MockLogger{})
        adapter := executors.NewBackwardCompatibilityAdapter(dockerExecutor)
        
        // Create original work execution core
        core := layer1.NewWorkExecutionCore()
        
        // Register adapted executor
        err := core.RegisterExecutor(docker.WorkTypeDocker, adapter)
        if err != nil {
                t.Fatalf("Failed to register adapted executor: %v", err)
        }
        
        // Test execution with original interface
        work := createDockerWork()
        workContext := &layer0.Context{}
        
        result, err := core.ExecuteWork(work, workContext)
        if err != nil {
                t.Fatalf("Backward compatibility execution failed: %v", err)
        }
        
        if result.Status != layer0.WorkStatusCompleted {
                t.Errorf("Expected completed status, got %s", result.Status)
        }
}

func TestConcurrentExecution(t *testing.T) {
        registry := executors.NewEnhancedWorkRegistry()
        dockerExecutor := docker.NewDockerExecutor(&MockDockerClient{}, &MockLogger{})
        
        err := registry.RegisterExecutor(docker.WorkTypeDocker, dockerExecutor)
        if err != nil {
                t.Fatalf("Failed to register executor: %v", err)
        }
        
        // Execute multiple works concurrently
        numWorkers := 10
        results := make(chan error, numWorkers)
        
        for i := 0; i < numWorkers; i++ {
                go func(id int) {
                        work := createDockerWorkWithID(layer0.WorkID(fmt.Sprintf("concurrent-work-%d", id)))
                        executor, err := registry.GetExecutor(docker.WorkTypeDocker)
                        if err != nil {
                                results <- err
                                return
                        }
                        
                        ctx := context.Background()
                        workContext := &layer0.Context{}
                        
                        result, err := executor.Execute(ctx, work, workContext)
                        if err != nil {
                                results <- err
                                return
                        }
                        
                        if !result.Success {
                                results <- fmt.Errorf("execution failed: %s", result.Error)
                                return
                        }
                        
                        results <- nil
                }(i)
        }
        
        // Wait for all workers to complete
        for i := 0; i < numWorkers; i++ {
                select {
                case err := <-results:
                        if err != nil {
                                t.Errorf("Concurrent execution failed: %v", err)
                        }
                case <-time.After(30 * time.Second):
                        t.Fatal("Concurrent execution timed out")
                }
        }
}

// Helper functions

func testWorkExecution(t *testing.T, registry *executors.EnhancedWorkRegistry, validator *schemas.SchemaValidator, work layer0.Work) {
        // Validate work
        err := validator.ValidateWork(work)
        if err != nil {
                t.Fatalf("Work validation failed: %v", err)
        }
        
        // Get executor
        executor, err := registry.GetExecutor(work.GetType())
        if err != nil {
                t.Fatalf("Failed to get executor: %v", err)
        }
        
        // Execute work
        ctx := context.Background()
        workContext := &layer0.Context{}
        
        result, err := executor.Execute(ctx, work, workContext)
        if err != nil {
                t.Fatalf("Work execution failed: %v", err)
        }
        
        if !result.Success {
                t.Errorf("Expected successful execution, got failure: %s", result.Error)
        }
        
        // Validate result structure
        if result.Outputs == nil {
                t.Error("Expected outputs in result")
        }
        
        if result.Metrics.Duration <= 0 {
                t.Error("Expected positive execution duration")
        }
}

func createDockerWork() layer0.Work {
        return createDockerWorkWithID("docker-test-work")
}

func createDockerWorkWithID(id layer0.WorkID) layer0.Work {
        work := layer0.NewWork(id, docker.WorkTypeDocker, "Test Docker Work")
        config := work.GetConfiguration()
        config.Parameters["executor_config"] = map[string]interface{}{
                "image": "ubuntu:20.04",
                "command": []string{"echo", "hello"},
        }
        return layer0.Work{
                ID:            work.GetID(),
                Type:          work.GetType(),
                Status:        work.GetStatus(),
                Priority:      work.GetPriority(),
                Metadata:      work.GetMetadata(),
                Configuration: config,
                Input:         work.GetInput(),
                Output:        work.GetOutput(),
                Error:         work.GetError(),
        }
}

func createGRPCWork() layer0.Work {
        work := layer0.NewWork("grpc-test-work", grpc.WorkTypeGRPC, "Test gRPC Work")
        config := work.GetConfiguration()
        config.Parameters["executor_config"] = map[string]interface{}{
                "endpoint": "localhost:9090",
                "method":   "test.Service/TestMethod",
        }
        return layer0.Work{
                ID:            work.GetID(),
                Type:          work.GetType(),
                Status:        work.GetStatus(),
                Priority:      work.GetPriority(),
                Metadata:      work.GetMetadata(),
                Configuration: config,
                Input:         work.GetInput(),
                Output:        work.GetOutput(),
                Error:         work.GetError(),
        }
}

func createServerlessWork() layer0.Work {
        work := layer0.NewWork("serverless-test-work", serverless.WorkTypeServerless, "Test Serverless Work")
        config := work.GetConfiguration()
        config.Parameters["executor_config"] = map[string]interface{}{
                "provider": "aws",
                "function": "test-function",
                "region":   "us-east-1",
        }
        return layer0.Work{
                ID:            work.GetID(),
                Type:          work.GetType(),
                Status:        work.GetStatus(),
                Priority:      work.GetPriority(),
                Metadata:      work.GetMetadata(),
                Configuration: config,
                Input:         work.GetInput(),
                Output:        work.GetOutput(),
                Error:         work.GetError(),
        }
}

// Mock implementations for testing

type MockDockerClient struct{}

func (m *MockDockerClient) RunContainer(ctx context.Context, config docker.DockerWorkConfig, input interface{}) (interface{}, []executors.LogEntry, error) {
        return "mock docker output", []executors.LogEntry{
                {
                        Timestamp: time.Now(),
                        Level:     "INFO",
                        Message:   "Container started",
                        Source:    "docker",
                },
        }, nil
}

func (m *MockDockerClient) PullImage(ctx context.Context, image string) error {
        return nil
}

func (m *MockDockerClient) RemoveContainer(ctx context.Context, containerID string) error {
        return nil
}

type MockConnectionPool struct{}

func (m *MockConnectionPool) GetConnection(endpoint string, tlsConfig grpc.TLSConfig) (grpc.GRPCConnection, error) {
        return &MockGRPCConnection{}, nil
}

func (m *MockConnectionPool) ReleaseConnection(conn grpc.GRPCConnection) error {
        return nil
}

func (m *MockConnectionPool) Close() error {
        return nil
}

type MockGRPCConnection struct{}

func (m *MockGRPCConnection) Invoke(ctx context.Context, method string, request interface{}, response interface{}, metadata map[string]string) error {
        return nil
}

func (m *MockGRPCConnection) Close() error {
        return nil
}

type MockServerlessProvider struct {
        name string
}

func (m *MockServerlessProvider) InvokeFunction(ctx context.Context, config serverless.ServerlessWorkConfig, payload interface{}) (interface{}, []executors.LogEntry, error) {
        return "mock serverless response", []executors.LogEntry{
                {
                        Timestamp: time.Now(),
                        Level:     "INFO",
                        Message:   "Function invoked",
                        Source:    "serverless",
                },
        }, nil
}

func (m *MockServerlessProvider) ValidateConfig(config serverless.ServerlessWorkConfig) error {
        return nil
}

func (m *MockServerlessProvider) GetProviderName() string {
        return m.name
}

type MockLogger struct{}

func (m *MockLogger) Info(msg string, fields ...interface{})  {}
func (m *MockLogger) Error(msg string, fields ...interface{}) {}
func (m *MockLogger) Debug(msg string, fields ...interface{}) {}
func (m *MockLogger) Warn(msg string, fields ...interface{})  {}

type MockPlugin struct {
        *plugins.BasePlugin
        initialized bool
}

func NewMockPlugin() *MockPlugin {
        info := plugins.PluginInfo{
                Name:        "mock-plugin",
                Version:     "1.0.0",
                Author:      "Test",
                Description: "Mock plugin for testing",
                WorkTypes:   []string{"mock"},
                APIVersion:  "1.0",
        }

        metadata := executors.WorkMetadata{
                Name:        "Mock Plugin",
                Version:     "1.0.0",
                Author:      "Test",
                Description: "Mock plugin for testing",
                WorkTypes:   []string{"mock"},
        }

        schema := executors.WorkSchema{
                JSONSchema:    `{"type": "object"}`,
                Examples:      []executors.WorkDefinition{},
                Documentation: "Mock plugin schema",
        }

        return &MockPlugin{
                BasePlugin: plugins.NewBasePlugin(info, metadata, schema),
        }
}

func (p *MockPlugin) Execute(ctx context.Context, work layer0.Work, workContext *layer0.Context) (executors.WorkResult, error) {
        return executors.WorkResult{
                Success: true,
                Outputs: map[string]interface{}{"result": "mock"},
                Metrics: executors.ExecutionMetrics{
                        StartTime: time.Now(),
                        EndTime:   time.Now(),
                        Duration:  time.Millisecond,
                },
        }, nil
}

func (p *MockPlugin) Validate(work layer0.Work) error {
        return nil
}

func (p *MockPlugin) CanExecute(workType layer0.WorkType) bool {
        return workType == "mock"
}

func (p *MockPlugin) GetSupportedTypes() []layer0.WorkType {
        return []layer0.WorkType{"mock"}
}

func (p *MockPlugin) Initialize(config map[string]interface{}) error {
        p.initialized = true
        return nil
}

func (p *MockPlugin) Shutdown() error {
        p.initialized = false
        return nil
}
