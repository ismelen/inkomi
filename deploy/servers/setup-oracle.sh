#!/usr/bin/env bash
# Prepares the Oracle instance for the api stack (back/api/docker-compose.yml, deployed by
# .github/workflows/deploy-api.yml). Run it after the Ansible of utils/server-config, as root:
#
#   scp -P 2222 -r deploy admin@<oracle>:~/
#   ssh -p 2222 admin@<oracle>
#   sudo TS_AUTHKEY=tskey-auth-... API_DOMAIN=inkomi.example.com bash ~/deploy/servers/setup-oracle.sh
#
# Variables:
#   TS_AUTHKEY    Tailscale auth key (only the first time, while the node isn't in the tailnet)
#   API_DOMAIN    public hostname of the api stack, served by Caddy with HTTPS (optional: without
#                 it Caddy is untouched)
#   DEPLOY_USER   user of the pipeline (default deployer)
#   SWAP_SIZE     swap file size when there's none (default 4G)
#
# See deploy/SERVERS.md for the reasons behind each step.
set -euo pipefail
source "$(dirname "$0")/lib.sh"

DEPLOY_USER="${DEPLOY_USER:-deployer}"
SWAP_SIZE="${SWAP_SIZE:-4G}"

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

prepare_stack_dir() {
  log "Stack directory of $DEPLOY_USER"
  local home stack
  home="$(getent passwd "$DEPLOY_USER" | cut -d: -f6)"
  stack="$home/inkomi/back/api"
  install -d -o "$DEPLOY_USER" -g "$DEPLOY_USER" -m 700 "$home/inkomi" "$home/inkomi/back" "$stack"
  # the container runs as uid 10001
  install -d -o 10001 -g 10001 -m 700 "$stack/data"
  if [[ -f "$stack/.env" ]]; then
    chown "$DEPLOY_USER:$DEPLOY_USER" "$stack/.env"
    chmod 600 "$stack/.env"
  else
    warn "missing $stack/.env (template: back/api/.env.example)"
  fi
  [[ -f "$stack/firebase.json" ]] && chmod 600 "$stack/firebase.json" \
    || warn "missing $stack/firebase.json"
  # deploy/api/deploy.sh runs docker compose as the deploy user
  getent group docker >/dev/null && usermod -aG docker "$DEPLOY_USER"
}

# Daily copy of the SQLite databases (users, transactions) of the api stack, kept 14 days in
# /var/backups/inkomi.
install_backups() {
  log "Daily SQLite backups"
  cat > /usr/local/bin/inkomi-backup <<'EOF'
#!/usr/bin/env bash
set -euo pipefail
dest=/var/backups/inkomi/$(date +%Y-%m-%d)
mkdir -p "$dest"
for db in /home/*/inkomi/back/api/data/users.db /home/*/inkomi/back/api/data/transactions.db; do
  if [[ -f "$db" ]]; then sqlite3 "$db" ".backup '$dest/$(basename "$db")'"; fi
done
find /var/backups/inkomi -mindepth 1 -maxdepth 1 -type d -mtime +14 -exec rm -rf {} +
EOF
  chmod 755 /usr/local/bin/inkomi-backup
  echo '30 3 * * * root /usr/local/bin/inkomi-backup' > /etc/cron.d/inkomi-backup
  install -d -m 700 /var/backups/inkomi
}

configure_caddy() {
  if [[ -z "${API_DOMAIN:-}" ]]; then
    warn "API_DOMAIN not set: Caddy untouched"
    return
  fi
  # the api stack joins caddy_net with these aliases (back/api/docker-compose.yml)
  write_caddyfile "$API_DOMAIN {
    @api path /auth/* /users/* /books/* /uploads /uploads/* /ws /ws/* /preview /preview/* /app/* /files/* /wishlist /wishlist/* /recommendations /recommendations/*
    reverse_proxy @api inkomi-api:3000
    reverse_proxy inkomi-web:8080
}

:80 {
    respond \"Acceso denegado. Petición no reconocida.\" 403
}"
}

install_packages
remove_oci_reject_rules
ensure_swap "$SWAP_SIZE"
setup_tailscale inkomi-oracle
prepare_stack_dir
install_backups
configure_caddy

log "Done"
cat <<EOF

Pending:
  - Fill $(getent passwd "$DEPLOY_USER" | cut -d: -f6)/inkomi/back/api/.env (template: back/api/.env.example) and firebase.json
  - OCI Security List: TCP 80, 443 and 2222 open (UDP 41641 optional, direct Tailscale)
  - GitHub secrets: ORACLE_HOST, ORACLE_SSH_PORT=2222, ORACLE_USERNAME=$DEPLOY_USER, ORACLE_PRIVATE_KEY, GHCR_PULL_TOKEN
  - Run Actions → Deploy web (image) once, then Actions → Deploy api
EOF
