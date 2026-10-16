# App móvil (`front/mobile`) — reglas

## Stack y comandos

- Expo SDK 54 (dev build, nada de Expo Go: Google Sign-In, FCM y los módulos nativos lo necesitan), React Native 0.81, TypeScript estricto, expo-router 6, zustand 5 + immer, expo-sqlite, i18next.
- `npm start` (Metro) · `npx expo run:android` (dev build) · `npm run typecheck` (`tsc --noEmit`) · `npx prettier --check .`.
- Variables en `.env` (ver `.env.template`): `EXPO_PUBLIC_BACKEND_URL` (sin `/api`), `EXPO_PUBLIC_DROPBOX_API_KEY`, `EXPO_PUBLIC_GOOGLE_WEB_CLIENT_ID`. Solo se leen en `src/config.ts`.
- Entrada `index.ts`: registra el handler de FCM en background antes de `expo-router/entry`.

## Layout

```
app/_layout.tsx            fuentes, i18n, hidrata la sesión; Stack.Protected (auth) vs (app)
app/(auth)/login.tsx       login / registro / Google / Dropbox
app/(app)/_layout.tsx      con sesión: arranca WS + sync + push + perfil; monta SendingModal
app/(app)/*.tsx            index (Home), search, activity (En curso), history, settings,
                           options/[id] (item | grupo | "default"), group/[id]/reorder
app/oauth.tsx              redirect del OAuth de Dropbox
src/api/                   http-client, ws-client, un módulo por servicio, api-error
src/api/types/             espejo de los DTOs del backend
src/models/                tipos y lógica pura (lista, manga-config, ereaders, upload-tree, activity-view)
src/services/              SQLite, sync, subida, descarga, push, OAuth, pickers (sin React)
src/stores/                zustand por dominio (use-session, use-list, use-activity…)
src/components/ui/         primitivas (SText, SButton, SIcon, SModal, SSheet…)
src/components/<dominio>/  list, activity, history, books, upload, options, settings
src/theme/ · src/i18n/ · src/utils/
modules/file-handler/      módulo nativo Android: copia content:// a la caché
```

## Diseño

- Fuente de verdad: frame **"Mobile"** de `docs/ideas.excalidraw`. Monocromo Open Color; estados verde (enviado/subido), azul (fusión/enlaces), naranja (procesando/kepub), rojo (error).
- Colores, radios, espaciado y tipografía solo desde `src/theme`. Nada de colores ni tamaños de fuente sueltos en componentes.
- Iconos solo con `SIcon` (Material Symbols Rounded, viewBox `0 -960 960 960`). Un icono nuevo se añade a `src/components/icons/icons.ts` copiando el path de `google/material-design-icons/symbols/web/<name>/materialsymbolsrounded/<name>_24px.svg`.
- Textos siempre con i18n (`useTranslation`/`i18n.t`), en `es.json` (base y fallback) y `en.json` a la vez. Plurales con `_one`/`_other`.

## Contratos (el backend manda)

- Sin prefijo `/api`: el ingress enruta `/auth`, `/users`, `/books`, `/uploads`, `/ws` y `/app` (release-service, público).
- `src/api/types/*` copia nombres y casing exactos de los DTOs Go y cita el fichero de origen. Si cambia un DTO, se cambia aquí. Casing: **snake_case** en el body de `POST /uploads/request` (`SourceDTO`); **camelCase** en `ConfigDTO`, auth, users, `GET /uploads`, descarga y mensajes del WS.
- Errores `{error: string}`; el middleware JWT responde 401 en **texto plano**. Los mensajes exactos de los centinelas se traducen en `api/api-error.ts` (`errorKey`) a claves `errors.*`; un centinela nuevo del backend añade su entrada y su texto.
- Rangos de la configuración manga: `back/manga-worker/internal/domain/models/manga-config.go` (reflejados en `models/manga-config.ts`, que valida antes de enviar).

## Red

- Todo HTTP pasa por `src/api/http-client.ts`; nunca `fetch` suelto (excepto el intercambio OAuth con Dropbox, que no es nuestro backend).
- Refresh **single-flight**: el refresh token rota y es de un solo uso; dos refresh en paralelo cierran la sesión. Usar `refreshAccessToken`/`getValidAccessToken`, nunca llamar a `/auth/refresh` a mano. Un refresh rechazado (401) borra los tokens y la app vuelve al login.
- El WebSocket (`ws-client`) se autentica con la cabecera `Authorization: Bearer` (tercer argumento del `WebSocket` de RN), reconecta con backoff, se cierra en background y se reabre al volver a primer plano.

