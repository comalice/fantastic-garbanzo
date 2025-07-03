package layer0

import (
	"testing"
	"time"
)

func TestNewContext(t *testing.T) {
	id := ContextID("test-context")
	scope := ContextScopeWorkflow
	name := "Test Context"

	context := NewContext(id, scope, name)

	if context.GetID() != id {
		t.Errorf("Expected ID %s, got %s", id, context.GetID())
	}

	if context.GetScope() != scope {
		t.Errorf("Expected scope %s, got %s", scope, context.GetScope())
	}

	if context.GetMetadata().Name != name {
		t.Errorf("Expected name %s, got %s", name, context.GetMetadata().Name)
	}

	if context.Size() != 0 {
		t.Errorf("Expected empty context, got size %d", context.Size())
	}
}

func TestNewChildContext(t *testing.T) {
	id := ContextID("child-context")
	scope := ContextScopeState
	name := "Child Context"
	parentID := ContextID("parent-context")

	context := NewChildContext(id, scope, name, parentID)

	if context.GetParentID() == nil || *context.GetParentID() != parentID {
		t.Errorf("Expected parent ID %s, got %v", parentID, context.GetParentID())
	}
}

func TestContextSetAndGet(t *testing.T) {
	context := NewContext("test", ContextScopeWorkflow, "Test")
	key := "test-key"
	value := "test-value"

	// Set value
	newContext := context.Set(key, value)

	// Get value from new context
	retrievedValue, exists := newContext.Get(key)
	if !exists {
		t.Error("Key should exist in new context")
	}

	if retrievedValue != value {
		t.Errorf("Expected value %s, got %v", value, retrievedValue)
	}

	// Original context should remain unchanged (immutability)
	_, existsInOriginal := context.Get(key)
	if existsInOriginal {
		t.Error("Original context should remain unchanged")
	}
}

func TestContextDelete(t *testing.T) {
	context := NewContext("test", ContextScopeWorkflow, "Test")
	key := "test-key"
	value := "test-value"

	// Set and then delete
	contextWithValue := context.Set(key, value)
	contextWithoutValue := contextWithValue.Delete(key)

	// Check that key is deleted in new context
	_, exists := contextWithoutValue.Get(key)
	if exists {
		t.Error("Key should not exist after deletion")
	}

	// Original context with value should remain unchanged (immutability)
	_, existsInOriginal := contextWithValue.Get(key)
	if !existsInOriginal {
		t.Error("Original context should remain unchanged")
	}
}

func TestContextHas(t *testing.T) {
	context := NewContext("test", ContextScopeWorkflow, "Test")
	key := "test-key"
	value := "test-value"

	// Initially should not have key
	if context.Has(key) {
		t.Error("Context should not have key initially")
	}

	// After setting should have key
	contextWithValue := context.Set(key, value)
	if !contextWithValue.Has(key) {
		t.Error("Context should have key after setting")
	}
}

func TestContextKeys(t *testing.T) {
	context := NewContext("test", ContextScopeWorkflow, "Test")

	// Initially should have no keys
	keys := context.Keys()
	if len(keys) != 0 {
		t.Errorf("Expected 0 keys, got %d", len(keys))
	}

	// Add some keys
	contextWithKeys := context.Set("key1", "value1").Set("key2", "value2")
	keys = contextWithKeys.Keys()

	if len(keys) != 2 {
		t.Errorf("Expected 2 keys, got %d", len(keys))
	}

	// Check that keys are present
	keyMap := make(map[string]bool)
	for _, key := range keys {
		keyMap[key] = true
	}

	if !keyMap["key1"] || !keyMap["key2"] {
		t.Error("Expected keys key1 and key2 to be present")
	}
}

func TestContextSize(t *testing.T) {
	context := NewContext("test", ContextScopeWorkflow, "Test")

	// Initially should be empty
	if context.Size() != 0 {
		t.Errorf("Expected size 0, got %d", context.Size())
	}

	// Add values and check size
	contextWithValues := context.Set("key1", "value1").Set("key2", "value2")
	if contextWithValues.Size() != 2 {
		t.Errorf("Expected size 2, got %d", contextWithValues.Size())
	}
}

func TestContextClear(t *testing.T) {
	context := NewContext("test", ContextScopeWorkflow, "Test")

	// Add some values
	contextWithValues := context.Set("key1", "value1").Set("key2", "value2")
	if contextWithValues.Size() != 2 {
		t.Error("Context should have 2 values before clearing")
	}

	// Clear context
	clearedContext := contextWithValues.Clear()
	if clearedContext.Size() != 0 {
		t.Errorf("Expected cleared context to be empty, got size %d", clearedContext.Size())
	}

	// Original context should remain unchanged (immutability)
	if contextWithValues.Size() != 2 {
		t.Error("Original context should remain unchanged")
	}
}

