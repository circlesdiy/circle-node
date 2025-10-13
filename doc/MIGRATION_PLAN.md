# Migration Plan: Bootstrap Production-Ready Architecture

## Current State Analysis

**What exists:**
- Basic Go 1.24 app with HTMX + server-rendered templates
- Mock data-driven handlers (no database)
- Template system with layouts, components, pages
- Middleware (security, rate limiting)
- CSS build system with file watching
- Docker setup with PostgreSQL 16 & Redis 7 (for Docmost, not used by app)
- ~3,000 lines of Go code

**What's missing:**
- Database integration (pgx/v5)
- Redis caching (go-redis/v9)
- Authentication system (sessions + WebAuthn)
- Repository layer (all handlers use mock data)
- Service layer (business logic)
- Domain models matching ERD
- Migrations system
- YAML config with env overrides
- Graceful shutdown
- Template/static file embedding
- HTMX-aware response handling

## Migration Strategy: Incremental Refactor

**Phase 1: Foundation (Sessions 1-2)**
1. Update `go.mod` with dependencies (pgx, redis, zap, yaml.v3, WebAuthn libs)
2. Create YAML config system in `internal/config/config.go`
3. Build storage layer (`internal/storage/postgres.go`, `redis.go`)
4. Implement embedded migrations in `internal/storage/migrations/`
5. Create `internal/app/app.go` for bootstrap + graceful shutdown
6. Refactor `cmd/server/main.go` to use app bootstrap

**Phase 2: Core Infrastructure (Sessions 3-4)**
7. Build HTTP response helpers in `internal/http/response.go`
8. Create HTMX-aware middleware (detect `HX-Request` header)
9. Refactor template system to use embedding (`embed.FS`)
10. Add context helpers for auth data
11. Update middleware chain with logging (zap), CSRF, session validation

**Phase 3: Authentication Domain (Sessions 5-6)**
12. Create `internal/auth/` with full domain pattern:
    - `domain.go` - User, Session, Device, WebAuthnCredential models
    - `repository.go` - Database queries for auth tables
    - `service.go` - Session management, WebAuthn registration/verification
    - `handler.go` - Login, logout, register, WebAuthn challenge endpoints
13. Implement session middleware
14. Write initial migrations for User, Session, Device tables

**Phase 4: Profile Domain (Sessions 7-8)**
15. Create `internal/profile/` with domain pattern
16. Migrate existing profile handlers to use real database
17. Add profile repository with CRUD operations
18. Implement profile service with authorization checks
19. Update templates to work with real data

**Phase 5: Post Domain (Sessions 9-10)**
20. Create `internal/post/` with domain pattern
21. Implement post feed, create, edit, delete
22. Add reactions and comments as sub-resources
23. Integrate with circles for visibility

**Phase 6: Circle Domain (Sessions 11-12)**
24. Create `internal/circle/` with RBAC
25. Implement membership management
26. Add role/permission system
27. Integrate with post visibility

**Phase 7: Chat & Notifications (Sessions 13-14)**
28. Create `internal/chat/` with SSE support
29. Create `internal/notification/` with SSE
30. Implement real-time message delivery
31. Add read receipts and typing indicators

**Phase 8: Finalization (Session 15)**
32. Remove all mock data
33. Update Docker Compose to use app database
34. Create `Makefile` with common tasks
35. Test full stack with `docker-compose up`
36. Verify single binary contains all assets

## Key Decisions

- **Preserve working code:** Keep existing templates/handlers, refactor incrementally
- **Database-first:** Migrations follow ERD.md schema exactly
- **Standard library routing:** Use Go 1.22+ `"GET /posts/{id}"` patterns
- **Embed everything:** Templates, migrations, static files in binary
- **HTMX fragments:** Check header, return partial or full page
- **Session-based auth:** WebAuthn for passwordless, fallback to email magic links

## Success Metrics

✅ `go run cmd/server/main.go` starts with auto-migrations
✅ Login/logout with WebAuthn works
✅ Create post, add reaction, comment
✅ Join circle, view circle feed
✅ Send/receive messages with SSE
✅ `docker-compose up` runs full stack
✅ Single binary deployment ready

## Architecture Pattern

**Three-layer domain pattern:**
```
Handler (HTTP) → Service (business logic) → Repository (database)
```

Each domain (auth, post, profile, circle, chat, notification) has:
- `domain.go` - Models and business logic methods
- `handler.go` - HTTP handlers, registers routes with `*http.ServeMux`
- `service.go` - Business logic, authorization, validation
- `repository.go` - Database queries only

## Target Project Structure
```
cmd/server/main.go              # Entry point
internal/
  app/app.go                    # Bootstrap, dependency injection, graceful shutdown
  auth/                         # Session + WebAuthn authentication
  post/                         # Posts with reactions, comments
  profile/                      # User profiles
  circle/                       # Groups/communities with RBAC
  chat/                         # Messaging with SSE
  notification/                 # Notifications with SSE
  storage/
    postgres.go                 # Connection pool, embedded migrations
    redis.go                    # Cache client
    migrations/*.sql            # SQL migration files
  http/
    router.go                   # Route setup with net/http
    middleware.go               # Auth, logging, CSRF, security headers, rate limiting
    response.go                 # Response helpers
  template/
    template.go                 # Template compilation and rendering
  config/config.go              # YAML config with env overrides
web/
  templates/
    layouts/
    pages/
    components/
    partials/
  static/css/
  static/js/
deployments/docker/
  Dockerfile
  docker-compose.yml
config.example.yaml
Makefile
```
