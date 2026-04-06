# ADR-001: Local User Authentication

**Status:** Implemented  
**Date:** 2026-04-03  
**Deciders:** OpenVox View maintainers

---

## Context

OpenVox View currently has no authentication. All API endpoints under `/api/v1/` are publicly accessible to anyone who can reach the server. This is a significant security gap, especially for the Puppet CA endpoints (`/api/v1/ca/*`) which allow signing and revoking certificates.

The application is a single Go binary that embeds the Vue 3 frontend. Adding authentication requires:
- A place to store user accounts
- A session mechanism that survives page reloads without forcing re-login
- A toggle so trusted-network deployments can run without auth

---

## Decisions

| Question | Decision | Rationale |
|---|---|---|
| User storage | **SQLite** | Enables user management API/UI; single-file DB fits the single-binary model |
| Token strategy | **Short-lived access JWT + long-lived refresh token** | Avoids frequent re-logins while keeping access tokens small and revocable |
| Authorization model | **Admin flag** (`is_admin` boolean on user) | User management restricted to admins; all other endpoints available to any authenticated user. `--create-admin` sets `is_admin = true`, SAML auto-provisioned users default to `false`. |
| Auth optional | **Config toggle** (`auth.enabled`) | Trusted-network deployments should not be forced to manage users |

---

## Token Strategy

| Token | TTL | Storage | Purpose |
|---|---|---|---|
| **Access token** (JWT) | `15m` (configurable) | Frontend `localStorage` via Pinia | Sent as `Authorization: Bearer` on every API request |
| **Refresh token** (opaque) | `30d` (configurable) | SQLite `refresh_tokens` table + frontend `localStorage` | Exchanges for a new access token; revocable server-side |

**Refresh flow:**
1. Access token expires → frontend receives `401`
2. Axios response interceptor automatically calls `POST /api/v1/auth/refresh` with the refresh token
3. If valid: new access + refresh tokens returned, request retried transparently
4. If invalid/expired: auth store cleared, redirect to login

This gives users persistent sessions without re-login until the refresh token expires (30 days) or is explicitly revoked.

---

## Database Design

### SQLite file location

Configurable via `auth.db_path` (default: `data/openvoxview.db` relative to binary). Created on startup if absent.

### Schema

```sql
CREATE TABLE users (
    id           INTEGER PRIMARY KEY AUTOINCREMENT,
    username     TEXT    UNIQUE NOT NULL,
    email        TEXT,
    display_name TEXT,
    password_hash TEXT,                    -- bcrypt hash (NULL for SAML users)
    auth_source  TEXT NOT NULL DEFAULT 'local',  -- 'local' | 'saml' (ADR-002)
    is_admin     BOOLEAN NOT NULL DEFAULT FALSE, -- only admins can manage users
    created_at   DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at   DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE refresh_tokens (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id    INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash TEXT    UNIQUE NOT NULL,    -- SHA-256 of the raw token
    expires_at DATETIME NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    revoked_at DATETIME                   -- NULL = active
);

CREATE INDEX idx_refresh_tokens_user_id ON refresh_tokens(user_id);
CREATE INDEX idx_refresh_tokens_hash    ON refresh_tokens(token_hash);
```

`password_hash` is `NULL` for SAML-provisioned users (see ADR-002). The `auth_source` column distinguishes local vs. SAML users and is shared with ADR-002.

### First-run bootstrapping

A `--create-admin` CLI flag creates the first user interactively if the DB is empty:

```
openvoxview --create-admin
Username: admin
Password: ****
Confirm:  ****
Admin user created.
```

On startup, if `auth.enabled = true` and the users table is empty, the server logs a prominent warning.

---

## Architecture Changes

### New Go Package: `db/`

```
db/
├── db.go          # Open/migrate SQLite, exported DB handle
├── users.go       # User CRUD
└── tokens.go      # Refresh token CRUD + cleanup
```

Uses `database/sql` with raw SQL (no ORM). Schema migrations run automatically on startup via versioned `CREATE TABLE IF NOT EXISTS` statements.

**New Go dependency:** `modernc.org/sqlite` — pure Go SQLite driver, no CGO required. Critical for cross-platform builds (Linux/FreeBSD, AMD64/ARM/ARM64) without a C toolchain.

### `config/config.go` — New auth config section

```go
type AuthConfig struct {
    Enabled          bool   `mapstructure:"enabled"`
    JwtSecret        string `mapstructure:"jwt_secret"`
    AccessTokenTTL   int    `mapstructure:"access_token_ttl_minutes"`  // default: 15
    RefreshTokenTTL  int    `mapstructure:"refresh_token_ttl_days"`    // default: 30
    DbPath           string `mapstructure:"db_path"`                   // default: "data/openvoxview.db"
}
```

Environment variable equivalents:
- `OPENVOXVIEW_AUTH_ENABLED`
- `OPENVOXVIEW_AUTH_JWT_SECRET`
- `OPENVOXVIEW_AUTH_DB_PATH`