func TestContextMerge(t *testing.T) {
	context1 := NewContext("test1", ContextScopeWorkflow, "Test1")
	context2 := NewContext("test2", ContextScopeWorkflow, "Test2")

	// Add different values to each context
	context1WithValues := context1.Set("key1", "value1").Set("shared", "from1")
	context2WithValues := context2.Set("key2", "value2").Set("shared", "from2")

	// Merge context2 into context1
	mergedContext := context1WithValues.Merge(context2WithValues)

	// Check merged values
	if mergedContext.Size() != 3 {
		t.Errorf("Expected merged context to have 3 values, got %d", mergedContext.Size())
	}

	value1, exists1 := mergedContext.Get("key1")
	if !exists1 || value1 != "value1" {
		t.Error("key1 should exist with value1")
	}

	value2, exists2 := mergedContext.Get("key2")
	if !exists2 || value2 != "value2" {
		t.Error("key2 should exist with value2")
	}

	// Shared key should have value from context2 (merged context)
	sharedValue, existsShared := mergedContext.Get("shared")
	if !existsShared || sharedValue != "from2" {
		t.Error("shared key should have value from context2")
	}
}

func TestContextClone(t *testing.T) {
	original := NewContext("test", ContextScopeWorkflow, "Test")
	original.Metadata.Tags = []string{"tag1", "tag2"}
	original.Metadata.Properties["key"] = "value"
	originalWithValues := original.Set("data1", "value1").Set("data2", "value2")

	cloned := originalWithValues.Clone()

	// Verify clone has same values
	if cloned.GetID() != originalWithValues.GetID() {
		t.Error("Cloned context should have same ID")
	}

	if cloned.Size() != originalWithValues.Size() {
		t.Error("Cloned context should have same size")
	}

	// Verify data independence (modify clone)
	cloned = cloned.Set("data1", "modified").Set("new", "value")

	// Original should not be affected
	originalValue, _ := originalWithValues.Get("data1")
	if originalValue == "modified" {
		t.Error("Original context data should not be affected by clone modification")
	}

	_, hasNew := originalWithValues.Get("new")
	if hasNew {
		t.Error("Original context should not have new key from clone")
	}

	// Verify metadata independence
	cloned.Metadata.Tags[0] = "modified"
	cloned.Metadata.Properties["key"] = "modified"

	if originalWithValues.Metadata.Tags[0] == "modified" {
		t.Error("Original context tags should not be affected by clone modification")
	}

	if originalWithValues.Metadata.Properties["key"] == "modified" {
		t.Error("Original context properties should not be affected by clone modification")
	}
}

func TestContextValidate(t *testing.T) {
	// Valid context
	validContext := NewContext("test", ContextScopeWorkflow, "Test")
	if err := validContext.Validate(); err != nil {
		t.Errorf("Valid context should not return error: %v", err)
	}

	// Invalid contexts
	invalidContexts := []*Context{
		{ID: "", Scope: ContextScopeWorkflow, Metadata: ContextMetadata{Name: "Test"}},
		{ID: "test", Scope: "", Metadata: ContextMetadata{Name: "Test"}},
		{ID: "test", Scope: ContextScopeWorkflow, Metadata: ContextMetadata{Name: ""}},
	}

	for i, context := range invalidContexts {
		if err := context.Validate(); err == nil {
			t.Errorf("Invalid context %d should return error", i)
		}
	}
}

func TestContextTypedGetters(t *testing.T) {
	context := NewContext("test", ContextScopeWorkflow, "Test")

	// Test string getter
	contextWithString := context.Set("string_key", "test_string")
	strValue, exists := contextWithString.GetString("string_key")
	if !exists || strValue != "test_string" {
		t.Error("GetString should return correct string value")
	}

	// Test int getter
	contextWithInt := context.Set("int_key", 42)
	intValue, exists := contextWithInt.GetInt("int_key")
	if !exists || intValue != 42 {
		t.Error("GetInt should return correct int value")
	}

	// Test bool getter
	contextWithBool := context.Set("bool_key", true)
	boolValue, exists := contextWithBool.GetBool("bool_key")
	if !exists || !boolValue {
		t.Error("GetBool should return correct bool value")
	}

	// Test float64 getter
	contextWithFloat := context.Set("float_key", 3.14)
	floatValue, exists := contextWithFloat.GetFloat64("float_key")
	if !exists || floatValue != 3.14 {
		t.Error("GetFloat64 should return correct float64 value")
	}

	// Test type mismatch
	_, exists = contextWithString.GetInt("string_key")
	if exists {
		t.Error("GetInt should return false for string value")
	}
}

func TestContextConcurrency(t *testing.T) {
	context := NewContext("test", ContextScopeWorkflow, "Test")
	contextWithValue := context.Set("key", "value")

	// Test concurrent reads
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			value, exists := contextWithValue.Get("key")
			if !exists || value != "value" {
				t.Error("Concurrent read failed")
			}
			done <- true
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		select {
		case <-done:
		case <-time.After(time.Second):
			t.Error("Concurrent read test timed out")
		}
	}
}
