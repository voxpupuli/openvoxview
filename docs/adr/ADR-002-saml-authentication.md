# ADR-002: SAML Authentication (EntraID / ADFS)

**Status:** Implemented  
**Date:** 2026-04-03  
**Deciders:** OpenVox View maintainers  
**Depends on:** ADR-001 (shares SQLite DB, JWT issuance, refresh token mechanism, auth middleware, and frontend auth store)

---

## Context

Enterprise deployments need to integrate with corporate identity providers. Managing local user accounts (ADR-001) is operationally burdensome at scale and doesn't integrate with corporate MFA, conditional access policies, or account lifecycle management.

Both **Microsoft Entra ID (Azure AD)** and **ADFS** are targeted. SAML 2.0 is universally supported by both; OIDC/OAuth2 support in older ADFS versions (pre-4.0) is unreliable.

The SAML implementation builds directly on ADR-001's infrastructure: after the SAML assertion is validated, the same JWT access + refresh token pair is issued. The frontend auth flow is identical regardless of login method.

---

## Decisions

| Question | Decision | Rationale |
|---|---|---|
| Protocol | **SAML 2.0** | Universal EntraID + ADFS support; `crewjam/saml` is battle-tested |
| User provisioning | **Auto-provision on first SAML login** into SQLite `users` table | Eliminates manual account creation; attributes flow from IdP |
| SAML attributes mapped | **email, givenname, surname, displayname** | User profile fields; no group/role mapping needed |
| Token handoff SPA→SP | **Query param `?token=` on redirect** | Simple, consistent with single-binary model; frontend clears URL immediately |
| IdP metadata | **Fetched from URL at startup, cached** | Picks up IdP cert rotations automatically |

---

## SAML Attribute Mapping

The following SAML claim URIs are used. These are the standard Microsoft claim types emitted by both EntraID and ADFS by default:

| App field | SAML claim URI |
|---|---|
| `email` | `http://schemas.xmlsoap.org/ws/2005/05/identity/claims/emailaddress` |
| `given_name` | `http://schemas.xmlsoap.org/ws/2005/05/identity/claims/givenname` |
| `surname` | `http://schemas.xmlsoap.org/ws/2005/05/identity/claims/surname` |
| `display_name` | `http://schemas.microsoft.com/identity/claims/displayname` |
| `username` (identity) | `http://schemas.xmlsoap.org/ws/2005/05/identity/claims/emailaddress` (email used as username for SAML users) |

All attribute URIs are configurable to support non-Microsoft IdPs.

---

## Authentication Flow

```
Browser                  OpenVox View (SP)              EntraID / ADFS (IdP)
  |                             |                               |
  | GET /dashboard              |                               |
  |  (not authenticated)        |                               |
  |<-- 302 /login --------------|                               |
  |                             |                               |
  | Click "Login with SSO"      |                               |
  | GET /api/v1/auth/saml/login>|                               |
  |                             |-- SAMLRequest (redirect) ---> |
  |<-- 302 to IdP login --------|                               |
  |                             |                               |
  | User authenticates (+ MFA) --------------------------->     |
  |                             |                               |
  |<-- POST /api/v1/auth/saml/acs with SAMLResponse ------------|
  |                             |                               |
  |                             |-- validate signature          |
  |                             |-- check audience/conditions   |
  |                             |-- extract attributes          |
  |                             |-- upsert user in SQLite       |
  |                             |-- issue access + refresh JWT  |
  |<-- 302 /login?token=<jwt>&refresh=<token>                   |
  |                             |                               |
  | Frontend stores tokens      |                               |
  | Clears URL params           |                               |
  | Redirect to Dashboard       |                               |
```

---

## SQLite Schema Additions (extends ADR-001)

No new tables needed. The `users` table from ADR-001 already has `auth_source` and `email`/`display_name` columns. SAML users are distinguished by `auth_source = 'saml'` and have `password_hash = NULL`.

Two additional columns are added to `users`:

```sql
ALTER TABLE users ADD COLUMN given_name TEXT;
ALTER TABLE users ADD COLUMN surname    TEXT;
```

On each SAML login, the user record is **upserted** (insert or update on `email` conflict) so that profile changes in the IdP propagate automatically.

---

## Architecture Changes

### New Go Dependency

| Package | Purpose |
|---|---|
| `github.com/crewjam/saml` | SAML 2.0 SP implementation (signature verification, metadata generation, assertion parsing). Used in production by Grafana, HashiCorp Vault. |

### `config/config.go` — SAML sub-section under `AuthConfig`