## Envío

- `POST /uploads/request` → `PUT` binario a cada `url` (presignada, caduca en 15 min) → `PATCH /uploads/done/{id}` por cada fichero y cada item `library`. **Nunca** PATCH sobre una fila folder.
- Los hijos de una carpeta heredan dirección, `to_cloud` y config del padre. Un item manga solo manda `config_hash` propio si tiene opciones propias.
- Android: los `content://` se copian a la caché con `modules/file-handler` antes de subir, y la subida corre dentro de `react-native-background-actions` (`services/background-task.ts`).
- Los ítems confirmados salen de la lista; los fallidos se quedan.

## Estado

- zustand por dominio (`use-<dominio>.ts`), con immer si hay mutaciones anidadas. La lógica que no es de React va en `services/` o `models/`.
- **SQLite** (`services/activity-db.ts`) para la actividad: única fuente de En curso e Historial.
- **SecureStore** para secretos (tokens, `deviceId`). **AsyncStorage** solo para preferencias (`use-settings`) y el borrador de la lista (`use-list`), vía `persist`.
- Cerrar sesión (o perderla) borra SQLite y el borrador: los datos son por usuario.

## Sincronización

- Tiempo real: eventos del WS (estado, `result` de fusión, cierre de carpeta) y push FCM `sent` → upsert idempotente por `id` en SQLite. El estado nunca retrocede (`statusRank`), así que una respuesta vieja no deshace un evento nuevo.
- `GET /uploads` para reconstruir (BD vacía: primera página, guarda `historyCursor`) y reconciliar (`updatedSince=<syncedAt>`, paginando con `nextCursor`, guarda el `syncedAt` de la primera respuesta).
- **Abrir el WS primero y reconciliar en su `onOpen`**: así no se pierde nada entre la consulta y la suscripción.
- Un evento de un `id` desconocido dispara una reconciliación. Las subidas nunca confirmadas se purgan a los 20 min.
- Una carpeta sin fusión se queda en `pending_upload` en el servidor: su estado se deriva de los hijos (`models/upload-tree.ts`).

## Descargas

- Solo con el enlace prefirmado de `GET /uploads/{id}/download`, pedido en cada toque: caduca, nunca se cachea la URL. 409/404 se muestran y no se reintentan en bucle.
- Compartir = `Sharing.shareAsync`. Descargar = carpeta SAF elegida y recordada (Android) / share sheet (iOS).

## Actualizaciones de la app (sin Google Play)

- `.github/workflows/build-mobile.yml` compila el APK y publica una GitHub Release `v<expo.version>` con el asset `inkomi.apk`; solo publica si `expo.version` de `app.json` es mayor que la última release, así que cada versión nueva sube `version`.
- `back/release-service` expone `GET /app/latest` (versión, notas, tamaño, sha256) y `GET /app/download` (302 al APK). Sin token: se comprueba antes de iniciar sesión.
- `services/app-update.ts` compara con `appVersion` al arrancar y cada 6 h al volver a primer plano, descarga el APK y abre el instalador (`REQUEST_INSTALL_PACKAGES`). Solo Android.
- Todas las builds se firman con la misma clave; si cambia, Android no instala encima y hay que desinstalar.

## Nombres y estilo

- Ficheros kebab-case (`list-row.tsx`, `use-list.ts`); componentes PascalCase con export nombrado; hooks/stores `useX`; constantes camelCase.
- Prettier del proyecto (single quotes, 100 columnas). Comentarios solo cuando el código no se explica solo.
- Commits Conventional con scope `mobile`: `feat(mobile): ...`.

## Limitaciones conocidas

- El servidor mantiene una sola conexión WS por usuario: otro dispositivo conectado desplaza a este (la reconciliación compensa).
- Los ficheros intermedios de una fusión (hijos `joined`) se borran: solo son descargables los resultados `isResult`.
- No hay cancelación ni reintento de envíos en el servidor: "Cancelar" solo aborta lo no confirmado.
