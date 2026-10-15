#!/usr/bin/env bash
# Deploys the Inkomi backend to the k3s cluster. Run over SSH on the k3s server node (the Oracle
# instance) by .github/workflows/backend-k3s.yml, from a fresh copy of deploy/, after the
# pipeline pushed the images to the registry.
#
#   deploy.sh --tag <tag> --services <all|svc,svc,...> --registry ghcr.io/<owner> [--run <id>]
#   env: GHCR_USER / GHCR_TOKEN (optional) to create the image pull secret
#
# Steps:
#   1. Syncs the Secret <svc>-env of each service from ~/inkomi/back/<svc>/.env on this server.
#   2. Uploads the manga-worker preview samples to R2 when they changed.
#   3. Sandbox: deploys everything with the new images in a throwaway namespace (its own NATS,
#      Redis and databases, on the same nodes as production) and runs the e2e/integration tests
#      there. The namespace is deleted afterwards, whatever the result.
#   4. Production: applies the manifests with the new images and waits for the rollout; a failed
#      rollout is undone.
# It needs no sudo: kubectl reads the k3s kubeconfig (installed with --write-kubeconfig-mode 644).
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
ENV_DIR="${ENV_DIR:-$HOME/inkomi/back}"
PROD_NS="${PROD_NS:-inkomi}"
ALL_SERVICES=(auth-service library-service transaction-manager manga-worker epub-worker push-service)

TAG=""
SERVICES=""
REGISTRY=""
RUN_ID="$(date +%s)"

while [[ $# -gt 0 ]]; do
  case "$1" in
    --tag) TAG="$2"; shift 2 ;;
    --services) SERVICES="$2"; shift 2 ;;
    --registry) REGISTRY="$2"; shift 2 ;;
    --run) RUN_ID="$2"; shift 2 ;;
    *) echo "unknown argument $1" >&2; exit 2 ;;
  esac
done
[[ -n "$TAG" && -n "$SERVICES" && -n "$REGISTRY" ]] || { echo "--tag, --services and --registry are required" >&2; exit 2; }

export KUBECONFIG="${KUBECONFIG:-/etc/rancher/k3s/k3s.yaml}"

log() { printf '\n==> %s\n' "$*"; }
warn() { printf 'WARNING: %s\n' "$*" >&2; }
tag_var() { echo "TAG_$(echo "$1" | tr 'a-z-' 'A-Z_')"; }

# ─── checks ──────────────────────────────────────────────────────────────────────────────

check_cluster() {
  command -v kubectl >/dev/null || { echo "kubectl not found: is k3s installed? (deploy/SERVERS.md)" >&2; exit 1; }
  command -v envsubst >/dev/null || { echo "envsubst not found: sudo apt-get install -y gettext-base" >&2; exit 1; }
  kubectl get nodes -L inkomi.io/capacity
  if [[ -z "$(kubectl get nodes -l inkomi.io/capacity=high -o name)" ]]; then
    echo "no node labelled inkomi.io/capacity=high: manga-worker and epub-worker couldn't be scheduled" >&2
    exit 1
  fi
  kubectl apply -f "$ROOT/deploy/k8s/traefik-config.yaml"
}

# ─── manifests ───────────────────────────────────────────────────────────────────────────

# render <file> <namespace> prints the manifest with the namespace, registry and image tags set.
render() {
  local file="$1" vars='$NAMESPACE $REGISTRY $TAG_E2E'
  for svc in "${ALL_SERVICES[@]}"; do vars+=" \$$(tag_var "$svc")"; done
  NAMESPACE="$2" envsubst "$vars" < "$file"
}

current_tag() {
  kubectl -n "$PROD_NS" get deployment "$1" -o jsonpath='{.spec.template.spec.containers[0].image}' 2>/dev/null \
    | sed -n 's/.*://p'
}

prepare_namespace() {
  local ns="$1"
  kubectl create namespace "$ns" --dry-run=client -o yaml | kubectl apply -f - >/dev/null

  if [[ -n "${GHCR_TOKEN:-}" ]]; then
    kubectl -n "$ns" create secret docker-registry ghcr \
      --docker-server="${REGISTRY%%/*}" --docker-username="${GHCR_USER:-token}" --docker-password="$GHCR_TOKEN" \
      --dry-run=client -o yaml | kubectl apply -f - >/dev/null
  fi

  for svc in "${ALL_SERVICES[@]}"; do
    local args=() env_file="$ENV_DIR/$svc/.env"
    if [[ -f "$env_file" ]]; then
      args+=(--from-file=.env="$env_file")
    else
      warn "$env_file not found: $svc starts without its .env"
      args+=(--from-literal=.env=)
    fi
    if [[ "$svc" == "push-service" ]]; then
      if [[ -f "$ENV_DIR/push-service/firebase.json" ]]; then
        args+=(--from-file=firebase.json="$ENV_DIR/push-service/firebase.json")
      else
        warn "$ENV_DIR/push-service/firebase.json not found"
        args+=(--from-literal=firebase.json='{}')
      fi
    fi
    kubectl -n "$ns" create secret generic "$svc-env" "${args[@]}" --dry-run=client -o yaml \
      | kubectl apply -f - >/dev/null
  done
}

