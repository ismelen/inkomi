# Backend (`back/`) — reglas para Go

El backend es un solo módulo Go y un solo proceso: `back/api` (módulo `github.com/ismelen/inkomi/back/api`), con cada área funcional como módulo interno. Se despliega con Docker Compose.

## Arquitectura

- Hexagonal. Dependencias solo `ui → usecases → domain` e `infra → domain` (`infra` implementa los puertos); `shared` lo usa cualquiera. `usecases` nunca importa `infra` ni `ui`.
- DI manual por constructor; la composition root es `main.go` + `wires/`. Sin frameworks de DI.
- Toda dependencia externa (BD, cola, cloud, push, HTTP externo) va detrás de un puerto (interface) en el dominio.
- Un tipo / caso de uso / helper por fichero.

## Layout (`back/api`)

```
main.go                         arranque: NATS embebido, jobs.db, router, servidor; llama a wires.<Módulo>
wires/                          paquete wires: un fichero por módulo (auth.go, manga.go…) + deps.go (Deps compartidas, TimeoutMiddleware, CloseLogged)
migrations/                     users.sql, transactions.sql, joins.sql (embed)
samples/                        samples de la preview (embed; también los copia la imagen e2e)
e2e/                            tests e2e (build tag e2e) y cmd/upload-samples; imagen propia (e2e/Dockerfile)
internal/config/                config.Load() → config.Env único (ver .env.example). No usar os.Getenv fuera de aquí.
internal/platform/              lo compartido entre módulos: jobs, jobrun, jobcancel, queue (+ NATS embebido), cloud, storage, httpauth, httplog, fsutil, resources, sqlitedb
internal/modules/<módulo>/      auth, library, transactions, manga, epub, push, release
```

Cada módulo se divide en estas carpetas (solo las que tengan contenido):

```
domain/models/      entidades, enums, errores centinela (errors.go)
domain/ports/       interfaces, una por fichero
usecases/<caso>/    un paquete por UC;  usecases/mocks/
infra/<adaptador>/  repositories, datasources, queue, cloud, sockethub, dropbox…
shared/             utilidades sin lógica de negocio
ui/                 entrada HTTP: dtos, handlers, routes, middlewares, requtil
```

- La cola es un NATS con JetStream embebido en el proceso (`platform/queue.StartEmbeddedNATS`, sin puertos, store en `DATA_DIR/nats`); los adaptadores reciben la `*nats.Conn` y no la cierran.
- Sin Redis: el estado de los joins de epub va en SQLite (`SQLiteJoinStore`, `jobs.db`).
- Un módulo usa otro en proceso a través de un adaptador en su `infra` que implementa su propio puerto (p. ej. `push/infra/usersettings` llama a los UCs de auth). Nada de HTTP interno ni tokens de servicio.
- Un único router chi en `PORT` con todas las rutas, más `GET /healthz`.
- Los adaptadores de `platform` reciben su configuración por parámetro (nunca leen `config.Env`). Sus interfaces (`IQueue`, `ICloudStorage`) se usan directamente como puertos, sin redeclararlas; cada paquete comprueba su implementación con `var _ IX = (*Impl)(nil)`.
- Archivos en el disco del servidor (`STORAGE=local`, por defecto): `platform/storage.LocalStorage` implementa `cloud.ICloudStorage` en `DATA_DIR/files/<id>/<filename>` con enlaces duros (los temporales de los workers están en el mismo volumen, `TMPDIR`), y los clientes suben y bajan por `/files/{id}/{filename}` con URLs firmadas (HMAC con caducidad y, al subir, el tamaño declarado). `STORAGE=r2` vuelve a R2.
- Cada envío reserva su pico de disco al pedirse (`models.EstimateStoragePeak` + `storagequota.DiskQuota`); si no cabe, 507 `insufficient storage`. La reserva se libera cuando todo su árbol termina. Cada paso borra su entrada en cuanto guarda su salida (cbz → epub → bloque → kepub); la unión borra los capítulos de cada bloque al guardarlo y un reintento reutiliza los bloques ya guardados.

