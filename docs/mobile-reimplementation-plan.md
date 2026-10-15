# Reimplementación de `front/mobile` (diseño Mobile de `docs/ideas.excalidraw` + microservicios)

## Context

La app actual (`front/mobile`, Expo SDK 54, expo-router, zustand) habla con el monolito legacy `back/api` (`/api/transactions/*`, polling, sin auth, token de Dropbox crudo en cada envío). El backend ya está migrado a microservicios detrás del ingress (`deploy/k8s/ingress.yaml`, **sin prefijo `/api`**) y el diseño móvil del excalidraw (frame "Mobile", 16 pantallas) define otra UX: lista única con agrupaciones, opciones manga por elemento/grupo, envío por URLs presignadas y actividad en tiempo real. Objetivo: reescribir la app para que cuadre con ese diseño y con los contratos reales, y redactar `front/mobile/CLAUDE.md`.

Decisiones del usuario:
- Se quita el "Añadir por MD5" manual. El buscador junto al botón de adjuntar usa `GET /books/search`; "+" añade el resultado a la lista como source `library` (md5 en `filename`).
- La cola ("En curso") y el "Historial" viven en SQLite local del móvil y se actualizan en tiempo real (WS + FCM). Para reconstruirlos/sincronizarlos (instalación nueva, otro dispositivo, eventos perdidos con la app cerrada o el WS caído) se **añade un endpoint de listado en transaction-manager** (`GET /uploads`).
- Se quitan "Reintentar" (historial) y la X de "En cola" (no hay endpoint).

## Paso 0 — Guardar este plan

Copiar este documento tal cual a `docs/mobile-reimplementation-plan.md` antes de empezar (junto a `docs/microservices-migration-plan.md`).

## Parte 1 — Backend: `GET /uploads` en transaction-manager

Siguiendo `back/CLAUDE.md` (hexagonal, un tipo/UC por fichero, `Wrap`, camelCase en DTOs nuevos).

**Contrato** (Bearer; el ingress ya enruta `/uploads`):
```
GET /uploads?limit=50&cursor=<opaco>&updatedSince=<RFC3339>
200 {
  uploads: UploadDTO[],       // envíos raíz (parentId null), más recientes primero (createdAt desc, id desc)
  nextCursor: string | null,  // siguiente página de historial; null = no hay más
  syncedAt: string            // hora del servidor (RFC3339) para el siguiente updatedSince
}
UploadDTO = SourceRow sin userId/configKey (mismos campos camelCase que el `result` del WS:
  id, title, filename, kepubify, toCloud, isRtl, join, status, errorMsg, size, isFile,
  createdAt, updatedAt, completedAt, parentId, position, isResult)
  + children: SourceRow[]     // items (position asc) y resultados de fusión (isResult=true)
```
- `limit`: 1–100, defecto 50 → 400 si inválido. `cursor` codifica `createdAt|id` (base64url) → 400 si inválido.
- `updatedSince`: devuelve solo los árboles en los que **algún nodo** (raíz o hijo) tiene `updatedAt >= updatedSince` (`>=` porque SQLite guarda segundos; el cliente hace upsert idempotente). Combinable con limit/cursor.
- Se excluyen los árboles abandonados: ningún nodo ha salido de `pending_upload` (request hecho pero nunca confirmado). Ojo: la raíz de una carpeta sin fusión se queda en `pending_upload` para siempre, así que el filtro mira a los hijos.
- Solo filas del usuario del token.

