
package workflow

import (
	"fmt"
	"sync"

	"github.com/ubom/workflow/layer0"
	"github.com/ubom/workflow/executors"
	"github.com/ubom/workflow/plugins"
)

// UnifiedRegistry is the single registry for all executors (built-in and plugins)
type UnifiedRegistry struct {
	executors map[layer0.WorkType]executors.WorkExecutor
	schemas   map[layer0.WorkType]executors.WorkSchema
	metadata  map[layer0.WorkType]executors.WorkMetadata
	plugins   map[string]plugins.ExternalWorkPlugin
	mutex     sync.RWMutex
	logger    Logger
}

// Logger interface for logging
type Logger interface {
	Info(msg string, fields ...interface{})
	Error(msg string, fields ...interface{})
	Debug(msg string, fields ...interface{})
	Warn(msg string, fields ...interface{})
}

// NewUnifiedRegistry creates a new unified registry
func NewUnifiedRegistry(logger Logger) *UnifiedRegistry {
	return &UnifiedRegistry{
		executors: make(map[layer0.WorkType]executors.WorkExecutor),
		schemas:   make(map[layer0.WorkType]executors.WorkSchema),
		metadata:  make(map[layer0.WorkType]executors.WorkMetadata),
		plugins:   make(map[string]plugins.ExternalWorkPlugin),
		mutex:     sync.RWMutex{},
		logger:    logger,
	}
}

// RegisterExecutor registers a built-in executor
func (r *UnifiedRegistry) RegisterExecutor(workType layer0.WorkType, executor executors.WorkExecutor) error {
	if executor == nil {
		return fmt.Errorf("executor cannot be nil")
	}

	r.mutex.Lock()
	defer r.mutex.Unlock()

	if _, exists := r.executors[workType]; exists {
		return fmt.Errorf("executor for work type %s already registered", workType)
	}

	r.executors[workType] = executor
	r.schemas[workType] = executor.GetSchema()
	r.metadata[workType] = executor.GetMetadata()

	if r.logger != nil {
		r.logger.Info("Built-in executor registered", "workType", workType)
	}

	return nil
}

// RegisterPlugin registers a plugin executor
func (r *UnifiedRegistry) RegisterPlugin(name string, plugin plugins.ExternalWorkPlugin) error {
	if plugin == nil {
		return fmt.Errorf("plugin cannot be nil")
	}

	r.mutex.Lock()
	defer r.mutex.Unlock()

	if _, exists := r.plugins[name]; exists {
		return fmt.Errorf("plugin %s already registered", name)
	}

	// Register plugin
	r.plugins[name] = plugin

	// Register plugin's supported work types
	supportedTypes := plugin.GetSupportedTypes()
	for _, workType := range supportedTypes {
		if _, exists := r.executors[workType]; exists {
			r.logger.Warn("Plugin overriding existing executor", "plugin", name, "workType", workType)
		}
		
		// Create adapter for plugin
		adapter := &PluginAdapter{plugin: plugin}
		r.executors[workType] = adapter
		r.schemas[workType] = plugin.GetSchema()
		r.metadata[workType] = plugin.GetMetadata()
	}

	if r.logger != nil {
		r.logger.Info("Plugin registered", "name", name, "supportedTypes", supportedTypes)
	}

	return nil
}

// UnregisterExecutor removes an executor from the registry
func (r *UnifiedRegistry) UnregisterExecutor(workType layer0.WorkType) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if _, exists := r.executors[workType]; !exists {
		return fmt.Errorf("no executor registered for work type %s", workType)
	}

	delete(r.executors, workType)
	delete(r.schemas, workType)
	delete(r.metadata, workType)

	if r.logger != nil {
		r.logger.Info("Executor unregistered", "workType", workType)
	}

	return nil
}

// UnregisterPlugin removes a plugin and its executors from the registry
func (r *UnifiedRegistry) UnregisterPlugin(name string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	plugin, exists := r.plugins[name]
	if !exists {
		return fmt.Errorf("plugin %s not registered", name)
	}

	// Remove plugin's work types from executors
	supportedTypes := plugin.GetSupportedTypes()
	for _, workType := range supportedTypes {
		delete(r.executors, workType)
		delete(r.schemas, workType)
		delete(r.metadata, workType)
	}

	delete(r.plugins, name)

	if r.logger != nil {
		r.logger.Info("Plugin unregistered", "name", name)
	}

	return nil
}

// GetExecutor retrieves an executor for a work type
func (r *UnifiedRegistry) GetExecutor(workType layer0.WorkType) (executors.WorkExecutor, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	executor, exists := r.executors[workType]
	if !exists {
		return nil, fmt.Errorf("no executor registered for work type %s", workType)
	}

	return executor, nil
}

// GetSchema retrieves the schema for a work type
func (r *UnifiedRegistry) GetSchema(workType layer0.WorkType) (executors.WorkSchema, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	schema, exists := r.schemas[workType]
	if !exists {
		return executors.WorkSchema{}, fmt.Errorf("no schema found for work type %s", workType)
	}

	return schema, nil
}

// GetMetadata retrieves the metadata for a work type
func (r *UnifiedRegistry) GetMetadata(workType layer0.WorkType) (executors.WorkMetadata, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	metadata, exists := r.metadata[workType]
	if !exists {
		return executors.WorkMetadata{}, fmt.Errorf("no metadata found for work type %s", workType)
	}

	return metadata, nil
}

// GetSupportedWorkTypes returns all registered work types
func (r *UnifiedRegistry) GetSupportedWorkTypes() []layer0.WorkType {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	types := make([]layer0.WorkType, 0, len(r.executors))
	for workType := range r.executors {
		types = append(types, workType)
	}

	return types
}

// ValidateWork validates a work item against its schema
func (r *UnifiedRegistry) ValidateWork(work layer0.Work) error {
	executor, err := r.GetExecutor(work.GetType())
	if err != nil {
		return err
	}

	return executor.Validate(work)
}

// PluginAdapter adapts external plugins to the unified WorkExecutor interface
type PluginAdapter struct {
	plugin plugins.ExternalWorkPlugin
}

// Execute adapts plugin execution to unified interface
func (a *PluginAdapter) Execute(ctx context.Context, work layer0.Work, workContext *layer0.Context) (executors.WorkResult, error) {
	// Convert context for plugin (plugins expect layer0.Context, not pointer)
	var pluginContext layer0.Context
	if workContext != nil {
		pluginContext = *workContext
	}

	result, err := a.plugin.Execute(work, pluginContext)
	if err != nil {
		return executors.WorkResult{
			Success: false,
			Error:   err.Error(),
		}, err
	}

	return executors.WorkResult{
		Success: true,
		Outputs: map[string]interface{}{
			"result": result,
		},
	}, nil
}

// Validate delegates to plugin
func (a *PluginAdapter) Validate(work layer0.Work) error {
	return a.plugin.Validate(work)
}

// CanExecute delegates to plugin
func (a *PluginAdapter) CanExecute(workType layer0.WorkType) bool {
	return a.plugin.CanExecute(workType)
}

// GetSupportedTypes delegates to plugin
func (a *PluginAdapter) GetSupportedTypes() []layer0.WorkType {
	return a.plugin.GetSupportedTypes()
}

// GetSchema delegates to plugin
func (a *PluginAdapter) GetSchema() executors.WorkSchema {
	return a.plugin.GetSchema()
}

// GetMetadata delegates to plugin
func (a *PluginAdapter) GetMetadata() executors.WorkMetadata {
	return a.plugin.GetMetadata()
}