```go
type AuthConfig struct {
    // ... existing ADR-001 fields ...
    Saml SamlConfig `mapstructure:"saml"`
}

type SamlConfig struct {
    Enabled         bool   `mapstructure:"enabled"`
    IdpMetadataURL  string `mapstructure:"idp_metadata_url"`   // preferred: live URL
    IdpMetadataFile string `mapstructure:"idp_metadata_file"`  // fallback: local file
    SpEntityID      string `mapstructure:"sp_entity_id"`       // e.g. https://openvoxview.example.com
    SpAcsURL        string `mapstructure:"sp_acs_url"`         // e.g. https://openvoxview.example.com/api/v1/auth/saml/acs
    SpCertFile      string `mapstructure:"sp_cert_file"`
    SpKeyFile       string `mapstructure:"sp_key_file"`

    // Attribute claim URIs (defaults to Microsoft standard claims)
    AttrEmail       string `mapstructure:"attr_email"`
    AttrGivenName   string `mapstructure:"attr_given_name"`
    AttrSurname     string `mapstructure:"attr_surname"`
    AttrDisplayName string `mapstructure:"attr_display_name"`
}
```

Environment variable equivalents:
- `OPENVOXVIEW_AUTH_SAML_ENABLED`
- `OPENVOXVIEW_AUTH_SAML_IDP_METADATA_URL`
- `OPENVOXVIEW_AUTH_SAML_SP_ENTITY_ID`
- `OPENVOXVIEW_AUTH_SAML_SP_ACS_URL`
- `OPENVOXVIEW_AUTH_SAML_SP_CERT_FILE`
- `OPENVOXVIEW_AUTH_SAML_SP_KEY_FILE`

### `middleware/saml.go` — SP initialization

```go
func NewSamlServiceProvider(cfg *config.SamlConfig) (*saml.ServiceProvider, error) {
    // 1. Load SP cert + key from files
    // 2. Fetch/parse IdP metadata from URL or file
    // 3. Return configured saml.ServiceProvider
}
```

The `ServiceProvider` instance is created once at startup and injected into `AuthHandler`. If `idp_metadata_url` is set, metadata is re-fetched every hour in the background to pick up IdP certificate rotations.

### `handler/auth.go` — SAML endpoints (adds to ADR-001)

```go
// GET /api/v1/auth/saml/metadata
// Returns SP metadata XML — paste this into EntraID/ADFS app registration
func (h *AuthHandler) SamlMetadata(c *gin.Context)

// GET /api/v1/auth/saml/login  
// Builds SAML AuthnRequest, redirects browser to IdP
func (h *AuthHandler) SamlLogin(c *gin.Context)

// POST /api/v1/auth/saml/acs  (Assertion Consumer Service — registered in IdP)
// Validates SAMLResponse, upserts user, issues JWT pair, redirects to frontend
func (h *AuthHandler) SamlACS(c *gin.Context)
```

**ACS handler logic:**

```go
func (h *AuthHandler) SamlACS(c *gin.Context) {
    // 1. Parse and validate SAML response (crewjam/saml handles sig, audience, timing)
    assertion, err := h.sp.ParseResponse(c.Request, []string{relayState})

    // 2. Extract attributes
    email       := getAttr(assertion, h.cfg.Auth.Saml.AttrEmail)
    givenName   := getAttr(assertion, h.cfg.Auth.Saml.AttrGivenName)
    surname     := getAttr(assertion, h.cfg.Auth.Saml.AttrSurname)
    displayName := getAttr(assertion, h.cfg.Auth.Saml.AttrDisplayName)

    // 3. Upsert user in SQLite
    user, err := h.db.UpsertSamlUser(email, givenName, surname, displayName)

    // 4. Issue access + refresh tokens (same function as local login)
    tokens, err := h.issueTokenPair(user)

    // 5. Redirect frontend with tokens
    redirectURL := fmt.Sprintf("/login?token=%s&refresh=%s", tokens.AccessToken, tokens.RefreshToken)
    c.Redirect(http.StatusFound, redirectURL)
}
```

### `main.go` — SAML routes (public, outside JWT middleware)

```go
// These must be public — browser redirects, no token available yet
if cfg.Auth.Saml.Enabled {
    r.GET("/api/v1/auth/saml/metadata", authHandler.SamlMetadata)
    r.GET("/api/v1/auth/saml/login",    authHandler.SamlLogin)
    r.POST("/api/v1/auth/saml/acs",     authHandler.SamlACS)
}
```

### `handler/core.go` — Meta response update

Add `SamlEnabled bool` to the `/api/v1/meta` response so the frontend knows whether to show the SSO button:

