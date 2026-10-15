# Shared steps of setup-oracle.sh and setup-mobile.sh. Sourced, not executed.
# Every step is idempotent: running a setup script again only applies what's missing.

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
  warn "disable key expiry of '$1' in https://login.tailscale.com/admin/machines or the cluster breaks in 180 days"
}

# open_cluster_firewall: lets the cluster talk over Tailscale and Caddy (docker) reach Traefik.
open_cluster_firewall() {
  if ! command -v ufw >/dev/null; then
    warn "ufw not installed: skipping firewall rules"
    return
  fi
  log "UFW rules for the cluster"
  ufw allow in on tailscale0 comment 'k3s over tailscale' >/dev/null
  ufw allow from 10.42.0.0/16 comment 'k3s pods' >/dev/null
  ufw allow from 10.43.0.0/16 comment 'k3s services' >/dev/null
  # Caddy runs in docker and reaches Traefik's NodePort through the docker bridge gateway
  ufw allow proto tcp from 172.16.0.0/12 to any port 30080 comment 'caddy -> traefik' >/dev/null
  ufw reload >/dev/null
}

# caddy_gateway prints the ip Caddy (network caddy_net) uses to reach the host.
caddy_gateway() {
  docker network inspect caddy_net -f '{{(index .IPAM.Config 0).Gateway}}' 2>/dev/null
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
