Applies to: all files.

# Development Philosophy

Our engineering is guided by a set of core principles that favor clarity, stability, and deliberate design.

**1. Simplicity and First Principles**
Build on core language features and minimal, well-established libraries. Favor writing focused, internal solutions over importing large dependencies to solve small problems.

**2. Clarity and Maintainability**
Write self-documenting code through clear naming and logical structure. Avoid cleverness for its own sake; if a complex pattern is unavoidable, its rationale and function must be made obvious with targeted documentation.

**3. Readability Over Premature Optimization**
Write clear, straightforward code first. Use profilers to identify and resolve performance bottlenecks only when requirements are not met. Trust the compiler and optimize based on evidence, not assumptions.

**4. Orthogonality and Conceptual Integrity**
Model the code on the domain. Concepts distinct in reality must be distinct in the implementation (e.g., a `Template` is not an `Instance`). This separation of concerns creates a more rational and refactorable codebase.

**5. Modular and Composable Design**
Architect the system as a collection of composable modules, each with a single responsibility and a well-defined interface. Favor a monorepo until separation becomes a practical necessity.

**6. Data-Centric Design**
Define canonical data structures that are consistent across the database, server, and APIs. A stable, predictable data interface is paramount for system integrity and reduces complexity.

**7. Local-First and Developer-Centric**
Ensure the development environment is self-contained and easy to run locally. Provide tooling that supports a tight feedback loop for both manual development and automated CI/CD pipelines.