```go
type MetaResponse struct {
    CaEnabled       bool   `json:"CaEnabled"`
    CaReadOnly      bool   `json:"CaReadOnly"`
    UnreportedHours uint64 `json:"UnreportedHours"`
    StripPathPrefix string `json:"StripPathPrefix"`
    SamlEnabled     bool   `json:"SamlEnabled"`   // NEW
}
```

---

## Frontend Changes

No new dependencies. SAML is a browser redirect — the SPA only needs minor additions.

### `ui/src/client/models.ts`

```typescript
export interface ApiMeta {
  CaEnabled: boolean
  CaReadOnly: boolean
  UnreportedHours: number
  StripPathPrefix: string
  SamlEnabled: boolean      // NEW
}
```

### `ui/src/pages/LoginPage.vue` — SSO button + token extraction

```typescript
onMounted(async () => {
  // Case 1: returning from SAML ACS with tokens in URL
  if (route.query.token) {
    auth.setAuth({
      access_token:  route.query.token as string,
      refresh_token: route.query.refresh as string,
      // decode expiry from JWT claims
    })
    await router.replace({ name: 'Dashboard' })
    return
  }

  // Case 2: load meta to decide whether to show SSO button
  try {
    const meta = await Backend.getMeta()
    samlEnabled.value = meta.data.Data.SamlEnabled
  } catch { /* ignore, auth might be disabled */ }
})

function samlLogin() {
  // Full page navigation to trigger SAML redirect chain
  window.location.href = '/api/v1/auth/saml/login'
}
```

```vue
<template>
  <!-- Existing local login form -->
  <q-form @submit="localLogin"> ... </q-form>

  <q-separator v-if="samlEnabled" label="or" />

  <q-btn
    v-if="samlEnabled"
    label="Login with SSO"
    icon="corporate_fare"
    color="primary"
    class="full-width"
    @click="samlLogin"
  />
</template>
```

### User Management UI — SAML users are read-only

The IdP is the source of truth for SAML user profiles. When editing a SAML user (`auth_source = 'saml'`) in the User Management page:

- **Email** and **Display Name** fields are disabled (managed by IdP)
- **Password** fields are hidden (SAML users have no local password)
- **Save** button is hidden (no editable fields)
- An info banner explains that the profile is managed by the identity provider and updated on each login

No other frontend files need changes — the auth store, axios interceptors, router guard, and token refresh logic from ADR-001 all work unchanged for SAML-authenticated users.

---

## SP Certificate

The SAML SP requires an X.509 certificate for the SP metadata (and optionally signing AuthnRequests). This is independent of PuppetDB TLS certificates.

A `--generate-saml-cert` CLI flag generates a self-signed cert:

```
openvoxview --generate-saml-cert --output-dir /etc/openvoxview/
Generated: /etc/openvoxview/saml-sp.crt
Generated: /etc/openvoxview/saml-sp.key
```

The `.crt` content is embedded in the SP metadata XML that the IdP reads. Copy-paste it into the IdP's "SAML Signing Certificate" field during app registration.

---

## IdP Setup Guides

### Microsoft Entra ID

1. **Azure Portal** → Enterprise Applications → New Application → "Create your own application" (non-gallery)
2. **Single Sign-On** → SAML
3. **Basic SAML Configuration:**
   - Identifier (Entity ID): value of `sp_entity_id`
   - Reply URL (ACS): value of `sp_acs_url`
   - Sign on URL: `https://<host>/api/v1/auth/saml/login`
4. **Attributes & Claims** — ensure these are emitted (they are by default):
   - `emailaddress` → `user.mail`
   - `givenname` → `user.givenname`
   - `surname` → `user.surname`
   - `displayname` → `user.displayname`
5. **SAML Certificates** → copy the **App Federation Metadata Url** → use as `idp_metadata_url`
6. **Users and Groups** → assign the users/groups who should have access

### ADFS

1. **ADFS Management** → Trust Relationships → Relying Party Trusts → Add Relying Party Trust
2. Choose **"Import data about the relying party from a URL"** → enter `https://<host>/api/v1/auth/saml/metadata`
3. **Issuance Transform Rules** → Add Rule → "Send LDAP Attributes as Claims":
   | LDAP Attribute | Outgoing Claim Type |
   |---|---|
   | E-Mail-Addresses | E-Mail Address |
   | Given-Name | Given Name |
   | Surname | Surname |
   | Display-Name | `http://schemas.microsoft.com/identity/claims/displayname` |
4. No relying party token encryption needed unless required by policy

---

## Configuration Example

