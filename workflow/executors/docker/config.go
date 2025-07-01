
package docker

import (
        "fmt"
        "time"
)

// DockerWorkConfig represents simplified Docker-specific work configuration
type DockerWorkConfig struct {
        // Dockerfile-based configuration (preferred)
        DockerfilePath string            `json:"dockerfile_path,omitempty"`
        BuildContext   string            `json:"build_context,omitempty"`
        BuildArgs      map[string]string `json:"build_args,omitempty"`
        
        // Direct image configuration (fallback)
        Image       string            `json:"image,omitempty"`
        Command     []string          `json:"command,omitempty"`
        Args        []string          `json:"args,omitempty"`
        
        // Runtime configuration
        Environment map[string]string `json:"environment,omitempty"`
        Volumes     []VolumeMount     `json:"volumes,omitempty"`
        Resources   ResourceLimits    `json:"resources,omitempty"`
        Network     NetworkConfig     `json:"network,omitempty"`
        WorkingDir  string            `json:"working_dir,omitempty"`
        User        string            `json:"user,omitempty"`
        Privileged  bool              `json:"privileged,omitempty"`
        Timeout     time.Duration     `json:"timeout,omitempty"`
}

// VolumeMount represents a volume mount configuration
type VolumeMount struct {
        Source      string `json:"source"`
        Target      string `json:"target"`
        ReadOnly    bool   `json:"read_only,omitempty"`
        Type        string `json:"type,omitempty"` // bind, volume, tmpfs
}

// ResourceLimits defines resource constraints for the container
type ResourceLimits struct {
        CPULimit    string `json:"cpu_limit,omitempty"`    // e.g., "0.5" for 0.5 CPU
        MemoryLimit string `json:"memory_limit,omitempty"` // e.g., "512m" for 512MB
        CPURequest  string `json:"cpu_request,omitempty"`
        MemoryRequest string `json:"memory_request,omitempty"`
}

// NetworkConfig defines network configuration for the container
type NetworkConfig struct {
        Mode       string            `json:"mode,omitempty"`       // bridge, host, none, container
        Ports      []PortMapping     `json:"ports,omitempty"`
        DNS        []string          `json:"dns,omitempty"`
        ExtraHosts []string          `json:"extra_hosts,omitempty"`
        Labels     map[string]string `json:"labels,omitempty"`
}

// PortMapping represents a port mapping configuration
type PortMapping struct {
        HostPort      int    `json:"host_port"`
        ContainerPort int    `json:"container_port"`
        Protocol      string `json:"protocol,omitempty"` // tcp, udp
}

// DefaultDockerConfig returns a default Docker configuration
func DefaultDockerConfig() DockerWorkConfig {
        return DockerWorkConfig{
                BuildContext: ".",
                BuildArgs:    make(map[string]string),
                Environment:  make(map[string]string),
                Volumes:      []VolumeMount{},
                Resources: ResourceLimits{
                        CPULimit:    "1.0",
                        MemoryLimit: "512m",
                },
                Network: NetworkConfig{
                        Mode:  "bridge",
                        Ports: []PortMapping{},
                },
                Timeout: 5 * time.Minute,
        }
}

// Validate validates the Docker configuration
func (c *DockerWorkConfig) Validate() error {
        // Either Dockerfile path or image must be specified
        if c.DockerfilePath == "" && c.Image == "" {
                return fmt.Errorf("either dockerfile_path or image must be specified")
        }
        
        // If using Dockerfile, build context is required
        if c.DockerfilePath != "" && c.BuildContext == "" {
                return fmt.Errorf("build_context is required when using dockerfile_path")
        }
        
        // Validate volume mounts
        for _, volume := range c.Volumes {
                if volume.Source == "" || volume.Target == "" {
                        return fmt.Errorf("volume mount source and target are required")
                }
        }
        
        // Validate port mappings
        for _, port := range c.Network.Ports {
                if port.HostPort <= 0 || port.ContainerPort <= 0 {
                        return fmt.Errorf("port mappings must have positive port numbers")
                }
        }
        
        return nil
}
