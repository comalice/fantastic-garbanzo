
package layer1

import (
        "fmt"
        "sync"

        "github.com/ubom/workflow/layer0"
)

// StateMachineCore provides core state machine functionality
type StateMachineCore struct {
        states      map[layer0.StateID]layer0.State
        transitions map[layer0.TransitionID]layer0.Transition
        currentState *layer0.StateID
        mutex       sync.RWMutex
}

// StateMachineCoreInterface defines the contract for state machine operations
type StateMachineCoreInterface interface {
        AddState(state layer0.State) error
        RemoveState(stateID layer0.StateID) error
        GetState(stateID layer0.StateID) (layer0.State, error)
        GetAllStates() []layer0.State
        AddTransition(transition layer0.Transition) error
        RemoveTransition(transitionID layer0.TransitionID) error
        GetTransition(transitionID layer0.TransitionID) (layer0.Transition, error)
        GetTransitionsFromState(stateID layer0.StateID) []layer0.Transition
        GetTransitionsToState(stateID layer0.StateID) []layer0.Transition
        SetCurrentState(stateID layer0.StateID) error
        GetCurrentState() (*layer0.StateID, error)
        CanTransition(fromStateID, toStateID layer0.StateID) bool
        GetAvailableTransitions() []layer0.Transition
        ValidateStateMachine() error
}

// NewStateMachineCore creates a new state machine core
func NewStateMachineCore() *StateMachineCore {
        return &StateMachineCore{
                states:      make(map[layer0.StateID]layer0.State),
                transitions: make(map[layer0.TransitionID]layer0.Transition),
                mutex:       sync.RWMutex{},
        }
}

// AddState adds a state to the state machine
func (smc *StateMachineCore) AddState(state layer0.State) error {
        if err := state.Validate(); err != nil {
                return fmt.Errorf("invalid state: %w", err)
        }

        smc.mutex.Lock()
        defer smc.mutex.Unlock()

        if _, exists := smc.states[state.GetID()]; exists {
                return fmt.Errorf("state with ID %s already exists", state.GetID())
        }

        smc.states[state.GetID()] = state
        return nil
}

// RemoveState removes a state from the state machine
func (smc *StateMachineCore) RemoveState(stateID layer0.StateID) error {
        smc.mutex.Lock()
        defer smc.mutex.Unlock()

        if _, exists := smc.states[stateID]; !exists {
                return fmt.Errorf("state with ID %s does not exist", stateID)
        }

        // Check if state is referenced by any transitions
        for _, transition := range smc.transitions {
                if transition.GetFromStateID() == stateID || transition.GetToStateID() == stateID {
                        return fmt.Errorf("cannot remove state %s: referenced by transition %s", stateID, transition.GetID())
                }
        }

        // Check if it's the current state
        if smc.currentState != nil && *smc.currentState == stateID {
                return fmt.Errorf("cannot remove current state %s", stateID)
        }

        delete(smc.states, stateID)
        return nil
}

// GetState retrieves a state by ID
func (smc *StateMachineCore) GetState(stateID layer0.StateID) (layer0.State, error) {
        smc.mutex.RLock()
        defer smc.mutex.RUnlock()

        state, exists := smc.states[stateID]
        if !exists {
                return layer0.State{}, fmt.Errorf("state with ID %s does not exist", stateID)
        }

        return state, nil
}

// GetAllStates returns all states in the state machine
func (smc *StateMachineCore) GetAllStates() []layer0.State {
        smc.mutex.RLock()
        defer smc.mutex.RUnlock()

        states := make([]layer0.State, 0, len(smc.states))
        for _, state := range smc.states {
                states = append(states, state)
        }

        return states
}

// AddTransition adds a transition to the state machine
func (smc *StateMachineCore) AddTransition(transition layer0.Transition) error {
        if err := transition.Validate(); err != nil {
                return fmt.Errorf("invalid transition: %w", err)
        }

        smc.mutex.Lock()
        defer smc.mutex.Unlock()

        if _, exists := smc.transitions[transition.GetID()]; exists {
                return fmt.Errorf("transition with ID %s already exists", transition.GetID())
        }

        // Verify that from and to states exist
        if _, exists := smc.states[transition.GetFromStateID()]; !exists {
                return fmt.Errorf("from state %s does not exist", transition.GetFromStateID())
        }

        if _, exists := smc.states[transition.GetToStateID()]; !exists {
                return fmt.Errorf("to state %s does not exist", transition.GetToStateID())
        }

        smc.transitions[transition.GetID()] = transition
        return nil
}

// RemoveTransition removes a transition from the state machine
func (smc *StateMachineCore) RemoveTransition(transitionID layer0.TransitionID) error {
        smc.mutex.Lock()
        defer smc.mutex.Unlock()

        if _, exists := smc.transitions[transitionID]; !exists {
                return fmt.Errorf("transition with ID %s does not exist", transitionID)
        }

        delete(smc.transitions, transitionID)
        return nil
}

