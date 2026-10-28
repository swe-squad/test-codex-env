---
name: new-service
description: Add a new service (go-gin, python-fastapi, react-vite) to this monorepo under services/<name> from the repo's templates; CI/CD to GHCR and CapRover is shared and picks it up automatically. Optionally sets up its GitHub environment + CapRover secrets. Use when the user asks to create/scaffold/add a new service, project, app, API or frontend, e.g. "new-service go-gin payment-api" or "tạo service react-vite admin-web".
---

# New service

This repo is a monorepo: every service lives in `services/<name>/` and shares the workflows in
`.github/workflows/`. Adding a service = adding its directory; no workflow changes needed.
Run scripts from the repo root.

## Inputs

| Input | Required | Notes |
|---|---|---|
| template | yes | A directory under `template/` not starting with `_`. Today: `go-gin`, `python-fastapi`, `react-vite`. Map loose wording: "go"/"gin" → `go-gin`, "python"/"fastapi"/"py" → `python-fastapi`, "react"/"vite"/"frontend"/"fe" → `react-vite`. |
| service name | yes | lowercase, digits, `-`; 3–50 chars. Becomes the directory, GHCR image `ghcr.io/<owner>/<name>`, GitHub environment prefix, CapRover app suggestion. Normalize obvious cases (`Payment API` → `payment-api`) and tell the user. |

If template or name is missing, ask once for both. Do not ask about anything else up front.

## Steps

1. **Scaffold**
   ```bash
   scripts/new-service.sh <template> <name>
   ```
   Creates `services/<name>/` (refuses if it exists). Owner/repo come from the `origin` remote.

2. **Verify locally**: `make ci` inside `services/<name>` — the same command CI runs.
   Skip only if the toolchain is missing, and say so. Remove build output afterwards (`bin/`, `dist/`, `.venv/`).

3. **GitHub environment + secrets** — writes secrets to the GitHub repo, so **ask the user first**.
   The user should run it themselves so tokens never pass through the chat:
   ```bash
   CAPROVER_SERVER=... APP_NAME_DEVELOP=<name>-dev CAPROVER_TOKEN_DEVELOP=... \
     scripts/setup-service.sh <name>
   ```
   - Creates environment `<name>-develop`. With `APP_NAME_PRODUCTION` + `CAPROVER_TOKEN_PRODUCTION` it also
     creates `<name>-production` and adds the service to repo variable `PRODUCTION_SERVICES`.
   - For every name in `service.json` `buildArgVars` (react-vite: `VITE_API_BASE_URL`), pass `<VAR>_DEVELOP`
     / `<VAR>_PRODUCTION`; they become environment variables used as Docker build args.
   - Never echo tokens back or write them to files.

4. **Commit and deploy**: commit `services/<name>` (ask before pushing). Pushing to `develop`, or a PR
   merged into it, deploys every changed service; `/deploy` on a PR deploys the PR's services.

5. **Report**: path, template, image, port, what was verified, which secrets are still missing, and the
   CapRover checklist below.

## CapRover checklist (manual, per environment)

- Create the app (`<name>-dev`, and `<name>` for production if wanted) and enable **App Token** under Deployment.
- Set **Container HTTP Port** to `port` in `service.json` (go-gin `8080`, python-fastapi `8000`, react-vite `80`).
- Enable HTTPS for the app's subdomain.
- The GHCR registry in CapRover → Cluster must be able to read `ghcr.io/<owner>/*` (PAT with `read:packages`).
- Runtime env vars (DB URLs, etc.) go in CapRover App Configs, not in the image.

## How the pipeline works (for answering follow-up questions)

- `.github/scripts/changed-services.sh <base> <head>`: JSON list of `services/<x>` (with `service.json`)
  changed in `base...head`; empty/zero/unknown base → all services.
- `ci.yml`: on PRs, `make ci` per changed service (matrix) via reusable `_ci.yml`, which sets up the
  toolchain from `service.json` `stack`.
- `_build-deploy.yml` (reusable): `_ci.yml` → build `services/<svc>` → push `ghcr.io/<owner>/<svc>:<sha7>`
  + moving tag → CapRover using secrets of environment `<svc>-<develop|production>`. Queued per service+env.
- `deploy.yml`: push to `develop` → changed services to `<svc>-develop`; push to `main` → changed services
  that are in `PRODUCTION_SERVICES` to `<svc>-production`. Manual run takes a services list (empty = all).
- `pr-deploy.yml`: `/deploy` or `/deploy svc1 svc2` on a PR (owner/member/collaborator, not forks) →
  PR head to `<svc>-develop` (tag `pr-<n>`), reacts 🚀 and comments the result.
- `prune-ghcr.yml`: weekly, per service image keeps the 10 newest versions plus whatever `latest-*` points to.
- Codex review: root `AGENTS.md` review guidelines + the service's own `AGENTS.md`; enabled once for the repo.

## Adding a new template

Create `template/<name>/` with the app, `Dockerfile`, `.dockerignore`, `.gitignore`, `Makefile` with a `ci`
target, `service.json` (`stack` must be one `_ci.yml` knows — add a setup step there for a new stack — plus
`port`, optional `buildArgVars`) and `AGENTS.md` (stack notes + `## Review guidelines (<stack>)`).
Use `__SERVICE_NAME__` / `__GITHUB_OWNER__` placeholders; Go module paths `github.com/__GITHUB_OWNER__/__SERVICE_NAME__`
are rewritten to `github.com/<owner>/<repo>/services/<name>`. Update the template list here and in `README.md`.
