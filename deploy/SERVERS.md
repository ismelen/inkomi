# Servidores: lo que hay que hacer a mano

Guía de lo que **no** hace la pipeline (`.github/workflows/deploy-api.yml`) y hay que preparar en el servidor. La máquina base se configura con el Ansible de `utils/server-config` (usuarios, SSH en 2222, UFW, fail2ban, Docker y Caddy). Esta guía va encima de eso.

## Arquitectura

Un único servidor (Oracle, ARM) con Docker Compose (`back/api/docker-compose.yml`, proyecto `inkomi-api`):

- `api`: `back/api`, todos los módulos en un proceso, con NATS embebido (JetStream en disco) y las tres SQLite (`users.db`, `transactions.db`, `jobs.db`) en `data/`. Los archivos van al disco del servidor (`STORAGE=local`, ver más abajo) o a R2.
- `web`: la imagen `inkomi-web:latest` que publica `deploy-web.yml`.
- Entrada HTTP: Caddy (HTTPS) → las rutas de la API a `api` y el resto a `web`, en el mismo dominio (sin CORS).
- La pipeline entra por SSH (usuario `deployer`, sin sudo) y usa `docker compose`.

## Script de preparación

`deploy/servers/setup-oracle.sh` automatiza las secciones 2 a 4 y 8. Es idempotente: si se vuelve a ejecutar, solo aplica lo que falta. Se ejecuta como root **después del Ansible**.

```bash
scp -P 2222 -r deploy <admin>@<ip-oracle>:~/
ssh -p 2222 <admin>@<ip-oracle>
sudo TS_AUTHKEY=tskey-auth-... API_DOMAIN=inkomi.tudominio.com bash ~/deploy/servers/setup-oracle.sh
```

Hace: paquetes, quita los REJECT de OCI, swap, Tailscale, la carpeta del stack en `~/inkomi/back/api` (con `data/` de uid 10001 y `deployer` en el grupo `docker`), la copia diaria de las SQLite (`/var/backups/inkomi`, 14 días) y Caddy con HTTPS para `API_DOMAIN`.

Lo que sigue siendo manual: el cambio del Ansible (§1), la *Security List* de OCI, desactivar la expiración de clave en Tailscale, rellenar el `.env` (§4) y los secrets de GitHub (§6).

---

## 1. Ansible (`utils/server-config`)

Cambios necesarios antes de aplicarlo:

1. **Arquitectura de Docker.** `roles/docker/tasks/main.yml` tiene `arch=amd64` fijo y en una máquina ARM fallará. Cámbialo por:
   ```yaml
   repo: "deb [arch={{ 'arm64' if ansible_architecture == 'aarch64' else 'amd64' }} signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/ubuntu {{ ansible_distribution_release }} stable"
   ```
2. **`config/general.yml`**: comenta `allowed_web_ip` (servidor público).
3. **Usuario `deployer`**: es el de la pipeline. No necesita sudo; sí el grupo `docker` (`deploy.sh` usa `docker compose`, y si no puede prueba con `sudo -n`).

## 2. Oracle: particularidades

- **iptables de Oracle.** Las imágenes Ubuntu de OCI traen reglas propias en `/etc/iptables/rules.v4` que rechazan todo menos el 22. Con UFW encima bloquean el 80/443/2222. Revisa `sudo iptables -L INPUT -n --line-numbers` y elimina esas reglas (o vacía `rules.v4` y reinicia).
- **Security List / NSG de la VCN**: abre TCP 80, 443 y 2222. Opcional: UDP 41641 para que Tailscale conecte directo en vez de por relay.
- **Swap**: con 2 GB es imprescindible.
  ```bash
  sudo fallocate -l 4G /swapfile && sudo chmod 600 /swapfile && sudo mkswap /swapfile && sudo swapon /swapfile
  echo '/swapfile none swap sw 0 0' | sudo tee -a /etc/fstab
  ```

## 3. Tailscale (opcional)

Solo para administrar el servidor sin exponer más puertos (Tailscale SSH):

