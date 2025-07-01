
package schemas

import (
        "encoding/json"
        "fmt"
        "reflect"
        "regexp"
        "strconv"

        "github.com/ubom/workflow/layer0"
        "github.com/ubom/workflow/executors"
)

// SchemaValidator provides JSON schema validation for work definitions
type SchemaValidator struct {
        schemas map[layer0.WorkType]executors.WorkSchema
}

// NewSchemaValidator creates a new schema validator
func NewSchemaValidator() *SchemaValidator {
        return &SchemaValidator{
                schemas: make(map[layer0.WorkType]executors.WorkSchema),
        }
}

// RegisterSchema registers a schema for a work type
func (v *SchemaValidator) RegisterSchema(workType layer0.WorkType, schema executors.WorkSchema) {
        v.schemas[workType] = schema
}

// ValidateWork validates a work item against its registered schema
func (v *SchemaValidator) ValidateWork(work layer0.Work) error {
        schema, exists := v.schemas[work.GetType()]
        if !exists {
                return fmt.Errorf("no schema registered for work type %s", work.GetType())
        }

        // Parse the JSON schema
        var schemaObj map[string]interface{}
        if err := json.Unmarshal([]byte(schema.JSONSchema), &schemaObj); err != nil {
                return fmt.Errorf("invalid JSON schema for work type %s: %w", work.GetType(), err)
        }

        // Get executor config from work
        executorConfig, exists := work.GetConfiguration().Parameters["executor_config"]
        if !exists {
                return fmt.Errorf("executor_config not found in work parameters")
        }

        // Validate against schema
        return v.validateValue(executorConfig, schemaObj, "")
}

// ValidateConfiguration validates a configuration object against a schema
func (v *SchemaValidator) ValidateConfiguration(config interface{}, schemaJSON string) error {
        var schemaObj map[string]interface{}
        if err := json.Unmarshal([]byte(schemaJSON), &schemaObj); err != nil {
                return fmt.Errorf("invalid JSON schema: %w", err)
        }

        return v.validateValue(config, schemaObj, "")
}

// validateValue validates a value against a schema object
func (v *SchemaValidator) validateValue(value interface{}, schema map[string]interface{}, path string) error {
        // Check type
        if schemaType, exists := schema["type"]; exists {
                if err := v.validateType(value, schemaType.(string), path); err != nil {
                        return err
                }
        }

        // Check required fields for objects
        if required, exists := schema["required"]; exists && reflect.TypeOf(value).Kind() == reflect.Map {
                if err := v.validateRequired(value, required.([]interface{}), path); err != nil {
                        return err
                }
        }

        // Check properties for objects
        if properties, exists := schema["properties"]; exists && reflect.TypeOf(value).Kind() == reflect.Map {
                if err := v.validateProperties(value, properties.(map[string]interface{}), path); err != nil {
                        return err
                }
        }

        // Check items for arrays
        if items, exists := schema["items"]; exists && reflect.TypeOf(value).Kind() == reflect.Slice {
                if err := v.validateItems(value, items.(map[string]interface{}), path); err != nil {
                        return err
                }
        }

        // Check enum values
        if enum, exists := schema["enum"]; exists {
                if err := v.validateEnum(value, enum.([]interface{}), path); err != nil {
                        return err
                }
        }

        // Check pattern for strings
        if pattern, exists := schema["pattern"]; exists && reflect.TypeOf(value).Kind() == reflect.String {
                if err := v.validatePattern(value.(string), pattern.(string), path); err != nil {
                        return err
                }
        }

        // Check minimum/maximum for numbers
        if minimum, exists := schema["minimum"]; exists {
                if err := v.validateMinimum(value, minimum, path); err != nil {
                        return err
                }
        }

        if maximum, exists := schema["maximum"]; exists {
                if err := v.validateMaximum(value, maximum, path); err != nil {
                        return err
                }
        }

        return nil
}

// validateType validates the type of a value
func (v *SchemaValidator) validateType(value interface{}, expectedType string, path string) error {
        actualType := v.getJSONType(value)
        if actualType != expectedType {
                return fmt.Errorf("type mismatch at %s: expected %s, got %s", path, expectedType, actualType)
        }
        return nil
}

// validateRequired validates required fields in an object
func (v *SchemaValidator) validateRequired(value interface{}, required []interface{}, path string) error {
        valueMap, ok := value.(map[string]interface{})
        if !ok {
                return fmt.Errorf("expected object at %s", path)
        }

        for _, requiredField := range required {
                fieldName := requiredField.(string)
                if _, exists := valueMap[fieldName]; !exists {
                        return fmt.Errorf("required field %s missing at %s", fieldName, path)
                }
        }

        return nil
}

// validateProperties validates properties of an object
func (v *SchemaValidator) validateProperties(value interface{}, properties map[string]interface{}, path string) error {
        valueMap, ok := value.(map[string]interface{})
        if !ok {
                return fmt.Errorf("expected object at %s", path)
        }

        for fieldName, fieldValue := range valueMap {
                if propertySchema, exists := properties[fieldName]; exists {
                        fieldPath := path + "." + fieldName
                        if path == "" {
                                fieldPath = fieldName
                        }
                        if err := v.validateValue(fieldValue, propertySchema.(map[string]interface{}), fieldPath); err != nil {
                                return err
                        }
                }
        }

        return nil
}

