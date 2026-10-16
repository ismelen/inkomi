# Servidores: lo que hay que hacer a mano

Guía de lo que **no** hace la pipeline (`.github/workflows/backend-k3s.yml`) y hay que preparar en las máquinas. Las máquinas base se configuran con el Ansible de `utils/server-config` (usuarios, SSH en 2222, UFW, fail2ban, Docker y Caddy). Esta guía va encima de eso.

## Arquitectura

Un único clúster **k3s** con dos nodos (y los que se añadan después):

| Nodo | Rol k3s | Etiqueta | Qué corre | Entrada desde internet |
|---|---|---|---|---|
| Oracle (≈2 vCPU / 2 GB) | server (control plane) | `inkomi.io/capacity=low` | NATS, Redis, auth-service, library-service, transaction-manager, push-service | IP pública → Caddy (HTTPS) |
| Móvil (ARM64, ≥4 GB) | agent | `inkomi.io/capacity=high` | manga-worker, epub-worker | ninguna (opcional: Cloudflare Tunnel, ver §8) |

- `manga-worker` y `epub-worker` *prefieren* nodos `high`. Si no hay ninguno (o no tienen hueco), se colocan donde quepan, aunque sea en un `low`. El resto *prefiere* nodos `low`, así que los nodos grandes quedan libres para procesar imágenes. Son preferencias, no requisitos: el clúster funciona aunque solo haya un nodo.
- Los nodos se ven entre sí por **Tailscale** (el móvil está detrás del NAT de casa). Flannel va sobre la interfaz `tailscale0`.
- Entrada HTTP: el Caddy de cada nodo → Traefik de k3s (NodePort `30080`) → servicios por ruta (`deploy/k8s/ingress.yaml`). Cualquier nodo sirve todas las rutas.
- Las imágenes son multi-arch (amd64 + arm64) y se descargan de `ghcr.io`, así que da igual la arquitectura de Oracle y de los nodos futuros.
- La pipeline entra **solo en Oracle** (SSH 2222, usuario `deployer`, sin sudo) y usa `kubectl`.

## Scripts de preparación

Automatizan las secciones 2 a 8 y 10 de esta guía. Son idempotentes: si se vuelven a ejecutar, solo aplican lo que falta. Se ejecutan como root **después del Ansible**.

```bash
# 1. Oracle (server)
scp -P 2222 -r deploy <admin>@<ip-oracle>:~/
ssh -p 2222 <admin>@<ip-oracle>
sudo TS_AUTHKEY=tskey-auth-... DOMAIN=api.tudominio.com bash ~/deploy/servers/setup-oracle.sh
#    → imprime la IP de Tailscale de Oracle y el token para unir nodos

# 2. Móvil (agent) y nodos futuros
scp -P 2222 -r deploy <admin>@<ip-movil>:~/
ssh -p 2222 <admin>@<ip-movil>
sudo TS_AUTHKEY=tskey-auth-... K3S_URL=https://<ip-tailscale-oracle>:6443 K3S_TOKEN=<token>      bash ~/deploy/servers/setup-mobile.sh
#    otro nodo pequeño: añade NODE_CAPACITY=low
```

| Script | Qué hace |
|---|---|
| `setup-oracle.sh` | paquetes, quita los REJECT de OCI, swap, Tailscale, reglas UFW, k3s server con etiqueta `low`, carpetas de `.env` de `deployer`, copia diaria de las SQLite (`/var/backups/inkomi`, 14 días) y Caddy con HTTPS para `DOMAIN` |
| `setup-mobile.sh` | paquetes, comprobaciones (arquitectura, disco, cgroups), desactiva la suspensión, Tailscale, reglas UFW, k3s agent con etiqueta `high` y, opcionalmente, Cloudflare Tunnel y su Caddy (`CF_TUNNEL_TOKEN`, `TUNNEL_HOSTNAME`) |

Lo que sigue siendo manual: el cambio del Ansible (§1), la *Security List* de OCI, desactivar la expiración de clave en Tailscale, rellenar los `.env` (§6) y los secrets de GitHub (§9).

Las auth keys de Tailscale se crean en <https://login.tailscale.com/admin/settings/keys>. Usa una *reusable* si vas a añadir varios nodos. Tailscale SSH queda activado (`tailscale ssh <usuario>@<nodo>`), así que puedes administrar el móvil desde fuera de casa sin túnel.