**Ficheros** (`back/transaction-manager/`):
- `internal/domain/models/source-tree.go` (`SourceTree{Source; Children []Source}`), `source-list-filter.go` (`SourceListFilter{Limit, Cursor *SourceCursor, UpdatedSince *time.Time}` con `NewSourceListFilter` que valida) y `source-cursor.go` (encode/decode); errores `ErrInvalidCursor`, `ErrInvalidLimit` en `errors.go`.
- `internal/domain/ports/source-repository.go`: `ListTreesByUserId(ctx, userId, filter) ([]models.SourceTree, error)` y `Now(ctx) (time.Time, error)` (o reutilizar `CURRENT_TIMESTAMP` dentro de la misma consulta) + actualizar `usecases/mocks/mock-source-repository.go`.
- `internal/infra/repositories/sqlite-sources-repository.go`: 1 consulta para las raíces paginadas (con `EXISTS` sobre hijos para `updatedSince` y el filtro de abandonados) + 1 consulta `WHERE parentId IN (...)` para los hijos; reutilizar `sourceColumns`/`sourceFields`. Índice nuevo en `migrations/`: `(userId, parentId, createdAt DESC)`.
- Corregir de paso: `UpdateStatus`/`UpdateErrorMsg` no rellenan `completedAt` en estados terminales (`done|sent|failed`); hacerlo como ya hace `UpsertResult`, para que el historial muestre la hora de fin.
- `internal/usecases/listuploads/list-uploads-uc.go` (`ListUploadsUC.Execute(ctx, userId, filter) (*models.SourcePage, error)` → árboles + `nextCursor` + `syncedAt`).
- `internal/infra/api/dtos/upload-list-response.go` (`UploadListResponse`, `UploadResponse`, `NewUploadListResponse`), `handlers/uploads-handler.go` → `HandleList` (parsea query, mapea errores a 400), `routes/upload-routes.go` → `r.Get("/", ...)`; cablear el UC en `main.go`.
- Tests: `list-uploads-uc_test.go` con mocks, test de repositorio (paginación, `updatedSince` disparado por un hijo, exclusión de abandonados, aislamiento por usuario, resultados de fusión como hijos) y ampliar `back/e2e/manga-flow_test.go` para comprobar que tras `sent` el envío aparece en `GET /uploads`.
- Documentar el endpoint en `docs/` si existe catálogo (el plan de migración) — solo una línea.

### `GET /uploads/{sourceId}/download` (enlace prefirmado para descargar/compartir)

El fichero final queda en R2 en `<id>/<filename>` (los workers borran solo los intermedios; push-service no borra nada), y `filename` se mantiene actualizado en la fila (`UpdateFile`/`UpsertResult`).
```
GET /uploads/{sourceId}/download   (Bearer)
200 { url: string, filename: string, size: number | null, expiresAt: string }
```
- Descargable = fila del usuario con `isFile` y `status ∈ {done, sent}` (los hijos de una carpeta fusionada acaban en `joined` y sus ficheros se borran → no descargables; sí lo son sus resultados `isResult`).
- Errores: 404 `source not found`; 409 `source not ready` (no terminal o `failed`/`joined`); 404 `file not available` si el `HEAD` a R2 falla (purgado). Errores centinela en `models/errors.go` y mapeo en el handler con `requtil.NewError`.
- `ports/cloud-storage.go`: añadir `GetDownloadUrl(id, filename string) (string, time.Time, error)`; `infra/cloud/cloudflare-r2-storage.go` con `PresignGetObject`, `ResponseContentDisposition: attachment; filename*=UTF-8''<filename>` y caducidad 15 min (constante compartida con la de PUT). Actualizar `mocks/mock-cloud-storage.go`.
- `internal/usecases/getdownload/get-download-uc.go` (`GetDownloadUC.Execute(ctx, userId, sourceId)`), DTO `download-response.go`, `HandleDownload` en `uploads-handler.go`, `r.Get("/{sourceId}/download", ...)` en `upload-routes.go`, cableado en `main.go`.
- Tests del UC (no encontrado, no listo, hijo `joined`, no disponible en R2, ok) y paso en el e2e: tras `sent`, pedir el enlace y hacer GET verificando que descarga el epub.

## Parte 2 — Móvil

## Contratos (fuente de verdad: backend)

Base `EXPO_PUBLIC_BACKEND_URL` (sin `/api`). Errores JSON `{error: string}`; 204 sin cuerpo; el middleware JWT responde 401 en **texto plano**.