// validateItems validates items in an array
func (v *SchemaValidator) validateItems(value interface{}, itemSchema map[string]interface{}, path string) error {
        valueSlice := reflect.ValueOf(value)
        if valueSlice.Kind() != reflect.Slice {
                return fmt.Errorf("expected array at %s", path)
        }

        for i := 0; i < valueSlice.Len(); i++ {
                item := valueSlice.Index(i).Interface()
                itemPath := fmt.Sprintf("%s[%d]", path, i)
                if err := v.validateValue(item, itemSchema, itemPath); err != nil {
                        return err
                }
        }

        return nil
}

// validateEnum validates enum values
func (v *SchemaValidator) validateEnum(value interface{}, enumValues []interface{}, path string) error {
        for _, enumValue := range enumValues {
                if reflect.DeepEqual(value, enumValue) {
                        return nil
                }
        }

        return fmt.Errorf("value at %s is not one of the allowed enum values", path)
}

// validatePattern validates string patterns
func (v *SchemaValidator) validatePattern(value, pattern, path string) error {
        matched, err := regexp.MatchString(pattern, value)
        if err != nil {
                return fmt.Errorf("invalid pattern %s: %w", pattern, err)
        }

        if !matched {
                return fmt.Errorf("value at %s does not match pattern %s", path, pattern)
        }

        return nil
}

// validateMinimum validates minimum values for numbers
func (v *SchemaValidator) validateMinimum(value interface{}, minimum interface{}, path string) error {
        valueFloat, err := v.toFloat64(value)
        if err != nil {
                return fmt.Errorf("cannot validate minimum for non-numeric value at %s", path)
        }

        minimumFloat, err := v.toFloat64(minimum)
        if err != nil {
                return fmt.Errorf("invalid minimum value in schema")
        }

        if valueFloat < minimumFloat {
                return fmt.Errorf("value at %s (%v) is less than minimum (%v)", path, valueFloat, minimumFloat)
        }

        return nil
}

// validateMaximum validates maximum values for numbers
func (v *SchemaValidator) validateMaximum(value interface{}, maximum interface{}, path string) error {
        valueFloat, err := v.toFloat64(value)
        if err != nil {
                return fmt.Errorf("cannot validate maximum for non-numeric value at %s", path)
        }

        maximumFloat, err := v.toFloat64(maximum)
        if err != nil {
                return fmt.Errorf("invalid maximum value in schema")
        }

        if valueFloat > maximumFloat {
                return fmt.Errorf("value at %s (%v) is greater than maximum (%v)", path, valueFloat, maximumFloat)
        }

        return nil
}

// getJSONType returns the JSON type of a value
func (v *SchemaValidator) getJSONType(value interface{}) string {
        if value == nil {
                return "null"
        }

        switch reflect.TypeOf(value).Kind() {
        case reflect.Bool:
                return "boolean"
        case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
                 reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
                 reflect.Float32, reflect.Float64:
                return "number"
        case reflect.String:
                return "string"
        case reflect.Slice, reflect.Array:
                return "array"
        case reflect.Map, reflect.Struct:
                return "object"
        default:
                return "unknown"
        }
}

// toFloat64 converts a value to float64
func (v *SchemaValidator) toFloat64(value interface{}) (float64, error) {
        switch v := value.(type) {
        case float64:
                return v, nil
        case float32:
                return float64(v), nil
        case int:
                return float64(v), nil
        case int32:
                return float64(v), nil
        case int64:
                return float64(v), nil
        case string:
                return strconv.ParseFloat(v, 64)
        default:
                return 0, fmt.Errorf("cannot convert %T to float64", value)
        }
}

// ValidationRule represents a custom validation rule
type ValidationRule struct {
        Field         string      `json:"field"`
        Type          string      `json:"type"`
        Required      bool        `json:"required"`
        Pattern       string      `json:"pattern,omitempty"`
        MinValue      interface{} `json:"min_value,omitempty"`
        MaxValue      interface{} `json:"max_value,omitempty"`
        AllowedValues []interface{} `json:"allowed_values,omitempty"`
}

// ValidateWithRules validates a value against custom validation rules
func (v *SchemaValidator) ValidateWithRules(value interface{}, rules []ValidationRule) error {
        valueMap, ok := value.(map[string]interface{})
        if !ok {
                return fmt.Errorf("expected object for rule validation")
        }

        for _, rule := range rules {
                fieldValue, exists := valueMap[rule.Field]
                
                // Check required
                if rule.Required && !exists {
                        return fmt.Errorf("required field %s is missing", rule.Field)
                }

                if !exists {
                        continue // Skip validation for optional missing fields
                }

                // Validate type
                if rule.Type != "" {
                        if err := v.validateType(fieldValue, rule.Type, rule.Field); err != nil {
                                return err
                        }
                }

                // Validate pattern
                if rule.Pattern != "" && reflect.TypeOf(fieldValue).Kind() == reflect.String {
                        if err := v.validatePattern(fieldValue.(string), rule.Pattern, rule.Field); err != nil {
                                return err
                        }
                }

                // Validate min/max values
                if rule.MinValue != nil {
                        if err := v.validateMinimum(fieldValue, rule.MinValue, rule.Field); err != nil {
                                return err
                        }
                }

                if rule.MaxValue != nil {
                        if err := v.validateMaximum(fieldValue, rule.MaxValue, rule.Field); err != nil {
                                return err
                        }
                }

                // Validate allowed values
                if len(rule.AllowedValues) > 0 {
                        if err := v.validateEnum(fieldValue, rule.AllowedValues, rule.Field); err != nil {
                                return err
                        }
                }
        }

        return nil
}
