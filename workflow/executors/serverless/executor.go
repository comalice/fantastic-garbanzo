
package serverless

import (
        "context"
        "encoding/json"
        "fmt"
        "time"

        "github.com/ubom/workflow/layer0"
        "github.com/ubom/workflow/executors"
)

const (
        WorkTypeServerless layer0.WorkType = "serverless"
)

// ServerlessExecutor implements work execution using serverless functions
type ServerlessExecutor struct {
        *executors.BaseExecutor
        providers map[string]ServerlessProvider
        logger    Logger
}

// ServerlessProvider interface for different cloud providers (allows mocking)
type ServerlessProvider interface {
        InvokeFunction(ctx context.Context, config ServerlessWorkConfig, payload interface{}) (interface{}, []executors.LogEntry, error)
        ValidateConfig(config ServerlessWorkConfig) error
        GetProviderName() string
}

// Logger interface for logging (allows mocking)
type Logger interface {
        Info(msg string, fields ...interface{})
        Error(msg string, fields ...interface{})
        Debug(msg string, fields ...interface{})
        Warn(msg string, fields ...interface{})
}

// NewServerlessExecutor creates a new serverless executor
func NewServerlessExecutor(providers map[string]ServerlessProvider, logger Logger) *ServerlessExecutor {
        baseExecutor := executors.NewBaseExecutor(
                "Serverless Executor",
                "2.0.0",
                "UBOM Workflow Engine",
                "Executes work using serverless functions across cloud providers",
                []layer0.WorkType{WorkTypeServerless},
        )
        
        return &ServerlessExecutor{
                BaseExecutor: baseExecutor,
                providers:    providers,
                logger:       logger,
        }
}

// Execute executes work using serverless functions
func (e *ServerlessExecutor) Execute(ctx context.Context, work layer0.Work, workContext *layer0.Context) (executors.WorkResult, error) {
        startTime := time.Now()
        
        // Parse serverless configuration
        config, err := e.parseConfig(work)
        if err != nil {
                return executors.WorkResult{
                        Success: false,
                        Error:   fmt.Sprintf("failed to parse serverless configuration: %v", err),
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
                        Error:   fmt.Sprintf("invalid serverless configuration: %v", err),
                        Metrics: executors.ExecutionMetrics{
                                StartTime: startTime,
                                EndTime:   time.Now(),
                                Duration:  time.Since(startTime),
                        },
                }, err
        }

        // Get provider
        provider, exists := e.providers[config.Provider]
        if !exists {
                return executors.WorkResult{
                        Success: false,
                        Error:   fmt.Sprintf("unsupported provider: %s", config.Provider),
                        Metrics: executors.ExecutionMetrics{
                                StartTime: startTime,
                                EndTime:   time.Now(),
                                Duration:  time.Since(startTime),
                        },
                }, fmt.Errorf("unsupported provider: %s", config.Provider)
        }

        e.logger.Info("Starting serverless function invocation", "provider", config.Provider, "function", config.Function, "workID", work.GetID())

        // Set timeout context
        if ctx == nil {
                ctx = context.Background()
        }
        if config.Timeout > 0 {
                var cancel context.CancelFunc
                ctx, cancel = context.WithTimeout(ctx, config.Timeout)
                defer cancel()
        }

        // Invoke function
        response, logs, err := provider.InvokeFunction(ctx, config, work.GetInput())
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
                e.logger.Error("Serverless function invocation failed", "error", err, "workID", work.GetID())
                result.Success = false
                result.Error = err.Error()
                return result, err
        }

        e.logger.Info("Serverless function invocation completed", "workID", work.GetID(), "duration", duration)
        
        result.Success = true
        result.Outputs = map[string]interface{}{
                "response": response,
        }

        return result, nil
}

// Validate validates a work item for serverless execution
func (e *ServerlessExecutor) Validate(work layer0.Work) error {
        if work.GetType() != WorkTypeServerless {
                return fmt.Errorf("work type %s is not supported by serverless executor", work.GetType())
        }

        config, err := e.parseConfig(work)
        if err != nil {
                return fmt.Errorf("failed to parse serverless configuration: %w", err)
        }

        if err := config.Validate(); err != nil {
                return err
        }

        // Validate with provider
        provider, exists := e.providers[config.Provider]
        if !exists {
                return fmt.Errorf("unsupported provider: %s", config.Provider)
        }

        return provider.ValidateConfig(config)
}

// GetSchema returns the JSON schema for serverless work
func (e *ServerlessExecutor) GetSchema() executors.WorkSchema {
        return executors.WorkSchema{
                JSONSchema:    ServerlessWorkSchema,
                Examples:      e.getExamples(),
                Documentation: "Serverless executor invokes cloud functions on AWS Lambda, Google Cloud Functions, or Azure Functions",
        }
}



// parseConfig parses serverless configuration from work parameters
func (e *ServerlessExecutor) parseConfig(work layer0.Work) (ServerlessWorkConfig, error) {
        config := DefaultServerlessConfig()
        
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
                return config, fmt.Errorf("failed to unmarshal serverless config: %w", err)
        }

        return config, nil
}

// getExamples returns example work definitions
func (e *ServerlessExecutor) getExamples() []executors.WorkDefinition {
        return []executors.WorkDefinition{
                {
                        Name:        "AWS Lambda Function",
                        Description: "Invoke an AWS Lambda function",
                        Configuration: map[string]interface{}{
                                "executor_config": map[string]interface{}{
                                        "provider": "aws",
                                        "function": "my-lambda-function",
                                        "region":   "us-east-1",
                                        "timeout":  "30s",
                                        "credentials": map[string]interface{}{
                                                "type":       "aws_iam",
                                                "access_key": "AKIAIOSFODNN7EXAMPLE",
                                                "secret_key": "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
                                        },
                                },
                        },
                        Input:          map[string]interface{}{"message": "Hello Lambda!"},
                        ExpectedOutput: map[string]interface{}{"response": "Function executed successfully"},
                },
                {
                        Name:        "Google Cloud Function",
                        Description: "Invoke a Google Cloud Function",
                        Configuration: map[string]interface{}{
                                "executor_config": map[string]interface{}{
                                        "provider":   "gcp",
                                        "function":   "my-cloud-function",
                                        "region":     "us-central1",
                                        "project_id": "my-project-123",
                                        "credentials": map[string]interface{}{
                                                "type":     "service_account",
                                                "key_file": "/path/to/service-account.json",
                                        },
                                },
                        },
                        Input:          map[string]interface{}{"data": "process this"},
                        ExpectedOutput: map[string]interface{}{"response": "Data processed"},
                },
                {
                        Name:        "Azure Function",
                        Description: "Invoke an Azure Function",
                        Configuration: map[string]interface{}{
                                "executor_config": map[string]interface{}{
                                        "provider": "azure",
                                        "function": "my-azure-function",
                                        "credentials": map[string]interface{}{
                                                "type": "managed_identity",
                                        },
                                },
                        },
                        Input:          map[string]interface{}{"input": "test data"},
                        ExpectedOutput: map[string]interface{}{"response": "Azure function result"},
                },
        }
}
