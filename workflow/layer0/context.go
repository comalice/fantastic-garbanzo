package layer0

import (
	"fmt"
	"sync"
	"time"
)

// ContextID represents a unique identifier for a context
type ContextID string

// ContextScope defines the scope of the context
type ContextScope string

const (
	ContextScopeGlobal   ContextScope = "global"
	ContextScopeWorkflow ContextScope = "workflow"
	ContextScopeState    ContextScope = "state"
	ContextScopeWork     ContextScope = "work"
)

// ContextMetadata contains metadata about a context
type ContextMetadata struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Tags        []string          `json:"tags"`
	Properties  map[string]string `json:"properties"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

// Context represents an atomic context in the workflow system
// It provides a thread-safe way to store and retrieve data across workflow execution
type Context struct {
	ID       ContextID              `json:"id"`
	Scope    ContextScope           `json:"scope"`
	Metadata ContextMetadata        `json:"metadata"`
	Data     map[string]interface{} `json:"data"`
	ParentID *ContextID             `json:"parent_id,omitempty"`
	mutex    sync.RWMutex           `json:"-"` // Not serialized
}

// ContextInterface defines the contract for context operations
type ContextInterface interface {
	GetID() ContextID
	GetScope() ContextScope
	GetMetadata() ContextMetadata
	GetParentID() *ContextID
	Get(key string) (interface{}, bool)
	Set(key string, value interface{}) *Context
	Delete(key string) *Context
	Has(key string) bool
	Keys() []string
	Size() int
	Clear() *Context
	Merge(other *Context) *Context
	Clone() *Context
	Validate() error
}

// NewContext creates a new context with the given parameters
func NewContext(id ContextID, scope ContextScope, name string) *Context {
	now := time.Now()
	return &Context{
		ID:    id,
		Scope: scope,
		Metadata: ContextMetadata{
			Name:        name,
			Description: "",
			Tags:        []string{},
			Properties:  make(map[string]string),
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		Data:     make(map[string]interface{}),
		ParentID: nil,
		mutex:    sync.RWMutex{},
	}
}

// NewChildContext creates a new child context with the given parent
func NewChildContext(id ContextID, scope ContextScope, name string, parentID ContextID) *Context {
	context := NewContext(id, scope, name)
	context.ParentID = &parentID
	return context
}

// GetID returns the context ID
func (c *Context) GetID() ContextID {
	return c.ID
}

// GetScope returns the context scope
func (c *Context) GetScope() ContextScope {
	return c.Scope
}

// GetMetadata returns the context metadata
func (c *Context) GetMetadata() ContextMetadata {
	return c.Metadata
}

// GetParentID returns the parent context ID
func (c *Context) GetParentID() *ContextID {
	return c.ParentID
}

// Get retrieves a value from the context
func (c *Context) Get(key string) (interface{}, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	value, exists := c.Data[key]
	return value, exists
}

// cloneLocked creates a deep copy of the context without acquiring locks
// This method assumes the caller already holds the appropriate locks
func (c *Context) cloneLocked() *Context {
	metadata := ContextMetadata{
		Name:        c.Metadata.Name,
		Description: c.Metadata.Description,
		Tags:        make([]string, len(c.Metadata.Tags)),
		Properties:  make(map[string]string),
		CreatedAt:   c.Metadata.CreatedAt,
		UpdatedAt:   c.Metadata.UpdatedAt,
	}

	copy(metadata.Tags, c.Metadata.Tags)
	for k, v := range c.Metadata.Properties {
		metadata.Properties[k] = v
	}

	data := make(map[string]interface{})
	for k, v := range c.Data {
		data[k] = v // Shallow copy of values
	}

	var parentID *ContextID
	if c.ParentID != nil {
		id := *c.ParentID
		parentID = &id
	}

	return &Context{
		ID:       c.ID,
		Scope:    c.Scope,
		Metadata: metadata,
		Data:     data,
		ParentID: parentID,
		mutex:    sync.RWMutex{},
	}
}

// Set creates a new context with the key-value pair set (immutable)
func (c *Context) Set(key string, value interface{}) *Context {
	c.mutex.RLock()
	newContext := c.cloneLocked()
	c.mutex.RUnlock()

	// No need to lock newContext as it's not shared yet
	newContext.Data[key] = value
	newContext.Metadata.UpdatedAt = time.Now()
	return newContext
}

// Delete creates a new context with the key removed (immutable)
func (c *Context) Delete(key string) *Context {
	c.mutex.RLock()
	newContext := c.cloneLocked()
	c.mutex.RUnlock()

	// No need to lock newContext as it's not shared yet
	delete(newContext.Data, key)
	newContext.Metadata.UpdatedAt = time.Now()
	return newContext
}

// Has checks if a key exists in the context
func (c *Context) Has(key string) bool {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	_, exists := c.Data[key]
	return exists
}

// Keys returns all keys in the context
func (c *Context) Keys() []string {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	keys := make([]string, 0, len(c.Data))
	for key := range c.Data {
		keys = append(keys, key)
	}
	return keys
}

// Size returns the number of key-value pairs in the context
func (c *Context) Size() int {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	return len(c.Data)
}

// Clear creates a new context with all data cleared (immutable)
func (c *Context) Clear() *Context {
	c.mutex.RLock()
	newContext := c.cloneLocked()
	c.mutex.RUnlock()

	// No need to lock newContext as it's not shared yet
	newContext.Data = make(map[string]interface{})
	newContext.Metadata.UpdatedAt = time.Now()
	return newContext
}

// Merge creates a new context with data from another context merged in (immutable)
func (c *Context) Merge(other *Context) *Context {
	c.mutex.RLock()
	newContext := c.cloneLocked()
	c.mutex.RUnlock()

	other.mutex.RLock()
	for key, value := range other.Data {
		newContext.Data[key] = value
	}
	other.mutex.RUnlock()

	newContext.Metadata.UpdatedAt = time.Now()
	return newContext
}

// Clone creates a deep copy of the context
func (c *Context) Clone() *Context {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	return c.cloneLocked()
}

// Validate checks if the context is valid
func (c *Context) Validate() error {
	if c.ID == "" {
		return fmt.Errorf("context ID cannot be empty")
	}

	if c.Scope == "" {
		return fmt.Errorf("context scope cannot be empty")
	}

	if c.Metadata.Name == "" {
		return fmt.Errorf("context name cannot be empty")
	}

	return nil
}

// GetString retrieves a string value from the context
func (c *Context) GetString(key string) (string, bool) {
	value, exists := c.Get(key)
	if !exists {
		return "", false
	}

	if str, ok := value.(string); ok {
		return str, true
	}

	return "", false
}

// GetInt retrieves an int value from the context
func (c *Context) GetInt(key string) (int, bool) {
	value, exists := c.Get(key)
	if !exists {
		return 0, false
	}

	if i, ok := value.(int); ok {
		return i, true
	}

	return 0, false
}

// GetBool retrieves a bool value from the context
func (c *Context) GetBool(key string) (bool, bool) {
	value, exists := c.Get(key)
	if !exists {
		return false, false
	}

	if b, ok := value.(bool); ok {
		return b, true
	}

	return false, false
}

// GetFloat64 retrieves a float64 value from the context
func (c *Context) GetFloat64(key string) (float64, bool) {
	value, exists := c.Get(key)
	if !exists {
		return 0.0, false
	}

	if f, ok := value.(float64); ok {
		return f, true
	}

	return 0.0, false
}
