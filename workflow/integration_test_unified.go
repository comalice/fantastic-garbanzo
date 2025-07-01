package workflow

import (
	"context"
	"testing"
	"time"

	"github.com/ubom/workflow/layer0"
	"github.com/ubom/workflow/executors"
	"github.com/ubom/workflow/executors/docker"
	"github.com/ubom/workflow/executors/grpc"
	"github.com/ubom/workflow/executors/serverless"
	"github.com/ubom/workflow/overlays"
	"github.com/ubom/workflow/plugins"
)

// MockLogger for testing
type MockLogger struct{}

func (m *MockLogger) Info(msg string, fields ...interface{})  {}
func (m *MockLogger) Error(msg string, fields ...interface{}) {}
func (m *MockLogger) Debug(msg string, fields ...interface{}) {}
func (m *MockLogger) Warn(msg string, fields ...interface{})  {}

// MockMetricsCollector for testing
type MockMetricsCollector struct{}

func (m *MockMetricsCollector) RecordExecution(workType layer0.WorkType, duration time.Duration, success bool) {}
func (m *MockMetricsCollector) RecordResourceUsage(cpu float64, memory int64) {}

// MockDockerUtils for testing
type MockDockerUtils struct{}

func (m *MockDockerUtils) RunContainer(ctx context.Context, config docker.DockerWorkConfig, input interface{}) (interface{}, []executors.LogEntry, error) {
	return "docker output", []executors.LogEntry{
		{
			Timestamp: time.Now(),
			Level:     "INFO",
			Message:   "Container executed successfully",
			Source:    "mock-docker",
		},
	}, nil
}

func (m *MockDockerUtils) BuildImage(ctx context.Context, config docker.DockerWorkConfig) (string, error) {
	return "test-image:latest", nil
}

func (m *MockDockerUtils) PullImage(ctx context.Context, image string) error {
	return nil
}

func (m *MockDockerUtils) RemoveContainer(ctx context.Context, containerID string) error {
	return nil
}

func (m *MockDockerUtils) IsDockerAvailable() bool {
	return true
}

func (m *MockDockerUtils) GetDockerVersion() (string, error) {
	return "20.10.0", nil
}

// MockConnectionPool for gRPC testing
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

// MockGRPCConnection for testing
type MockGRPCConnection struct{}

func (m *MockGRPCConnection) Invoke(ctx context.Context, method string, request interface{}, response interface{}, metadata map[string]string) error {
	return nil
}

func (m *MockGRPCConnection) Close() error {
	return nil
}

// MockServerlessProvider for testing
type MockServerlessProvider struct{}

func (m *MockServerlessProvider) InvokeFunction(ctx context.Context, config serverless.ServerlessWorkConfig, payload interface{}) (interface{}, []executors.LogEntry, error) {
	return "serverless output", []executors.LogEntry{
		{
			Timestamp: time.Now(),
			Level:     "INFO",
			Message:   "Function executed successfully",
			Source:    "mock-serverless",
		},
	}, nil
}

func (m *MockServerlessProvider) ValidateConfig(config serverless.ServerlessWorkConfig) error {
	return nil
}

func (m *MockServerlessProvider) GetProviderName() string {
	return "mock-aws"
}

// TestUnifiedArchitecture tests the new unified architecture with overlay pattern
func TestUnifiedArchitecture(t *testing.T) {
	logger := &MockLogger{}
	
	// Create unified registry
	registry := NewUnifiedRegistry(logger)
	
	// Create base executors
	dockerExecutor := docker.NewDockerExecutor(logger)
	dockerExecutor.SetDockerUtils(&MockDockerUtils{}) // Set mock for testing
	
	grpcExecutor := grpc.NewGRPCExecutor(&MockConnectionPool{}, logger)
	
	serverlessExecutor := serverless.NewServerlessExecutor(
		map[string]serverless.ServerlessProvider{
			"aws": &MockServerlessProvider{},
		},
		logger,
	)
	
	// Test 1: Register base executors
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
	
	// Test 2: Verify registration
	supportedTypes := registry.GetSupportedWorkTypes()
	expectedTypes := []layer0.WorkType{docker.WorkTypeDocker, grpc.WorkTypeGRPC, serverless.WorkTypeServerless}
	
	if len(supportedTypes) != len(expectedTypes) {
		t.Fatalf("Expected %d supported types, got %d", len(expectedTypes), len(supportedTypes))
	}
	
	// Test 3: Test overlay pattern with metrics and logging
	metricsCollector := &MockMetricsCollector{}
	
	// Create overlay chain: Base Executor -> Metrics -> Logging -> Retry
	metricsOverlay := overlays.NewMetricsOverlay(metricsCollector)
	loggingOverlay := overlays.NewLoggingOverlay(logger)
	retryOverlay := overlays.NewRetryOverlay(3, time.Millisecond*100, "linear")
	
	// Chain the overlays
	metricsOverlay.SetNext(dockerExecutor)
	loggingOverlay.SetNext(metricsOverlay)
	retryOverlay.SetNext(loggingOverlay)
	
	// Register the overlay-enhanced executor
	err = registry.RegisterExecutor("docker-enhanced", retryOverlay)
	if err != nil {
		t.Fatalf("Failed to register enhanced Docker executor: %v", err)
	}
	
	// Test 4: Execute work with enhanced executor
	work := layer0.NewWork("test-work", "docker-enhanced", "Test Enhanced Docker Work")
	work = work.SetInput(map[string]interface{}{"test": "data"})
	
	// Set executor config
	config := work.GetConfiguration()
	config.Parameters["executor_config"] = map[string]interface{}{
		"image": "ubuntu:20.04",
		"command": []string{"echo"},
		"args": []string{"Hello World"},
	}
	work = work.SetConfiguration(config)
	
	enhancedExecutor, err := registry.GetExecutor("docker-enhanced")
	if err != nil {
		t.Fatalf("Failed to get enhanced executor: %v", err)
	}
	
	ctx := context.Background()
	workContext := layer0.NewContext()
	
	result, err := enhancedExecutor.Execute(ctx, work, workContext)
	if err != nil {
		t.Fatalf("Enhanced executor execution failed: %v", err)
	}
	
	if !result.Success {
		t.Fatalf("Expected successful execution, got: %s", result.Error)
	}
	
	// Test 5: Verify metrics and logs are present
	if len(result.Logs) == 0 {
		t.Error("Expected logs from overlay chain")
	}
	
	if result.Metrics.Duration == 0 {
		t.Error("Expected metrics from overlay chain")
	}
	
	// Test 6: Test schema and metadata retrieval
	schema, err := registry.GetSchema(docker.WorkTypeDocker)
	if err != nil {
		t.Fatalf("Failed to get schema: %v", err)
	}
	
	if schema.Documentation == "" {
		t.Error("Expected schema documentation")
	}
	
	metadata, err := registry.GetMetadata(docker.WorkTypeDocker)
	if err != nil {
		t.Fatalf("Failed to get metadata: %v", err)
	}
	
	if metadata.Name == "" {
		t.Error("Expected metadata name")
	}
	
	// Test 7: Test work validation
	err = registry.ValidateWork(work)
	if err != nil {
		t.Fatalf("Work validation failed: %v", err)
	}
	
	t.Log("Unified architecture test completed successfully")
}

