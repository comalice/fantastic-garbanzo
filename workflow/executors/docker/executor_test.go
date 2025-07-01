
package docker

import (
        "context"
        "testing"

        "github.com/ubom/workflow/layer0"
        "github.com/ubom/workflow/executors"
)

// MockDockerUtils implements DockerUtilsInterface for testing
type MockDockerUtils struct {
        runContainerFunc func(ctx context.Context, config DockerWorkConfig, input interface{}) (interface{}, []executors.LogEntry, error)
}

func (m *MockDockerUtils) RunContainer(ctx context.Context, config DockerWorkConfig, input interface{}) (interface{}, []executors.LogEntry, error) {
        if m.runContainerFunc != nil {
                return m.runContainerFunc(ctx, config, input)
        }
        return "mock output", []executors.LogEntry{}, nil
}

func (m *MockDockerUtils) BuildImage(ctx context.Context, config DockerWorkConfig) (string, error) {
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

// MockLogger implements Logger for testing
type MockLogger struct{}

func (m *MockLogger) Info(msg string, fields ...interface{})  {}
func (m *MockLogger) Error(msg string, fields ...interface{}) {}
func (m *MockLogger) Debug(msg string, fields ...interface{}) {}
func (m *MockLogger) Warn(msg string, fields ...interface{})  {}

func TestDockerExecutor_Execute(t *testing.T) {
        logger := &MockLogger{}
        executor := NewDockerExecutor(logger)
        
        // Mock the docker utils for testing
        executor.dockerUtils = &MockDockerUtils{}

        // Create test work
        work := layer0.NewWork("test-work", WorkTypeDocker, "Test Docker Work")
        work = work.SetInput(map[string]interface{}{"test": "data"})
        
        // Set executor config
        config := work.GetConfiguration()
        config.Parameters["executor_config"] = map[string]interface{}{
                "image": "ubuntu:20.04",
                "command": []string{"echo", "hello"},
        }
        work = layer0.Work{
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

        ctx := context.Background()
        workContext := &layer0.Context{}

        result, err := executor.Execute(ctx, work, workContext)
        if err != nil {
                t.Fatalf("Execute failed: %v", err)
        }

        if !result.Success {
                t.Errorf("Expected success, got failure: %s", result.Error)
        }
}

func TestDockerExecutor_Validate(t *testing.T) {
        logger := &MockLogger{}
        executor := NewDockerExecutor(logger)
        executor.dockerUtils = &MockDockerUtils{}

        // Test valid work
        work := layer0.NewWork("test-work", WorkTypeDocker, "Test Docker Work")
        config := work.GetConfiguration()
        config.Parameters["executor_config"] = map[string]interface{}{
                "image": "ubuntu:20.04",
        }
        work = layer0.Work{
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

        err := executor.Validate(work)
        if err != nil {
                t.Errorf("Validate failed for valid work: %v", err)
        }

        // Test invalid work type
        invalidWork := layer0.NewWork("test-work", "invalid", "Test Invalid Work")
        err = executor.Validate(invalidWork)
        if err == nil {
                t.Error("Expected validation error for invalid work type")
        }
}

func TestDockerExecutor_CanExecute(t *testing.T) {
        logger := &MockLogger{}
        executor := NewDockerExecutor(logger)
        executor.dockerUtils = &MockDockerUtils{}

        if !executor.CanExecute(WorkTypeDocker) {
                t.Error("Expected executor to handle Docker work type")
        }

        if executor.CanExecute("invalid") {
                t.Error("Expected executor to reject invalid work type")
        }
}
