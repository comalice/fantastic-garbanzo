
package layer1

import (
        "fmt"
        "time"

        "github.com/ubom/workflow/layer0"
)

// WorkflowDefinitionID represents a unique identifier for a workflow definition
type WorkflowDefinitionID string

// WorkflowDefinitionVersion represents a version of a workflow definition
type WorkflowDefinitionVersion string

// WorkflowDefinitionStatus represents the status of a workflow definition
type WorkflowDefinitionStatus string

const (
        WorkflowDefinitionStatusDraft     WorkflowDefinitionStatus = "draft"
        WorkflowDefinitionStatusActive    WorkflowDefinitionStatus = "active"
        WorkflowDefinitionStatusInactive  WorkflowDefinitionStatus = "inactive"
        WorkflowDefinitionStatusDeprecated WorkflowDefinitionStatus = "deprecated"
)

// WorkflowDefinitionMetadata contains metadata about a workflow definition
type WorkflowDefinitionMetadata struct {
        Name        string            `json:"name"`
        Description string            `json:"description"`
        Tags        []string          `json:"tags"`
        Properties  map[string]string `json:"properties"`
        CreatedAt   time.Time         `json:"created_at"`
        UpdatedAt   time.Time         `json:"updated_at"`
        CreatedBy   string            `json:"created_by"`
        UpdatedBy   string            `json:"updated_by"`
}

// WorkflowDefinition represents a complete workflow definition
type WorkflowDefinition struct {
        ID               WorkflowDefinitionID      `json:"id"`
        Version          WorkflowDefinitionVersion `json:"version"`
        Status           WorkflowDefinitionStatus  `json:"status"`
        Metadata         WorkflowDefinitionMetadata `json:"metadata"`
        StateMachine     *StateMachineCore         `json:"state_machine"`
        InitialStateID   layer0.StateID            `json:"initial_state_id"`
        FinalStateIDs    []layer0.StateID          `json:"final_state_ids"`
        ErrorStateIDs    []layer0.StateID          `json:"error_state_ids"`
        GlobalContext    *layer0.Context           `json:"global_context"`
        Configuration    WorkflowConfiguration     `json:"configuration"`
}

// WorkflowConfiguration contains configuration for workflow execution
type WorkflowConfiguration struct {
        MaxConcurrentInstances int           `json:"max_concurrent_instances"`
        DefaultTimeoutSeconds   int           `json:"default_timeout_seconds"`
        RetryPolicy             RetryPolicy   `json:"retry_policy"`
        CompensationEnabled     bool          `json:"compensation_enabled"`
        PersistenceEnabled      bool          `json:"persistence_enabled"`
        LoggingLevel           string        `json:"logging_level"`
        Environment            map[string]string `json:"environment"`
}

// RetryPolicy defines retry behavior for workflow operations
type RetryPolicy struct {
        MaxRetries        int           `json:"max_retries"`
        InitialDelay      time.Duration `json:"initial_delay"`
        MaxDelay          time.Duration `json:"max_delay"`
        BackoffMultiplier float64       `json:"backoff_multiplier"`
        RetryableErrors   []string      `json:"retryable_errors"`
}

// WorkflowDefinitionInterface defines the contract for workflow definition operations
type WorkflowDefinitionInterface interface {
        GetID() WorkflowDefinitionID
        GetVersion() WorkflowDefinitionVersion
        GetStatus() WorkflowDefinitionStatus
        GetMetadata() WorkflowDefinitionMetadata
        GetStateMachine() *StateMachineCore
        GetInitialStateID() layer0.StateID
        GetFinalStateIDs() []layer0.StateID
        GetErrorStateIDs() []layer0.StateID
        GetGlobalContext() layer0.Context
        GetConfiguration() WorkflowConfiguration
        SetStatus(status WorkflowDefinitionStatus) WorkflowDefinition
        SetStateMachine(stateMachine *StateMachineCore) WorkflowDefinition
        SetInitialStateID(stateID layer0.StateID) WorkflowDefinition
        AddFinalStateID(stateID layer0.StateID) WorkflowDefinition
        AddErrorStateID(stateID layer0.StateID) WorkflowDefinition
        UpdateGlobalContext(context layer0.Context) WorkflowDefinition
        UpdateConfiguration(config WorkflowConfiguration) WorkflowDefinition
        Validate() error
        Clone() WorkflowDefinition
        IsActive() bool
        CanExecute() bool
}

