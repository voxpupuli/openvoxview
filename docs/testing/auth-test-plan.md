# Authentication Test Plan (ADR-001, ADR-002, ADR-003, ADR-004)

Manual test plan for authentication, user management, and security hardening.

## Authentication Flow

- [x] **Login with valid credentials** — should redirect to Dashboard
- [x] **Login with wrong password** — should show "Invalid username or password"
- [x] **Login with non-existent user** — should show "Invalid username or password" (same message, no user enumeration)
- [x] **Login with empty fields** — should show validation error
- [x] **Rate limiting** — try 6 rapid failed logins from the same IP, 6th should return "Too many login attempts"

## Session Persistence

- [x] **Page reload after login** — should stay authenticated (not redirected to login)
- [x] **Close browser tab and reopen** — should still be logged in (refresh token in localStorage)
- [x] **Wait 15+ minutes** (or set `access_token_ttl_minutes: 1` for testing) — next API call should silently refresh the token without redirecting to login

## Logout

- [x] **Click user icon → Logout** — should redirect to login page
- [x] **After logout, press browser back button** — should not access protected content (redirected to login)
- [x] **After logout, reuse old access token via curl** — should get 401

## Token Security

- [x] **Access API without token**: `curl http://<host>/api/v1/view/node_overview` — should return 401
- [x] **Access API with valid token**: `curl -H "Authorization: Bearer <token>" http://<host>/api/v1/view/node_overview` — should return data
- [x] **Access API with garbage token**: `curl -H "Authorization: Bearer abc123" ...` — should return 401
- [x] **Public endpoints work without token**: `curl http://<host>/api/v1/version` and `curl http://<host>/api/v1/meta` — should return data

## User Management API

- [x] **List users**: `GET /api/v1/auth/users` — should return the admin user
- [x] **Create user**: `POST /api/v1/auth/users` with `{"username":"testuser","password":"test12345678","email":"test@example.com"}` — should return 201
- [x] **Create duplicate username** — should return 409 Conflict
- [x] **Create user with short password** (< 8 chars) — should return 400
- [x] **Update user**: `PUT /api/v1/auth/users/2` with `{"display_name":"Test User"}` — should update
- [x] **Delete user**: `DELETE /api/v1/auth/users/2` — should delete
- [x] **Self-delete guard**: try deleting your own user ID — should return 403 "cannot delete your own account"
- [x] **Admin-only guard**: as non-admin user, `GET /api/v1/auth/users` — should return 403 "admin access required"
- [x] **Admin-only guard**: as non-admin user, `POST /api/v1/auth/users` — should return 403
- [x] **Admin-only guard**: as non-admin user, `PUT /api/v1/auth/users/:id` — should return 403
- [x] **Admin-only guard**: as non-admin user, `DELETE /api/v1/auth/users/:id` — should return 403
- [x] **Create user with is_admin**: `POST /api/v1/auth/users` with `{"username":"adminuser","password":"test12345678","is_admin":true}` — should create admin user
- [x] **Update user is_admin**: `PUT /api/v1/auth/users/:id` with `{"is_admin":true}` — should promote user to admin
- [x] **Self-demote guard**: try setting your own `is_admin` to `false` — should return 403 "cannot remove your own admin role"
- [x] **Password min length on update**: `PUT /api/v1/auth/users/:id` with `{"password":"short"}` — should return 400
- [x] **Login as newly created user** — should work

## Auth Disabled Mode

- [x] **Set `auth.enabled: false`**, restart — all pages accessible without login, no user menu shown, no login redirect

## CLI

- [x] **`--create-admin`** — creates user interactively, confirm it shows in `GET /api/v1/auth/users`
- [x] **`--create-admin` sets is_admin** — created user should have `is_admin: true`
- [x] **`--create-admin` with mismatched passwords** — should abort with error

## UI Elements

- [x] **User menu visible** in top-right toolbar when logged in (account icon)
- [x] **User menu shows username/display name and email**
- [x] **All existing pages still work** after login: Dashboard, Nodes, Facts, Reports, Query, CA (if enabled), predefined views

## User Management UI (ADR-003)

### Visibility
- [x] **"Users" menu item visible** in sidebar only when auth is enabled **and user is admin**
- [x] **"Users" menu item hidden** for non-admin users
- [x] **"Users" menu item hidden** when auth is disabled

