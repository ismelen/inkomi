# Helpers of setup-oracle.sh. Sourced, not executed.
# Every step is idempotent: running the setup script again only applies what's missing.

log() { printf '\n\033[1m==> %s\033[0m\n' "$*"; }
warn() { printf '\033[33mWARNING: %s\033[0m\n' "$*" >&2; }
die() { printf '\033[31mERROR: %s\033[0m\n' "$*" >&2; exit 1; }

require_root() {
  [[ $EUID -eq 0 ]] || die "run it as root: sudo -E bash $0"
}

require_var() {
  [[ -n "${!1:-}" ]] || die "missing $1 ($2)"
}

install_packages() {
  log "Packages"
  apt-get update -qq
  DEBIAN_FRONTEND=noninteractive apt-get install -y -qq curl ca-certificates gettext-base jq sqlite3 >/dev/null
}

# ensure_swap <size>: creates /swapfile when the machine has no swap.
ensure_swap() {
  if [[ -n "$(swapon --show --noheadings)" ]]; then
    log "Swap already configured"
    return
  fi
  log "Creating a $1 swap file"
  fallocate -l "$1" /swapfile
  chmod 600 /swapfile
  mkswap /swapfile >/dev/null
  swapon /swapfile
  grep -q '^/swapfile ' /etc/fstab || echo '/swapfile none swap sw 0 0' >> /etc/fstab
  sysctl -q vm.swappiness=10
  echo 'vm.swappiness=10' > /etc/sysctl.d/90-inkomi-swap.conf
}

# setup_tailscale <hostname>: joins the tailnet with TS_AUTHKEY and prints the node's ip.
setup_tailscale() {
  log "Tailscale"
  command -v tailscale >/dev/null || curl -fsSL https://tailscale.com/install.sh | sh
  if ! tailscale ip -4 >/dev/null 2>&1; then
    require_var TS_AUTHKEY "Tailscale auth key: https://login.tailscale.com/admin/settings/keys"
    tailscale up --authkey "$TS_AUTHKEY" --hostname "$1" --ssh
  fi
  TS_IP="$(tailscale ip -4 | head -n1)"
  echo "tailscale ip: $TS_IP"
  warn "disable key expiry of '$1' in https://login.tailscale.com/admin/machines or it leaves the tailnet in 180 days"
}

# write_caddyfile <site block>: rewrites /opt/caddy/Caddyfile (created by Ansible) in place,
# keeping the bind-mounted file, and reloads Caddy.
write_caddyfile() {
  local caddyfile=/opt/caddy/Caddyfile
  [[ -f "$caddyfile" ]] || { warn "$caddyfile not found (Ansible caddy role not applied): skipping Caddy"; return; }
  log "Caddy"
  cp "$caddyfile" "$caddyfile.bak.$(date +%Y%m%d%H%M%S)"
  printf '%s\n' "$1" > "$caddyfile"
  if docker exec caddy caddy reload --config /etc/caddy/Caddyfile --adapter caddyfile; then
    echo "Caddy reloaded"
  else
    warn "Caddy didn't accept the new config: check docker logs caddy"
  fi
  warn "copy $caddyfile into utils/server-config/roles/caddy/files/Caddyfile or the next Ansible run overwrites it"
}
