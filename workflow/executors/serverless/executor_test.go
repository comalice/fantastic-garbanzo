
package serverless

import (
        "context"
        "testing"

        "github.com/ubom/workflow/layer0"
        "github.com/ubom/workflow/executors"
)

// MockServerlessProvider implements ServerlessProvider for testing
type MockServerlessProvider struct {
        name string
}

func (m *MockServerlessProvider) InvokeFunction(ctx context.Context, config ServerlessWorkConfig, payload interface{}) (interface{}, []executors.LogEntry, error) {
        return "mock response", []executors.LogEntry{}, nil
}

func (m *MockServerlessProvider) ValidateConfig(config ServerlessWorkConfig) error {
        return nil
}

func (m *MockServerlessProvider) GetProviderName() string {
        return m.name
}

// MockLogger implements Logger for testing
type MockLogger struct{}

func (m *MockLogger) Info(msg string, fields ...interface{})  {}
func (m *MockLogger) Error(msg string, fields ...interface{}) {}
func (m *MockLogger) Debug(msg string, fields ...interface{}) {}
func (m *MockLogger) Warn(msg string, fields ...interface{})  {}

func TestServerlessExecutor_Execute(t *testing.T) {
        providers := map[string]ServerlessProvider{
                "aws": &MockServerlessProvider{name: "aws"},
        }
        logger := &MockLogger{}
        executor := NewServerlessExecutor(providers, logger)

        // Create test work
        work := layer0.NewWork("test-work", WorkTypeServerless, "Test Serverless Work")
        work = work.SetInput(map[string]interface{}{"test": "data"})
        
        // Set executor config
        config := work.GetConfiguration()
        config.Parameters["executor_config"] = map[string]interface{}{
                "provider": "aws",
                "function": "test-function",
                "region":   "us-east-1",
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

func TestServerlessExecutor_CanExecute(t *testing.T) {
        providers := map[string]ServerlessProvider{
                "aws": &MockServerlessProvider{name: "aws"},
        }
        logger := &MockLogger{}
        executor := NewServerlessExecutor(providers, logger)

        if !executor.CanExecute(WorkTypeServerless) {
                t.Error("Expected executor to handle serverless work type")
        }

        if executor.CanExecute("invalid") {
                t.Error("Expected executor to reject invalid work type")
        }
}
