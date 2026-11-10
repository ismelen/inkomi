# Backend (`back/`) — reglas para Go

El backend existe en dos versiones con la misma funcionalidad, rutas y contratos:

- `microservices/`: un módulo Go por servicio (`auth-service`, `library-service`, `transaction-manager`, `manga-worker`, `epub-worker`, `push-service`, `release-service`), más `common` y `e2e`. Se despliega en k3s.
- `monolith/`: un solo módulo y un solo proceso con los mismos servicios como módulos internos. Se despliega con Docker Compose.

Son **copias independientes**: el monolito no importa nada de `microservices/`. Un cambio de comportamiento se hace en las dos (y el e2e de `microservices/e2e` se pasa contra las dos).

## Arquitectura

- Hexagonal. Dependencias solo `infra → usecases → domain`; `shared` lo usa cualquiera. `usecases` nunca importa `infra`.
- DI manual por constructor; `main.go` es la única composition root. Sin frameworks de DI.
- Toda dependencia externa (BD, cola, cloud, push, HTTP externo) va detrás de un puerto (interface) en el dominio.
- Un tipo / caso de uso / helper por fichero.

## Layout de un microservicio

```
main.go · go.mod (github.com/ismelen/inkomi/back/microservices/<svc>) · Dockerfile · migrations/
internal/config/            config.Load() → config.Env (AppConfig tipado, godotenv). No usar os.Getenv fuera de aquí.
internal/domain/models/     entidades, enums, errores centinela (errors.go)
internal/domain/ports/      interfaces, una por fichero
internal/usecases/<caso>/   un paquete por UC;  internal/usecases/mocks/
internal/infra/api/{dtos,handlers,routes,middlewares,requtil,validation}
internal/infra/<adaptador>/ queue, cloud, repositories, datasources, sockethub…
internal/shared/            utilidades sin lógica de negocio
```

## Layout del monolito (`back/monolith`, módulo `github.com/ismelen/inkomi/back/monolith`)

```
main.go, wire-<módulo>.go       composition root: un fichero por módulo; deps compartidas en main.go
migrations/                     users.sql, transactions.sql, joins.sql (embed)
internal/config/                config.Load() → config.Env único (la unión de variables, ver .env.example)
internal/platform/              equivalente a common: jobs, jobrun, queue (+ NATS embebido), cloud, httpauth, fsutil, sqlitedb
internal/modules/<módulo>/      auth, library, transactions, manga, epub, push, release; cada uno con el layout de un microservicio (domain, usecases, infra, shared) sin config
```

- La cola es un NATS con JetStream embebido en el proceso (`platform/queue.StartEmbeddedNATS`, sin puertos, store en `DATA_DIR/nats`); los adaptadores reciben la `*nats.Conn` y no la cierran.
- Sin Redis: el estado de los joins de epub va en SQLite (`SQLiteJoinStore`, `jobs.db`).
- Un módulo usa otro en proceso a través de un adaptador en su `infra` que implementa su propio puerto (p. ej. `push/infra/usersettings` llama a los UCs de auth). Nada de HTTP interno ni tokens de servicio.
- Un único router chi en `PORT` con las mismas rutas que el Ingress, más `GET /healthz`.
- Archivos en el disco del servidor (`STORAGE=local`, por defecto): `platform/storage.LocalStorage` implementa `cloud.ICloudStorage` en `DATA_DIR/files/<id>/<filename>` con enlaces duros (los temporales de los workers están en el mismo volumen, `TMPDIR`), y los clientes suben y bajan por `/files/{id}/{filename}` con URLs firmadas (HMAC con caducidad y, al subir, el tamaño declarado). `STORAGE=r2` vuelve a R2.
- Cada envío reserva su pico de disco al pedirse (`models.EstimateStoragePeak` + `storagequota.DiskQuota`); si no cabe, 507 `insufficient storage`. La reserva se libera cuando todo su árbol termina. Cada paso borra su entrada en cuanto guarda su salida (cbz → epub → bloque → kepub); la unión borra los capítulos de cada bloque al guardarlo y un reintento reutiliza los bloques ya guardados.