```yaml
auth:
  enabled: true
  jwt_secret: "minimum-32-character-random-secret-here"
  access_token_ttl_minutes: 15
  refresh_token_ttl_days: 30
  db_path: "data/openvoxview.db"

  # Local users still work alongside SAML — useful for a break-glass admin
  users: []  # managed via API

  saml:
    enabled: true
    idp_metadata_url: "https://login.microsoftonline.com/<tenant-id>/federationmetadata/2007-06/federationmetadata.xml"
    sp_entity_id: "https://openvoxview.example.com"
    sp_acs_url:   "https://openvoxview.example.com/api/v1/auth/saml/acs"
    sp_cert_file: "/etc/openvoxview/saml-sp.crt"
    sp_key_file:  "/etc/openvoxview/saml-sp.key"
    # Attribute URIs below are the Microsoft defaults — only change if your IdP differs
    attr_email:        "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/emailaddress"
    attr_given_name:   "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/givenname"
    attr_surname:      "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/surname"
    attr_display_name: "http://schemas.microsoft.com/identity/claims/displayname"
```

---

## Security Considerations

- **Assertion validation**: `crewjam/saml` validates signature, issuer, audience restriction, `NotBefore`/`NotOnOrAfter`, and `InResponseTo` (replay prevention) by default — do not bypass these checks
- **ACS URL must be HTTPS**: EntraID and ADFS reject HTTP ACS URLs; document this as a hard requirement
- **Token query params**: `?token=` and `?refresh=` are short-lived (access token TTL) and cleared from the URL immediately by the SPA. They will appear in browser history — a known trade-off of this approach
- **RelayState**: Managed by `crewjam/saml` to prevent CSRF on the ACS endpoint
- **SAML + local coexistence**: Operators can have both enabled simultaneously. A "break-glass" local admin account is recommended so access is not lost if the IdP is unreachable
- **IdP metadata refresh**: Re-fetch metadata hourly to automatically handle IdP signing certificate rotations without a restart
- **Attribute assertion**: If a SAML login arrives with no `email` attribute, reject it with a clear error — email is the user's primary identity key

---

## Consequences

### Positive
- Users authenticate with corporate credentials; inherits MFA and conditional access from the IdP
- Account lifecycle (onboarding/offboarding) managed centrally in the IdP
- Profile attributes (name, email) stay current via upsert on each login
- Local accounts remain available as a fallback

### Negative / Trade-offs
- SAML app registration in EntraID/ADFS requires IdP admin access — operational overhead for initial setup
- `crewjam/saml` adds a significant dependency; XML parsing is a known attack surface (mitigated by the library's signature enforcement)
- `?token=` in URL is a minor security imperfection; acceptable for v1
- ACS endpoint must be reachable by the IdP (requires public HTTPS URL or network path from IdP to SP)
- SAML metadata fetch on startup adds a network dependency; implement graceful degradation if URL unreachable (fall back to `idp_metadata_file`, warn loudly)

### Future upgrade path
- Implement SAML Single Logout (SLO) so logout in OpenVox View propagates to the IdP session
- Replace `?token=` handoff with HttpOnly short-lived cookie for improved security
- Add SAML group → role mapping when RBAC is introduced
- Add OIDC/OAuth2 as an alternative IdP protocol for non-ADFS cloud environments

---

## Implementation Checklist

**Backend**
- [ ] Add `github.com/crewjam/saml` to `go.mod`
- [ ] Add `SamlConfig` struct to `config/config.go` with Viper bindings
- [ ] Add `given_name` / `surname` columns to `users` table migration in `db/users.go`
- [ ] Add `UpsertSamlUser()` to `db/users.go`
- [ ] Create `middleware/saml.go` — `NewSamlServiceProvider()`, hourly metadata refresh goroutine
- [ ] Add `SamlMetadata`, `SamlLogin`, `SamlACS` to `handler/auth.go`
- [ ] Add `SamlEnabled` to `MetaResponse` in `handler/core.go` (or `handler/view.go`)
- [ ] Register SAML routes in `main.go` (public, conditional on `cfg.Auth.Saml.Enabled`)
- [ ] Add `--generate-saml-cert` CLI flag
- [ ] Update `CONFIGURATION.md` with SAML config options and IdP setup guides

**Frontend**
- [ ] Add `SamlEnabled` to `ApiMeta` in `ui/src/client/models.ts`
- [ ] Add SAML token extraction + SSO button to `ui/src/pages/LoginPage.vue`
- [ ] No other frontend changes needed (auth store, interceptors, and router guard from ADR-001 are protocol-agnostic)
