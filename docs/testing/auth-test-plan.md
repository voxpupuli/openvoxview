# Authentication Test Plan (ADR-001)

Manual test plan for local user authentication.

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
- [x] **Login as newly created user** — should work

## Auth Disabled Mode

- [x] **Set `auth.enabled: false`**, restart — all pages accessible without login, no user menu shown, no login redirect

## CLI

- [x] **`--create-admin`** — creates user interactively, confirm it shows in `GET /api/v1/auth/users`
- [x] **`--create-admin` with mismatched passwords** — should abort with error

## UI Elements

- [x] **User menu visible** in top-right toolbar when logged in (account icon)
- [x] **User menu shows username/display name and email**
- [x] **All existing pages still work** after login: Dashboard, Nodes, Facts, Reports, Query, CA (if enabled), predefined views

## User Management UI (ADR-003)

### Visibility
- [x] **"Users" menu item visible** in sidebar when auth is enabled
- [ ] **"Users" menu item hidden** when auth is disabled

### User Table
- [ ] **Users page loads** — shows table with all users, columns: Username, Display Name, Email, Auth Source, Created, Actions
- [ ] **Refresh button** — reloads the user list

### Create User
- [x] **Click "Add User"** — opens create dialog with empty fields
- [x] **Submit with valid data** — user created, success notification, table refreshes
- [x] **Submit with missing username** — shows "Username is required" error
- [x] **Submit with short password** (< 8 chars) — shows "Password must be at least 8 characters"
- [x] **Submit with mismatched passwords** — shows "Passwords do not match"
- [x] **Submit with duplicate username** — shows "Username already exists"
- [x] **Cancel button** — closes dialog without creating

### Edit User
- [x] **Click edit icon on a user row** — opens dialog pre-populated with user data
- [x] **Username field is read-only** in edit mode
- [ ] **Update display name / email** — saves, success notification, table refreshes
- [ ] **Change password** — confirm password field appears, saves new password
- [x] **Leave password blank** — keeps existing password unchanged
- [x] **Cancel button** — closes dialog without saving

### Delete User
- [ ] **Click delete icon** — shows confirmation dialog with username
- [ ] **Confirm delete** — user removed, success notification, table refreshes
- [ ] **Cancel delete** — dialog closes, user not deleted
- [ ] **Delete button disabled on own row** — tooltip shows "Cannot delete your own account"

### i18n
- [ ] **German locale** — all labels, buttons, and messages display in German