## Módulo compartido de los microservicios (`back/microservices/common`)

- Módulo `github.com/ismelen/inkomi/back/microservices/common`: cualquier código que necesiten varios servicios. Contiene los mensajes y subjects de la cola (`jobs.Msg`, `jobs.Subject`), el consumo de jobs (`jobrun`: `Delivery`, `DeliveryHandler`, `Handle`, `Permanent`/`For`) y lo compartido de infraestructura: `queue` (`IQueue` + `NatsQueue` sobre JetStream), `cloud` (`ICloudStorage` + `CloudflareR2Storage` con claves `id/filename`; además de los streams expone `UploadFile`/`DownloadFile` para ficheros locales), `httpauth` (JWT RS*), `fsutil`. Las dependencias externas solo en esos subpaquetes de adaptador; lo específico de un servicio vive en ese servicio.
- Los adaptadores de `common` reciben su configuración por parámetro (nunca leen el `config.Env` de un servicio). Las interfaces de `common` (`IQueue`, `ICloudStorage`) son los puertos de cola y storage de los servicios: se usan directamente, sin redeclararlas en `domain/ports`. Cada paquete comprueba su implementación con `var _ IX = (*Impl)(nil)`.
- Cada servicio lo importa con `require github.com/ismelen/inkomi/back/microservices/common v0.0.0` + `replace github.com/ismelen/inkomi/back/microservices/common => ../common`. Sin `go.work`.
- La traducción a tipos propios del servicio (p. ej. subject → `SourceStatus`) vive en el servicio, no en `common`.
- El Dockerfile de un servicio que lo use se construye con contexto `back/microservices/` para poder copiar `common/`.

## Nombres

- Ficheros `kebab-case.go`; paquetes en minúsculas, una palabra (`uploaddone`, `requtil`).
- `New<Tipo>` devuelve puntero (o `(*T, error)` si puede fallar). Casos de uso: `<X>UC` con `Execute(ctx, ...)`.
- Handlers: `<Recurso>Handler.Handle<Acción>`. Rutas: `Setup<Recurso>Routes`. DTOs: `<X>Request` / `<X>Response`.
- Toda interface empieza por `I` (puertos, interfaces de `common`, internas de un adaptador y genéricas): `IQueue`, `ICloudStorage`, `ISourceRepository`, `IMirror`, `IFilter[T, R]`. Las implementaciones y los mocks usan el nombre sin la `I`: `<Tecnología><Nombre>` (`NatsQueue`, `SQLiteSourceRepository`) y `Mock<Nombre>` (`MockQueue`).
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
- Tags JSON: un único estilo por servicio, el que espera el front (los servicios nuevos usan camelCase).
- Para trabajo en background desde un handler: `context.WithoutCancel(r.Context())` + goroutine.

## Patrones existentes (reutilizar antes de inventar)

`shared/filter` (Chain of Responsibility genérico: embeber `filter.Base[T,R]`, encadenar con `filter.Use`), Strategy detrás de puertos (mirrors de libgen),
callbacks/Observer (`SetOnUpdate`, `OnFreeSpace`), Builders de epub, Worker Pool (`imagev2`), `singleflight`, genéricos (`Wrap[R]`, `Queue[T]`, `AtomicList[T]`).
Usar genéricos solo cuando eliminan duplicación real.
Workers: los UCs exponen `Execute(ctx, msg jobs.Msg, delivery jobrun.Delivery) error` y `main` los suscribe con `jobrun.Handle(queue, uc.Execute)`, que parsea el mensaje y publica `job.step.error` (errores `jobrun.Permanent` al momento; el resto en el último intento).

## Errores, concurrencia, observabilidad

