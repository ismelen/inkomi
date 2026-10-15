#!/usr/bin/env bash
# Prepares the Oracle instance as the k3s server node (control plane, inkomi.io/capacity=low) and
# the public entry point. Run it after the Ansible of utils/server-config, as root:
#
#   scp -P 2222 -r deploy admin@<oracle>:~/
#   ssh -p 2222 admin@<oracle>
#   sudo TS_AUTHKEY=tskey-auth-... DOMAIN=api.example.com bash ~/deploy/servers/setup-oracle.sh
#
# Variables:
#   TS_AUTHKEY    Tailscale auth key (only the first time, while the node isn't in the tailnet)
#   DOMAIN        public hostname served by Caddy with HTTPS (optional: without it Caddy is untouched)
#   DEPLOY_USER   user of the pipeline (default deployer)
#   SWAP_SIZE     swap file size when there's none (default 4G)
#
# See deploy/SERVERS.md for the reasons behind each step.
set -euo pipefail
source "$(dirname "$0")/lib.sh"

DEPLOY_USER="${DEPLOY_USER:-deployer}"
SWAP_SIZE="${SWAP_SIZE:-4G}"
SERVICES=(auth-service library-service transaction-manager manga-worker epub-worker push-service)

require_root
id "$DEPLOY_USER" >/dev/null 2>&1 || die "user $DEPLOY_USER doesn't exist: apply the Ansible first"
command -v docker >/dev/null || die "docker not installed: apply the Ansible first"

# OCI's Ubuntu images ship REJECT rules in /etc/iptables/rules.v4 that drop everything but ssh,
# whatever UFW allows. Only those REJECT lines go; the InstanceServices chain (metadata, boot
# volume) stays.
remove_oci_reject_rules() {
  local rules=/etc/iptables/rules.v4
  if [[ -f "$rules" ]] && grep -q -- '-j REJECT --reject-with icmp-host-prohibited' "$rules"; then
    log "Removing OCI's default REJECT rules"
    cp "$rules" "$rules.bak.$(date +%Y%m%d%H%M%S)"
    sed -i '/-j REJECT --reject-with icmp-host-prohibited/d' "$rules"
  fi
  for chain in INPUT FORWARD; do
    while iptables -C "$chain" -j REJECT --reject-with icmp-host-prohibited 2>/dev/null; do
      iptables -D "$chain" -j REJECT --reject-with icmp-host-prohibited
    done
  done
}

install_k3s_server() {
  if command -v k3s >/dev/null && systemctl is-active -q k3s; then
    log "k3s server already running"
  else
    log "Installing k3s server"
    curl -sfL https://get.k3s.io | sh -s - server \
      --write-kubeconfig-mode 644 \
      --disable servicelb \
      --node-ip "$TS_IP" \
      --advertise-address "$TS_IP" \
      --tls-san "$TS_IP" \
      --flannel-iface tailscale0 \
      --node-label inkomi.io/capacity=low \
      --kube-proxy-arg nodeport-addresses=172.16.0.0/12 \
      --kubelet-arg system-reserved=memory=300Mi \
      --kubelet-arg 'eviction-hard=memory.available<150Mi'
  fi

  log "Waiting for the node"
  for _ in $(seq 1 60); do
    k3s kubectl get node "$(hostname)" >/dev/null 2>&1 && break
    sleep 5
  done
  k3s kubectl wait --for=condition=Ready "node/$(hostname)" --timeout=300s
  k3s kubectl label node "$(hostname)" inkomi.io/capacity=low --overwrite >/dev/null
  k3s kubectl apply -f "$(dirname "$0")/../k8s/traefik-config.yaml" 2>/dev/null \
    || warn "traefik-config.yaml not next to this script: the pipeline applies it on the first deploy"
}

prepare_env_dirs() {
  log "Env directories of $DEPLOY_USER"
  local home base
  home="$(getent passwd "$DEPLOY_USER" | cut -d: -f6)"
  base="$home/inkomi/back"
  for svc in "${SERVICES[@]}"; do
    install -d -o "$DEPLOY_USER" -g "$DEPLOY_USER" -m 700 "$base/$svc"
    if [[ -f "$base/$svc/.env" ]]; then
      chown "$DEPLOY_USER:$DEPLOY_USER" "$base/$svc/.env"
      chmod 600 "$base/$svc/.env"
    else
      warn "missing $base/$svc/.env (template: back/$svc/.env.example)"
    fi
  done
  chown "$DEPLOY_USER:$DEPLOY_USER" "$home/inkomi" "$base"
  chmod 700 "$home/inkomi"
  [[ -f "$base/push-service/firebase.json" ]] && chmod 600 "$base/push-service/firebase.json" \
    || warn "missing $base/push-service/firebase.json"
}

# Daily copy of the SQLite databases (users, transactions) kept 14 days in /var/backups/inkomi.
install_backups() {
  log "Daily SQLite backups"
  cat > /usr/local/bin/inkomi-backup <<'EOF'
#!/usr/bin/env bash
# Copies the SQLite databases of the inkomi namespace from the k3s local-path volumes.
set -euo pipefail
dest=/var/backups/inkomi/$(date +%Y-%m-%d)
mkdir -p "$dest"
for db in $(find /var/lib/rancher/k3s/storage -path '*_inkomi_*' \( -name 'users.db' -o -name 'transactions.db' \)); do
  sqlite3 "$db" ".backup '$dest/$(basename "$db")'"
done
find /var/backups/inkomi -mindepth 1 -maxdepth 1 -type d -mtime +14 -exec rm -rf {} +
EOF
  chmod 755 /usr/local/bin/inkomi-backup
  echo '30 3 * * * root /usr/local/bin/inkomi-backup' > /etc/cron.d/inkomi-backup
  install -d -m 700 /var/backups/inkomi
}

configure_caddy() {
  if [[ -z "${DOMAIN:-}" ]]; then
    warn "DOMAIN not set: Caddy untouched"
    return
  fi
  local gateway
  gateway="$(caddy_gateway)" || true
  [[ -n "$gateway" ]] || { warn "docker network caddy_net not found: skipping Caddy"; return; }
  write_caddyfile "$DOMAIN {
    reverse_proxy $gateway:30080
}

:80 {
    respond \"Acceso denegado. Petición no reconocida.\" 403
}"
}

install_packages
remove_oci_reject_rules
ensure_swap "$SWAP_SIZE"
setup_tailscale inkomi-oracle
open_cluster_firewall
install_k3s_server
prepare_env_dirs
install_backups
configure_caddy

log "Done"
cat <<EOF

Oracle tailscale ip : $TS_IP
Join token (setup-mobile.sh K3S_TOKEN):
  $(cat /var/lib/rancher/k3s/server/node-token)

Pending:
  - Fill $(getent passwd "$DEPLOY_USER" | cut -d: -f6)/inkomi/back/<service>/.env and push-service/firebase.json
  - OCI Security List: TCP 80, 443 and 2222 open (UDP 41641 optional, direct Tailscale)
  - GitHub secrets: ORACLE_HOST, ORACLE_SSH_PORT=2222, ORACLE_USERNAME=$DEPLOY_USER, ORACLE_PRIVATE_KEY, GHCR_PULL_TOKEN
  - Run setup-mobile.sh on the phone, then: kubectl get nodes -L inkomi.io/capacity
EOF