### `handler/auth.go` — Auth endpoints

```
POST   /api/v1/auth/login    – local credential check → access + refresh tokens
POST   /api/v1/auth/refresh  – exchange refresh token → new access + refresh tokens (rotation)
POST   /api/v1/auth/logout   – revoke current refresh token
GET    /api/v1/auth/me       – return current user profile from token claims
```

**User management endpoints** (also in `handler/auth.go`, require auth + admin):

```
GET    /api/v1/auth/users        – list all users (admin only)
POST   /api/v1/auth/users        – create user (admin only)
PUT    /api/v1/auth/users/:id    – update user (admin only)
DELETE /api/v1/auth/users/:id    – delete user (admin only)
```

Login response body:

```json
{
  "Data": {
    "access_token": "eyJ...",
    "refresh_token": "opaque-random-64-bytes-hex",
    "expires_in": 900
  }
}
```

### `middleware/auth.go` — JWT validation

```go
func JWTAuthMiddleware(cfg *config.Config) gin.HandlerFunc {
    return func(c *gin.Context) {
        if !cfg.Auth.Enabled {
            c.Next()
            return
        }
        token := extractBearerToken(c)
        claims, err := validateToken(token, cfg.Auth.JwtSecret)
        if err != nil {
            c.AbortWithStatusJSON(http.StatusUnauthorized, NewErrorResponse(err))
            return
        }
        c.Set("user_id",  claims.Subject)   // user ID as string
        c.Set("username", claims["username"])
        c.Next()
    }
}
```

### JWT claims

```json
{
  "sub": "42",
  "username": "alice",
  "email": "alice@example.com",
  "display_name": "Alice Smith",
  "is_admin": true,
  "iat": 1234567890,
  "exp": 1234568790
}
```

### `main.go` — Route structure

```go
// Public (no auth)
r.POST("/api/v1/auth/login",   authHandler.Login)
r.POST("/api/v1/auth/refresh", authHandler.Refresh)
r.GET("/api/v1/version",       ...)
r.GET("/api/v1/meta",          ...)  // public so frontend can detect auth state

// Protected
api := r.Group("/api/v1/")
api.Use(middleware.JWTAuthMiddleware(cfg))
{
    api.GET("view/*",  ...)
    api.POST("pdb/*",  ...)
    api.POST("ca/*",   ...)
    api.POST("auth/logout",         authHandler.Logout)
    api.GET("auth/me",              authHandler.Me)
    api.GET("auth/users",           authHandler.ListUsers)
    api.POST("auth/users",          authHandler.CreateUser)
    api.PUT("auth/users/:id",       authHandler.UpdateUser)
    api.DELETE("auth/users/:id",    authHandler.DeleteUser)
}
```

---

## Frontend Changes

### New: `ui/src/stores/auth.ts` (Pinia, persisted)

```typescript
export const useAuthStore = defineStore('auth', () => {
  const accessToken  = ref<string | null>(null)
  const refreshToken = ref<string | null>(null)
  const username     = ref<string | null>(null)
  const email        = ref<string | null>(null)
  const displayName  = ref<string | null>(null)
  const expiresAt    = ref<number | null>(null)   // Unix seconds

  const isAuthenticated = computed(() =>
    !!accessToken.value && !!expiresAt.value && Date.now() < expiresAt.value * 1000
  )

  function setAuth(data: LoginResponse) { /* populate from response */ }
  function clearAuth() { /* null everything */ }

  return { accessToken, refreshToken, username, email, displayName,
           expiresAt, isAuthenticated, setAuth, clearAuth }
}, { persist: true })
```

### New: `ui/src/pages/LoginPage.vue`

- Centered Quasar `q-card` with logo
- Username + password inputs
- "Login" button → `Backend.login()` → `auth.setAuth()` → redirect to Dashboard
- Error message on `401`
- Conditionally shows "Login with SSO" button when SAML is enabled (see ADR-002)

### New: `ui/src/layouts/AuthLayout.vue`

Minimal layout (no sidebar/header) used only for the login page.

### `ui/src/router/routes.ts` — Guard + login route

```typescript
// Add before existing routes
{
  path: '/login',
  component: () => import('layouts/AuthLayout.vue'),
  children: [{ name: 'Login', path: '', component: () => import('pages/LoginPage.vue') }],
  meta: { public: true }
}

// Navigation guard
router.beforeEach((to) => {
  const auth = useAuthStore()
  if (!to.meta?.public && !auth.isAuthenticated) {
    return { name: 'Login' }
  }
})
```

### `ui/src/boot/axios.ts` — Interceptors

