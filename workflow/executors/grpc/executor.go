
package grpc

import (
        "context"
        "encoding/json"
        "fmt"
        "time"

        "github.com/ubom/workflow/layer0"
        "github.com/ubom/workflow/executors"
)

const (
        WorkTypeGRPC layer0.WorkType = "grpc"
)

// GRPCExecutor implements work execution using gRPC calls
type GRPCExecutor struct {
        *executors.BaseExecutor
        connectionPool ConnectionPool
        logger         Logger
}

// ConnectionPool interface for managing gRPC connections (allows mocking)
type ConnectionPool interface {
        GetConnection(endpoint string, tlsConfig TLSConfig) (GRPCConnection, error)
        ReleaseConnection(conn GRPCConnection) error
        Close() error
}

// GRPCConnection interface for gRPC connections (allows mocking)
type GRPCConnection interface {
        Invoke(ctx context.Context, method string, request interface{}, response interface{}, metadata map[string]string) error
        Close() error
}

// Logger interface for logging (allows mocking)
type Logger interface {
        Info(msg string, fields ...interface{})
        Error(msg string, fields ...interface{})
        Debug(msg string, fields ...interface{})
        Warn(msg string, fields ...interface{})
}

// NewGRPCExecutor creates a new gRPC executor
func NewGRPCExecutor(connectionPool ConnectionPool, logger Logger) *GRPCExecutor {
        baseExecutor := executors.NewBaseExecutor(
                "gRPC Executor",
                "2.0.0",
                "UBOM Workflow Engine",
                "Executes work using gRPC service calls",
                []layer0.WorkType{WorkTypeGRPC},
        )
        
        return &GRPCExecutor{
                BaseExecutor:   baseExecutor,
                connectionPool: connectionPool,
                logger:         logger,
        }
}

// Execute executes work using gRPC calls
func (e *GRPCExecutor) Execute(ctx context.Context, work layer0.Work, workContext *layer0.Context) (executors.WorkResult, error) {
        startTime := time.Now()
        
        // Parse gRPC configuration
        config, err := e.parseConfig(work)
        if err != nil {
                return executors.WorkResult{
                        Success: false,
                        Error:   fmt.Sprintf("failed to parse gRPC configuration: %v", err),
                        Metrics: executors.ExecutionMetrics{
                                StartTime: startTime,
                                EndTime:   time.Now(),
                                Duration:  time.Since(startTime),
                        },
                }, err
        }

        // Validate configuration
        if err := config.Validate(); err != nil {
                return executors.WorkResult{
                        Success: false,
                        Error:   fmt.Sprintf("invalid gRPC configuration: %v", err),
                        Metrics: executors.ExecutionMetrics{
                                StartTime: startTime,
                                EndTime:   time.Now(),
                                Duration:  time.Since(startTime),
                        },
                }, err
        }

        e.logger.Info("Starting gRPC call", "endpoint", config.Endpoint, "method", config.Method, "workID", work.GetID())

        // Set timeout context
        if ctx == nil {
                ctx = context.Background()
        }
        if config.Timeout > 0 {
                var cancel context.CancelFunc
                ctx, cancel = context.WithTimeout(ctx, config.Timeout)
                defer cancel()
        }

        // Get connection
        conn, err := e.connectionPool.GetConnection(config.Endpoint, config.TLS)
        if err != nil {
                return executors.WorkResult{
                        Success: false,
                        Error:   fmt.Sprintf("failed to get gRPC connection: %v", err),
                        Metrics: executors.ExecutionMetrics{
                                StartTime: startTime,
                                EndTime:   time.Now(),
                                Duration:  time.Since(startTime),
                        },
                }, err
        }
        defer e.connectionPool.ReleaseConnection(conn)

        // Execute with retry
        var response interface{}
        var logs []executors.LogEntry
        
        err = e.executeWithRetry(ctx, conn, config, work.GetInput(), &response, &logs)
        endTime := time.Now()
        duration := endTime.Sub(startTime)

        result := executors.WorkResult{
                Logs: logs,
                Metrics: executors.ExecutionMetrics{
                        StartTime: startTime,
                        EndTime:   endTime,
                        Duration:  duration,
                },
        }

        if err != nil {
                e.logger.Error("gRPC call failed", "error", err, "workID", work.GetID())
                result.Success = false
                result.Error = err.Error()
                return result, err
        }

        e.logger.Info("gRPC call completed", "workID", work.GetID(), "duration", duration)
        
        result.Success = true
        result.Outputs = map[string]interface{}{
                "response": response,
        }

        return result, nil
}

// Validate validates a work item for gRPC execution
func (e *GRPCExecutor) Validate(work layer0.Work) error {
        if work.GetType() != WorkTypeGRPC {
                return fmt.Errorf("work type %s is not supported by gRPC executor", work.GetType())
        }

        config, err := e.parseConfig(work)
        if err != nil {
                return fmt.Errorf("failed to parse gRPC configuration: %w", err)
        }

        return config.Validate()
}

