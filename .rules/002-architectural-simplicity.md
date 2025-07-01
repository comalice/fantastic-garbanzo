Applies to: *.go

Architectural Simplicity

These rules prevent over-abstraction and unnecessary complexity by establishing clear criteria for architectural decisions. They are informed by lessons learned from refactoring overly complex systems.

1. Interface Creation Criteria Create interfaces only when you have multiple concrete implementations or when testing requires mocking. A single implementation does not justify an interface. Ask: "Do I have 2+ real implementations today, or will I definitely need them within the next sprint?"

2. Abstraction Layer Justification Each abstraction layer must solve a concrete problem, not a hypothetical one. Before adding abstraction, document the specific problem it solves and the cost of not having it. Prefer composition over inheritance, and concrete types over interfaces until proven otherwise.

3. Registry and Factory Pattern Boundaries Use registries only when you need runtime discovery of implementations. Use factories only when object creation is complex or requires dependency injection. A simple constructor call is preferable to a factory method. Avoid registry-of-registries or factory-of-factories patterns.

4. Struct Design and Initialization Favor struct literals over constructor functions unless initialization requires validation or complex setup. Avoid getter/setter methods for simple data access - use public fields. If you find yourself writing more than 3 constructor variants, reconsider the design.

5. Metadata and Configuration Consolidation Consolidate similar metadata structures rather than creating type-specific variants. If two structs share 80% of their fields, they should likely be the same struct with optional fields or a discriminator. Avoid parallel hierarchies of configuration objects.

6. Red Flags for Over-Engineering Stop and reconsider when you encounter: more than 5 interfaces in a single package, constructor functions that only call struct literals, registries with fewer than 3 registered types, metadata structs that differ only in naming, or abstraction layers with 1:1 mappings to concrete implementations.

7. Decision Tree for Complexity Before adding complexity, ask in order: (1) Can this be solved with existing language features? (2) Can this be solved with a simple function? (3) Can this be solved with a struct and methods? (4) Do I really need an interface? (5) Do I really need a registry/factory? Each "yes" should stop the progression.

8. Refactoring Triggers Refactor toward simplicity when: interface count exceeds concrete type count in a package, boilerplate code (getters/setters/constructors) exceeds business logic, or when adding a new feature requires touching more than 3 abstraction layers. Complexity should serve the domain, not the architecture.