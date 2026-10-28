# test-codex-env

Monorepo of services deployed to CapRover. Each service lives in `services/<name>/` and has its own
`AGENTS.md` with stack-specific commands, conventions and review checks; read it before working on that service.

## Layout

- `services/<name>/` — one deployable service. `service.json` declares `stack` (`go` | `python` | `node`),
  container `port`, and optional `buildArgVars` (GitHub variables passed as Docker build args).
  `make ci` is what CI runs for that service.
- `template/` — templates for new services; `scripts/new-service.sh` copies one into `services/`.
- `.github/workflows/` — shared pipelines for every service (see below). `.github/scripts/changed-services.sh`
  decides which services a change touches.

## Branches and deploy

- Work on feature branches, open PRs into `develop`. CI runs only for services changed in the PR.
- Comment `/deploy` on a PR to deploy its changed services (or `/deploy <svc> ...`) to their develop environment.
- Merge to `develop` deploys the changed services to `<svc>-develop`.
  Merge to `main` deploys changed services to `<svc>-production`, only for services listed in the repo
  variable `PRODUCTION_SERVICES`; others are skipped.
- Workflows: `ci.yml` (PR checks), `deploy.yml` (push), `pr-deploy.yml` (`/deploy`), `prune-ghcr.yml` (weekly),
  reusable `_ci.yml` and `_build-deploy.yml`.
- Runtime config comes from env vars set in CapRover App Configs. Never commit `.env`, `captain.json` or secrets.

## Review guidelines

- Write review comments in Vietnamese; keep code identifiers and error messages as-is.
- Only flag issues worth fixing before merge. No style nits that the linter/formatter already covers.
- Treat as P0: leaked secrets/tokens/credentials, auth or permission bypass, injection (SQL, command, path traversal), data loss.
- Treat as P1:
  - Unhandled errors or swallowed exceptions on I/O, DB and HTTP calls.
  - Missing input validation on request handlers.
  - Behavior change without a matching test.
  - Changes to a service's `Dockerfile`, `service.json` or `.github/workflows/*` that could break build or deploy
    (wrong port, missing build arg, widened `permissions`, untrusted input interpolated into `run:`).
  - A change in one service that silently depends on another service's code instead of its API.
  - New dependencies that are unmaintained, duplicate an existing one, or are pulled in for something trivial.
- Logs must not contain secrets or personal data.
- Note when a PR adds an env var but not to the service's `.env.example` / README.