// GetTransition retrieves a transition by ID
func (smc *StateMachineCore) GetTransition(transitionID layer0.TransitionID) (layer0.Transition, error) {
        smc.mutex.RLock()
        defer smc.mutex.RUnlock()

        transition, exists := smc.transitions[transitionID]
        if !exists {
                return layer0.Transition{}, fmt.Errorf("transition with ID %s does not exist", transitionID)
        }

        return transition, nil
}

// GetTransitionsFromState returns all transitions from a specific state
func (smc *StateMachineCore) GetTransitionsFromState(stateID layer0.StateID) []layer0.Transition {
        smc.mutex.RLock()
        defer smc.mutex.RUnlock()

        var transitions []layer0.Transition
        for _, transition := range smc.transitions {
                if transition.GetFromStateID() == stateID {
                        transitions = append(transitions, transition)
                }
        }

        return transitions
}

// GetTransitionsToState returns all transitions to a specific state
func (smc *StateMachineCore) GetTransitionsToState(stateID layer0.StateID) []layer0.Transition {
        smc.mutex.RLock()
        defer smc.mutex.RUnlock()

        var transitions []layer0.Transition
        for _, transition := range smc.transitions {
                if transition.GetToStateID() == stateID {
                        transitions = append(transitions, transition)
                }
        }

        return transitions
}

// SetCurrentState sets the current state of the state machine
func (smc *StateMachineCore) SetCurrentState(stateID layer0.StateID) error {
        smc.mutex.Lock()
        defer smc.mutex.Unlock()

        if _, exists := smc.states[stateID]; !exists {
                return fmt.Errorf("state with ID %s does not exist", stateID)
        }

        smc.currentState = &stateID
        return nil
}

// GetCurrentState returns the current state ID
func (smc *StateMachineCore) GetCurrentState() (*layer0.StateID, error) {
        smc.mutex.RLock()
        defer smc.mutex.RUnlock()

        if smc.currentState == nil {
                return nil, fmt.Errorf("no current state set")
        }

        return smc.currentState, nil
}

// CanTransition checks if a transition from one state to another is possible
func (smc *StateMachineCore) CanTransition(fromStateID, toStateID layer0.StateID) bool {
        smc.mutex.RLock()
        defer smc.mutex.RUnlock()

        for _, transition := range smc.transitions {
                if transition.GetFromStateID() == fromStateID && transition.GetToStateID() == toStateID {
                        return true
                }
        }

        return false
}

// GetAvailableTransitions returns all transitions available from the current state
func (smc *StateMachineCore) GetAvailableTransitions() []layer0.Transition {
        smc.mutex.RLock()
        defer smc.mutex.RUnlock()

        if smc.currentState == nil {
                return []layer0.Transition{}
        }

        return smc.getTransitionsFromStateUnsafe(*smc.currentState)
}

// getTransitionsFromStateUnsafe is an internal method that doesn't acquire locks
func (smc *StateMachineCore) getTransitionsFromStateUnsafe(stateID layer0.StateID) []layer0.Transition {
        var transitions []layer0.Transition
        for _, transition := range smc.transitions {
                if transition.GetFromStateID() == stateID {
                        transitions = append(transitions, transition)
                }
        }
        return transitions
}

// ValidateStateMachine validates the entire state machine
func (smc *StateMachineCore) ValidateStateMachine() error {
        smc.mutex.RLock()
        defer smc.mutex.RUnlock()

        if len(smc.states) == 0 {
                return fmt.Errorf("state machine must have at least one state")
        }

        // Check for initial states
        hasInitialState := false
        for _, state := range smc.states {
                if state.GetType() == layer0.StateTypeInitial {
                        hasInitialState = true
                        break
                }
        }

        if !hasInitialState {
                return fmt.Errorf("state machine must have at least one initial state")
        }

        // Validate all states
        for _, state := range smc.states {
                if err := state.Validate(); err != nil {
                        return fmt.Errorf("invalid state %s: %w", state.GetID(), err)
                }
        }

        // Validate all transitions
        for _, transition := range smc.transitions {
                if err := transition.Validate(); err != nil {
                        return fmt.Errorf("invalid transition %s: %w", transition.GetID(), err)
                }

                // Verify referenced states exist
                if _, exists := smc.states[transition.GetFromStateID()]; !exists {
                        return fmt.Errorf("transition %s references non-existent from state %s", transition.GetID(), transition.GetFromStateID())
                }

                if _, exists := smc.states[transition.GetToStateID()]; !exists {
                        return fmt.Errorf("transition %s references non-existent to state %s", transition.GetID(), transition.GetToStateID())
                }
        }

        return nil
}
