# Agent Rules & Development Workflow

This document dictates the deterministic development flow that all autonomous agents must follow when working on the `socialfoodie` repository. These rules apply to all tasks, implementation, testing, and git operations.

## 1. Branching & Git Workflow
*   **Branch Enforcement:** The repository uses two primary branches: `master` and `dev`. All development work MUST be done on the `dev` branch.
*   **Pre-Implementation Check:** Always verify you are on the `dev` branch before starting any implementation or before committing, unless explicitly instructed otherwise by the user.
*   **Commit Scope:** Only commit the specific files you have touched or modified for your task. Verify the `git status` or `git diff` before committing to ensure no unintended files are included.
*   **Conventional Commits:** Commit messages must be concise, describing exactly what was done, and must follow the conventional commits standard (e.g., `feat:`, `chore:`, `fix:`, `refactor:`).
*   **Push Policy:** Create commits locally but **DO NOT** push to the remote repository. Only push when explicitly demanded by the user.

## 2. Testing & Build Verification
*   **Test Coverage:** Always write unit tests to maximize coverage for new features or modifications.
*   **Pre-Commit Verification:** Before committing any changes, you MUST:
    1. Run all unit tests locally.
    2. Build the project locally.
    3. Build and test using Docker to verify containerized environments are not broken.
    4. Update the `README.md` or any other relevant documentation if the changes require it.
*   **No Broken Commits:** Never commit code that fails to build or fails tests.

## 3. Bug Fixes & Troubleshooting Protocol
*   When the user reports a bug, says "something is wrong", or requests a fix, follow this strict protocol:
    1. **Investigate:** Analyze logs, code, or run commands to understand the root cause.
    2. **Report & Plan:** Present your findings and propose a clear plan of action to the user.
    3. **Wait for Approval:** DO NOT execute the fix until the user explicitly approves the plan.

## 4. Architecture & Scalability
*   **Scalable Design:** All implementation and planning must be scalable and scale-directed. Always consider how a feature or database query will perform as the data volume and agent queries grow. Build for an event-driven, microservices-ready environment.

## 5. Infrastructure & Tooling Standards
*   **Database Migrations:** Schema changes must be version-controlled using a migration tool (e.g., `golang-migrate`). Never manually apply raw SQL changes without a tracked migration file.
*   **Integration & Contract Testing:** Because the system spans asynchronous boundaries (Python to Go via RabbitMQ), integration or schema contract tests must be written to ensure message payloads do not break between services.
*   **Secret Management:** Never commit secrets, API keys, or connection strings. Use `.env` files locally and Docker Compose `env_file` configurations. Ensure all secret files are added to `.gitignore`.