// NewWorkflowDefinition creates a new workflow definition
func NewWorkflowDefinition(id WorkflowDefinitionID, version WorkflowDefinitionVersion, name string) WorkflowDefinition {
        now := time.Now()
        return WorkflowDefinition{
                ID:      id,
                Version: version,
                Status:  WorkflowDefinitionStatusDraft,
                Metadata: WorkflowDefinitionMetadata{
                        Name:        name,
                        Description: "",
                        Tags:        []string{},
                        Properties:  make(map[string]string),
                        CreatedAt:   now,
                        UpdatedAt:   now,
                        CreatedBy:   "",
                        UpdatedBy:   "",
                },
                StateMachine:   NewStateMachineCore(),
                InitialStateID: "",
                FinalStateIDs:  []layer0.StateID{},
                ErrorStateIDs:  []layer0.StateID{},
                GlobalContext:  layer0.NewContext(layer0.ContextID(string(id)+"-global"), layer0.ContextScopeGlobal, "Global Context"),
                Configuration: WorkflowConfiguration{
                        MaxConcurrentInstances: 100,
                        DefaultTimeoutSeconds:  3600, // 1 hour
                        RetryPolicy: RetryPolicy{
                                MaxRetries:        3,
                                InitialDelay:      time.Second,
                                MaxDelay:          time.Minute,
                                BackoffMultiplier: 2.0,
                                RetryableErrors:   []string{},
                        },
                        CompensationEnabled: true,
                        PersistenceEnabled:  true,
                        LoggingLevel:       "INFO",
                        Environment:        make(map[string]string),
                },
        }
}

// GetID returns the workflow definition ID
func (wd WorkflowDefinition) GetID() WorkflowDefinitionID {
        return wd.ID
}

// GetVersion returns the workflow definition version
func (wd WorkflowDefinition) GetVersion() WorkflowDefinitionVersion {
        return wd.Version
}

// GetStatus returns the workflow definition status
func (wd WorkflowDefinition) GetStatus() WorkflowDefinitionStatus {
        return wd.Status
}

// GetMetadata returns the workflow definition metadata
func (wd WorkflowDefinition) GetMetadata() WorkflowDefinitionMetadata {
        return wd.Metadata
}

// GetStateMachine returns the state machine
func (wd WorkflowDefinition) GetStateMachine() *StateMachineCore {
        return wd.StateMachine
}

// GetInitialStateID returns the initial state ID
func (wd WorkflowDefinition) GetInitialStateID() layer0.StateID {
        return wd.InitialStateID
}

// GetFinalStateIDs returns the final state IDs
func (wd WorkflowDefinition) GetFinalStateIDs() []layer0.StateID {
        return wd.FinalStateIDs
}

// GetErrorStateIDs returns the error state IDs
func (wd WorkflowDefinition) GetErrorStateIDs() []layer0.StateID {
        return wd.ErrorStateIDs
}

// GetGlobalContext returns the global context
func (wd WorkflowDefinition) GetGlobalContext() *layer0.Context {
        return wd.GlobalContext
}

// GetConfiguration returns the workflow configuration
func (wd WorkflowDefinition) GetConfiguration() WorkflowConfiguration {
        return wd.Configuration
}

// SetStatus creates a new workflow definition with updated status (immutable)
func (wd WorkflowDefinition) SetStatus(status WorkflowDefinitionStatus) WorkflowDefinition {
        newWd := wd.Clone()
        newWd.Status = status
        newWd.Metadata.UpdatedAt = time.Now()
        return newWd
}

