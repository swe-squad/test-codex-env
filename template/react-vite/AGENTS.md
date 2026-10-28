# __SERVICE_NAME__

React + TypeScript SPA built with Vite, served by nginx (`nginx.conf`, SPA fallback, `/health`).

## Commands

- Install: `npm ci`
- Dev: `npm run dev`
- Test: `npm test` (vitest + Testing Library)
- Lint: `npm run lint` (oxlint)
- Build: `npm run build`

## Conventions

- Config is baked in at build time through `VITE_*` env vars, read only in `src/config.ts`.
  New `VITE_*` vars must also be added to the `Dockerfile` ARGs, `service.json` `buildArgVars`,
  and as GitHub environment variables.
- Function components and hooks only; no class components.
- Tests query by role/label (`getByRole`), not by CSS class or test id when avoidable.

## Review guidelines (React)

- Secrets in `VITE_*` vars or source: everything there ships to the browser.
- `useEffect` with missing/incorrect dependencies, fetches without cleanup/abort.
- `dangerouslySetInnerHTML` or unescaped user content.
- Large new dependencies that noticeably grow the bundle.