```typescript
// REQUEST: inject access token
api.interceptors.request.use((config) => {
  const auth = useAuthStore()
  if (auth.accessToken) config.headers.Authorization = `Bearer ${auth.accessToken}`
  return config
})

// RESPONSE: silent token refresh on 401, then retry once
let isRefreshing = false
let failedQueue: Array<{ resolve: Function; reject: Function }> = []

api.interceptors.response.use(null, async (error: AxiosError) => {
  if (error.response?.status !== 401) return Promise.reject(error)

  const auth = useAuthStore()
  if (!auth.refreshToken) { auth.clearAuth(); router.push({ name: 'Login' }); return }

  if (isRefreshing) {
    return new Promise((resolve, reject) => failedQueue.push({ resolve, reject }))
  }

  isRefreshing = true
  try {
    const res = await Backend.refreshToken(auth.refreshToken)
    auth.setAuth(res.data.Data)
    failedQueue.forEach(p => p.resolve())
    return api.request(error.config!)  // retry original request
  } catch {
    auth.clearAuth()
    router.push({ name: 'Login' })
  } finally {
    isRefreshing = false
    failedQueue = []
  }
})
```

### `ui/src/client/backend.ts` — New auth methods

```typescript
login(username: string, password: string): AxiosPromise<BaseResponse<LoginResponse>>
refreshToken(token: string): AxiosPromise<BaseResponse<LoginResponse>>
logout(): AxiosPromise<BaseResponse<null>>
getMe(): AxiosPromise<BaseResponse<UserProfile>>
getUsers(): AxiosPromise<BaseResponse<UserProfile[]>>
createUser(data: CreateUserRequest): AxiosPromise<BaseResponse<UserProfile>>
updateUser(id: number, data: UpdateUserRequest): AxiosPromise<BaseResponse<UserProfile>>
deleteUser(id: number): AxiosPromise<BaseResponse<null>>
```

### `ui/src/layouts/MainLayout.vue` — User menu

Add to toolbar: avatar showing `displayName`, dropdown with "My Account" and "Logout". Logout calls `Backend.logout()` then `auth.clearAuth()` then navigates to Login.

---

## Configuration Example

```yaml
auth:
  enabled: true
  jwt_secret: "minimum-32-character-random-secret-here"
  access_token_ttl_minutes: 15
  refresh_token_ttl_days: 30
  db_path: "data/openvoxview.db"
```

---

## Security Considerations

- **JWT secret**: Warn loudly at startup if shorter than 32 characters or if using a placeholder value
- **Refresh token rotation**: Each use of a refresh token invalidates it and issues a new one. If an old token is presented (reuse detection), revoke the entire token family for that user
- **bcrypt cost**: Use cost factor 12 (current recommended default)
- **Rate limiting**: `POST /api/v1/auth/login` — 5 attempts per IP per minute (in-memory, resets on restart)
- **HTTPS**: Access and refresh tokens in `localStorage` are only safe over HTTPS; document this requirement
- **DB file permissions**: The SQLite file contains password hashes; restrict to `0600`
- **Self-delete guard**: Prevent a user from deleting their own account via the management API
- **Self-demote guard**: Prevent an admin from removing their own admin flag (avoids lockout)
- **Admin-only user management**: Only users with `is_admin = true` can access user management endpoints (list, create, update, delete). The `is_admin` flag is included in JWT claims so the check doesn't require a DB call.

---

## Consequences

### Positive
- SQLite enables a full user management API (list/create/update/delete users) now and a management UI later
- Refresh token rotation gives persistent sessions with server-side revocability
- No external infrastructure; SQLite file ships alongside the binary
- Auth-disabled mode requires zero config changes for existing deployments

### Negative / Trade-offs
- Adds `modernc.org/sqlite` as a significant new dependency (~8 MB to binary size)
- SQLite file must be on persistent storage (not suitable for stateless container without a volume mount)
- In-memory rate limiter resets on restart (acceptable for v1)

### Future upgrade path
- Add TOTP/MFA per user
- Replace in-memory rate limiter with persistent counter in SQLite
- Add audit log table for login/logout/user-change events

---

## Implementation Checklist

**Backend**
- [x] Add `modernc.org/sqlite` to `go.mod`
- [x] Add `golang-jwt/jwt/v5` to `go.mod`
- [x] Create `db/` package: `db.go`, `users.go`, `tokens.go`
- [x] Add `AuthConfig` to `config/config.go` with Viper bindings and env vars
- [x] Create `middleware/auth.go` (JWT validation, auth-disabled passthrough)
- [x] Create `handler/auth.go` (login, refresh, logout, me, user CRUD)
- [x] Modify `main.go`: open DB, init handlers, register routes (public vs. protected)
- [x] Add `--create-admin` CLI flag
- [x] Update `CONFIGURATION.md`

**Frontend**
- [x] Create `ui/src/stores/auth.ts`
- [x] Create `ui/src/pages/LoginPage.vue`
- [x] Create `ui/src/layouts/AuthLayout.vue`
- [x] Update `ui/src/router/routes.ts` (login route + `beforeEach` guard)
- [x] Update `ui/src/boot/axios.ts` (request injector + 401 refresh interceptor)
- [x] Update `ui/src/client/backend.ts` (auth methods)
- [x] Update `ui/src/client/models.ts` (LoginResponse, UserProfile, etc.)
- [x] Update `ui/src/layouts/MainLayout.vue` (user menu + logout)