wait_rollouts() {
  local ns="$1"
  kubectl -n "$ns" rollout status statefulset/nats --timeout=300s || return 1
  for deploy in redis "${ALL_SERVICES[@]}"; do
    kubectl -n "$ns" rollout status "deployment/$deploy" --timeout=600s || return 1
  done
}

dump_state() {
  local ns="$1"
  kubectl -n "$ns" get pods -o wide || true
  for deploy in "${ALL_SERVICES[@]}"; do
    echo "----- $deploy"
    kubectl -n "$ns" logs "deployment/$deploy" --tail=80 2>&1 || true
  done
}

# run_job applies a Job manifest, waits for it and leaves its logs in $JOB_LOGS.
run_job() {
  local ns="$1" file="$2" name="$3" timeout="$4"
  kubectl -n "$ns" delete job "$name" --ignore-not-found --wait=true >/dev/null
  render "$file" "$ns" | kubectl apply -f - >/dev/null

  local result="timeout"
  for ((i = 0; i < timeout / 5; i++)); do
    if [[ "$(kubectl -n "$ns" get job "$name" -o jsonpath='{.status.succeeded}')" == "1" ]]; then
      result="succeeded"; break
    fi
    if [[ -n "$(kubectl -n "$ns" get job "$name" -o jsonpath='{.status.failed}')" ]]; then
      result="failed"; break
    fi
    sleep 5
  done

  JOB_LOGS="$(kubectl -n "$ns" logs "job/$name" --tail=-1 2>&1 || true)"
  echo "$JOB_LOGS"
  echo "job $name: $result"
  [[ "$result" == "succeeded" ]]
}

# ─── steps ───────────────────────────────────────────────────────────────────────────────

check_cluster

DEPLOYED=()
for svc in "${ALL_SERVICES[@]}"; do
  if [[ "$SERVICES" == "all" || ",$SERVICES," == *",$svc,"* ]]; then
    DEPLOYED+=("$svc")
    export "$(tag_var "$svc")=$TAG"
    continue
  fi
  current="$(current_tag "$svc")"
  if [[ -z "$current" ]]; then
    echo "$svc isn't deployed yet: the first deploy must be 'all'" >&2
    exit 1
  fi
  export "$(tag_var "$svc")=$current"
done
export TAG_E2E="$TAG" REGISTRY

log "Production secrets ($PROD_NS)"
prepare_namespace "$PROD_NS"

log "Preview samples"
SAMPLES_CHANGED="false"
run_job "$PROD_NS" "$ROOT/deploy/k8s/samples-job.yaml" upload-samples 300
if grep -q "samples-changed" <<<"$JOB_LOGS"; then SAMPLES_CHANGED="true"; fi

SANDBOX_NS="inkomi-sandbox-$RUN_ID"
log "Sandbox $SANDBOX_NS"
cleanup_sandbox() { kubectl delete namespace "$SANDBOX_NS" --wait=false --ignore-not-found >/dev/null 2>&1 || true; }
trap cleanup_sandbox EXIT

prepare_namespace "$SANDBOX_NS"
render "$ROOT/deploy/k8s/inkomi.yaml" "$SANDBOX_NS" | kubectl apply -f -
if ! wait_rollouts "$SANDBOX_NS"; then
  dump_state "$SANDBOX_NS"
  echo "sandbox rollout failed: production is untouched" >&2
  exit 1
fi

log "E2E and integration tests"
if ! run_job "$SANDBOX_NS" "$ROOT/deploy/k8s/e2e-job.yaml" e2e 900; then
  dump_state "$SANDBOX_NS"
  echo "e2e tests failed: production is untouched" >&2
  exit 1
fi
cleanup_sandbox
trap - EXIT

log "Production ($PROD_NS): ${DEPLOYED[*]}"
render "$ROOT/deploy/k8s/inkomi.yaml" "$PROD_NS" | kubectl apply -f -
render "$ROOT/deploy/k8s/ingress.yaml" "$PROD_NS" | kubectl apply -f -

if ! wait_rollouts "$PROD_NS"; then
  for svc in "${DEPLOYED[@]}"; do
    warn "rolling back $svc"
    kubectl -n "$PROD_NS" rollout undo "deployment/$svc" || true
  done
  dump_state "$PROD_NS"
  exit 1
fi

if [[ "$SAMPLES_CHANGED" == "true" && " ${DEPLOYED[*]} " != *" manga-worker "* ]]; then
  log "Restarting manga-worker to drop its cached samples"
  kubectl -n "$PROD_NS" rollout restart deployment/manga-worker
  kubectl -n "$PROD_NS" rollout status deployment/manga-worker --timeout=600s
fi

log "Deployed ${DEPLOYED[*]} ($TAG)"