---

## 1. Ansible (`utils/server-config`)

Cambios necesarios antes de aplicarlo a estas máquinas:

1. **Arquitectura de Docker.** `roles/docker/tasks/main.yml` tiene `arch=amd64` fijo y en el móvil (y en Oracle si es ARM) fallará. Cámbialo por:
   ```yaml
   repo: "deb [arch={{ 'arm64' if ansible_architecture == 'aarch64' else 'amd64' }} signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/ubuntu {{ ansible_distribution_release }} stable"
   ```
   Si el móvil no es Ubuntu (Debian, postmarketOS…) la URL `linux/ubuntu` tampoco valdrá.
2. **`config/general.yml`**
   - Oracle: comenta `allowed_web_ip` (servidor público).
   - Móvil: el tráfico web llega por el túnel (localhost). Ojo: los puertos que publica Docker (Caddy 80/443) se saltan UFW, así que en la LAN de casa siguen abiertos.
3. **Usuario `deployer`**: es el de la pipeline. No necesita sudo (`deploy.sh` no lo usa) ni el grupo `docker`. Necesita que exista `~/inkomi/back/` con los `.env` (paso 6).
4. **UFW** para el tráfico del clúster (en ambos nodos):
   ```bash
   sudo ufw allow in on tailscale0
   sudo ufw allow from 10.42.0.0/16   # pods
   sudo ufw allow from 10.43.0.0/16   # services
   ```

## 2. Oracle: particularidades

- **iptables de Oracle.** Las imágenes Ubuntu de OCI traen reglas propias en `/etc/iptables/rules.v4` que rechazan todo menos el 22. Con UFW encima bloquean el 80/443/2222. Revisa `sudo iptables -L INPUT -n --line-numbers` y elimina esas reglas (o vacía `rules.v4` y reinicia).
- **Security List / NSG de la VCN**: abre TCP 80, 443 y 2222. Opcional: UDP 41641 para que Tailscale conecte directo en vez de por relay.
- **Swap**: con 2 GB es imprescindible.
  ```bash
  sudo fallocate -l 4G /swapfile && sudo chmod 600 /swapfile && sudo mkswap /swapfile && sudo swapon /swapfile
  echo '/swapfile none swap sw 0 0' | sudo tee -a /etc/fstab
  ```
- **La API antigua** (`back/api` con docker compose, más otel/prometheus/tempo/loki/grafana) no cabe junto a k3s en 2 GB. Cuando migres, para ese stack (`docker compose down` en `~/inkomi/back/api`).

## 3. Móvil: comprobaciones previas

- Kernel compatible con k3s. Tras instalarlo: `sudo k3s check-config`. Fíjate en cgroups (memory, cpuset), overlayfs, iptables/nftables y vxlan.
- Que no se suspenda: desactiva el ahorro de energía y la suspensión, y déjalo siempre enchufado (vigila la batería hinchada a largo plazo).
- Almacenamiento: manga-worker usa hasta 4 GB de temporales por conversión (`emptyDir`) y las imágenes ocupan espacio. Deja unos 10 GB libres.

## 4. Tailscale (en todos los nodos)

```bash
curl -fsSL https://tailscale.com/install.sh | sh
sudo tailscale up --ssh          # Tailscale SSH: administrar el nodo sin abrir puertos
tailscale ip -4      # anota la IP de cada nodo (100.x.y.z)
```

En la consola de Tailscale, desactiva la expiración de clave de ambos nodos (*Disable key expiry*). Si no, el clúster se rompe a los 180 días.

## 5. k3s

### Oracle (server)

```bash
ORACLE_TS_IP=$(tailscale ip -4)
curl -sfL https://get.k3s.io | sh -s - server \
  --write-kubeconfig-mode 644 \
  --disable servicelb \
  --node-ip "$ORACLE_TS_IP" \
  --advertise-address "$ORACLE_TS_IP" \
  --flannel-iface tailscale0 \
  --node-label inkomi.io/capacity=low \
  --kube-proxy-arg nodeport-addresses=172.16.0.0/12 \
  --kubelet-arg system-reserved=memory=300Mi \
  --kubelet-arg eviction-hard=memory.available<150Mi

sudo cat /var/lib/rancher/k3s/server/node-token   # token para unir nodos
```

