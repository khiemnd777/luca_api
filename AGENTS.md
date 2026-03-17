# AGENTS.md

## Project Context
`luca_api` is a **Go** backend for Luca's production and operations platform. It is structured as a **modular monolith** with a gateway that boots and proxies multiple modules. The codebase is organized around clear module boundaries and layered backend design.

Primary stack:
- **Go**
- **Fiber**
- **PostgreSQL**
- **Ent ORM**
- **Flyway**
- **Redis**
- **JWT**
- **RBAC**
- **WebSocket**
- **YAML configuration**
- **Zap logging**
- **Cron jobs / workers**
- **Excel import/export**

## Architecture Rules
- Keep the existing module-based architecture intact.
- Follow the established layering: `handler -> service -> repository`.
- Put shared concerns in `shared/`, not inside feature modules.
- Treat `modules/main` as the core business domain and other modules as platform/domain support modules.
- Do not introduce a new architectural style unless explicitly requested.
- Prefer extending an existing module over creating a new one when the concern already has a clear owner.

## Ownership Map
- `gateway/`: bootstraps modules and reverse proxies routes.
- `modules/auth`: authentication.
- `modules/token`: refresh/access token lifecycle.
- `modules/profile`: current-user profile flows.
- `modules/user`: user management.
- `modules/rbac`: role/permission management.
- `modules/auditlog`: audit trail.
- `modules/notification`: notifications and unread state.
- `modules/realtime`: websocket delivery and broadcast.
- `modules/search`: centralized search and permission guards.
- `modules/metadata`: metadata collections, fields, mapping, import/export engine.
- `modules/photo`: photo upload, resize, metadata, cleanup.
- `modules/folder`: folder organization for photos.
- `modules/attribute`: attribute and option management.
- `modules/main`: department-scoped operational domain.

## `modules/main` Rules
- `modules/main` is department-aware. Assume department membership and scoped access matter.
- Before changing a business feature, inspect its `registry.go` and route registration path.
- Reuse the existing custom-field system instead of inventing one-off flexible JSON handling.
- Respect existing relation registrars in `modules/main/features/__relation`.
- If a feature already has import support, keep new fields compatible with import flows.
- If a feature affects reporting or lifecycle state, check whether dashboard jobs or stats rebuild logic also need updates.

## Data and Persistence Rules
- Use **Ent** patterns already present in the repo for normal entity access.
- Use **Flyway** SQL migrations for schema changes.
- Never make schema-only changes without checking whether generated Ent code, repositories, DTOs, handlers, and import/export paths also need updates.
- If adding fields to business entities, inspect:
- Ent schema / generated usage
- Flyway migration requirements
- repository queries
- service validation/business rules
- handler request/response structs
- search indexing hooks
- metadata/custom-field interactions
- dashboard/reporting dependencies

## API and Module Conventions
- Keep handlers thin.
- Put business rules in services.
- Put data access in repositories.
- Reuse existing middleware for auth, department membership, and internal-only routes.
- Match existing route shape and naming patterns.
- Avoid introducing inconsistent response formats; follow shared app/error helpers.
- Prefer explicit domain naming over generic helpers.

## Auth, Security, and Permissions
- Preserve **JWT** auth flow and internal-service boundaries.
- Do not bypass RBAC checks.
- When adding endpoints, decide explicitly whether they are:
- public
- authenticated
- internal-only
- department-member protected
- Search results must respect existing permission guard behavior.
- Delivery/public QR flows are special-case public endpoints; treat them carefully.

## Redis, Realtime, and Events
- Use existing cache helpers before adding direct Redis logic.
- Invalidate cache keys when mutating cached entities.
- For cross-module side effects, check existing pub/sub topics before inventing new integration patterns.
- If a change affects user-facing live updates, inspect `notification`, `realtime`, and `search` modules for follow-on updates.
- Keep event names and payloads consistent with current conventions.

## Files, Media, and Background Jobs
- Photo flows include upload, resize, metadata extraction, soft delete, and cleanup.
- If changing file handling, preserve storage path conventions and cleanup behavior.
- If adding scheduled behavior, integrate with the existing cron/job registration approach.
- If changing time-sensitive data like order codes or dashboard rebuild logic, inspect related jobs.

## Import/Export and Metadata
- This repo has a metadata-driven import/export system. Do not hardcode import behavior when a reusable mapping/profile path already exists.
- Validate identifier safety and SQL safety carefully in metadata/import code.
- When adding importable fields, check whether they belong in core fields, metadata fields, or external mappings.

## Coding Style
- Write idiomatic **Go**.
- Keep functions focused and direct.
- Prefer consistency with surrounding code over personal style preferences.
- Avoid premature abstraction.
- Avoid large rewrites when a targeted fix is enough.
- Add comments only where logic is non-obvious.
- Preserve ASCII unless the file already uses non-ASCII intentionally.

## What To Inspect Before Editing
For most non-trivial changes, inspect:
- module `main.go`
- relevant `registry.go`
- handler/service/repository chain
- related middleware
- config structs and `config.yaml.tmpl`
- migrations
- shared helpers that already solve the same problem
- cron/jobs if lifecycle or reporting is involved

## Expectations For Changes
- Make the smallest coherent change that fully solves the problem.
- Keep cross-module behavior aligned.
- Do not leave partially wired features.
- If you add a field or feature, update all layers that expose or depend on it.
- If tests exist nearby, update or add them.
- If no tests exist, at least reason through affected flows and note risk areas.

## Avoid
- Do not bypass module boundaries casually.
- Do not duplicate shared utilities inside modules.
- Do not add new frameworks or infrastructure without strong justification.
- Do not mix unrelated refactors into a feature change.
- Do not silently break cache behavior, search indexing, realtime notifications, or dashboard aggregates.

## Default Mindset
Act like a senior backend engineer working inside an existing production codebase:
- preserve architecture
- respect module ownership
- trace side effects
- update all dependent layers
- prefer pragmatic, maintainable changes