**auth-service** (`back/auth-service/internal/infra/api/routes/*`, DTOs en `.../dtos/`):
- `POST /auth/register|login {email,password,deviceId}`, `POST /auth/google {idToken,deviceId}`, `POST /auth/dropbox {refreshToken,deviceId}` → `TokensResponse {accessToken, accessTokenExpiresAt, refreshToken, refreshTokenExpiresAt}` (RFC3339).
- `POST /auth/refresh {refreshToken}` (rotado, de un solo uso) · `POST /auth/logout {refreshToken}` → 204.
- Bearer: `GET /users/me` → `{id, email?, emailVerified, hasPassword, providers: ('google'|'dropbox')[], dropbox: {linked, email?, folder}}`; `PUT /users/me/fcm-token {token}` (`""` = borrar); `PUT|DELETE /users/me/dropbox {refreshToken}`; `PUT /users/me/dropbox/folder {folder}` → `{folder}`.
- Mensajes de error exactos en `back/auth-service/internal/infra/api/handlers/api-errors.go` → mapear a claves i18n.

**library-service**: `GET /books/search?q&lang&fmt=epub,cbz,pdf` (Bearer) → `{title, author, pages, language, extension, md5, coverUrl}[]`; `md5` puede ser `""` (no añadible); 400 `empty query`, 503 `no libgen mirror available`. Sin paginación ni tamaño.

**transaction-manager** (`back/transaction-manager/internal/infra/api/`):
- `POST /uploads/request` body `{ereader_key, sources: SourceDTO[], config: Record<key, ConfigDTO>}` → `SourceDTO[]` con `id` y `url` (solo `type:"file"`). SourceDTO snake_case: `title, type ('file'|'folder'|'library'), size?, filename, config_hash? (clave del mapa), reading_direction ('ltr'|'rtl'), to_cloud, should_join, items[]`. Hijos heredan dirección/to_cloud/config del padre.
- `PUT <url>` bytes crudos (R2, caduca en 15 min) → `PATCH /uploads/done/{id}` (204) por cada file e item library. **Nunca** sobre una fila folder.
- `ConfigDTO` camelCase (`config-dto.go`): `ereaderKey` (obligatorio), `rotateDoublePage, cropTolerance(-1|[0,1]), enlarge, embed, gamma[0.1,3], brightness[-1,1], contrast[0,2], grayscale, blackPointThreshold<whitePointThreshold, dithering, palette8bit, quality[1,100], sharpen{enabled,radius,x1,y2,y3,m1,m2}`. Validar en cliente con los rangos de `back/manga-worker/internal/domain/models/manga-config.go`.
- `GET /ws` con header `Authorization: Bearer` (RN: `new WebSocket(url, undefined, {headers})`). Una conexión por usuario, sin snapshot ni pings. Mensajes sin `type`: `{id,status,filename?,size?,error?}` · `{id,status,result: SourceRow}` (resultado de fusión) · `{id,status:'done'|'failed',error?}` (cierre de carpeta fusionada). Status: `pending_upload|queued|processing|joining|joined|kepubifying|done|sent|failed`.
- EReaderKey: Kindle `K1 K2 KDX K34 K57 KPW KV KPW34 K810 KO K11 KPW5 KS KCS`; Kobo (kepubify) `KoMT KoG KoGHD KoA KoAHD KoAH2O KoAO KoN KoC KoCC KoL KoLC KoF KoS KoE` (`back/transaction-manager/internal/domain/models/ereader.go`, con resoluciones).
- FCM data: `{type:'sent', id, filename, toCloud, parentId?}`.
- **Nuevo** `GET /uploads` (Parte 1) para sincronizar cola e historial, y `GET /uploads/{id}/download` para el enlace prefirmado de descarga.

### Sincronización local ↔ servidor
- SQLite local (`sources` con los campos de `UploadDTO` + `children` como filas con `parentId`; tabla `sync_state` con `syncedAt`, `historyCursor`, `historyExhausted`). Es la única fuente de las pantallas En curso / Historial.
- Al enviar: insertar los sources devueltos por `/uploads/request` (`pending_upload` → `queued` tras el PATCH).
- Tiempo real: eventos WS → upsert por `id` (forma A: status/filename/size/error; forma B: upsert del `result` como hijo; forma C: cierre de la carpeta). FCM `sent` en segundo plano → upsert del mismo modo.
- Reconciliación: al iniciar sesión con BD vacía → `GET /uploads?limit=50` (reconstrucción). En cada (re)conexión del WS y al volver a primer plano → **abrir WS primero** y luego `GET /uploads?updatedSince=<syncedAt>` (paginando con `nextCursor` hasta agotar), upsert de árboles completos y guardar el nuevo `syncedAt`. Historial: al llegar al final de lo local, pedir la siguiente página con `historyCursor`.
- Cerrar sesión borra la BD local (los datos son por usuario).

