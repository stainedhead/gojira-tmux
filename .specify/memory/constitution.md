# gojira-tmux Constitution

This constitution governs development practices for the gojira-tmux project.

> **IMPORTANT FOR AI TOOLS**: This file references [AGENTS.md](../../AGENTS.md) as the single source of truth for all development rules, conventions, and practices. **Do not modify this file** to add or change rules. All updates to project guidelines must be made in AGENTS.md.

## Core Principles

See [AGENTS.md](../../AGENTS.md) for complete details on all principles below.

### I. Test-Driven Development (NON-NEGOTIABLE)

TDD is mandatory for all code changes. The red-green-refactor cycle must be strictly followed:
- Tests written and approved before implementation
- Tests must fail before implementation (red)
- Minimum code to pass (green)
- Refactor while maintaining green

### II. Clean Architecture

All code follows Clean Architecture with clear layer separation:
- Domain layer has no external dependencies
- Dependencies point inward
- Interfaces defined in domain, implemented in adapters

### III. Code Quality Gates

All changes must pass before merge:
- All tests passing
- Linting clean
- Successful build on all target platforms

### IV. CI/CD Pipeline

GitHub Actions enforces quality gates and produces multi-platform binaries distributed via GitHub Packages.

## Governance

- AGENTS.md is the authoritative source for all development rules
- This constitution references but does not duplicate AGENTS.md content
- Amendments require updates to AGENTS.md, not this file
- All PRs must verify compliance with AGENTS.md standards

**Version**: 1.0.0 | **Ratified**: 2026-01-01