```bash
curl -fsSL https://tailscale.com/install.sh | sh
sudo tailscale up --ssh
```

Las auth keys se crean en <https://login.tailscale.com/admin/settings/keys>. En la consola, desactiva la expiración de clave del nodo (*Disable key expiry*).

## 4. Variables de entorno

```
/home/deployer/inkomi/back/api/.env            ← plantilla: back/api/.env.example
/home/deployer/inkomi/back/api/firebase.json
/home/deployer/inkomi/back/api/data/           ← propiedad de uid 10001 (el usuario del contenedor)
/home/deployer/inkomi/back/api/.deploy.env     ← opcional: API_DOMAIN=<dominio> para el Caddy del host (§5)
```

```bash
sudo install -d -o 10001 -g 10001 -m 700 ~deployer/inkomi/back/api/data
chmod 700 ~/inkomi && chmod 600 ~/inkomi/back/api/.env ~/inkomi/back/api/firebase.json
```

- `deploy.sh` no despliega si faltan `JWT_PRIVATE_KEY`, `JWT_PUBLIC_KEY`, `TOKEN_ENCRYPTION_KEY`, `GOOGLE_CLIENT_IDS` o las `R2_*` (las usan la subida de samples y los tests e2e).
- `TOKEN_ENCRYPTION_KEY`: no la cambies después, o los tokens de Dropbox guardados quedan ilegibles.
- Si cambias el `.env` sin desplegar código: `cd ~/inkomi/back/api && docker compose up -d api`.

## 5. Caddy

**Caddy en Docker** (servidores montados con el Ansible, `roles/caddy/files/Caddyfile`): `setup-oracle.sh` con `API_DOMAIN` escribe:

```caddyfile
inkomi.tudominio.com {
    @api path /auth/* /users/* /books/* /uploads /uploads/* /ws /ws/* /preview /preview/* /app/* /files/* /wishlist /wishlist/*
    reverse_proxy @api inkomi-api:3000
    reverse_proxy inkomi-web:8080
}
```

El compose une `api` y `web` a `caddy_net` con los alias `inkomi-api` e `inkomi-web`. Copia ese Caddyfile a `utils/server-config/roles/caddy/files/Caddyfile` o el siguiente Ansible lo sobrescribe.

**Caddy instalado en el host** (systemd): tras un despliegue sano, `deploy.sh` renderiza `deploy/api/Caddyfile` con el `API_DOMAIN` de `.deploy.env`, lo instala en `/etc/caddy/Caddyfile` (con copia `.bak-*` del anterior) y recarga Caddy. Las rutas de la API van a `127.0.0.1:3001` y el resto a la web en `127.0.0.1:8081`.

## 6. GitHub

Secrets del repositorio (o del environment `production`, al que puedes poner aprobación manual):

| Secret | Valor |
|---|---|
| `ORACLE_HOST` | IP pública o dominio de Oracle |
| `ORACLE_SSH_PORT` | `2222` (por defecto 22) |
| `ORACLE_USERNAME` | `deployer` |
| `ORACLE_PRIVATE_KEY` | clave privada SSH de `deployer` |
| `GHCR_PULL_TOKEN` | opcional: PAT (classic) con `read:packages`; sin él se usa el token del job |

*Actions → Deploy api (Docker Compose)* (`.github/workflows/deploy-api.yml`):

1. Tests (con libvips) e imágenes `inkomi-api` e `inkomi-e2e` (`linux/arm64`, en un runner ARM) en ghcr.io.
2. En el servidor, `deploy/api/deploy.sh`: sube los samples de la preview a R2, arranca la imagen nueva como proyecto compose desechable (datos y red propios) y pasa los tests e2e contra él. Solo si pasan, actualiza producción, espera al healthcheck (`/healthz`) y, si falla, vuelve a la versión anterior (`.deployed-tag`).

