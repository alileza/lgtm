<!--
Sync Impact Report:
- Version change: Template → 1.0.0 (Initial constitution)
- Added principles: All 5 core principles based on Go philosophy
- Added sections: Development Workflow, Architecture Constraints
- Templates requiring updates: ✅ Updated plan-template.md references
- Follow-up TODOs: None - all placeholders filled
-->

# LGTM Constitution

## Core Principles

### I. Clear is Better than Clever
Code must be readable and maintainable over being clever or compact. Optimize for the reader, not the writer. Explicit is better than implicit. Avoid magic numbers, unclear variable names, or overly complex one-liners. When in doubt, choose clarity.

**Rationale**: Code is read far more often than it is written. Clear code reduces bugs, speeds up debugging, and enables faster onboarding.

### II. Lazy and Pragmatic
Only add what you need, when you need it. Don't build for hypothetical future requirements. Start with the simplest solution that works. Add complexity only when the current solution becomes insufficient.

**Rationale**: YAGNI (You Aren't Gonna Need It) prevents over-engineering and reduces maintenance burden. Premature optimization and abstraction are the root of many software problems.

### III. Minimal Dependencies
Prefer standard library solutions over third-party dependencies. Each dependency must justify its inclusion by solving a significant problem that would be costly to implement in-house. Avoid framework-heavy solutions when simple approaches suffice.

**Rationale**: Dependencies introduce security risks, version conflicts, and maintenance overhead. The best dependency is no dependency.

### IV. Clear Boundaries
Split files and modules only when boundaries are clear and necessary. Organize code by functionality, not by artificial layers. Keep related code together until separation becomes obviously beneficial.

**Rationale**: Premature abstraction creates confusing code organization. Natural boundaries emerge from understanding the problem domain, not from following rigid patterns.

### V. Simple Directory Structure
Avoid complex directory hierarchies and framework-imposed structures. Use flat structures when possible. Create subdirectories only when grouping provides clear organizational benefit.

**Rationale**: Deep directory trees make navigation difficult and often reflect over-engineered architectures. Simple structures are easier to understand and maintain.

## Architecture Constraints

**File Organization**: Keep related functionality in the same file until it becomes unwieldy (>500-1000 lines depending on complexity). Split based on clear functional boundaries, not abstract patterns.

**Testing**: Write tests that focus on behavior, not implementation details. Test the public interface. Integration tests over unit tests when they provide better confidence with similar effort.

**Configuration**: Prefer convention over configuration. Use environment variables or simple config files. Avoid complex configuration systems unless absolutely necessary.

## Development Workflow

**Code Review**: Focus on clarity, simplicity, and correctness. Reject unnecessary complexity. Ask "Is this the simplest solution that works?"

**Refactoring**: Refactor when code becomes hard to understand or modify, not according to arbitrary patterns. Let the code tell you when it needs to be split or reorganized.

**Documentation**: Code should be self-documenting through clear names and structure. Add comments only for non-obvious business logic or complex algorithms.

## Governance

This constitution supersedes all other practices and conventions. All code changes must align with these principles. Complexity must be explicitly justified and approved.

When principles conflict, prioritize in this order: Clarity > Simplicity > Performance > Convenience.

All team members are responsible for upholding these principles in code reviews and design discussions.

**Version**: 1.0.0 | **Ratified**: 2025-11-14 | **Last Amended**: 2025-11-14
