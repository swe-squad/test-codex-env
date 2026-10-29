#!/bin/sh
# Run by the nginx image entrypoint before nginx.conf is rendered: refuse to start without the API upstream,
# so a missing App Config fails the deploy instead of every /api/ request
set -eu
if [ -z "${PING_API_URL:-}" ]; then
  echo "PING_API_URL is required, e.g. http://srv-captain--ping-api-dev:8080" >&2
  exit 1
fi