## Deseados

- `wishlist`: libros y autores que quiere cada usuario, en `wishlist.db` (`migrations/wishlist.sql`), con autocompletado desde Open Library (`platform/openlibrary`, con un único límite de peticiones para todo el proceso). Los duplicados se detectan por la clave de Open Library o por título y autor normalizados. `GET /wishlist/works` pagina los libros de un autor (por clave o nombre), con orden y, si hay idioma, título y portada de la edición en ese idioma. Los detalles de cada obra deseada (portada, sinopsis, temas) se guardan en `wishlist_work` (compartida entre usuarios): se buscan al añadir (un libro escrito a mano toma la clave de la mejor coincidencia), `BackfillWorksUC` reintenta cada 6 h los que faltan y `GET /wishlist/works/{workKey}` los sirve.
- `library`: `GET /books/search` responde `{books, nextPage}`. Las páginas de libgen+ traen 25 resultados aunque se pidan más, así que `PlusMirror` lee 4 páginas por búsqueda y `SearchBooksUC` lee otra tanda si los filtros (idioma, formato) dejan menos de 10 libros.

## Nombres

- Ficheros `kebab-case.go`; paquetes en minúsculas, una palabra (`uploaddone`, `requtil`).
- `New<Tipo>` devuelve puntero (o `(*T, error)` si puede fallar). Casos de uso: `<X>UC` con `Execute(ctx, ...)`.
- Handlers: `<Recurso>Handler.Handle<Acción>`. Rutas: `Setup<Recurso>Routes`. DTOs: `<X>Request` / `<X>Response`.
- Toda interface empieza por `I` (puertos, interfaces de `platform`, internas de un adaptador y genéricas): `IQueue`, `ICloudStorage`, `ISourceRepository`, `IMirror`, `IFilter[T, R]`. Las implementaciones y los mocks usan el nombre sin la `I`: `<Tecnología><Nombre>` (`NatsQueue`, `SQLiteSourceRepository`) y `Mock<Nombre>` (`MockQueue`).
- Errores `Err<Causa>`; enums `<Tipo><Valor>`; constantes en camelCase/PascalCase, nunca `UPPER_SNAKE`.
- Acrónimos en mayúsculas en el código nuevo (`MD5`, `URL`), salvo `Id`: siempre `Id` / `xId` (`UserId`, `ParentId`, `GetByIdAndUserId`), nunca `ID`. Receptor de 1–2 letras, siempre puntero y el mismo en todo el tipo.
- Inglés en código, comentarios y mensajes de error. Revisa las erratas en nombres de paquetes y tipos.

## Dominio

- Entidades con `New<X>` que fija invariantes; la lógica de la entidad vive en sus métodos, no en los handlers.
- Enums: preferir `type X string` + constantes. Si hace falta orden, `int32`+`iota` con `String`/`MarshalJSON`/`UnmarshalJSON`/`IsValid`.
- Puertos sin comentario de cabecera (`// X is the port implemented by...`): el nombre ya lo explica. Solo se comentan los métodos con semántica no evidente (p. ej. qué devuelve en un caso especial).
- El estado compartido entre goroutines se encapsula en tipos seguros (atomic, mutex); no se exponen los campos crudos.

## Casos de uso

- Las dependencias son campos privados con tipo de puerto. `ctx` siempre es el primer parámetro.
- Nada de `http.Request`, SQL ni SDKs. Un UC puede recibir otro UC por constructor.
- Flujo común con un paso variable: Template Method por embedding (`BaseTransactionUC` + `processor`).
- El archivo que contiene el método `Execute` y la declaración del struct de UC debe llamarse `<action>-uc.go` y el test `<action>-uc_test.go`

## HTTP (`chi`)

