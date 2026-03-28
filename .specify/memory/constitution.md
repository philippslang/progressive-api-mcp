<!--
SYNC IMPACT REPORT
==================
Version change: [unversioned template] → 1.0.0
Modified principles: N/A (initial constitution — all principles are new)

Added sections:
- Core Principles (5 principles)
  - I. Code Quality Standards
  - II. Test-First Development (NON-NEGOTIABLE)
  - III. Integration & Contract Testing
  - IV. Performance Requirements
  - V. Simplicity & Maintainability
- Quality Gates & CI/CD
- Development Workflow
- Governance

Removed sections: N/A (initial ratification)

Templates requiring updates:
- ✅ .specify/templates/plan-template.md — Constitution Check section references these principles
- ✅ .specify/templates/spec-template.md — Success Criteria section aligns with performance targets
- ✅ .specify/templates/tasks-template.md — Task phases align with test-first discipline
- ✅ No command files contained agent-specific references requiring update

Follow-up TODOs:
- None. All placeholders resolved.
-->

# Prograpimcp Constitution

## Core Principles

### I. Code Quality Standards

Every unit of code committed to this project MUST meet the following quality bar:

- Functions and methods MUST have a single, clearly stated responsibility (Single Responsibility Principle).
- Cyclomatic complexity per function MUST NOT exceed 10; values above 7 require justification in a code review comment.
- Public APIs and exported symbols MUST have documentation comments explaining purpose, parameters, and return values.
- Magic numbers and string literals MUST be extracted into named constants or configuration.
- Dead code, commented-out blocks, and TODO/FIXME notes MUST NOT be merged to the main branch without an associated tracked issue.
- Linting and static analysis MUST pass with zero warnings at merge time; suppression annotations require inline justification.

**Rationale**: Maintainable, readable code reduces defect rates and onboarding time. These rules are enforceable automatically and leave no room for subjective interpretation.

### II. Test-First Development (NON-NEGOTIABLE)

Test-Driven Development (TDD) is MANDATORY for all feature work:

- Tests MUST be written and reviewed before implementation begins (Red phase).
- The failing test suite MUST be demonstrated (or committed) before the implementation PR is opened.
- The Green phase MUST produce the minimal implementation that makes tests pass.
- Refactoring (Refactor phase) MUST leave the test suite green.
- Unit test coverage for new code MUST be ≥ 80%; critical paths (auth, data mutations, public API handlers) MUST reach ≥ 95%.
- Tests MUST be deterministic: no sleep-based timing, no dependency on external network in unit tests, no random seeds without fixed values.

**Rationale**: Tests written after the fact verify coincidental behavior, not intended contracts. Pre-implementation tests force precise requirement articulation and catch design issues before code is written.

### III. Integration & Contract Testing

Integration and contract tests MUST accompany every inter-component boundary:

- Every public API endpoint or service interface MUST have at least one contract test covering the happy path and at least one error path.
- Contract tests MUST run against real dependencies (real database, real message broker); mocks are only permitted for third-party external services outside our control.
- Integration test suites MUST be runnable in CI with deterministic outcomes using containerized dependencies (e.g., Docker Compose).
- Schema or interface changes MUST include a contract test update in the same commit — no schema drift.
- Test data MUST be isolated per test run; shared fixture mutation between tests is forbidden.

**Rationale**: Mocked integration tests have historically masked production failures (mock/prod divergence). Real-dependency tests are slower but provide genuine confidence in deployment.

### IV. Performance Requirements

Performance is a first-class feature, not an afterthought:

- Each feature specification MUST define at least one measurable performance success criterion (e.g., p95 latency ≤ 200ms, throughput ≥ 1000 req/s).
- Performance benchmarks MUST be included in the test suite for any hot-path operation and MUST be tracked across releases.
- Regressions exceeding 10% in any tracked benchmark MUST be investigated and addressed before merge.
- Database queries MUST be reviewed for index usage; N+1 query patterns are forbidden.
- Memory allocations in tight loops MUST be minimized; profiling output MUST be attached to PRs that touch performance-sensitive paths.
- Async/non-blocking I/O MUST be used for all network and disk operations.

