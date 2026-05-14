# Security Analysis: OpenVox View

**Date:** 2026-04-04
**Last Updated:** 2026-04-04
**Status:** All findings resolved (5 fixed, 1 accepted risk)

## Vuln 1: Wildcard CORS Allows Cross-Origin Authenticated Requests — `main.go:240-243`

* **Status: FIXED**
* **Severity: HIGH**
* **Fix:** CORS is now configurable via `cors_origin` config. Defaults to disabled (no CORS headers). Allowed headers restricted to `Authorization, Content-Type`.

---

## Vuln 2: SAML Authentication Passes Tokens in Redirect URL — `handler/auth.go:416`

* **Status: ACCEPTED RISK (Low)**
* **Original Severity: HIGH** → **Revised: LOW**
* **Rationale:** Tokens are placed in the URL hash fragment (`/ui/?#/login?token=...&refresh=...`), not in query parameters. Fragments have key properties that mitigate the originally reported risks:
  - Server access logs: **Not affected** — browsers never send fragments to servers
  - Proxy logs: **Not affected** — proxies do not see the fragment
  - Referer header: **Not affected** — fragments are stripped from Referer
  - Browser history: **Affected** — but requires physical access to the user's machine
* The frontend extracts and clears the tokens from the URL immediately. The remaining risk (local browser history) is acceptable for most deployments.

---

## Vuln 3: No Authorization on User Management — `main.go:213-220`

* **Status: FIXED**
* **Severity: HIGH**
* **Fix:** Added `AdminRequiredMiddleware()`, `is_admin` field in DB and JWT claims. User management endpoints now gated behind admin check. `--create-admin` CLI sets admin flag. Self-demote guard prevents admins from removing their own admin role.

---

## Vuln 4: IDP-Initiated SAML Enabled — `middleware/saml.go:81`

* **Status: FIXED**
* **Severity: MEDIUM**
* **Fix:** `AllowIDPInitiated` set to `false`. SAML assertions now require matching `InResponseTo` request ID.

---

## Vuln 5: SAML Cookie Missing Secure Flag — `handler/auth.go:338`

* **Status: FIXED**
* **Severity: MEDIUM**
* **Fix:** Cookie `Secure` flag now set to `true` on both set and clear operations.

---

## Vuln 6: Password Minimum Length Not Enforced on Update — `handler/auth.go:219-222`

* **Status: FIXED**
* **Severity: MEDIUM**
* **Fix:** `UpdateUser` handler now validates `len(*req.Password) >= 8` when password is provided.

---

## Summary

| # | Finding | Severity | Status |
|---|---------|----------|--------|
| 1 | Wildcard CORS with auth headers | HIGH | FIXED |
| 2 | SAML tokens in redirect URL fragment | LOW | ACCEPTED RISK |
| 3 | No authorization on user management | HIGH | FIXED |
| 4 | IDP-initiated SAML allows assertion replay | MEDIUM | FIXED |
| 5 | SAML cookie missing Secure flag | MEDIUM | FIXED |
| 6 | Password min length not enforced on update | MEDIUM | FIXED |
