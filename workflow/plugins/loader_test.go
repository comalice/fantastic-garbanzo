
package plugins

import (
	"context"
	"testing"

	"github.com/ubom/workflow/layer0"
	"github.com/ubom/workflow/executors"
)

// MockPlugin implements ExternalWorkPlugin for testing
type MockPlugin struct {
	*BasePlugin
	initialized bool
}

func NewMockPlugin() *MockPlugin {
	info := PluginInfo{
		Name:        "mock-plugin",
		Version:     "1.0.0",
		Author:      "Test",
		Description: "Mock plugin for testing",
		WorkTypes:   []string{"mock"},
		APIVersion:  "1.0",
	}

	metadata := executors.WorkMetadata{
		Name:        "Mock Plugin",
		Version:     "1.0.0",
		Author:      "Test",
		Description: "Mock plugin for testing",
		WorkTypes:   []string{"mock"},
	}

	schema := executors.WorkSchema{
		JSONSchema:    `{"type": "object"}`,
		Examples:      []executors.WorkDefinition{},
		Documentation: "Mock plugin schema",
	}

	return &MockPlugin{
		BasePlugin: NewBasePlugin(info, metadata, schema),
	}
}

func (p *MockPlugin) Execute(ctx context.Context, work layer0.Work, workContext *layer0.Context) (executors.WorkResult, error) {
	return executors.WorkResult{
		Success: true,
		Outputs: map[string]interface{}{"result": "mock"},
	}, nil
}

func (p *MockPlugin) Validate(work layer0.Work) error {
	return nil
}

func (p *MockPlugin) CanExecute(workType layer0.WorkType) bool {
	return workType == "mock"
}

func (p *MockPlugin) GetSupportedTypes() []layer0.WorkType {
	return []layer0.WorkType{"mock"}
}

func (p *MockPlugin) Initialize(config map[string]interface{}) error {
	p.initialized = true
	return nil
}

func (p *MockPlugin) Shutdown() error {
	p.initialized = false
	return nil
}

// MockLogger implements Logger for testing
type MockLogger struct{}

func (m *MockLogger) Info(msg string, fields ...interface{})  {}
func (m *MockLogger) Error(msg string, fields ...interface{}) {}
func (m *MockLogger) Debug(msg string, fields ...interface{}) {}
func (m *MockLogger) Warn(msg string, fields ...interface{})  {}

func TestDefaultPluginLoader_RegisterPlugin(t *testing.T) {
	logger := &MockLogger{}
	loader := NewDefaultPluginLoader(logger)

	factory := func() ExternalWorkPlugin {
		return NewMockPlugin()
	}

	err := loader.RegisterPlugin("mock-plugin", factory)
	if err != nil {
		t.Fatalf("RegisterPlugin failed: %v", err)
	}

	// Test duplicate registration
	err = loader.RegisterPlugin("mock-plugin", factory)
	if err == nil {
		t.Error("Expected error for duplicate plugin registration")
	}
}

func TestDefaultPluginLoader_GetPlugin(t *testing.T) {
	logger := &MockLogger{}
	loader := NewDefaultPluginLoader(logger)

	factory := func() ExternalWorkPlugin {
		return NewMockPlugin()
	}

	err := loader.RegisterPlugin("mock-plugin", factory)
	if err != nil {
		t.Fatalf("RegisterPlugin failed: %v", err)
	}

	plugin, err := loader.GetPlugin("mock-plugin")
	if err != nil {
		t.Fatalf("GetPlugin failed: %v", err)
	}

	if plugin == nil {
		t.Error("Expected plugin, got nil")
	}

	// Test non-existent plugin
	_, err = loader.GetPlugin("non-existent")
	if err == nil {
		t.Error("Expected error for non-existent plugin")
	}
}

func TestDefaultPluginLoader_InitializePlugin(t *testing.T) {
	logger := &MockLogger{}
	loader := NewDefaultPluginLoader(logger)

	factory := func() ExternalWorkPlugin {
		return NewMockPlugin()
	}

	err := loader.RegisterPlugin("mock-plugin", factory)
	if err != nil {
		t.Fatalf("RegisterPlugin failed: %v", err)
	}

	config := map[string]interface{}{"test": "config"}
	err = loader.InitializePlugin("mock-plugin", config)
	if err != nil {
		t.Fatalf("InitializePlugin failed: %v", err)
	}

	status, err := loader.GetPluginStatus("mock-plugin")
	if err != nil {
		t.Fatalf("GetPluginStatus failed: %v", err)
	}

	if status.Status != "initialized" {
		t.Errorf("Expected status 'initialized', got '%s'", status.Status)
	}
}