- Handler: `func(r *http.Request) (*R, error)`, registrado con `requtil.Wrap[R]`. Wrap se encarga de: error → `{"error": msg}` con el status de `*ApiError` o 500; `nil` → 204; `FileResponse` → descarga; resto → 200 JSON.
- Errores de cliente con `requtil.NewError(http.StatusBadRequest, "...")`.
- El handler solo parsea, valida la forma, llama al UC y mapea al DTO. Mapeos: `dto.To<Modelo>()` y `New<X>Response(model)`.
- Tags JSON: un único estilo por módulo, el que espera el front (los módulos nuevos usan camelCase).
- Para trabajo en background desde un handler: `context.WithoutCancel(r.Context())` + goroutine.

## Patrones existentes (reutilizar antes de inventar)

`shared/filter` (Chain of Responsibility genérico: embeber `filter.Base[T,R]`, encadenar con `filter.Use`), Strategy detrás de puertos (mirrors de libgen),
callbacks/Observer (`SetOnUpdate`, `OnFreeSpace`), Builders de epub, Worker Pool (`imagev2`), `singleflight`, genéricos (`Wrap[R]`, `Queue[T]`, `AtomicList[T]`).
Usar genéricos solo cuando eliminan duplicación real.
Workers: los UCs exponen `Execute(ctx, msg jobs.Msg, delivery jobrun.Delivery) error` y `wires` los suscribe con `jobrun.Handle(queue, uc.Execute)`, que parsea el mensaje y publica `job.step.error` (errores `jobrun.Permanent` al momento; el resto en el último intento).

## Errores, concurrencia, observabilidad

- Nada de `panic` salvo en el arranque desde `main`/`wires` (`log.Fatal`). Envolver con `fmt.Errorf("ctx: %w", err)` y comparar con `errors.Is/As`.
- Mensajes de error en minúscula y sin punto final. Nunca ignorar errores en silencio: `_ = f()` con un comentario.
- Estado compartido con `atomic`, `sync.Map` o `RWMutex` + `defer Unlock()`. No llamar callbacks con un lock tomado. Los getters devuelven copias (`slices.Clone`).
- Las goroutines largas y los `select` sobre canales respetan `ctx.Done()`. Los procesos periódicos usan `Start(ctx, interval)` con `time.Ticker`.
- Sin tracing (OTel).

## Tests

- testify: `require` para las precondiciones, `assert` para el resto. Nombre: `Test<Tipo>_<Método>_<Escenario>_<Resultado>`.
- Comentarios `// Arrange` `// Act` `// Assert`; `t.Parallel()`; table-driven con `t.Run`; `t.Helper()`; `t.TempDir()`.
- Paquete externo `<pkg>_test` por defecto.
- Mocks escritos a mano (sin generadores) con campos `XxxFn` y un comportamiento por defecto inocuo.
- Todo UC nuevo lleva tests. Los adaptadores se prueban con dependencias embebidas (SQLite, NATS embebido, ficheros reales).
- Antes de terminar, en `back/api`: `gofmt -l .`, `go vet ./...`, `go vet -tags e2e ./e2e/...` y `go test ./... -race`.

## Despliegue (Docker Compose)

- `back/api/Dockerfile` (contexto `back/api/`, cgo + libvips; la pipeline solo construye `linux/arm64`, en un runner ARM nativo, porque el servidor es ARM) con las etapas `test` y `runtime`; `back/api/.env.example` con todas las variables (una variable nueva va también ahí).
- `back/api/e2e/Dockerfile` (contexto `back/api/`, sin cgo): `e2e.test` y `upload-samples`.
- `back/api/docker-compose.yml` (proyecto `inkomi-api`: `api` + `web` en la red `caddy_net`, perfil `e2e`), desplegado por `deploy/api/deploy.sh` desde `.github/workflows/deploy-api.yml`: e2e en un proyecto compose desechable y, si pasan, producción con healthcheck y rollback. Servidor: `deploy/SERVERS.md`.
- Desarrollo local: `go run .` (o `air`) en `back/api` con libvips instalado; la web con `npm run dev` (proxy a `localhost:3000`).

## Commits

Conventional Commits con `api` o el módulo como scope: `feat(api): ...`, `fix(transactions): ...`, `feat(manga): ...`.

## Comments

Only put comments when its necessary if the code itself isn't explanatory.
