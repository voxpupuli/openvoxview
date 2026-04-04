# ADR-003: User Management UI

**Status:** Implemented  
**Date:** 2026-04-03  
**Deciders:** OpenVox View maintainers  
**Depends on:** ADR-001 (backend user CRUD API, auth store, auth-enabled detection)

---

## Context

ADR-001 introduced local user authentication with a full backend CRUD API for user management (`GET/POST/PUT/DELETE /api/v1/auth/users`). The frontend client methods (`Backend.getUsers()`, `createUser()`, `updateUser()`, `deleteUser()`) and TypeScript models (`UserProfile`, `CreateUserRequest`, `UpdateUserRequest`) also already exist.

Currently, user management is only possible via CLI (`--create-admin`) or direct API calls with curl. Administrators need a UI to manage users without leaving the browser.

---

## Decision

Add a **User Management page** at route `/users` within the existing `MainLayout`, following the same patterns used by the CA Overview page (q-table, q-dialog for confirmations, Notify for feedback). The page is only visible in the sidebar when `auth.enabled` is `true`.

No backend changes are required — all API endpoints already exist.

---

## UI Design

### Page: `UserManagementPage.vue`

A single page with:

1. **Users table** (q-table) showing all users
2. **"Add User" button** in the table header — opens a create dialog
3. **Row actions**: Edit and Delete buttons per row
4. **Self-delete protection**: Delete button disabled on the current user's row

### Table columns

| Column | Field | Sortable | Notes |
|---|---|---|---|
| Username | `username` | Yes | Primary identifier |
| Display Name | `display_name` | Yes | |
| Email | `email` | Yes | |
| Auth Source | `auth_source` | Yes | "local" or "saml" (future ADR-002) |
| Created | `created_at` | Yes | Formatted date |
| Actions | — | No | Edit / Delete buttons |

### Create User Dialog

Quasar `q-dialog` with a form:

| Field | Type | Validation |
|---|---|---|
| Username | q-input, text | Required |
| Email | q-input, email | Optional |
| Display Name | q-input, text | Optional |
| Password | q-input, password | Required, min 8 chars |
| Confirm Password | q-input, password | Must match Password |

On submit: `Backend.createUser()` → success notification → reload table.  
On 409: show "Username already exists" error.

### Edit User Dialog

Same q-dialog, pre-populated with existing values. Username is read-only (not editable after creation). Password fields are optional — leave blank to keep current password.

On submit: `Backend.updateUser(id, data)` → success notification → reload table.

### Delete Confirmation

Quasar `q.dialog()` confirm pattern (same as CA page uses for sign/revoke/clean):

```
"Delete user <username>? This action cannot be undone."
[Cancel] [Delete (red)]
```

On confirm: `Backend.deleteUser(id)` → success notification → reload table.  
Self-delete: button is disabled with a tooltip "Cannot delete your own account". Backend also enforces this (403).

---

## Architecture Changes

### Frontend only — no backend changes

#### New: `ui/src/pages/admin/UserManagementPage.vue`

Single-file Vue component containing:
- `q-table` with columns definition, pagination, and row template
- Three functions: `loadUsers()`, `openCreateDialog()`, `openEditDialog(user)`, `confirmDelete(user)`
- Create/Edit dialog as inline `q-dialog` with `v-model` toggle (same pattern as other pages)
- Current user ID from `useAuthStore()` to disable self-delete

#### Update: `ui/src/router/routes.ts`

Add route under the existing MainLayout children:

```typescript
{
  name: 'UserManagement',
  path: 'users',
  component: () => import('pages/admin/UserManagementPage.vue'),
}
```

#### Update: `ui/src/layouts/MainLayout.vue`

Add sidebar menu item, conditionally shown when `auth.authEnabled` is true:

```vue
<q-item clickable :to="{ name: 'UserManagement' }" v-if="auth.authEnabled">
  <q-item-section avatar>
    <q-icon name="manage_accounts" />
  </q-item-section>
  <q-item-section>
    <q-item-label>Users</q-item-label>
  </q-item-section>
</q-item>
```

Placed after the CA menu item in the sidebar.

---

## What Already Exists (no changes needed)

| Layer | Component | Status |
|---|---|---|
| Backend API | `GET/POST/PUT/DELETE /api/v1/auth/users` | Done (ADR-001) |
| Frontend client | `Backend.getUsers()`, `createUser()`, `updateUser()`, `deleteUser()` | Done (ADR-001) |
| TypeScript models | `UserProfile`, `CreateUserRequest`, `UpdateUserRequest` | Done (ADR-001) |
| Auth store | `useAuthStore()` with user ID for self-delete check | Done (ADR-001) |
| Auth middleware | JWT validation on all user management endpoints | Done (ADR-001) |

---

## Implementation Checklist

- [x] Create `ui/src/pages/admin/UserManagementPage.vue`
- [x] Add `UserManagement` route to `ui/src/router/routes.ts`
- [x] Add "Users" menu item to `ui/src/layouts/MainLayout.vue` sidebar (conditional on `auth.authEnabled`)
- [x] Build and verify (`yarn build`, `yarn lint`)
- [x] Add test cases to `docs/testing/auth-test-plan.md`
- [x] Add i18n keys for all labels, buttons, and messages (en-US, de-DE)
