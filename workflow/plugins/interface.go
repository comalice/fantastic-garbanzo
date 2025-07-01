
package plugins

import (
        "fmt"

        "github.com/ubom/workflow/executors"
)

// ExternalWorkPlugin defines the interface for external work plugins
type ExternalWorkPlugin interface {
        executors.WorkExecutor
        
        // Plugin lifecycle methods
        Initialize(config map[string]interface{}) error
        Shutdown() error
        HealthCheck() error
        
        // Plugin information
        GetPluginInfo() PluginInfo
}

// PluginInfo contains information about a plugin
type PluginInfo struct {
        Name        string   `json:"name"`
        Version     string   `json:"version"`
        Author      string   `json:"author"`
        Description string   `json:"description"`
        WorkTypes   []string `json:"work_types"`
        APIVersion  string   `json:"api_version"`
        Homepage    string   `json:"homepage,omitempty"`
        License     string   `json:"license,omitempty"`
}

// PluginFactory is a function that creates a new plugin instance
type PluginFactory func() ExternalWorkPlugin

// PluginRegistry manages plugin registration and discovery
type PluginRegistry interface {
        RegisterPlugin(name string, factory PluginFactory) error
        UnregisterPlugin(name string) error
        GetPlugin(name string) (ExternalWorkPlugin, error)
        ListPlugins() []PluginInfo
        LoadPluginFromFile(path string) error
        LoadPluginsFromDirectory(dir string) error
}

// PluginConfig represents configuration for a plugin
type PluginConfig struct {
        Name     string                 `json:"name"`
        Enabled  bool                   `json:"enabled"`
        Config   map[string]interface{} `json:"config"`
        Priority int                    `json:"priority"`
}

// PluginStatus represents the status of a plugin
type PluginStatus struct {
        Name        string `json:"name"`
        Status      string `json:"status"` // loaded, initialized, error, disabled
        Error       string `json:"error,omitempty"`
        LastChecked string `json:"last_checked"`
        Version     string `json:"version"`
}

// PluginManager manages the lifecycle of plugins
type PluginManager interface {
        LoadPlugin(config PluginConfig) error
        UnloadPlugin(name string) error
        InitializePlugin(name string, config map[string]interface{}) error
        ShutdownPlugin(name string) error
        GetPluginStatus(name string) (PluginStatus, error)
        ListPluginStatuses() []PluginStatus
        HealthCheckAll() map[string]error
}

// BasePlugin provides a base implementation for plugins
type BasePlugin struct {
        info     PluginInfo
        metadata executors.WorkMetadata
        schema   executors.WorkSchema
}

// NewBasePlugin creates a new base plugin
func NewBasePlugin(info PluginInfo, metadata executors.WorkMetadata, schema executors.WorkSchema) *BasePlugin {
        return &BasePlugin{
                info:     info,
                metadata: metadata,
                schema:   schema,
        }
}

// GetPluginInfo returns plugin information
func (p *BasePlugin) GetPluginInfo() PluginInfo {
        return p.info
}

// GetMetadata returns work metadata
func (p *BasePlugin) GetMetadata() executors.WorkMetadata {
        return p.metadata
}

// GetSchema returns work schema
func (p *BasePlugin) GetSchema() executors.WorkSchema {
        return p.schema
}

// Initialize provides a default implementation (can be overridden)
func (p *BasePlugin) Initialize(config map[string]interface{}) error {
        return nil
}

// Shutdown provides a default implementation (can be overridden)
func (p *BasePlugin) Shutdown() error {
        return nil
}

// HealthCheck provides a default implementation (can be overridden)
func (p *BasePlugin) HealthCheck() error {
        return nil
}

// PluginError represents an error that occurred in a plugin
type PluginError struct {
        PluginName string
        Operation  string
        Err        error
}

// Error implements the error interface
func (e *PluginError) Error() string {
        return fmt.Sprintf("plugin %s error in %s: %v", e.PluginName, e.Operation, e.Err)
}

// Unwrap returns the underlying error
func (e *PluginError) Unwrap() error {
        return e.Err
}

// NewPluginError creates a new plugin error
func NewPluginError(pluginName, operation string, err error) *PluginError {
        return &PluginError{
                PluginName: pluginName,
                Operation:  operation,
                Err:        err,
        }
}