## Diseño → pantallas (expo-router)

Estilo del excalidraw: monocromo Open Color — texto/primario `#1e1e1e`, secundario `#868e96`, bordes `#ced4da`, superficies `#e9ecef`, fondo `#ffffff`; estados verde `#2f9e44` (enviado/subido), azul `#1971c2` (fusión/enlaces), naranja `#f08c00` (procesando/kepub), rojo `#e03131` (error). Iconos Material Symbols Rounded (ya hay `SIcon` + `icons.ts`; sustituir paths por: settings, history, progress_activity, attach_file, search, tune, expand_more, drag_indicator, chevron_left, close, checklist, radio_button_(un)checked, delete, drive_file_move). Cabeceras de sección en mayúsculas 11px gris, switches negros, botones primarios negros con texto blanco, modales centrados sobre scrim.

| Ruta | Pantalla | Contenido / comportamiento |
|---|---|---|
| `app/(auth)/login.tsx` | Accede a Inkomi | Pestañas Iniciar sesión / Registrarse (registro: "Mínimo 8 caracteres", "Crear cuenta"); "Continuar con Google" (`@react-native-google-signin/google-signin` → idToken → `/auth/google`); "Continuar con Dropbox" (PKCE existente → refresh token → `/auth/dropbox`, que además vincula la carpeta). Se muestra sin sesión o tras "Cerrar sesión". |
| `app/(app)/index.tsx` | Home "INKOMI" | Header: historial, en curso, ajustes. Botón adjuntar → sheet "Añadir" (Archivos: epub/cbz/pdf · Carpeta: crea agrupación con todos sus ficheros). Barra de búsqueda → `/search`. Lista: elementos sueltos (título, ext · tamaño, `tune` → Opciones) y agrupaciones (fondo gris, `settings` → modal Agrupación, `expand_more` plega; hijos indentados). Pie: switch Dropbox (`to_cloud`, valor inicial de Ajustes; si no vinculado → flujo de vinculación) + "SEND". Mantener pulsado → modo selección. |
| (modo selección en Home) | — | "N elementos seleccionados", X = cancelar selección, checklist = (de)seleccionar todos, radios; pie Delete (modal "¿Borrar N elementos?", solo quita de la lista) / Move (modal "Mover N elementos a": Sin agrupar, grupos existentes con nº archivos, "+ Nueva agrupación"). Asa de arrastre en hijo → Reordenar. |
| `app/(app)/group/[id]/reorder.tsx` | Reordenar | "<grupo> · N archivos", lista arrastrable con número de posición, "Ordenar por nombre", Guardar. |
| `app/(app)/options/[id].tsx` | Opciones | Título, Dirección (Izq→Der / Der→Izq), "Restablecer". Si cbz/pdf o grupo: IMAGEN (gamma, brillo, contraste, punto negro/blanco, tolerancia de recorte, calidad) y AJUSTES (rotar dobles, ampliar, rellenar=embed, escala de grises, dithering, paleta 8 bits, enfocar + sub-opciones sharpen plegables). epub: solo título y dirección. Reutilizada para "Manga por defecto" (Ajustes) y "Opciones manga (todos)" (grupo). |
| modal Agrupación | — | Título, "Fusionar en un libro" (`should_join`), dirección, "Opciones manga (todos)" → Opciones del grupo; Cancelar / Guardar. |
| `app/(app)/search.tsx` | Búsqueda | Input con X, chips Todos/EPUB/CBZ/PDF (→ `fmt`), filas (portada `coverUrl` o placeholder, título, "autor · idioma", chip extensión, páginas), botón ✓/+ (alterna añadir/quitar en la lista), barra inferior "N añadidos a la lista" + "Listo". Estados vacío/503. |
| modal Enviando | — | Progreso por fichero (Subido verde / % / En espera), "No cierres la app hasta terminar la subida", Cancelar (aborta lo no confirmado) / "Ver en cola" (→ En curso). Al terminar se vacía la lista. |
| `app/(app)/activity.tsx` | En curso | Tarjetas por envío no terminal: título, contador hijos `x/n`, línea de estado coloreada, barra y chips de paso Bajar(library)/Procesar · Fusionar · Kepub · Subir derivados del `status`; "En cola" para queued. Pie "Actualización en tiempo real". Sin X. |
| `app/(app)/history.tsx` | Historial | Agrupado HOY / AYER / ESTA SEMANA / ANTERIOR; título, "Fusionado · 120 MB · 10:32" / "N libros · MB · hora", estado Enviado (punto verde) o "Error: …" rojo. Sin Reintentar. Paginación local + `historyCursor` remoto. Envío de un solo fichero terminado: iconos Descargar y Compartir en la fila. Envío con varios ficheros (carpeta o resultados de fusión): tocar la fila abre el sheet Resultados. |
| sheet Resultados | — | Lista de ficheros descargables del envío (resultados `isResult` de una fusión, o hijos de una carpeta sin fusión) con nombre, tamaño y botones Descargar / Compartir; los no descargables (fallidos) muestran su error. |
| `app/(app)/settings.tsx` | Ajustes | CUENTA (email + Cerrar sesión), LECTOR (selector e-reader con "W × H px · kepubify automático"), DROPBOX (Conectado/Desconectar o Conectar, carpeta editable, "Guardar en Dropbox por defecto"), Notificaciones (switch → `PUT /users/me/fcm-token` con token o `""`), Idioma (es/en), Manga por defecto → Opciones, versión. |

