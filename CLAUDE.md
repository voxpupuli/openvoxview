# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What This Project Is

OpenVox View is a web-based dashboard for PuppetDB/OpenVoxDB infrastructure monitoring, inspired by Puppetboard. It consists of a Go backend API server and a Vue 3 + Quasar frontend.

## Build & Development Commands

### Full Build
```bash
make all        # Build everything (frontend first, then backend binary)
make ui         # Build frontend only (yarn install + yarn build)
make backend    # Build Go binary (requires frontend to be built first)
```

### Development (Hot Reload)
```bash
make develop-backend          # Go backend with hot reload via `air`
make develop-frontend         # Frontend dev server at http://localhost:9000 (proxies to backend at :5000)
make develop-backend-crafty   # Backend with Puppet TLS settings (PUPPETDB_TLS_IGNORE=true, port 8081)
```

### Frontend (from `ui/` directory)
```bash
yarn lint     # ESLint
yarn format   # Prettier
yarn test     # (stubbed, no tests currently)
```

### Go Backend
```bash
go build ./...   # Build
go vet ./...     # Vet
```

## Architecture

### Data Flow
Frontend → Go API (`/api/v1/*`) → PuppetDB or Puppet CA

The backend embeds the built frontend assets via `uiserve.go` so the whole app ships as a single binary.

### Backend (Go + Gin)

- **[main.go](main.go)** — Gin router setup, CORS config, route registration for all `/api/v1/*` endpoints
- **[uiserve.go](uiserve.go)** — Serves embedded Vue build assets for non-API routes
- **[config/config.go](config/config.go)** — Viper-based config loading from `config.yaml` or environment variables (see CONFIGURATION.md)
- **[handler/](handler/)** — HTTP handlers grouped by domain:
  - `view.go` — Node overview, metrics, predefined views, fact rendering
  - `pdb.go` — PuppetDB query execution, query history (in-memory), predefined queries
  - `ca.go` — Puppet CA certificate management (sign/revoke/clean)
  - `core.go` — Shared response helpers
- **[model/](model/)** — Domain structs (nodes, facts, events, certificates, metrics)
- **[puppetdb/client.go](puppetdb/client.go)** — PuppetDB HTTP client (supports TLS + client certs)
- **[puppetca/client.go](puppetca/client.go)** — Puppet CA HTTP client

### Frontend (Vue 3 + Quasar + TypeScript, in `ui/src/`)

- **[ui/src/pages/](ui/src/pages/)** — Route-level page components
- **[ui/src/components/](ui/src/components/)** — Reusable Vue components
- **[ui/src/stores/](ui/src/stores/)** — Pinia state stores
- **[ui/src/client/](ui/src/client/)** — Axios-based API client for the Go backend
- **[ui/src/puppet/](ui/src/puppet/)** — Puppet-domain utilities
- **[ui/src/router/](ui/src/router/)** — Vue Router config
- **[ui/src/i18n/](ui/src/i18n/)** — i18n translations

### API Structure
All backend endpoints are under `/api/v1/`:
- `GET /meta` — App metadata (CA enabled flag, unreported hours threshold, etc.)
- `GET /version` — Version info
- `/view/*` — Dashboard views (nodes, metrics, facts)
- `/pdb/*` — PuppetDB query execution and history
- `/ca/*` — Certificate management

### Key Config Points
- Default backend port: **5000**
- Config file: `config.yaml` (see CONFIGURATION.md for all options)
- TLS is fully supported for both PuppetDB and Puppet CA connections
- Query history is stored in-memory in `PdbHandler`