// TestPluginIntegration tests plugin integration with the unified registry
func TestPluginIntegration(t *testing.T) {
	logger := &MockLogger{}
	registry := NewUnifiedRegistry(logger)
	
	// Create a mock plugin
	plugin := &MockPlugin{
		info: plugins.PluginInfo{
			Name:        "test-plugin",
			Version:     "1.0.0",
			Author:      "Test Author",
			Description: "Test plugin for unified architecture",
			WorkTypes:   []string{"test-work"},
		},
	}
	
	// Register plugin
	err := registry.RegisterPlugin("test-plugin", plugin)
	if err != nil {
		t.Fatalf("Failed to register plugin: %v", err)
	}
	
	// Test plugin execution
	work := layer0.NewWork("test-work", "test-work", "Test Plugin Work")
	
	executor, err := registry.GetExecutor("test-work")
	if err != nil {
		t.Fatalf("Failed to get plugin executor: %v", err)
	}
	
	ctx := context.Background()
	workContext := layer0.NewContext()
	
	result, err := executor.Execute(ctx, work, workContext)
	if err != nil {
		t.Fatalf("Plugin execution failed: %v", err)
	}
	
	if !result.Success {
		t.Fatalf("Expected successful plugin execution, got: %s", result.Error)
	}
	
	// Unregister plugin
	err = registry.UnregisterPlugin("test-plugin")
	if err != nil {
		t.Fatalf("Failed to unregister plugin: %v", err)
	}
	
	// Verify plugin is removed
	_, err = registry.GetExecutor("test-work")
	if err == nil {
		t.Error("Expected error when getting unregistered plugin executor")
	}
	
	t.Log("Plugin integration test completed successfully")
}

// MockPlugin for testing
type MockPlugin struct {
	info plugins.PluginInfo
}

func (p *MockPlugin) Execute(ctx context.Context, work layer0.Work, workContext *layer0.Context) (executors.WorkResult, error) {
	return executors.WorkResult{
		Success: true,
		Outputs: map[string]interface{}{
			"result": "plugin output",
		},
		Logs: []executors.LogEntry{
			{
				Timestamp: time.Now(),
				Level:     "INFO",
				Message:   "Plugin executed successfully",
				Source:    "mock-plugin",
			},
		},
		Metrics: executors.ExecutionMetrics{
			StartTime: time.Now(),
			EndTime:   time.Now(),
			Duration:  time.Millisecond * 100,
		},
	}, nil
}

func (p *MockPlugin) Validate(work layer0.Work) error {
	return nil
}

func (p *MockPlugin) CanExecute(workType layer0.WorkType) bool {
	return workType == "test-work"
}

func (p *MockPlugin) GetSupportedTypes() []layer0.WorkType {
	return []layer0.WorkType{"test-work"}
}

func (p *MockPlugin) GetSchema() executors.WorkSchema {
	return executors.WorkSchema{
		JSONSchema:    "{}",
		Examples:      []executors.WorkDefinition{},
		Documentation: "Mock plugin schema",
	}
}

func (p *MockPlugin) GetMetadata() executors.WorkMetadata {
	return executors.WorkMetadata{
		Name:        p.info.Name,
		Version:     p.info.Version,
		Author:      p.info.Author,
		Description: p.info.Description,
		WorkTypes:   p.info.WorkTypes,
	}
}

func (p *MockPlugin) Initialize(config map[string]interface{}) error {
	return nil
}

func (p *MockPlugin) Shutdown() error {
	return nil
}

func (p *MockPlugin) HealthCheck() error {
	return nil
}

func (p *MockPlugin) GetPluginInfo() plugins.PluginInfo {
	return p.info
}
