# Web (`front/web`) — reglas

Mismo producto que `front/mobile` (mismo funcionamiento y contratos con el backend); lee también su `CLAUDE.md`. Aquí solo lo que cambia en la web.

## Stack y comandos

- Vite + React 19 + TypeScript estricto, react-router 7, zustand 5 + immer, Tailwind CSS v4, i18next, `idb` (IndexedDB), dnd-kit.
- `npm run dev` (proxy de `/auth`, `/users`, `/books`, `/uploads`, `/ws` y `/app` a los servicios locales por sus puertos, o todo a `DEV_BACKEND` si se define) · `npm run build` (incluye `tsc`) · `npm run typecheck` · `npx prettier --check .`.
- Sin backend: `npm run mock` (backend falso en memoria, `dev/mock-backend.mjs`, puerto 8080) y `npm run dev:mock` (Vite con `.env.mock`). Si cambia un contrato del backend, actualiza también el mock.
- Variables en `.env` (ver `.env.template`): `VITE_BACKEND_URL` (vacío = mismo origen), `VITE_DROPBOX_API_KEY`, `VITE_GOOGLE_WEB_CLIENT_ID`. Solo se leen en `src/config.ts`.

## Layout

```
src/main.tsx · src/router.tsx    sin sesión cualquier ruta muestra el login; con sesión AppLayout
src/pages/                       list, activity, search, settings, login, oauth-dropbox
src/components/layout/           AppLayout (arranca WS + sync + perfil), Sidebar, BottomNav, PageHeader
src/components/ui/               primitivas (Button, Modal, Menu, Switch, Slider, TextInput…)
src/components/<dominio>/        list (filas, dnd, modales), options, activity, history, upload
src/api/ · src/models/ · src/utils/ · src/i18n/locales/   espejo de móvil (ver abajo)
src/services/ · src/stores/      como en móvil, con las adaptaciones de plataforma
```

## Código compartido con móvil

- `api/types/*`, `api/{auth,books,uploads,users}-api.ts`, `api/api-error.ts`, `models/*` y `utils/format.ts` son copia de los de móvil. Un cambio de DTO o de lógica pura se hace en los dos. Única diferencia: `FileItem` guarda `file: File` en vez de `uri`.
- `es.json`/`en.json` parten de los de móvil más las claves de la web (`nav`, `list`, `md5`…).

## Diseño

- Fuente de verdad: frame **"Web"** de `docs/ideas.excalidraw` (escritorio, ≥1024 px). En tablet la sidebar queda en iconos y el panel de opciones pasa a modal; en teléfono hay barra de navegación inferior y se siguen los patrones del frame "Mobile".
- Colores solo con los tokens de `src/index.css` (`bg`, `fg`, `muted`, `line`, `strong`, `surface`, `subtle`, `primary`, `success`, `info`, `warning`, `danger`…), que cambian en tema oscuro (`data-theme` en `<html>`; "Sistema" sigue `prefers-color-scheme`). Nada de colores sueltos.
- Iconos solo con `Icon` (`components/icons/icons.ts`, mismos paths que móvil).
- Mensajes con `useToast`/`showError` en lugar de los `Alert` de la app.

## Plataforma (diferencias con móvil)

- Tokens y `deviceId` en `localStorage`; todas las pestañas comparten sesión (evento `storage`). El refresh, además de single-flight, va dentro de `navigator.locks` (`inkomi-refresh`) y relee el token: otra pestaña puede haberlo rotado ya.
- WebSocket: el navegador no manda cabeceras, así que el token va como subprotocolo `['bearer', token]` (lo acepta `back/common/httpauth` y el upgrader del transaction-manager lo devuelve). Solo la pestaña con el lock `inkomi-ws` abre el WS y sincroniza; las demás ven los cambios por la BD compartida (`BroadcastChannel`). El backoff solo se reinicia si la conexión duró >30 s (el servidor admite una conexión por usuario).
- Actividad en IndexedDB (`services/activity-db.ts`, misma API que la versión SQLite). El borrador de la lista se persiste en IndexedDB sin JSON para conservar los `File`. Preferencias en `localStorage` (el tema se aplica antes de React en `index.html`).
- Subida: `PUT` con `XMLHttpRequest` (progreso y `abort`); `beforeunload` avisa mientras sube. Necesita CORS del bucket R2 (`deploy/SERVERS.md`).
- Descarga: se pide el enlace en cada clic y se navega a él. No hay "Compartir".
- Google: Google Identity Services (botón propio de Google). Dropbox: PKCE con redirección completa a `/oauth/dropbox` (verifier e intención en `sessionStorage`).
- Notificaciones: `Notification` del navegador cuando un fichero pasa a `sent` por el WS, solo en la pestaña líder y sin foco. Sin FCM web.

## Despliegue

- `Dockerfile` (build con Node, nginx sin root en 8080 con fallback a `index.html`) y `deploy/k8s/web/web.yaml`; pipeline manual `.github/workflows/deploy-web.yml`. Mismo dominio que la API.

## Nombres y estilo

- Como en móvil: ficheros kebab-case, componentes PascalCase con export nombrado, stores `use-<dominio>.ts`. Prettier del proyecto. Commits con scope `web`: `feat(web): ...`.