**Migración desde el nombre anterior:** si el servidor aún tiene el stack en `~/inkomi/back/monolith` (proyecto `inkomi-monolith`), el primer `deploy.sh` para ese proyecto, mueve la carpeta a `~/inkomi/back/api` (con `.env`, `firebase.json` y `data/`), cambia `MONOLITH_DOMAIN` por `API_DOMAIN` en `.deploy.env` y etiqueta la imagen desplegada como `inkomi-api` para que el rollback siga funcionando. Si `~/inkomi/back/api` ya existía sin `.env` (la carpeta de la API antigua), primero la aparta como `~/inkomi/back/api-legacy-<fecha>`, sin borrar nada.

### Web (`front/web`)

*Actions → Deploy web (image)* (`.github/workflows/deploy-web.yml`) construye la imagen nginx y la publica como `inkomi-web:<sha>` y `inkomi-web:latest`. El stack usa `latest` y la recoge en el siguiente *Deploy api*, así que publícala al menos una vez antes del primero.

Variables del repositorio (públicas, van dentro del bundle): `DROPBOX_APP_KEY` y `GOOGLE_WEB_CLIENT_ID`. Además, una vez:

- **R2** (solo con `STORAGE=r2`): el navegador sube con `PUT` a las URLs prefirmadas. Política CORS del bucket: `AllowedOrigins: ["https://<dominio>"]`, `AllowedMethods: ["PUT"]`, `AllowedHeaders: ["content-type"]`.
- **Dropbox** (app console → OAuth 2 → Redirect URIs): `https://<dominio>/oauth/dropbox`.
- **Google** (client id web, el mismo de `GOOGLE_CLIENT_IDS`): añadir `https://<dominio>` a *Authorized JavaScript origins*.

## 7. Comprobaciones tras el primer despliegue

```bash
cd ~/inkomi/back/api
docker compose ps
docker compose logs -f api
docker compose ls                               # solo inkomi-api
curl -I https://inkomi.tudominio.com/healthz    # 200
```

Si un despliegue falla en los tests e2e, producción no se toca. Los logs del sandbox salen en el log del job de GitHub.

## 8. Copias de seguridad

`setup-oracle.sh` instala `/usr/local/bin/inkomi-backup` (cron diario a las 3:30): `sqlite3 .backup` de `users.db` y `transactions.db` en `/var/backups/inkomi/<fecha>`, 14 días. Están en el mismo disco: para sobrevivir a su pérdida, súbelas fuera (R2) o haz snapshots del volumen.

## 9. Archivos en el disco del servidor

La api guarda los archivos en `~/inkomi/back/api/data/files` (`STORAGE=local`) en lugar de R2. La app y la web suben y bajan por `https://<dominio>/files/...` con URLs firmadas; el Caddyfile ya enruta `/files/*` a la api. Variables del `.env` (ver `back/api/.env.example`): `PUBLIC_URL` (obligatoria), `FILES_RETENTION`, `DISK_RESERVE_MB`, `STORAGE_EPUB_RATIO` y `LIBRARY_ESTIMATE_MB`.

- Cada envío reserva el máximo de disco que va a ocupar a la vez. Si no cabe (descontando `DISK_RESERVE_MB`), el servidor responde 507 y la app dice que se intente más tarde.
- Lo que nadie toca en `FILES_RETENTION` (7 días) se borra: resultados ya enviados y subidas abandonadas.
- Con 6 GB libres en el volumen de arranque cabe poco más que una carpeta grande. Para más, crea un volumen de bloque en OCI (la capa gratuita llega a 200 GB en total), móntalo y apunta `data/` a él:
  ```bash
  sudo mkfs.ext4 /dev/oracleoci/oraclevdb
  sudo mkdir -p /mnt/inkomi && sudo mount /dev/oracleoci/oraclevdb /mnt/inkomi
  echo '/dev/oracleoci/oraclevdb /mnt/inkomi ext4 defaults,_netdev,nofail 0 2' | sudo tee -a /etc/fstab
  cd ~/inkomi/back/api && docker compose stop api
  sudo rsync -a data/ /mnt/inkomi/data/ && sudo mv data data.old && ln -s /mnt/inkomi/data data
  docker compose up -d api
  ```