- Nada de `panic` salvo en el arranque desde `main` (`log.Fatal`). Envolver con `fmt.Errorf("ctx: %w", err)` y comparar con `errors.Is/As`.
- Mensajes de error en minúscula y sin punto final. Nunca ignorar errores en silencio: `_ = f()` con un comentario.
- Estado compartido con `atomic`, `sync.Map` o `RWMutex` + `defer Unlock()`. No llamar callbacks con un lock tomado. Los getters devuelven copias (`slices.Clone`).
- Las goroutines largas y los `select` sobre canales respetan `ctx.Done()`. Los procesos periódicos usan `Start(ctx, interval)` con `time.Ticker`.
- Sin tracing (OTel) en ninguna de las dos versiones.

## Tests

- testify: `require` para las precondiciones, `assert` para el resto. Nombre: `Test<Tipo>_<Método>_<Escenario>_<Resultado>`.
- Comentarios `// Arrange` `// Act` `// Assert`; `t.Parallel()`; table-driven con `t.Run`; `t.Helper()`; `t.TempDir()`.
- Paquete externo `<pkg>_test` por defecto.
- Mocks escritos a mano (sin generadores) con campos `XxxFn` y un comportamiento por defecto inocuo.
- Todo UC nuevo lleva tests. Los adaptadores se prueban con dependencias embebidas (SQLite, NATS embebido, ficheros reales).
- Antes de terminar: `gofmt -l .`, `go vet ./...` y `go test ./... -race` en el servicio tocado.

## Despliegue de los microservicios (k3s)

- Cada servicio tiene su `Dockerfile` (contexto `back/microservices/`, multi-arch amd64/arm64) con las etapas `test` (vet + tests) y `runtime` (alpine, usuario sin root, `/app/.env` montado desde el Secret `<svc>-env`), y su `.env.example` con todas las variables. Una variable nueva va también al `.env.example`.
- Manifiestos en `deploy/k8s/` (renderizados con envsubst por `deploy/scripts/deploy.sh`); pipeline manual en `.github/workflows/deploy-microservices.yml`; preparación de los nodos en `deploy/SERVERS.md`.
- Los nodos llevan la etiqueta `inkomi.io/capacity` (`high`/`low`): los servicios que procesan imágenes prefieren `high` y el resto `low` (siempre `preferredDuringScheduling`, nunca obligatorio, para que se coloquen donde quepan).
- Un servicio nuevo: añadirlo a `ALL_SERVICES` en `deploy/scripts/deploy.sh`, a la lista del job `plan` del workflow, un fichero `deploy/k8s/app/<servicio>.yaml` (y una ruta en `ingress.yaml` si es público).
- Las direcciones dentro del clúster (NATS, Redis, auth-service interno, rutas de BD) se fijan como `env` en el manifiesto; godotenv no sobrescribe el entorno, así que mandan sobre el `.env`.
- Los tests e2e/integración viven en `back/microservices/e2e` (build tag `e2e`) y se ejecutan como Job en un namespace sandbox del clúster antes de tocar producción.

## Despliegue del monolito (Docker Compose)

- `back/monolith/Dockerfile` (contexto `back/monolith/`, cgo + libvips; la pipeline solo construye `linux/arm64`, en un runner ARM nativo, porque el servidor es ARM) con las etapas `test` y `runtime`; `back/monolith/.env.example` con todas las variables (una variable nueva va también ahí).
- `back/monolith/docker-compose.yml` (monolith + web en la red `caddy_net`, perfil `e2e`), desplegado por `deploy/monolith/deploy.sh` desde `.github/workflows/deploy-monolith.yml`: e2e en un proyecto compose desechable y, si pasan, producción con healthcheck y rollback. Servidor: `deploy/SERVERS.md` §12.
- Desarrollo local: `go run .` (o `air`) en `back/monolith` con libvips instalado, y la web con `DEV_BACKEND=http://localhost:3000`.

## Commits

Conventional Commits con el servicio como scope: `feat(epub-worker): ...`, `fix(transaction-manager): ...`, `feat(monolith): ...`.

## Comments

Only put comments when its necessary if the code itself isn't explanatory.