## Arquitectura de código (`front/mobile/src`)

- `api/http-client.ts`: fetch con base URL, Bearer, parseo `{error}`/texto plano, refresh **single-flight** ante 401 (rotación de un solo uso), cierre de sesión si el refresh falla. Un módulo por servicio: `api/auth-api.ts`, `api/users-api.ts`, `api/books-api.ts`, `api/uploads-api.ts` (request, done, list), `api/ws-client.ts` (reconexión con backoff, reconecta en `AppState` active y tras refresh). Tipos de contrato en `api/types/*.ts` copiando nombres/casing exactos del backend.
- `models/`: `ListItem` (file local: uri, nombre, ext, size, configKey?, title, rtl; library: md5, title, ext) y `Group` (title, join, rtl, configKey?, itemIds ordenados); `MangaConfig` con defaults y validación de rangos; `ereaders.ts` (keys, nombre, resolución, kepub).
- `services/`: `session-storage` (SecureStore: tokens, `deviceId` uuid), `dropbox-oauth` (reutilizar PKCE de `src/services/auth-serivice.ts`, solo para obtener refresh token; renombrar el fichero), `google-signin`, `upload-service` (construye `UploadRequestDTO` desde la lista → `/uploads/request` → `FileSystem.createUploadTask` PUT binario con progreso, concurrencia limitada → `PATCH done`; items library solo PATCH; Android: copiar `content://` a caché con el módulo existente `modules/file-handler` y mantener vivo con `react-native-background-actions`), `activity-db` (expo-sqlite: tablas `sources` y `sync_state` descritas arriba, upsert idempotente por `id`, consultas de cola —árboles no terminales— e historial —terminales, paginado por `createdAt`—; reemplaza el esquema actual), `sync-service` (reconstrucción y reconciliación con `GET /uploads`, orquesta WS + foreground), `download-service` (`GET /uploads/{id}/download` → descarga a caché con `File.downloadFileAsync` de `expo-file-system`; **Compartir** = `Sharing.shareAsync` de `expo-sharing` (ya instalado); **Descargar** = guardar en una carpeta elegida por el usuario vía SAF en Android (`Directory.pickDirectoryAsync`, recordando la carpeta) y share sheet "Guardar en Archivos" en iOS; si el enlace caduca o da 409/404, mostrar el error y no reintentar en bucle), `push-service` (permiso, token FCM, `onTokenRefresh`, handler data `sent`).
- `stores/` (zustand + immer): `useSession`, `useProfile` (`/users/me`), `useList` (borrador persistido en AsyncStorage: items, grupos, selección, configs), `useSettings` (ereader, toCloud por defecto, idioma, manga por defecto), `useActivity` (lee SQLite, aplica eventos WS/FCM), `useSearch`.
- `components/`: primitivas `SText`, `SButton`, `SSwitch`, `SSlider`, `SSegmented`, `SChip`, `SModal`, `SSheet`, `SIcon`, `Header` con back; de dominio `ListRow`, `GroupRow`, `ActivityCard`, `StepChips`, `HistoryRow`, `BookResultRow`, `UploadProgressRow`.
- `theme/` (tokens del diseño: colores, radios, espaciado, tipografía Inter existente), `i18n/` (reescribir es/en con todos los textos del diseño; español como base).
- `app/_layout.tsx`: fuentes, i18n, hidrata sesión; guardia `(auth)` vs `(app)`; inicia WS + push al tener sesión.

