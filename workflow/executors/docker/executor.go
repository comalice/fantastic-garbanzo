
package docker

import (
        "context"
        "encoding/json"
        "fmt"
        "time"

        "github.com/ubom/workflow/layer0"
        "github.com/ubom/workflow/executors"
)

const (
        WorkTypeDocker layer0.WorkType = "docker"
)

// DockerExecutor implements work execution using Docker containers
type DockerExecutor struct {
        *executors.BaseExecutor
        dockerUtils DockerUtilsInterface
        logger      Logger
}

// Logger interface for logging (allows mocking)
type Logger interface {
        Info(msg string, fields ...interface{})
        Error(msg string, fields ...interface{})
        Debug(msg string, fields ...interface{})
        Warn(msg string, fields ...interface{})
}

// NewDockerExecutor creates a new Docker executor
func NewDockerExecutor(logger Logger) *DockerExecutor {
        baseExecutor := executors.NewBaseExecutor(
                "Docker Executor",
                "2.0.0",
                "UBOM Workflow Engine",
                "Executes work using Docker containers with Dockerfile support",
                []layer0.WorkType{WorkTypeDocker},
        )
        
        return &DockerExecutor{
                BaseExecutor: baseExecutor,
                dockerUtils:  NewDockerUtils(logger),
                logger:       logger,
        }
}

// SetDockerUtils sets the docker utils (for testing)
func (e *DockerExecutor) SetDockerUtils(utils DockerUtilsInterface) {
        e.dockerUtils = utils
}

// Execute executes work using Docker containers
func (e *DockerExecutor) Execute(ctx context.Context, work layer0.Work, workContext *layer0.Context) (executors.WorkResult, error) {
        startTime := time.Now()
        
        // Parse Docker configuration
        config, err := e.parseConfig(work)
        if err != nil {
                return executors.WorkResult{
                        Success: false,
                        Error:   fmt.Sprintf("failed to parse Docker configuration: %v", err),
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
                        Error:   fmt.Sprintf("invalid Docker configuration: %v", err),
                        Metrics: executors.ExecutionMetrics{
                                StartTime: startTime,
                                EndTime:   time.Now(),
                                Duration:  time.Since(startTime),
                        },
                }, err
        }

        e.logger.Info("Starting Docker container execution", "image", config.Image, "workID", work.GetID())

        // Set timeout context
        if ctx == nil {
                ctx = context.Background()
        }
        if config.Timeout > 0 {
                var cancel context.CancelFunc
                ctx, cancel = context.WithTimeout(ctx, config.Timeout)
                defer cancel()
        }

        // Execute container using simplified Docker utilities
        output, logs, err := e.dockerUtils.RunContainer(ctx, config, work.GetInput())
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
                e.logger.Error("Docker container execution failed", "error", err, "workID", work.GetID())
                result.Success = false
                result.Error = err.Error()
                return result, err
        }

        e.logger.Info("Docker container execution completed", "workID", work.GetID(), "duration", duration)
        
        result.Success = true
        result.Outputs = map[string]interface{}{
                "result": output,
        }

        return result, nil
}

// Validate validates a work item for Docker execution
func (e *DockerExecutor) Validate(work layer0.Work) error {
        if work.GetType() != WorkTypeDocker {
                return fmt.Errorf("work type %s is not supported by Docker executor", work.GetType())
        }

        config, err := e.parseConfig(work)
        if err != nil {
                return fmt.Errorf("failed to parse Docker configuration: %w", err)
        }

        return config.Validate()
}

// GetSchema returns the JSON schema for Docker work
func (e *DockerExecutor) GetSchema() executors.WorkSchema {
        return executors.WorkSchema{
                JSONSchema:    DockerWorkSchema,
                Examples:      e.getExamples(),
                Documentation: "Docker executor runs work in containerized environments using Docker",
        }
}



// parseConfig parses Docker configuration from work parameters
func (e *DockerExecutor) parseConfig(work layer0.Work) (DockerWorkConfig, error) {
        config := DefaultDockerConfig()
        
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
                return config, fmt.Errorf("failed to unmarshal Docker config: %w", err)
        }

        return config, nil
}

// getExamples returns example work definitions
func (e *DockerExecutor) getExamples() []executors.WorkDefinition {
        return []executors.WorkDefinition{
                {
                        Name:        "Dockerfile-based Build",
                        Description: "Build and run a container from Dockerfile",
                        Configuration: map[string]interface{}{
                                "executor_config": map[string]interface{}{
                                        "dockerfile_path": "./Dockerfile",
                                        "build_context":   ".",
                                        "build_args": map[string]string{
                                                "VERSION": "1.0.0",
                                                "ENV":     "production",
                                        },
                                        "environment": map[string]string{
                                                "APP_ENV": "production",
                                        },
                                },
                        },
                        Input:          map[string]interface{}{"data": "input data"},
                        ExpectedOutput: map[string]interface{}{"result": "processed output"},
                },
                {
                        Name:        "Simple Python Script (Direct Image)",
                        Description: "Run a Python script using existing image",
                        Configuration: map[string]interface{}{
                                "executor_config": map[string]interface{}{
                                        "image":   "python:3.9-slim",
                                        "command": []string{"python", "-c"},
                                        "args":    []string{"print('Hello from Docker!')"},
                                },
                        },
                        Input:          nil,
                        ExpectedOutput: map[string]interface{}{"result": "Hello from Docker!\n"},
                },
                {
                        Name:        "Data Processing with Volumes",
                        Description: "Process data with volume mounts and resource limits",
                        Configuration: map[string]interface{}{
                                "executor_config": map[string]interface{}{
                                        "dockerfile_path": "./processors/Dockerfile",
                                        "build_context":   "./processors",
                                        "environment": map[string]string{
                                                "INPUT_FORMAT":  "json",
                                                "OUTPUT_FORMAT": "csv",
                                        },
                                        "volumes": []map[string]interface{}{
                                                {
                                                        "source":    "/host/data",
                                                        "target":    "/app/data",
                                                        "read_only": true,
                                                },
                                        },
                                        "resources": map[string]interface{}{
                                                "cpu_limit":    "2.0",
                                                "memory_limit": "1g",
                                        },
                                },
                        },
                        Input:          map[string]interface{}{"data": "sample data"},
                        ExpectedOutput: map[string]interface{}{"result": "processed data"},
                },
        }
}
