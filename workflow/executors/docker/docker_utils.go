package docker

import (
        "context"
        "fmt"
        "os/exec"
        "strings"
        "time"

        "github.com/ubom/workflow/executors"
)

// DockerUtilsInterface defines the interface for Docker operations
type DockerUtilsInterface interface {
        RunContainer(ctx context.Context, config DockerWorkConfig, input interface{}) (interface{}, []executors.LogEntry, error)
        BuildImage(ctx context.Context, config DockerWorkConfig) (string, error)
        PullImage(ctx context.Context, image string) error
        RemoveContainer(ctx context.Context, containerID string) error
        IsDockerAvailable() bool
        GetDockerVersion() (string, error)
}

// DockerUtils provides simplified Docker operations using standard Docker CLI
type DockerUtils struct {
        logger Logger
}

// NewDockerUtils creates a new Docker utilities instance
func NewDockerUtils(logger Logger) *DockerUtils {
        return &DockerUtils{
                logger: logger,
        }
}

// BuildImage builds a Docker image from Dockerfile
func (d *DockerUtils) BuildImage(ctx context.Context, config DockerWorkConfig) (string, error) {
        if config.DockerfilePath == "" {
                return "", fmt.Errorf("dockerfile_path is required for building")
        }

        // Generate unique image tag
        imageTag := fmt.Sprintf("ubom-workflow-%d", time.Now().Unix())
        
        // Prepare build command
        args := []string{"build", "-t", imageTag}
        
        // Add build args
        for key, value := range config.BuildArgs {
                args = append(args, "--build-arg", fmt.Sprintf("%s=%s", key, value))
        }
        
        // Add Dockerfile path
        args = append(args, "-f", config.DockerfilePath)
        
        // Add build context
        args = append(args, config.BuildContext)
        
        d.logger.Info("Building Docker image", "dockerfile", config.DockerfilePath, "context", config.BuildContext, "tag", imageTag)
        
        // Execute build command
        cmd := exec.CommandContext(ctx, "docker", args...)
        output, err := cmd.CombinedOutput()
        
        if err != nil {
                d.logger.Error("Docker build failed", "error", err, "output", string(output))
                return "", fmt.Errorf("docker build failed: %w\nOutput: %s", err, string(output))
        }
        
        d.logger.Info("Docker image built successfully", "tag", imageTag)
        return imageTag, nil
}

