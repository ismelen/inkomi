#!/usr/bin/env bash
# Prepares the phone (ARM64 Linux) as a k3s agent node with inkomi.io/capacity=high, where
# manga-worker and epub-worker run. Run it after the Ansible of utils/server-config, as root:
#
#   scp -P 2222 -r deploy admin@<phone>:~/
#   sudo TS_AUTHKEY=tskey-auth-... K3S_URL=https://100.x.y.z:6443 K3S_TOKEN=K10... \
#        bash ~/deploy/servers/setup-mobile.sh
#
# Variables:
#   TS_AUTHKEY       Tailscale auth key (only the first time)
#   K3S_URL          https://<oracle tailscale ip>:6443 (printed by setup-oracle.sh)
#   K3S_TOKEN        join token (printed by setup-oracle.sh)
#   NODE_CAPACITY    label of the node (default high)
#   CF_TUNNEL_TOKEN  optional: token of a Cloudflare tunnel created in the dashboard. Not needed
#                    for the app (Oracle is the entry point and the cluster routes to the phone);
#                    see deploy/SERVERS.md.
#   TUNNEL_HOSTNAME  public hostname of that tunnel, to configure Caddy behind it
#
# Also valid for any future node: change NODE_CAPACITY.
set -euo pipefail
source "$(dirname "$0")/lib.sh"

NODE_CAPACITY="${NODE_CAPACITY:-high}"
MIN_FREE_GB=10

require_root
require_var K3S_URL "https://<oracle tailscale ip>:6443, printed by setup-oracle.sh"
require_var K3S_TOKEN "printed by setup-oracle.sh"

check_machine() {
  log "Machine"
  echo "arch: $(uname -m), kernel: $(uname -r), memory: $(free -h | awk '/Mem:/ {print $2}')"
  [[ "$(uname -m)" == "aarch64" || "$(uname -m)" == "x86_64" ]] || die "unsupported architecture $(uname -m)"
  local free_gb
  free_gb="$(df -BG --output=avail /var | tail -n1 | tr -dc '0-9')"
  (( free_gb >= MIN_FREE_GB )) || warn "only ${free_gb}G free in /var (manga-worker needs ~${MIN_FREE_GB}G for images and temp files)"
  grep -qw memory /sys/fs/cgroup/cgroup.controllers 2>/dev/null \
    || grep -q '^memory' /proc/cgroups 2>/dev/null \
    || die "the kernel has no memory cgroup: k3s can't limit pods"
}

# A server must never suspend: mask sleep targets and ignore the power and suspend keys.
keep_awake() {
  log "Disabling suspend"
  systemctl mask -q sleep.target suspend.target hibernate.target hybrid-sleep.target
  mkdir -p /etc/systemd/logind.conf.d
  cat > /etc/systemd/logind.conf.d/90-inkomi-server.conf <<'EOF'
[Login]
HandleSuspendKey=ignore
HandleLidSwitch=ignore
HandleLidSwitchExternalPower=ignore
HandlePowerKey=ignore
IdleAction=ignore
EOF
  systemctl reload systemd-logind 2>/dev/null || true
}

install_k3s_agent() {
  if command -v k3s >/dev/null && systemctl is-active -q k3s-agent; then
    log "k3s agent already running"
    return
  fi
  log "Installing k3s agent ($NODE_CAPACITY)"
  curl -sfL https://get.k3s.io | K3S_URL="$K3S_URL" K3S_TOKEN="$K3S_TOKEN" sh -s - agent \
    --node-ip "$TS_IP" \
    --flannel-iface tailscale0 \
    --node-label "inkomi.io/capacity=$NODE_CAPACITY" \
    --kube-proxy-arg nodeport-addresses=172.16.0.0/12 \
    --kubelet-arg system-reserved=memory=300Mi

  log "Kernel check"
  k3s check-config 2>&1 | grep -E 'missing|fail' && warn "k3s check-config reports missing features (above)" || true
}

# Optional second entry point through a Cloudflare tunnel managed from the dashboard
# (Zero Trust → Networks → Tunnels): the tunnel's public hostname must point to http://localhost:80.
setup_tunnel() {
  if [[ -z "${CF_TUNNEL_TOKEN:-}" ]]; then
    log "No CF_TUNNEL_TOKEN: skipping the Cloudflare tunnel"
    return
  fi
  log "Cloudflare tunnel"
  if ! command -v cloudflared >/dev/null; then
    local arch=arm64
    [[ "$(uname -m)" == "x86_64" ]] && arch=amd64
    curl -fsSL "https://github.com/cloudflare/cloudflared/releases/latest/download/cloudflared-linux-$arch" -o /usr/local/bin/cloudflared
    chmod 755 /usr/local/bin/cloudflared
  fi
  if ! systemctl is-active -q cloudflared; then
    cloudflared service install "$CF_TUNNEL_TOKEN"
  fi

  if [[ -n "${TUNNEL_HOSTNAME:-}" ]]; then
    local gateway
    gateway="$(caddy_gateway)" || true
    [[ -n "$gateway" ]] || { warn "docker network caddy_net not found: skipping Caddy"; return; }
    # Behind the tunnel the client ip comes in Cf-Connecting-Ip; the preview rate limit needs it.
    write_caddyfile "{
    servers {
        trusted_proxies static private_ranges
        client_ip_headers Cf-Connecting-Ip X-Forwarded-For
    }
}

http://$TUNNEL_HOSTNAME {
    reverse_proxy $gateway:30080
}

:80 {
    respond \"Acceso denegado. Petición no reconocida.\" 403
}"
  fi
}

install_packages
check_machine
keep_awake
setup_tailscale "inkomi-$(hostname)"
open_cluster_firewall
install_k3s_agent
setup_tunnel

log "Done"
cat <<EOF

Node: $(hostname)  tailscale ip: $TS_IP  capacity: $NODE_CAPACITY
Check from Oracle:
  kubectl get nodes -L inkomi.io/capacity -o wide
If the label is missing (it's only applied when the node registers):
  kubectl label node $(hostname) inkomi.io/capacity=$NODE_CAPACITY --overwrite
EOF
