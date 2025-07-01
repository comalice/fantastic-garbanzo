
package grpc

import (
        "context"
        "testing"

        "github.com/ubom/workflow/layer0"
)

// MockConnectionPool implements ConnectionPool for testing
type MockConnectionPool struct{}

func (m *MockConnectionPool) GetConnection(endpoint string, tlsConfig TLSConfig) (GRPCConnection, error) {
        return &MockGRPCConnection{}, nil
}

func (m *MockConnectionPool) ReleaseConnection(conn GRPCConnection) error {
        return nil
}

func (m *MockConnectionPool) Close() error {
        return nil
}

// MockGRPCConnection implements GRPCConnection for testing
type MockGRPCConnection struct{}

func (m *MockGRPCConnection) Invoke(ctx context.Context, method string, request interface{}, response interface{}, metadata map[string]string) error {
        return nil
}

func (m *MockGRPCConnection) Close() error {
        return nil
}

// MockLogger implements Logger for testing
type MockLogger struct{}

func (m *MockLogger) Info(msg string, fields ...interface{})  {}
func (m *MockLogger) Error(msg string, fields ...interface{}) {}
func (m *MockLogger) Debug(msg string, fields ...interface{}) {}
func (m *MockLogger) Warn(msg string, fields ...interface{})  {}

func TestGRPCExecutor_Execute(t *testing.T) {
        pool := &MockConnectionPool{}
        logger := &MockLogger{}
        executor := NewGRPCExecutor(pool, logger)

        // Create test work
        work := layer0.NewWork("test-work", WorkTypeGRPC, "Test gRPC Work")
        work = work.SetInput(map[string]interface{}{"test": "data"})
        
        // Set executor config
        config := work.GetConfiguration()
        config.Parameters["executor_config"] = map[string]interface{}{
                "endpoint": "localhost:9090",
                "method":   "test.Service/TestMethod",
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

func TestGRPCExecutor_CanExecute(t *testing.T) {
        pool := &MockConnectionPool{}
        logger := &MockLogger{}
        executor := NewGRPCExecutor(pool, logger)

        if !executor.CanExecute(WorkTypeGRPC) {
                t.Error("Expected executor to handle gRPC work type")
        }

        if executor.CanExecute("invalid") {
                t.Error("Expected executor to reject invalid work type")
        }
}
