
package plugins

import (
        "fmt"
        "os"
        "path/filepath"
        "plugin"
        "sync"
        "time"

        "github.com/ubom/workflow/layer0"
)

// DefaultPluginLoader implements plugin loading functionality
type DefaultPluginLoader struct {
        plugins map[string]*LoadedPlugin
        mutex   sync.RWMutex
        logger  Logger
}

// LoadedPlugin represents a loaded plugin with its metadata
type LoadedPlugin struct {
        Plugin     ExternalWorkPlugin
        Info       PluginInfo
        Status     string
        Error      string
        LoadedAt   time.Time
        ConfigHash string
}

// Logger interface for logging
type Logger interface {
        Info(msg string, fields ...interface{})
        Error(msg string, fields ...interface{})
        Debug(msg string, fields ...interface{})
        Warn(msg string, fields ...interface{})
}

// NewDefaultPluginLoader creates a new plugin loader
func NewDefaultPluginLoader(logger Logger) *DefaultPluginLoader {
        return &DefaultPluginLoader{
                plugins: make(map[string]*LoadedPlugin),
                mutex:   sync.RWMutex{},
                logger:  logger,
        }
}

// RegisterPlugin registers a plugin factory
func (l *DefaultPluginLoader) RegisterPlugin(name string, factory PluginFactory) error {
        l.mutex.Lock()
        defer l.mutex.Unlock()

        if _, exists := l.plugins[name]; exists {
                return fmt.Errorf("plugin %s is already registered", name)
        }

        // Create plugin instance
        plugin := factory()
        if plugin == nil {
                return fmt.Errorf("plugin factory returned nil for %s", name)
        }

        // Get plugin info
        info := plugin.GetPluginInfo()
        if info.Name == "" {
                info.Name = name
        }

        l.plugins[name] = &LoadedPlugin{
                Plugin:   plugin,
                Info:     info,
                Status:   "loaded",
                LoadedAt: time.Now(),
        }

        l.logger.Info("Plugin registered", "name", name, "version", info.Version)
        return nil
}

// UnregisterPlugin unregisters a plugin
func (l *DefaultPluginLoader) UnregisterPlugin(name string) error {
        l.mutex.Lock()
        defer l.mutex.Unlock()

        loadedPlugin, exists := l.plugins[name]
        if !exists {
                return fmt.Errorf("plugin %s is not registered", name)
        }

        // Shutdown plugin if it's initialized
        if loadedPlugin.Status == "initialized" {
                if err := loadedPlugin.Plugin.Shutdown(); err != nil {
                        l.logger.Warn("Error shutting down plugin", "name", name, "error", err)
                }
        }

        delete(l.plugins, name)
        l.logger.Info("Plugin unregistered", "name", name)
        return nil
}

// GetPlugin retrieves a plugin by name
func (l *DefaultPluginLoader) GetPlugin(name string) (ExternalWorkPlugin, error) {
        l.mutex.RLock()
        defer l.mutex.RUnlock()

        loadedPlugin, exists := l.plugins[name]
        if !exists {
                return nil, fmt.Errorf("plugin %s is not registered", name)
        }

        if loadedPlugin.Status == "error" {
                return nil, fmt.Errorf("plugin %s is in error state: %s", name, loadedPlugin.Error)
        }

        return loadedPlugin.Plugin, nil
}

// ListPlugins returns information about all loaded plugins
func (l *DefaultPluginLoader) ListPlugins() []PluginInfo {
        l.mutex.RLock()
        defer l.mutex.RUnlock()

        infos := make([]PluginInfo, 0, len(l.plugins))
        for _, loadedPlugin := range l.plugins {
                infos = append(infos, loadedPlugin.Info)
        }

        return infos
}

// LoadPluginFromFile loads a plugin from a shared library file
func (l *DefaultPluginLoader) LoadPluginFromFile(path string) error {
        l.logger.Info("Loading plugin from file", "path", path)

        // Open the plugin file
        p, err := plugin.Open(path)
        if err != nil {
                return fmt.Errorf("failed to open plugin file %s: %w", path, err)
        }

        // Look for the plugin factory function
        factorySymbol, err := p.Lookup("NewPlugin")
        if err != nil {
                return fmt.Errorf("plugin %s does not export NewPlugin function: %w", path, err)
        }

        // Cast to plugin factory
        factory, ok := factorySymbol.(func() ExternalWorkPlugin)
        if !ok {
                return fmt.Errorf("plugin %s NewPlugin function has wrong signature", path)
        }

        // Get plugin name from filename
        name := filepath.Base(path)
        name = name[:len(name)-len(filepath.Ext(name))] // Remove extension

        // Register the plugin
        return l.RegisterPlugin(name, factory)
}