### User Table
- [x] **Users page loads** — shows table with all users, columns: Username, Display Name, Email, Auth Source, Admin, Created, Actions
- [x] **Admin column** — shows check icon for admin users, empty for non-admins
- [x] **Refresh button** — reloads the user list

### Create User
- [x] **Click "Add User"** — opens create dialog with empty fields
- [x] **Admin toggle** — create dialog has admin toggle, defaults to off
- [x] **Submit with valid data** — user created, success notification, table refreshes
- [x] **Submit with admin toggle on** — user created with `is_admin: true`
- [x] **Submit with missing username** — shows "Username is required" error
- [x] **Submit with short password** (< 8 chars) — shows "Password must be at least 8 characters"
- [x] **Submit with mismatched passwords** — shows "Passwords do not match"
- [x] **Submit with duplicate username** — shows "Username already exists"
- [x] **Cancel button** — closes dialog without creating

### Edit User
- [x] **Click edit icon on a user row** — opens dialog pre-populated with user data
- [x] **Username field is read-only** in edit mode
- [x] **Admin toggle pre-populated** — reflects current is_admin state
- [x] **Admin toggle on own account disabled** — shows warning banner "Cannot remove your own admin role"
- [x] **Change admin toggle on other user** — saves, user promoted/demoted
- [x] **Update display name / email** — saves, success notification, table refreshes
- [x] **Change password** — confirm password field appears, saves new password
- [x] **Leave password blank** — keeps existing password unchanged
- [x] **Cancel button** — closes dialog without saving

### Delete User
- [x] **Click delete icon** — shows confirmation dialog with username
- [x] **Confirm delete** — user removed, success notification, table refreshes
- [x] **Cancel delete** — dialog closes, user not deleted
- [x] **Delete button disabled on own row** — tooltip shows "Cannot delete your own account"

### i18n
- [x] **German locale** — all labels, buttons, and messages display in German

## SAML Authentication (ADR-002)

### Configuration
- [x] **App Federation Metadata URL** — must use the app-specific URL with `?appid=` parameter
- [x] **Missing SP cert files** — server should fail to start with clear error
- [x] **Missing IdP metadata URL and file** — server should fail to start with clear error
- [x] **Invalid IdP metadata URL** — server should fail to start with clear error

### SSO Login Flow
- [x] **SSO button visible** on login page when `auth.saml.enabled: true`
- [x] **SSO button hidden** when `auth.saml.enabled: false`
- [x] **Click "Login with SSO"** — redirects to EntraID login
- [x] **Authenticate at EntraID** — redirected back to OpenVox View, logged in, lands on Dashboard
- [x] **SAML user auto-provisioned** — user appears in `GET /api/v1/auth/users` with `auth_source: saml`
- [x] **SAML user profile attributes** — email, display name, given name, surname populated from IdP claims
- [x] **Repeat SAML login** — profile attributes updated (upsert), no duplicate user created
- [x] **SAML user cannot local login** — attempting local login with SAML user's email should fail

### Session Behavior
- [x] **Token refresh works for SAML users** — after access token expires, silent refresh works
- [x] **Logout works for SAML users** — click Logout, redirected to login page
- [x] **After SAML logout, press back button** — should not access protected content

### Coexistence with Local Auth
- [x] **Local login still works** when SAML is enabled — both buttons on login page
- [x] **Local user and SAML user can coexist** — different auth_source values
- [x] **Break-glass admin** — local admin account works even if IdP is unreachable

### SP Metadata
- [x] **`GET /api/v1/auth/saml/metadata`** — returns valid XML with SP entity ID, ACS URL, and certificate

### CLI
- [x] **`--generate-saml-cert`** — creates `saml-sp.crt` and `saml-sp.key` in current directory

### User Management UI
- [x] **SAML users visible** in Users table with auth source "saml"
- [x] **SAML user edit dialog** — shows info banner "managed by identity provider"
- [x] **SAML user email/display name disabled** — fields are greyed out, not editable
- [x] **SAML user password fields hidden** — no password fields shown for SAML users
- [x] **SAML user admin toggle editable** — admin toggle works for SAML users
- [x] **SAML user save button visible** — save button shown to allow admin toggle changes
- [x] **SAML users deletable** — can delete a SAML-provisioned user