- `--write-kubeconfig-mode 644`: el usuario `deployer` puede usar `kubectl` sin sudo (la pipeline lo necesita).
- `--disable servicelb`: Caddy ya ocupa el 80/443 del host. Traefik va por NodePort (`deploy/k8s/traefik-config.yaml`, que aplica la pipeline).
- `nodeport-addresses=172.16.0.0/12`: los NodePort solo escuchan en los bridges de Docker, por donde llega Caddy, y no en la IP pública.
- Comprueba que `envsubst` existe (`sudo apt-get install -y gettext-base`).

### Móvil y nodos futuros (agent)

```bash
NODE_TS_IP=$(tailscale ip -4)
curl -sfL https://get.k3s.io | K3S_URL="https://<ORACLE_TS_IP>:6443" K3S_TOKEN="<node-token>" sh -s - agent \
  --node-ip "$NODE_TS_IP" \
  --flannel-iface tailscale0 \
  --node-label inkomi.io/capacity=high \
  --kube-proxy-arg nodeport-addresses=172.16.0.0/12
```

Comprobar desde Oracle:

```bash
kubectl get nodes -L inkomi.io/capacity -o wide
```

**Añadir un nodo más adelante:** Ansible, Tailscale y el comando de agent con la etiqueta que le toque (`high` si tiene RAM para procesar imágenes). Para cambiar la etiqueta de un nodo existente:

```bash
kubectl label node <nodo> inkomi.io/capacity=high --overwrite
```

Con más nodos `high` puedes subir `replicas` de `manga-worker` y `epub-worker` en `deploy/k8s/app/manga-worker.yaml` y `epub-worker.yaml`: los consumidores de NATS reparten los trabajos.

## 6. Variables de entorno (en Oracle)

La pipeline crea los Secrets a partir de estos ficheros. Cada servicio tiene su `back/<servicio>/.env.example` con todas las variables explicadas.

```
/home/deployer/inkomi/back/auth-service/.env
/home/deployer/inkomi/back/library-service/.env
/home/deployer/inkomi/back/transaction-manager/.env
/home/deployer/inkomi/back/manga-worker/.env        ← también lo usan la subida de samples y los tests e2e (R2)
/home/deployer/inkomi/back/epub-worker/.env
/home/deployer/inkomi/back/push-service/.env
/home/deployer/inkomi/back/push-service/firebase.json
/home/deployer/inkomi/back/release-service/.env      ← opcional: todas las variables tienen valor por defecto
```

```bash
chmod 700 ~/inkomi && chmod 600 ~/inkomi/back/*/.env ~/inkomi/back/push-service/firebase.json
```

Claves que tienen que cuadrar entre servicios:

- `JWT_PUBLIC_KEY` es la misma en auth-service, library-service y transaction-manager, y es la pareja de `JWT_PRIVATE_KEY` de auth-service.
- La pública de `SERVICE_PRIVATE_KEY` (push-service) va en `SERVICE_PUBLIC_KEYS` de auth-service con la clave `"push-service"`.
- `TOKEN_ENCRYPTION_KEY` (auth-service): no la cambies después, o los tokens de Dropbox guardados quedan ilegibles.
- Las direcciones internas (NATS, Redis, auth-service, rutas de BD, puertos) las fija el manifiesto de cada servicio en `deploy/k8s/app/` y mandan sobre el `.env`.

Si cambias un `.env` sin desplegar código, reinicia el servicio para que lo lea:

```bash
kubectl -n inkomi rollout restart deployment/<servicio>
```

## 7. Caddy (`roles/caddy/files/Caddyfile`)

Caddy corre en Docker (red `caddy_net`) y llega a Traefik por la puerta de enlace de esa red:

```bash
docker network inspect caddy_net -f '{{(index .IPAM.Config 0).Gateway}}'   # p. ej. 172.18.0.1
```

**Oracle** (público, HTTPS automático):

```caddyfile
api.tudominio.com {
    reverse_proxy 172.18.0.1:30080
}
```

**Móvil** (detrás del túnel; el HTTPS lo pone Cloudflare). La IP real del cliente llega en `Cf-Connecting-Ip` y hace falta para el límite por IP de `/preview`:

```caddyfile
{
    servers {
        trusted_proxies static private_ranges
        client_ip_headers Cf-Connecting-Ip X-Forwarded-For
    }
}

http://app.tudominio.com {
    reverse_proxy 172.18.0.1:30080
}
```