// SetStateMachine creates a new workflow definition with updated state machine (immutable)
func (wd WorkflowDefinition) SetStateMachine(stateMachine *StateMachineCore) WorkflowDefinition {
        newWd := wd.Clone()
        newWd.StateMachine = stateMachine
        newWd.Metadata.UpdatedAt = time.Now()
        return newWd
}

// SetInitialStateID creates a new workflow definition with updated initial state ID (immutable)
func (wd WorkflowDefinition) SetInitialStateID(stateID layer0.StateID) WorkflowDefinition {
        newWd := wd.Clone()
        newWd.InitialStateID = stateID
        newWd.Metadata.UpdatedAt = time.Now()
        return newWd
}

// AddFinalStateID creates a new workflow definition with an additional final state ID (immutable)
func (wd WorkflowDefinition) AddFinalStateID(stateID layer0.StateID) WorkflowDefinition {
        newWd := wd.Clone()
        newWd.FinalStateIDs = append(newWd.FinalStateIDs, stateID)
        newWd.Metadata.UpdatedAt = time.Now()
        return newWd
}

// AddErrorStateID creates a new workflow definition with an additional error state ID (immutable)
func (wd WorkflowDefinition) AddErrorStateID(stateID layer0.StateID) WorkflowDefinition {
        newWd := wd.Clone()
        newWd.ErrorStateIDs = append(newWd.ErrorStateIDs, stateID)
        newWd.Metadata.UpdatedAt = time.Now()
        return newWd
}

// UpdateGlobalContext creates a new workflow definition with updated global context (immutable)
func (wd WorkflowDefinition) UpdateGlobalContext(context *layer0.Context) WorkflowDefinition {
        newWd := wd.Clone()
        newWd.GlobalContext = context
        newWd.Metadata.UpdatedAt = time.Now()
        return newWd
}

// UpdateConfiguration creates a new workflow definition with updated configuration (immutable)
func (wd WorkflowDefinition) UpdateConfiguration(config WorkflowConfiguration) WorkflowDefinition {
        newWd := wd.Clone()
        newWd.Configuration = config
        newWd.Metadata.UpdatedAt = time.Now()
        return newWd
}

// IsActive checks if the workflow definition is active
func (wd WorkflowDefinition) IsActive() bool {
        return wd.Status == WorkflowDefinitionStatusActive
}

// CanExecute checks if the workflow definition can be executed
func (wd WorkflowDefinition) CanExecute() bool {
        return wd.Status == WorkflowDefinitionStatusActive && wd.Validate() == nil
}

// Clone creates a deep copy of the workflow definition
func (wd WorkflowDefinition) Clone() WorkflowDefinition {
        metadata := WorkflowDefinitionMetadata{
                Name:        wd.Metadata.Name,
                Description: wd.Metadata.Description,
                Tags:        make([]string, len(wd.Metadata.Tags)),
                Properties:  make(map[string]string),
                CreatedAt:   wd.Metadata.CreatedAt,
                UpdatedAt:   wd.Metadata.UpdatedAt,
                CreatedBy:   wd.Metadata.CreatedBy,
                UpdatedBy:   wd.Metadata.UpdatedBy,
        }
        
        copy(metadata.Tags, wd.Metadata.Tags)
        for k, v := range wd.Metadata.Properties {
                metadata.Properties[k] = v
        }
        
        finalStateIDs := make([]layer0.StateID, len(wd.FinalStateIDs))
        copy(finalStateIDs, wd.FinalStateIDs)
        
        errorStateIDs := make([]layer0.StateID, len(wd.ErrorStateIDs))
        copy(errorStateIDs, wd.ErrorStateIDs)
        
        // Clone configuration
        retryPolicy := RetryPolicy{
                MaxRetries:        wd.Configuration.RetryPolicy.MaxRetries,
                InitialDelay:      wd.Configuration.RetryPolicy.InitialDelay,
                MaxDelay:          wd.Configuration.RetryPolicy.MaxDelay,
                BackoffMultiplier: wd.Configuration.RetryPolicy.BackoffMultiplier,
                RetryableErrors:   make([]string, len(wd.Configuration.RetryPolicy.RetryableErrors)),
        }
        copy(retryPolicy.RetryableErrors, wd.Configuration.RetryPolicy.RetryableErrors)
        
        environment := make(map[string]string)
        for k, v := range wd.Configuration.Environment {
                environment[k] = v
        }
        
        configuration := WorkflowConfiguration{
                MaxConcurrentInstances: wd.Configuration.MaxConcurrentInstances,
                DefaultTimeoutSeconds:  wd.Configuration.DefaultTimeoutSeconds,
                RetryPolicy:            retryPolicy,
                CompensationEnabled:    wd.Configuration.CompensationEnabled,
                PersistenceEnabled:     wd.Configuration.PersistenceEnabled,
                LoggingLevel:          wd.Configuration.LoggingLevel,
                Environment:           environment,
        }
        
        return WorkflowDefinition{
                ID:               wd.ID,
                Version:          wd.Version,
                Status:           wd.Status,
                Metadata:         metadata,
                StateMachine:     wd.StateMachine, // Shallow copy - state machine is managed separately
                InitialStateID:   wd.InitialStateID,
                FinalStateIDs:    finalStateIDs,
                ErrorStateIDs:    errorStateIDs,
                GlobalContext:    wd.GlobalContext.Clone(),
                Configuration:    configuration,
        }
}

