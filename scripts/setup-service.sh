#!/usr/bin/env bash
# Create the GitHub environments + CapRover secrets for one service of this monorepo.
#
#   scripts/setup-service.sh <service-name>
#
# Values are read from env vars; any that are unset are skipped (set them later in Settings → Environments):
#   CAPROVER_SERVER                         e.g. https://captain.example.com
#   APP_NAME_DEVELOP      CAPROVER_TOKEN_DEVELOP
#   APP_NAME_PRODUCTION   CAPROVER_TOKEN_PRODUCTION   (optional: without both, only develop is set up)
#   <VAR>_DEVELOP  <VAR>_PRODUCTION         for each name in service.json buildArgVars, e.g. VITE_API_BASE_URL_DEVELOP
#
# Creates environment <service>-develop (and <service>-production when its values are given, adding the
# service to the repo variable PRODUCTION_SERVICES so pushes to main deploy it). Safe to re-run.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
[[ $# -eq 1 ]] || { sed -n '2,13p' "$0" >&2; exit 1; }
SERVICE="$1"
[[ -f "${ROOT}/services/${SERVICE}/service.json" ]] || { echo "services/${SERVICE}/service.json not found" >&2; exit 1; }

REPO="$(git -C "${ROOT}" remote get-url origin | sed -E 's#^(git@github\.com:|https://github\.com/)##; s#\.git$##')"
gh auth status >/dev/null

ENVS=(develop)
if [[ -n "${APP_NAME_PRODUCTION:-}" && -n "${CAPROVER_TOKEN_PRODUCTION:-}" ]]; then
  ENVS+=(production)
else
  echo "Production values not set: setting up develop only"
fi

mapfile -t BUILD_ARG_VARS < <(python3 -c 'import json,sys; print("\n".join(json.load(open(sys.argv[1])).get("buildArgVars", [])))' "${ROOT}/services/${SERVICE}/service.json")

for ENV in "${ENVS[@]}"; do
  SUFFIX="${ENV^^}"
  GH_ENV="${SERVICE}-${ENV}"
  gh api --silent -X PUT "repos/${REPO}/environments/${GH_ENV}"
  echo "Environment ${GH_ENV}:"
  for pair in "CAPROVER_SERVER:CAPROVER_SERVER" "CAPROVER_TOKEN:CAPROVER_TOKEN_${SUFFIX}" "APP_NAME:APP_NAME_${SUFFIX}"; do
    secret="${pair%%:*}"; var="${pair#*:}"
    if [[ -n "${!var:-}" ]]; then
      gh secret set "${secret}" --repo "${REPO}" --env "${GH_ENV}" --body "${!var}"
    else
      echo "  - secret ${secret} skipped (\$${var} not set)"
    fi
  done
  for name in "${BUILD_ARG_VARS[@]}"; do
    [[ -n "${name}" ]] || continue
    var="${name}_${SUFFIX}"
    if [[ -n "${!var:-}" ]]; then
      gh variable set "${name}" --repo "${REPO}" --env "${GH_ENV}" --body "${!var}"
    else
      echo "  - variable ${name} skipped (\$${var} not set)"
    fi
  done
done

if [[ " ${ENVS[*]} " == *" production "* ]]; then
  current="$(gh variable get PRODUCTION_SERVICES --repo "${REPO}" 2>/dev/null || true)"
  if [[ ",${current// /}," != *",${SERVICE},"* ]]; then
    gh variable set PRODUCTION_SERVICES --repo "${REPO}" --body "${current:+${current},}${SERVICE}"
    echo "Added ${SERVICE} to PRODUCTION_SERVICES"
  fi
fi

echo "Done. Push to develop (or comment /deploy on a PR) to deploy ${SERVICE}."