**Rationale**: Retrofitting performance into a shipped system is costly and often architecturally disruptive. Treating performance as a gate criterion ensures it is addressed at the lowest possible cost.

### V. Simplicity & Maintainability

The simplest correct solution MUST be chosen over a clever or over-engineered one:

- YAGNI (You Aren't Gonna Need It): Features or abstractions not required by a current, tracked requirement MUST NOT be added.
- A new abstraction layer (interface, repository, service wrapper) MUST be justified with at least two concrete existing use cases — not hypothetical future ones.
- Dependencies MUST be evaluated for necessity; adding a dependency solely for convenience is forbidden when a standard-library solution is adequate.
- Code duplication up to three instances is acceptable; abstraction is REQUIRED only at the fourth instance (Rule of Three).
- All complexity deviations from these rules MUST be documented in the plan's Complexity Tracking table.

**Rationale**: Accidental complexity is the leading cause of maintenance burden. Explicit justification gates prevent well-intentioned engineers from gradually accreting unnecessary infrastructure.

## Quality Gates & CI/CD

Every pull request MUST pass the following automated gates before merge:

- **Lint**: Zero warnings; suppression annotations require inline comments.
- **Unit Tests**: All pass; coverage thresholds enforced (≥ 80% overall, ≥ 95% critical paths).
- **Contract Tests**: All pass against real containerized dependencies.
- **Benchmark Guard**: No tracked benchmark regresses > 10%.
- **Static Analysis / Security Scan**: Zero high/critical findings.
- **Build**: Project builds successfully with no errors.

Manual gates required before merge to `main`:
- At least one peer code review approval.
- Architecture review required for changes touching public API contracts, data models, or performance-critical paths.

## Development Workflow

The standard feature lifecycle MUST follow these steps in order:

1. **Spec** — Feature specification created with measurable acceptance criteria and performance success criteria (`/speckit.specify`).
2. **Clarify** — Ambiguities resolved before planning begins (`/speckit.clarify`).
3. **Plan** — Technical approach documented; Constitution Check passed (`/speckit.plan`).
4. **Tasks** — Implementation broken into independently testable, priority-ordered tasks (`/speckit.tasks`).
5. **Test-first** — Failing tests written and committed before implementation begins.
6. **Implement** — Minimal code to make tests green (`/speckit.implement`).
7. **Quality Gate** — All CI gates pass; peer review complete.
8. **Merge** — Squash-merge to `main` with a conventional commit message.

Hotfixes MUST follow the same workflow with an expedited review (single reviewer minimum). No exceptions for "quick fixes" — shortcuts introduce the bugs that hotfixes must later repair.

## Governance

This Constitution supersedes all other development practices, style guides, and informal team agreements. In case of conflict, this document governs.

**Amendment procedure**:
1. Propose the amendment in a dedicated PR with a summary of the change and its rationale.
2. At least two maintainers MUST review and approve.
3. A migration plan MUST accompany any amendment that invalidates existing code or tests.
4. The Constitution version MUST be bumped following semantic rules (see versioning policy below).
5. Dependent templates (`.specify/templates/`) MUST be updated in the same PR.

**Versioning policy**:
- **MAJOR**: Backward-incompatible governance change — principle removed, redefined, or existing code is retroactively non-compliant.
- **MINOR**: New principle or section added, or materially expanded guidance that does not invalidate existing work.
- **PATCH**: Clarification, wording improvement, or non-semantic refinement.

**Compliance review**: All PRs and code reviews MUST explicitly verify compliance with Core Principles. Reviewers are empowered to block merges that violate this Constitution, regardless of deadline pressure.

**Version**: 1.0.0 | **Ratified**: 2026-03-28 | **Last Amended**: 2026-03-28
