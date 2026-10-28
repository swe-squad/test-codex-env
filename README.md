# test-codex-env

Monorepo of services: Docker + GitHub Actions → GHCR → CapRover, one pipeline for every service.
Only services changed by a push or PR are tested and deployed.

| Template | Stack | Port | `make ci` |
|---|---|---|---|
| `go-gin` | Go 1.26, Gin, distroless image | 8080 | `go vet`, `go test -race`, build |
| `python-fastapi` | Python 3.14, FastAPI, uvicorn | 8000 | ruff, pytest |
| `react-vite` | React 19, TypeScript, Vite, vitest, nginx | 80 | oxlint, vitest, build |

Services today: `ping-api` (go-gin, `GET /api/v1/ping`), `ping-web` (react-vite, `GET /ping`).

## New service

With the `new-service` skill (Claude Code `/new-service go-gin payment-api`, Codex `$new-service ...`), or:

```bash
scripts/new-service.sh go-gin payment-api        # -> services/payment-api

# GitHub environment payment-api-develop + CapRover secrets
CAPROVER_SERVER=https://captain.example.com \
APP_NAME_DEVELOP=payment-api-dev CAPROVER_TOKEN_DEVELOP=... \
  scripts/setup-service.sh payment-api

git add services/payment-api && git commit -m "feat: add payment-api" && git push origin develop
```

End-to-end walkthrough (CapRover → skill → GitHub → deploy → checks): [docs/deploy-flow.md](docs/deploy-flow.md).

## Layout

```
services/<name>/            # one service: code, Dockerfile, Makefile (ci target), service.json, AGENTS.md
  service.json              # {"stack": "go|python|node", "port": 8080, "buildArgVars": ["VITE_API_BASE_URL"]}
template/                   # templates; _shared is copied into every new service first
.github/
  scripts/changed-services.sh   # which services/<name> changed between two commits
  workflows/
    ci.yml                  # PR -> make ci for changed services
    deploy.yml              # push develop/main -> deploy changed services
    pr-deploy.yml           # `/deploy [svc...]` comment on a PR -> develop
    prune-ghcr.yml          # weekly cleanup of every service image
    _ci.yml                 # reusable: set up toolchain from service.json, run make ci
    _build-deploy.yml       # reusable: CI -> build -> ghcr.io/<owner>/<svc> -> CapRover
scripts/
  new-service.sh            # scaffold services/<name>
  setup-service.sh          # GitHub environments + secrets for one service
.agents/skills/new-service/ # skill (Codex); .claude/skills symlinks here for Claude Code
AGENTS.md                   # repo rules + Codex review guidelines; each service adds its own AGENTS.md
```

## Deploy flow

| Trigger | Services | Environment | Image tags |
|---|---|---|---|
| Push / merge into `develop` | changed ones | `<svc>-develop` | `<sha7>`, `latest-develop` |
| Push / merge into `main` | changed ones listed in repo variable `PRODUCTION_SERVICES` | `<svc>-production` | `<sha7>`, `latest-main` |
| Comment `/deploy` on a PR | changed in the PR (or `/deploy svc1 svc2`) | `<svc>-develop` | `<sha7>`, `pr-<n>` |
| Actions → *Deploy* → Run workflow | input list (empty = all) | `main` → production, else develop | `<sha7>`, `latest-<branch>` |

- Each environment holds `CAPROVER_SERVER`, `APP_NAME`, `CAPROVER_TOKEN` secrets; variables named in
  `buildArgVars` (e.g. `VITE_API_BASE_URL`) are passed as Docker build args.
- Production is opt-in per service: `setup-service.sh` creates `<svc>-production` and adds the service to
  `PRODUCTION_SERVICES` only when `APP_NAME_PRODUCTION` and `CAPROVER_TOKEN_PRODUCTION` are given.
- Every deploy runs `make ci` first; deploys of the same service + environment are queued, never parallel.
- `/deploy` only works for owners/members/collaborators and not for PRs from forks.
- In CapRover set *Container HTTP Port* to the service's `port`.

## AI code review

Root `AGENTS.md` has `## Review guidelines`; each service's `AGENTS.md` adds stack-specific ones. Codex code
review follows both. Turn it on once: Codex settings → Code review → enable `swe-squad/test-codex-env`.