// Validate checks if the workflow definition is valid
func (wd WorkflowDefinition) Validate() error {
        if wd.ID == "" {
                return fmt.Errorf("workflow definition ID cannot be empty")
        }
        
        if wd.Version == "" {
                return fmt.Errorf("workflow definition version cannot be empty")
        }
        
        if wd.Status == "" {
                return fmt.Errorf("workflow definition status cannot be empty")
        }
        
        if wd.Metadata.Name == "" {
                return fmt.Errorf("workflow definition name cannot be empty")
        }
        
        if wd.StateMachine == nil {
                return fmt.Errorf("workflow definition must have a state machine")
        }
        
        // Validate state machine
        if err := wd.StateMachine.ValidateStateMachine(); err != nil {
                return fmt.Errorf("invalid state machine: %w", err)
        }
        
        // Validate initial state
        if wd.InitialStateID == "" {
                return fmt.Errorf("workflow definition must have an initial state")
        }
        
        if _, err := wd.StateMachine.GetState(wd.InitialStateID); err != nil {
                return fmt.Errorf("initial state %s does not exist in state machine: %w", wd.InitialStateID, err)
        }
        
        // Validate final states
        for _, stateID := range wd.FinalStateIDs {
                if _, err := wd.StateMachine.GetState(stateID); err != nil {
                        return fmt.Errorf("final state %s does not exist in state machine: %w", stateID, err)
                }
        }
        
        // Validate error states
        for _, stateID := range wd.ErrorStateIDs {
                if _, err := wd.StateMachine.GetState(stateID); err != nil {
                        return fmt.Errorf("error state %s does not exist in state machine: %w", stateID, err)
                }
        }
        
        // Validate global context
        if err := wd.GlobalContext.Validate(); err != nil {
                return fmt.Errorf("invalid global context: %w", err)
        }
        
        // Validate configuration
        if wd.Configuration.MaxConcurrentInstances <= 0 {
                return fmt.Errorf("max concurrent instances must be positive")
        }
        
        if wd.Configuration.DefaultTimeoutSeconds <= 0 {
                return fmt.Errorf("default timeout seconds must be positive")
        }
        
        if wd.Configuration.RetryPolicy.MaxRetries < 0 {
                return fmt.Errorf("max retries cannot be negative")
        }
        
        if wd.Configuration.RetryPolicy.InitialDelay < 0 {
                return fmt.Errorf("initial delay cannot be negative")
        }
        
        if wd.Configuration.RetryPolicy.MaxDelay < wd.Configuration.RetryPolicy.InitialDelay {
                return fmt.Errorf("max delay cannot be less than initial delay")
        }
        
        if wd.Configuration.RetryPolicy.BackoffMultiplier <= 0 {
                return fmt.Errorf("backoff multiplier must be positive")
        }
        
        return nil
}