// LoadPluginsFromDirectory loads all plugins from a directory
func (l *DefaultPluginLoader) LoadPluginsFromDirectory(dir string) error {
        l.logger.Info("Loading plugins from directory", "dir", dir)

        // Check if directory exists
        if _, err := os.Stat(dir); os.IsNotExist(err) {
                l.logger.Warn("Plugin directory does not exist", "dir", dir)
                return nil
        }

        // Walk through the directory
        return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
                if err != nil {
                        l.logger.Error("Error walking plugin directory", "path", path, "error", err)
                        return nil // Continue walking
                }

                // Skip directories and non-.so files
                if info.IsDir() || filepath.Ext(path) != ".so" {
                        return nil
                }

                // Load the plugin
                if err := l.LoadPluginFromFile(path); err != nil {
                        l.logger.Error("Failed to load plugin", "path", path, "error", err)
                        // Continue loading other plugins
                }

                return nil
        })
}

// InitializePlugin initializes a plugin with configuration
func (l *DefaultPluginLoader) InitializePlugin(name string, config map[string]interface{}) error {
        l.mutex.Lock()
        defer l.mutex.Unlock()

        loadedPlugin, exists := l.plugins[name]
        if !exists {
                return fmt.Errorf("plugin %s is not registered", name)
        }

        if loadedPlugin.Status == "initialized" {
                return fmt.Errorf("plugin %s is already initialized", name)
        }

        // Initialize the plugin
        if err := loadedPlugin.Plugin.Initialize(config); err != nil {
                loadedPlugin.Status = "error"
                loadedPlugin.Error = err.Error()
                return NewPluginError(name, "initialize", err)
        }

        loadedPlugin.Status = "initialized"
        loadedPlugin.Error = ""
        l.logger.Info("Plugin initialized", "name", name)
        return nil
}

// ShutdownPlugin shuts down a plugin
func (l *DefaultPluginLoader) ShutdownPlugin(name string) error {
        l.mutex.Lock()
        defer l.mutex.Unlock()

        loadedPlugin, exists := l.plugins[name]
        if !exists {
                return fmt.Errorf("plugin %s is not registered", name)
        }

        if loadedPlugin.Status != "initialized" {
                return fmt.Errorf("plugin %s is not initialized", name)
        }

        // Shutdown the plugin
        if err := loadedPlugin.Plugin.Shutdown(); err != nil {
                loadedPlugin.Status = "error"
                loadedPlugin.Error = err.Error()
                return NewPluginError(name, "shutdown", err)
        }

        loadedPlugin.Status = "loaded"
        loadedPlugin.Error = ""
        l.logger.Info("Plugin shut down", "name", name)
        return nil
}

// GetPluginStatus returns the status of a plugin
func (l *DefaultPluginLoader) GetPluginStatus(name string) (PluginStatus, error) {
        l.mutex.RLock()
        defer l.mutex.RUnlock()

        loadedPlugin, exists := l.plugins[name]
        if !exists {
                return PluginStatus{}, fmt.Errorf("plugin %s is not registered", name)
        }

        return PluginStatus{
                Name:        name,
                Status:      loadedPlugin.Status,
                Error:       loadedPlugin.Error,
                LastChecked: time.Now().Format(time.RFC3339),
                Version:     loadedPlugin.Info.Version,
        }, nil
}

// ListPluginStatuses returns the status of all plugins
func (l *DefaultPluginLoader) ListPluginStatuses() []PluginStatus {
        l.mutex.RLock()
        defer l.mutex.RUnlock()

        statuses := make([]PluginStatus, 0, len(l.plugins))
        for name, loadedPlugin := range l.plugins {
                statuses = append(statuses, PluginStatus{
                        Name:        name,
                        Status:      loadedPlugin.Status,
                        Error:       loadedPlugin.Error,
                        LastChecked: time.Now().Format(time.RFC3339),
                        Version:     loadedPlugin.Info.Version,
                })
        }

        return statuses
}

// HealthCheckAll performs health checks on all initialized plugins
func (l *DefaultPluginLoader) HealthCheckAll() map[string]error {
        l.mutex.RLock()
        defer l.mutex.RUnlock()

        results := make(map[string]error)
        for name, loadedPlugin := range l.plugins {
                if loadedPlugin.Status == "initialized" {
                        if err := loadedPlugin.Plugin.HealthCheck(); err != nil {
                                results[name] = err
                                l.logger.Warn("Plugin health check failed", "name", name, "error", err)
                        }
                }
        }

        return results
}

// GetPluginForWorkType finds a plugin that can handle the given work type
func (l *DefaultPluginLoader) GetPluginForWorkType(workType layer0.WorkType) (ExternalWorkPlugin, error) {
        l.mutex.RLock()
        defer l.mutex.RUnlock()

        for _, loadedPlugin := range l.plugins {
                if loadedPlugin.Status == "initialized" && loadedPlugin.Plugin.CanExecute(workType) {
                        return loadedPlugin.Plugin, nil
                }
        }

        return nil, fmt.Errorf("no plugin found for work type %s", workType)
}