## Limpieza

Eliminar lo ligado al monolito: `app/(tabs)/*`, `send-*.tsx`, `src/services/transaction-service.ts`, `transaction-request-dto.ts`, `useQueue`, `useSender`, `useLibgen`, `useMonitoredFolders`, `useVersionhecker`, modelos `transaction*`, componentes `queue/`, `senders/`, `home/`, `new-version-modal`, `modules/StorageManager.ts`, `react-native-dotenv`, check-mirror/app-version. Añadir dependencias: `@react-native-google-signin/google-signin` (+ plugin en `app.json`), slider (`@react-native-community/slider`), drag list (`react-native-draggable-flatlist` sobre gesture-handler/reanimated). `.env.tempate` → `.env.template` con `EXPO_PUBLIC_BACKEND_URL`, `EXPO_PUBLIC_DROPBOX_API_KEY`, `EXPO_PUBLIC_GOOGLE_WEB_CLIENT_ID`. Reescribir `README.md` (breve) y añadir script `typecheck` (`tsc --noEmit`).

## `front/mobile/CLAUDE.md`

En español, estilo de `back/CLAUDE.md` (reglas concisas): stack y comandos; layout de carpetas; contratos (sin `/api`, casing snake_case en uploads / camelCase en config, auth y WS; tipos en `src/api/types` como espejo del backend y quién es la fuente de verdad); reglas de red (todo por `http-client`, nunca fetch suelto; refresh single-flight; WS con header); flujo de envío (request → PUT → PATCH, nunca PATCH sobre folder); estado (zustand por dominio, SQLite para actividad, SecureStore para secretos, AsyncStorage solo preferencias/borrador); UI (tokens de `theme`, iconos solo vía `SIcon`, sin colores sueltos, textos siempre por i18n es/en); nombres (ficheros kebab-case, componentes PascalCase, hooks `useX`); referencia al frame "Mobile" de `docs/ideas.excalidraw` como fuente del diseño; sincronización (SQLite local como fuente de las pantallas, WS/FCM en tiempo real, `GET /uploads` para reconstruir/reconciliar, abrir WS antes de reconciliar); descarga/compartir solo vía enlace prefirmado de `GET /uploads/{id}/download` (nunca cachear la URL, caduca); limitaciones conocidas (una conexión WS por usuario, los ficheros intermedios de una fusión no son descargables).

## Verificación

0. Backend: `go test ./...` en `back/transaction-manager` y el e2e de `back/e2e` (requiere el stack local); `curl -H "Authorization: Bearer …" /uploads?limit=2` para ver paginación y `updatedSince`.
1. `npx tsc --noEmit` y `npx prettier --check .` en `front/mobile`.
2. Backend local (docker compose / e2e de `back/e2e`) y `npx expo run:android` (dev build; Google Sign-In y FCM no funcionan en Expo Go).
3. Flujo manual: registro → login → `/users/me` en Ajustes; vincular Dropbox y cambiar carpeta; añadir cbz sueltos + carpeta como agrupación con "Fusionar", ajustar Opciones, reordenar, mover/borrar en modo selección; buscar "quijote", añadir resultado; SEND → modal Enviando → En curso actualizándose por WS hasta `sent`; Historial persistido tras reiniciar; borrar datos de la app / instalar en otro dispositivo y comprobar que cola e historial se reconstruyen; matar la app durante un procesado y al volver comprobar que la reconciliación trae los cambios perdidos; desde Historial descargar y compartir un envío simple y, desde Resultados, uno fusionado; push recibido con la app en segundo plano; expirar access token (15 min) y comprobar refresh transparente; Cerrar sesión vuelve a Login.
