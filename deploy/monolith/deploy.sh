#!/usr/bin/env bash
# Deploys the Inkomi monolith with Docker Compose on the Oracle server. Run over SSH by
# .github/workflows/deploy-monolith.yml after the pipeline pushed the images to the registry.
#
#   deploy.sh --tag <tag> --registry ghcr.io/<owner> [--run <id>]
#   env: GHCR_USER / GHCR_TOKEN (optional) to log in to the registry
#
# The stack lives in ~/inkomi/back/monolith (MONOLITH_DIR): docker-compose.yml (copied by the
# pipeline), .env and firebase.json (prepared by hand, see deploy/SERVERS.md) and data/.
#
# Steps:
#   1. Uploads the manga-worker preview samples to R2 when they changed.
#   2. Sandbox: starts the new image as a throwaway compose project (its own data dir and network)
#      and runs the e2e tests against it. It is removed afterwards, whatever the result.
#   3. Production: starts the new image and waits for its healthcheck; on failure it goes back to
#      the previous tag.
set -euo pipefail

MONOLITH_DIR="${MONOLITH_DIR:-$HOME/inkomi/back/monolith}"
PROJECT="inkomi-monolith"
HEALTH_TIMEOUT="${HEALTH_TIMEOUT:-180}"

TAG=""
REGISTRY=""
RUN_ID="$(date +%s)"

while [[ $# -gt 0 ]]; do
  case "$1" in
    --tag) TAG="$2"; shift 2 ;;
    --registry) REGISTRY="$2"; shift 2 ;;
    --run) RUN_ID="$2"; shift 2 ;;
    *) echo "unknown argument $1" >&2; exit 2 ;;
  esac
done
[[ -n "$TAG" && -n "$REGISTRY" ]] || { echo "--tag and --registry are required" >&2; exit 2; }

log() { printf '\n==> %s\n' "$*"; }
warn() { printf 'WARNING: %s\n' "$*" >&2; }

# docker runs as the deploy user when it is in the docker group, otherwise through sudo -n.
if docker info >/dev/null 2>&1; then
  DOCKER=(docker)
else
  DOCKER=(sudo -n docker)
fi

compose() { "${DOCKER[@]}" compose --project-directory "$MONOLITH_DIR" -f "$MONOLITH_DIR/docker-compose.yml" "$@"; }

# ─── checks ──────────────────────────────────────────────────────────────────────────────

check_host() {
  "${DOCKER[@]}" info >/dev/null || { echo "docker is not reachable (add the deploy user to the docker group)" >&2; exit 1; }
  [[ -f "$MONOLITH_DIR/.env" ]] || { echo "$MONOLITH_DIR/.env not found (see back/monolith/.env.example)" >&2; exit 1; }
  if [[ ! -f "$MONOLITH_DIR/firebase.json" ]]; then
    warn "$MONOLITH_DIR/firebase.json not found: push notifications can't start"
    echo '{}' > "$MONOLITH_DIR/firebase.json"
  fi
  "${DOCKER[@]}" network inspect "${CADDY_NETWORK:-caddy_net}" >/dev/null 2>&1 \
    || "${DOCKER[@]}" network create "${CADDY_NETWORK:-caddy_net}" >/dev/null
  mkdir -p "$MONOLITH_DIR/data"
}

login() {
  if [[ -n "${GHCR_TOKEN:-}" ]]; then
    echo "$GHCR_TOKEN" | "${DOCKER[@]}" login "${REGISTRY%%/*}" -u "${GHCR_USER:-token}" --password-stdin >/dev/null
  fi
}

# ─── samples + sandbox ───────────────────────────────────────────────────────────────────

upload_samples() {
  log "Preview samples"
  REGISTRY="$REGISTRY" TAG="$TAG" compose --profile e2e run --rm --no-deps --entrypoint /app/upload-samples e2e \
    || warn "upload-samples failed: the preview may use stale samples"
}

run_sandbox() {
  local project="$PROJECT-sandbox-$RUN_ID" network="inkomi-sandbox-$RUN_ID" data
  data="$(mktemp -d)"
  # the container runs as uid 10001
  chmod 777 "$data"
  log "Sandbox $project"

  "${DOCKER[@]}" network create "$network" >/dev/null
  local status=0
  REGISTRY="$REGISTRY" TAG="$TAG" DATA_PATH="$data" CADDY_NETWORK="$network" \
    compose -p "$project" --profile e2e run --rm e2e || status=$?

  if [[ $status -ne 0 ]]; then
    log "Sandbox logs"
    REGISTRY="$REGISTRY" TAG="$TAG" DATA_PATH="$data" CADDY_NETWORK="$network" \
      compose -p "$project" logs --tail 200 monolith || true
  fi
  REGISTRY="$REGISTRY" TAG="$TAG" DATA_PATH="$data" CADDY_NETWORK="$network" \
    compose -p "$project" --profile e2e down -v --remove-orphans >/dev/null 2>&1 || true
  "${DOCKER[@]}" network rm "$network" >/dev/null 2>&1 || true
  # the files belong to uid 10001; a leftover temp dir is harmless
  rm -rf "$data" 2>/dev/null || true

  [[ $status -eq 0 ]] || { echo "e2e tests failed in the sandbox: production untouched" >&2; exit 1; }
}

# ─── production ──────────────────────────────────────────────────────────────────────────

wait_healthy() {
  local id deadline=$((SECONDS + HEALTH_TIMEOUT)) health
  id="$(compose ps -q monolith)"
  while (( SECONDS < deadline )); do
    health="$("${DOCKER[@]}" inspect -f '{{.State.Health.Status}}' "$id" 2>/dev/null || echo unknown)"
    [[ "$health" == "healthy" ]] && return 0
    sleep 3
  done
  return 1
}

deploy_production() {
  local previous=""
  [[ -f "$MONOLITH_DIR/.deployed-tag" ]] && previous="$(cat "$MONOLITH_DIR/.deployed-tag")"
  log "Production $TAG (previous: ${previous:-none})"

  REGISTRY="$REGISTRY" TAG="$TAG" compose pull monolith web
  REGISTRY="$REGISTRY" TAG="$TAG" compose up -d --remove-orphans monolith web

  if wait_healthy; then
    echo "$TAG" > "$MONOLITH_DIR/.deployed-tag"
    "${DOCKER[@]}" image prune -f >/dev/null || true
    log "Deployed $TAG"
    return 0
  fi

  REGISTRY="$REGISTRY" TAG="$TAG" compose logs --tail 200 monolith || true
  if [[ -n "$previous" ]]; then
    warn "monolith $TAG is not healthy: rolling back to $previous"
    REGISTRY="$REGISTRY" TAG="$previous" compose up -d monolith web
  fi
  exit 1
}

check_host
login
upload_samples
run_sandbox
deploy_production
