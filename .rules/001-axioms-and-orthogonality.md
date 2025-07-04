# Atomic and Orthogonal Design Methodology Prompt Template

## System Context
You are tasked with designing a **[SYSTEM_TYPE]** for **[DOMAIN/USE_CASE]**. Apply the atomic and orthogonal design methodology to create a systematic, well-structured architecture.

## Design Methodology Instructions

### Phase 1: Axiomatic Foundation (Atomic Components)
Start with the most fundamental, indivisible components. Think axiomatically - what are the absolute basics that cannot be broken down further?

**Instructions:**
1. Identify the **atomic concepts** - the smallest, most fundamental units in **[DOMAIN]**
2. Each atomic component should be:
   - **Independent**: Can exist and be understood without reference to other components
   - **Indivisible**: Cannot be meaningfully broken into smaller parts within this domain
   - **Self-contained**: Has clear boundaries and responsibilities
3. List 5-8 atomic components maximum
4. For each atomic component, define:
   - **Purpose**: What fundamental need it serves
   - **Properties**: Its essential characteristics
   - **Boundaries**: What it does and does NOT handle

**Template:**
```
## Atomic Components for [SYSTEM_TYPE]

### [Atomic Component 1]
- **Purpose**: [Single, clear responsibility]
- **Properties**: [Essential characteristics]
- **Boundaries**: [What it handles / what it doesn't]

### [Atomic Component 2]
...
```

### Phase 2: Orthogonal Design (Independence Verification)
Ensure components are truly orthogonal - they should not need to know about each other's internal workings.

**Instructions:**
1. **Independence Test**: Each component should be replaceable without affecting others
2. **Knowledge Isolation**: Component A should not need to understand Component B's internals
3. **Interface Definition**: Define clean, minimal interfaces between components
4. **Dependency Analysis**: Verify no circular dependencies or tight coupling

**Verification Questions:**
- Can I change the internal implementation of Component X without affecting Component Y?
- Does Component A need to know how Component B works internally?
- Are the interfaces between components minimal and well-defined?

### Phase 3: Compositional Architecture (Building Upward)
Show how atomic components combine to create higher-level functionality.

**Instructions:**
1. **Composition Rules**: Define how atomic components can be combined
2. **Emergent Properties**: Identify what new capabilities emerge from combinations
3. **Composition Patterns**: Show common ways components work together
4. **Scalability**: Demonstrate how compositions can be further composed

**Template:**
```
## Compositional Patterns

### [Composition Pattern 1]
- **Components Used**: [List of atomic components]
- **Interaction Model**: [How they work together]
- **Emergent Capability**: [What new functionality emerges]
- **Interface**: [How this composition appears to external users]

### [Composition Pattern 2]
...
```

### Phase 4: Layer-Based Progression
Organize the system into clear architectural layers, each building on the previous.

**Instructions:**
1. **Layer 1 - Atomic**: Individual components with no dependencies
2. **Layer 2 - Compositional**: Combinations of atomic components
3. **Layer 3 - Operational**: Complete workflows/processes using compositions
4. **Layer 4 - System**: Full system integration and orchestration

Each layer should:
- Only depend on layers below it
- Provide clear abstractions to layers above it
- Have well-defined responsibilities

**Template:**
```
## Architectural Layers

### Layer 1: Atomic Components
[List atomic components]

### Layer 2: Compositional Units  
[Show how atomics combine]

### Layer 3: Operational Processes
[Show complete workflows/operations]

### Layer 4: System Integration
[Show full system coordination]
```

### Phase 5: Separation of Concerns
Maintain clear boundaries and responsibilities at each level.

**Instructions:**
1. **Single Responsibility**: Each component/layer has one primary concern
2. **Clear Interfaces**: Well-defined boundaries between concerns
3. **Minimal Coupling**: Reduce dependencies between different concerns
4. **High Cohesion**: Related functionality stays together

**Concern Categories to Consider:**
- **Data Management**: How information is stored and accessed
- **Processing Logic**: How operations are performed
- **Control Flow**: How execution is managed and coordinated
- **Communication**: How components interact
- **Configuration**: How behavior is customized
- **Error Handling**: How failures are managed

## Design Principles to Follow

### 1. Concept Before Implementation
- Focus on **what** needs to be done before **how** it's implemented
- Define clear conceptual models before technical details
- Ensure concepts are domain-appropriate and intuitive

### 2. Orthogonality Maintenance
- Components should be **independently replaceable**
- Changes to one component should not require changes to others
- Interfaces should be **stable and minimal**

### 3. Compositional Thinking
- Simple components should combine to create complex behaviors
- Compositions should be **predictable** and **understandable**
- Higher-level abstractions should **hide complexity** appropriately

### 4. Progressive Complexity
- Start simple and add complexity gradually
- Each layer should be **fully functional** at its level
- Avoid **premature optimization** or over-engineering

## Output Structure

Provide your design in this format:

```markdown
# [SYSTEM_TYPE] Design: [PROJECT_NAME]

## 1. Atomic Components
[List and define atomic components]

## 2. Orthogonality Analysis
[Verify independence and clean interfaces]

## 3. Compositional Patterns
[Show how components combine]

## 4. Architectural Layers
[Present layer-based progression]

## 5. Separation of Concerns
[Define clear responsibilities]

## 6. System Integration
[Show how everything works together]
```

## Quality Checklist

Before finalizing your design, verify:

- [ ] **Atomic components are truly indivisible** in this domain
- [ ] **Components are orthogonal** - can be changed independently
- [ ] **Compositions create meaningful abstractions** without leaking complexity
- [ ] **Layers have clear responsibilities** and dependencies flow downward only
- [ ] **Concerns are properly separated** with minimal coupling
- [ ] **The design is conceptually clean** before implementation details
- [ ] **The system can evolve** - new requirements can be accommodated cleanly

## Usage Instructions

1. **Replace placeholders**: Substitute [SYSTEM_TYPE], [DOMAIN], [PROJECT_NAME] with your specific context
2. **Apply systematically**: Follow each phase in order
3. **Iterate if needed**: Refine atomic components if compositions reveal issues
4. **Maintain discipline**: Resist the urge to jump to implementation details
5. **Validate orthogonality**: Regularly check that components remain independent

This methodology ensures robust, maintainable, and scalable system designs that can evolve cleanly over time.
