# ADR-004: Configurable CORS Policy

**Status:** Proposed  
**Date:** 2026-04-04  
**Deciders:** OpenVox View maintainers  
**Depends on:** ADR-001 (authentication, JWT Bearer tokens)

---

## Context

OpenVox View currently sets a wildcard CORS policy on all API responses:

```go
c.Header("Access-Control-Allow-Origin", "*")
c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE")
c.Header("Access-Control-Allow-Headers", "Authorization, *")
```

This was added to support local development, where the Quasar dev server (`http://localhost:9000`) and the Go backend (`http://localhost:5000`) run on different ports. The browser treats different ports as different origins and blocks cross-origin API responses unless the server explicitly allows it via CORS headers.

In production, OpenVox View is deployed behind a reverse proxy (e.g., Apache) that serves both the SPA and the API under a single origin (e.g., `https://openvoxview.example.com`). In this setup, all requests are same-origin and CORS headers are unnecessary.

The wildcard `*` is a security risk: any website can make authenticated API requests to OpenVox View and read the responses if the user has an active session. This was identified as **Vuln 1** in the security review (`docs/security/security-review-2026-04-04.md`).

---

## Decision

Replace the hardcoded wildcard CORS middleware with a **configurable `cors_origin`** setting. Behavior:

| `cors_origin` value | Behavior |
|---|---|
| Empty / not set (default) | No CORS headers sent. Same-origin requests work normally. Cross-origin requests are blocked by the browser. |
| Specific origin (e.g., `http://localhost:9000`) | CORS headers set with that exact origin. Only that origin can make cross-origin requests. |

This keeps the development workflow functional while eliminating the security risk in production.

---

## Configuration

### `config.yaml`

```yaml
cors_origin: ""  # default: empty = no CORS headers
```

Development example:

```yaml
cors_origin: "http://localhost:9000"
```

### Environment variable

```
OPENVOXVIEW_CORS_ORIGIN=http://localhost:9000
```

---

## Architecture Changes

### `config/config.go`

Add field to `Config` struct:

```go
CorsOrigin string `mapstructure:"cors_origin"`
```

Add Viper default and env binding:

```go
viper.SetDefault("cors_origin", "")
viper.BindEnv("cors_origin", "OPENVOXVIEW_CORS_ORIGIN")
```

### `main.go`

Replace the current `AllowCORS` middleware:

**Before:**
```go
func AllowCORS(c *gin.Context) {
    c.Header("Access-Control-Allow-Origin", "*")
    c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE")
    c.Header("Access-Control-Allow-Headers", "Authorization, *")
    if c.Request.Method == http.MethodOptions {
        c.Status(http.StatusNoContent)
        return
    }
    c.Next()
}
```

**After:**
```go
func CORSMiddleware(allowedOrigin string) gin.HandlerFunc {
    return func(c *gin.Context) {
        if allowedOrigin == "" {
            c.Next()
            return
        }
        c.Header("Access-Control-Allow-Origin", allowedOrigin)
        c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE")
        c.Header("Access-Control-Allow-Headers", "Authorization, Content-Type")
        if c.Request.Method == http.MethodOptions {
            c.Status(http.StatusNoContent)
            return
        }
        c.Next()
    }
}
```

Usage in `main.go`:

```go
r.Use(CORSMiddleware(cfg.CorsOrigin))
```

Key changes:
- No headers sent when `cors_origin` is empty (production default)
- Explicit origin instead of wildcard `*`
- `Access-Control-Allow-Headers` lists only required headers (`Authorization`, `Content-Type`) instead of wildcard `*`

---

## Security Considerations

- **Default-secure:** An unconfigured deployment sends no CORS headers, so the browser's same-origin policy is fully enforced.
- **No wildcard:** Even when configured, the origin is explicit — never `*`. This ensures only the specified origin can read API responses.
- **Header allowlist:** Only `Authorization` and `Content-Type` are allowed, not `*`.
- **Startup log:** When `cors_origin` is set, log a message so operators are aware: `CORS: allowing origin <value>`.

---

## Implementation Checklist

- [ ] Add `CorsOrigin` field to `Config` struct in `config/config.go`
- [ ] Add Viper default and `OPENVOXVIEW_CORS_ORIGIN` env binding
- [ ] Replace `AllowCORS` function with `CORSMiddleware` in `main.go`
- [ ] Update `CONFIGURATION.md` with `cors_origin` setting
- [ ] Build and verify (`go vet`, `go build`)