// RunContainer runs a Docker container with the given configuration
func (d *DockerUtils) RunContainer(ctx context.Context, config DockerWorkConfig, input interface{}) (interface{}, []executors.LogEntry, error) {
        var imageName string
        var err error
        
        // Build image if Dockerfile is provided, otherwise use existing image
        if config.DockerfilePath != "" {
                imageName, err = d.BuildImage(ctx, config)
                if err != nil {
                        return nil, nil, err
                }
                // Clean up built image after execution
                defer d.cleanupImage(imageName)
        } else {
                imageName = config.Image
                // Pull image if it doesn't exist locally
                if err := d.pullImageIfNeeded(ctx, imageName); err != nil {
                        return nil, nil, err
                }
        }
        
        // Prepare run command
        args := []string{"run", "--rm"}
        
        // Add environment variables
        for key, value := range config.Environment {
                args = append(args, "-e", fmt.Sprintf("%s=%s", key, value))
        }
        
        // Add volume mounts
        for _, volume := range config.Volumes {
                mountStr := fmt.Sprintf("%s:%s", volume.Source, volume.Target)
                if volume.ReadOnly {
                        mountStr += ":ro"
                }
                args = append(args, "-v", mountStr)
        }
        
        // Add resource limits
        if config.Resources.CPULimit != "" {
                args = append(args, "--cpus", config.Resources.CPULimit)
        }
        if config.Resources.MemoryLimit != "" {
                args = append(args, "--memory", config.Resources.MemoryLimit)
        }
        
        // Add network configuration
        if config.Network.Mode != "" {
                args = append(args, "--network", config.Network.Mode)
        }
        
        // Add port mappings
        for _, port := range config.Network.Ports {
                portStr := fmt.Sprintf("%d:%d", port.HostPort, port.ContainerPort)
                if port.Protocol != "" {
                        portStr += "/" + port.Protocol
                }
                args = append(args, "-p", portStr)
        }
        
        // Add working directory
        if config.WorkingDir != "" {
                args = append(args, "-w", config.WorkingDir)
        }
        
        // Add user
        if config.User != "" {
                args = append(args, "-u", config.User)
        }
        
        // Add privileged mode
        if config.Privileged {
                args = append(args, "--privileged")
        }
        
        // Add image name
        args = append(args, imageName)
        
        // Add command and args
        if len(config.Command) > 0 {
                args = append(args, config.Command...)
        }
        if len(config.Args) > 0 {
                args = append(args, config.Args...)
        }
        
        d.logger.Info("Running Docker container", "image", imageName, "command", strings.Join(args, " "))
        
        // Execute run command
        cmd := exec.CommandContext(ctx, "docker", args...)
        
        // Set up input if provided
        if input != nil {
                inputStr := fmt.Sprintf("%v", input)
                cmd.Stdin = strings.NewReader(inputStr)
        }
        
        startTime := time.Now()
        output, err := cmd.CombinedOutput()
        duration := time.Since(startTime)
        
        // Create log entries
        logs := []executors.LogEntry{
                {
                        Timestamp: startTime,
                        Level:     "INFO",
                        Message:   fmt.Sprintf("Container execution started with command: %s", strings.Join(args, " ")),
                        Source:    "docker-executor",
                },
        }
        
        if err != nil {
                logs = append(logs, executors.LogEntry{
                        Timestamp: time.Now(),
                        Level:     "ERROR",
                        Message:   fmt.Sprintf("Container execution failed: %v", err),
                        Source:    "docker-executor",
                })
                d.logger.Error("Docker container execution failed", "error", err, "output", string(output), "duration", duration)
                return nil, logs, fmt.Errorf("docker run failed: %w\nOutput: %s", err, string(output))
        }
        
        logs = append(logs, executors.LogEntry{
                Timestamp: time.Now(),
                Level:     "INFO",
                Message:   fmt.Sprintf("Container execution completed successfully in %v", duration),
                Source:    "docker-executor",
        })
        
        d.logger.Info("Docker container execution completed", "duration", duration, "outputSize", len(output))
        
        // Return output as string
        return string(output), logs, nil
}

// pullImageIfNeeded pulls an image if it doesn't exist locally
func (d *DockerUtils) pullImageIfNeeded(ctx context.Context, image string) error {
        // Check if image exists locally
        cmd := exec.CommandContext(ctx, "docker", "image", "inspect", image)
        if err := cmd.Run(); err == nil {
                // Image exists locally
                return nil
        }
        
        d.logger.Info("Pulling Docker image", "image", image)
        
        // Pull the image
        cmd = exec.CommandContext(ctx, "docker", "pull", image)
        output, err := cmd.CombinedOutput()
        
        if err != nil {
                d.logger.Error("Docker pull failed", "image", image, "error", err, "output", string(output))
                return fmt.Errorf("docker pull failed for image %s: %w\nOutput: %s", image, err, string(output))
        }
        
        d.logger.Info("Docker image pulled successfully", "image", image)
        return nil
}

// cleanupImage removes a built image
func (d *DockerUtils) cleanupImage(imageTag string) {
        cmd := exec.Command("docker", "rmi", imageTag)
        if err := cmd.Run(); err != nil {
                d.logger.Warn("Failed to cleanup Docker image", "tag", imageTag, "error", err)
        } else {
                d.logger.Debug("Docker image cleaned up", "tag", imageTag)
        }
}

// PullImage pulls a Docker image
func (d *DockerUtils) PullImage(ctx context.Context, image string) error {
        return d.pullImageIfNeeded(ctx, image)
}

// RemoveContainer removes a container (no-op since we use --rm)
func (d *DockerUtils) RemoveContainer(ctx context.Context, containerID string) error {
        // No-op since we use --rm flag
        return nil
}

// IsDockerAvailable checks if Docker is available
func (d *DockerUtils) IsDockerAvailable() bool {
        cmd := exec.Command("docker", "version")
        return cmd.Run() == nil
}

// GetDockerVersion returns the Docker version
func (d *DockerUtils) GetDockerVersion() (string, error) {
        cmd := exec.Command("docker", "version", "--format", "{{.Server.Version}}")
        output, err := cmd.Output()
        if err != nil {
                return "", err
        }
        return strings.TrimSpace(string(output)), nil
}