Caddy sustituye el `X-Forwarded-For` de clientes no confiables por la IP real y Traefik lo conserva (`forwardedHeaders.insecure`), así que el límite por IP no se puede falsear.

## 8. Cloudflare Tunnel (móvil): opcional

**Para la app no hace falta.** Oracle es la entrada pública y, como los dos nodos están en el mismo clúster, una petición a `https://api.tudominio.com/preview` entra por Oracle y Traefik la manda al manga-worker del móvil por Tailscale. Para administrar el móvil tampoco: usa Tailscale SSH.

Un túnel en el móvil solo aporta una **segunda entrada pública** (otro hostname apuntando a la misma API). Como el control plane, NATS y las bases de datos están en Oracle, no da alta disponibilidad: si Oracle cae, la app cae igual.

Si aun así lo quieres, usa un **túnel con nombre gestionado desde el panel** (requiere un dominio en Cloudflare):

1. Cloudflare Zero Trust → *Networks* → *Tunnels* → *Create a tunnel* (cloudflared). Copia el token.
2. *Public hostname*: `movil.tudominio.com` → servicio `HTTP` → `localhost:80` (el Caddy del móvil).
3. En el móvil: `sudo CF_TUNNEL_TOKEN=... TUNNEL_HOSTNAME=movil.tudominio.com K3S_URL=... K3S_TOKEN=... bash ~/deploy/servers/setup-mobile.sh`. Instala `cloudflared` como servicio y configura Caddy para leer la IP real del cliente (`Cf-Connecting-Ip`), que necesita el límite por IP de `/preview`.

### ¿Y TryCloudflare (`cloudflared tunnel --url ...`)?

Sirve para **probar un rato**, no para dejarlo puesto:

- La URL (`*.trycloudflare.com`) es aleatoria y **cambia cada vez que se reinicia** `cloudflared` (corte de luz, reinicio del móvil…). La app tendría que cambiar de URL.
- No tiene garantía de disponibilidad, tiene límite de peticiones concurrentes y Cloudflare lo reserva para pruebas y desarrollo.
- Solo expone un servicio por proceso y no se puede proteger con Cloudflare Access.

Para echar un vistazo rápido a la previsualización desde el ereader sin dominio:

```bash
cloudflared tunnel --url http://localhost:80   # imprime https://<palabras>.trycloudflare.com
```

Con un Caddy que acepte cualquier host (`:80 { reverse_proxy <gateway>:30080 }`) bastaría con abrir `https://<palabras>.trycloudflare.com/preview`.

## 9. GitHub

Secrets del repositorio (o del environment `production`, al que puedes poner aprobación manual):

| Secret | Valor |
|---|---|
| `ORACLE_HOST` | IP pública de Oracle |
| `ORACLE_SSH_PORT` | `2222` |
| `ORACLE_USERNAME` | `deployer` |
| `ORACLE_PRIVATE_KEY` | clave privada SSH de `deployer` |
| `GHCR_PULL_TOKEN` | PAT (classic) con `read:packages`, para que los nodos descarguen las imágenes privadas de ghcr.io. Si haces públicos los paquetes `inkomi-*`, no hace falta. |

Primer despliegue: *Actions → Deploy backend (k3s) → Run workflow* con `all`. Después puedes desplegar solo algunos servicios (`manga-worker,epub-worker`).

## 10. Copias de seguridad

Las bases SQLite (usuarios en auth-service y transacciones en transaction-manager) viven en volúmenes `local-path` del nodo donde corren, en `/var/lib/rancher/k3s/storage/`. **Si se pierde ese disco se pierden los usuarios.** Programa una copia (cron + `sqlite3 .backup` y subida a R2, o snapshot del volumen de Oracle).

```bash
kubectl -n inkomi get pvc
sudo ls /var/lib/rancher/k3s/storage/
```

## 11. Comprobaciones tras el primer despliegue

```bash
kubectl -n inkomi get pods -o wide        # workers en el móvil, el resto en Oracle
kubectl -n inkomi logs deploy/manga-worker
curl -I https://api.tudominio.com/preview  # 200
kubectl get ns | grep sandbox              # vacío: la pipeline borra los sandboxes
```

Si un despliegue falla en los tests e2e, producción no se toca. Si falla el rollout de producción, la pipeline hace `rollout undo` de los servicios desplegados. Los logs de los pods salen en el log del job de GitHub.
