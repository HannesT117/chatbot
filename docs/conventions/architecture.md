# Architecture Conventions

## Design Principles

- Pull Complexity Downwards: Prioritize simple interfaces over simple implementations. Hide complexity in module internals, never expose it to callers.
- Design Deep Modules: Aim for modules with simple, narrow interfaces but significant internal functionality. Avoid shallow modules with large interfaces and minimal depth.
- Encapsulate Design Decisions: Each module should represent and encapsulate specific design decisions. The same design decision should never leak into multiple modules.
- Practice Strategic Programming: Invest in good design alongside working code. Don't just solve immediate problems; consider long-term system architecture.
- Design It Twice: For major design decisions, explore two radically different approaches before committing. Choose the superior design.
- Eliminate Special Cases: Design interfaces and logic to handle special cases automatically, without explicit error conditions or branching for edge cases.
- Use Clear, Obvious Name: Variable, function, and class names must unambiguously indicate purpose. Difficulty naming something signals a design problem.
- Write Valuable Comments: Comments must add information not deducible from code. Explain why and what's not obvious, never repeat the code itself.
- Avoid Over-Engineering: Build for known requirements, not speculative future ones. Solve problems when they arrive and you can see their actual shape.
- Manage Dependencies Ruthlessly: Minimize dependencies between modules. Dependencies and obscurity are the root causes of complexity.

## Red Flags

The following patterns are bad and MUST NOT appear in the code base.

- Shallow Module: Large or complex interface relative to the functionality provided.
- Information Leakage: A design decision or implementation detail is visible in multiple modules, creating coupling.
- Pass-Through Methods: Methods that merely forward calls to other methods with identical or near-identical signatures; they add no value.
- Temporal Decomposition: Modules split by execution order/timing rather than responsibility; related knowledge duplicated across modules (e.g., separate "train" and "test" dataset loaders instead of unified dataset module).
- Code Repetition: Non-trivial code fragments repeated in multiple places.
- General-Specific Mixture: Special-purpose logic mixed with general-purpose code rather than cleanly separated.
- Conjoined Methods: Two or more methods so interdependent they cannot be understood in isolation.
- Vague Names: Variable, function, or class names broad or ambiguous enough to refer to multiple concepts.
- Hard-to-Pick Names: Inability to name something concisely indicates unclear responsibility or poor design.
- Comment Repeats Code: Comments that simply restate what the code already expresses; they add no clarification or context.
- Non-Obvious Code: Logic is unclear or requires significant cognitive load to parse.
- Implementation Bleeds Into Interface: Internal implementation details or assumptions leak into the public API documentation or signatures.
- Overexposure: Internal data structures returned directly to callers, allowing unintended modification.
- Exception Proliferation: Too many exception types or conditions; exceptions used defensively rather than for genuinely exceptional circumstances.
- Unnecessary Inheritance or Decorators: Use of inheritance or decorator patterns where simpler approaches (composition, adding functionality directly) would suffice.