// GetSchema returns the JSON schema for gRPC work
func (e *GRPCExecutor) GetSchema() executors.WorkSchema {
        return executors.WorkSchema{
                JSONSchema:    GRPCWorkSchema,
                Examples:      e.getExamples(),
                Documentation: "gRPC executor makes remote procedure calls to gRPC services",
        }
}



// parseConfig parses gRPC configuration from work parameters
func (e *GRPCExecutor) parseConfig(work layer0.Work) (GRPCWorkConfig, error) {
        config := DefaultGRPCConfig()
        
        // Get executor config from work configuration
        executorConfig, exists := work.GetConfiguration().Parameters["executor_config"]
        if !exists {
                return config, fmt.Errorf("executor_config not found in work parameters")
        }

        // Convert to JSON and back to parse into struct
        configBytes, err := json.Marshal(executorConfig)
        if err != nil {
                return config, fmt.Errorf("failed to marshal executor config: %w", err)
        }

        if err := json.Unmarshal(configBytes, &config); err != nil {
                return config, fmt.Errorf("failed to unmarshal gRPC config: %w", err)
        }

        return config, nil
}

// executeWithRetry executes the gRPC call with retry logic
func (e *GRPCExecutor) executeWithRetry(ctx context.Context, conn GRPCConnection, config GRPCWorkConfig, request interface{}, response interface{}, logs *[]executors.LogEntry) error {
        var lastErr error
        
        for attempt := 1; attempt <= config.Retry.MaxAttempts; attempt++ {
                *logs = append(*logs, executors.LogEntry{
                        Timestamp: time.Now(),
                        Level:     "INFO",
                        Message:   fmt.Sprintf("Attempting gRPC call (attempt %d/%d)", attempt, config.Retry.MaxAttempts),
                        Source:    "grpc-executor",
                })

                err := conn.Invoke(ctx, config.Method, request, response, config.Metadata)
                if err == nil {
                        return nil
                }

                lastErr = err
                *logs = append(*logs, executors.LogEntry{
                        Timestamp: time.Now(),
                        Level:     "ERROR",
                        Message:   fmt.Sprintf("gRPC call failed (attempt %d/%d): %v", attempt, config.Retry.MaxAttempts, err),
                        Source:    "grpc-executor",
                })

                // Don't sleep after the last attempt
                if attempt < config.Retry.MaxAttempts {
                        delay := e.calculateRetryDelay(attempt, config.Retry)
                        *logs = append(*logs, executors.LogEntry{
                                Timestamp: time.Now(),
                                Level:     "INFO",
                                Message:   fmt.Sprintf("Retrying in %v", delay),
                                Source:    "grpc-executor",
                        })
                        
                        select {
                        case <-ctx.Done():
                                return ctx.Err()
                        case <-time.After(delay):
                                // Continue to next attempt
                        }
                }
        }

        return fmt.Errorf("gRPC call failed after %d attempts: %w", config.Retry.MaxAttempts, lastErr)
}

// calculateRetryDelay calculates the delay for the next retry attempt
func (e *GRPCExecutor) calculateRetryDelay(attempt int, retry RetryConfig) time.Duration {
        delay := retry.InitialDelay
        for i := 1; i < attempt; i++ {
                delay = time.Duration(float64(delay) * retry.Multiplier)
                if delay > retry.MaxDelay {
                        delay = retry.MaxDelay
                        break
                }
        }
        
        // Add jitter if enabled
        if retry.Jitter {
                jitter := time.Duration(float64(delay) * 0.1) // 10% jitter
                jitterFactor := float64(time.Now().UnixNano()%1000)/1000.0*2 - 1
                delay += time.Duration(float64(jitter) * jitterFactor)
        }
        
        return delay
}

// getExamples returns example work definitions
func (e *GRPCExecutor) getExamples() []executors.WorkDefinition {
        return []executors.WorkDefinition{
                {
                        Name:        "Simple gRPC Call",
                        Description: "Make a simple gRPC call to a service",
                        Configuration: map[string]interface{}{
                                "executor_config": map[string]interface{}{
                                        "endpoint": "localhost:9090",
                                        "method":   "myservice.MyService/GetUser",
                                        "timeout":  "30s",
                                },
                        },
                        Input:          map[string]interface{}{"user_id": "123"},
                        ExpectedOutput: map[string]interface{}{"response": map[string]interface{}{"user": "data"}},
                },
                {
                        Name:        "Secure gRPC Call",
                        Description: "Make a gRPC call with TLS and authentication",
                        Configuration: map[string]interface{}{
                                "executor_config": map[string]interface{}{
                                        "endpoint": "api.example.com:443",
                                        "method":   "api.v1.UserService/CreateUser",
                                        "tls": map[string]interface{}{
                                                "enabled": true,
                                        },
                                        "metadata": map[string]interface{}{
                                                "authorization": "Bearer token123",
                                        },
                                        "retry": map[string]interface{}{
                                                "max_attempts": 5,
                                                "initial_delay": "2s",
                                        },
                                },
                        },
                        Input:          map[string]interface{}{"name": "John Doe", "email": "john@example.com"},
                        ExpectedOutput: map[string]interface{}{"response": map[string]interface{}{"user_id": "456"}},
                },
        }
}
