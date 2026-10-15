# Backend (`back/`) — reglas para Go

## Arquitectura

- Hexagonal. Dependencias solo `infra → usecases → domain`; `shared` lo usa cualquiera. `usecases` nunca importa `infra`.
- DI manual por constructor; `main.go` es la única composition root. Sin frameworks de DI.
- Toda dependencia externa (BD, cola, cloud, push, HTTP externo) va detrás de un puerto (interface) en el dominio.
- Un tipo / caso de uso / helper por fichero.

## Layout (servicios nuevos; `api/` usa `domain/<contexto>/` con los puertos junto a las entidades)

```
main.go · go.mod (github.com/ismelen/inkomi/back/<svc>) · Dockerfile · migrations/
internal/config/            config.Load() → config.Env (AppConfig tipado, godotenv). No usar os.Getenv fuera de aquí.
internal/domain/models/     entidades, enums, errores centinela (errors.go)
internal/domain/ports/      interfaces, una por fichero
internal/usecases/<caso>/   un paquete por UC;  internal/usecases/mocks/
internal/infra/api/{dtos,handlers,routes,middlewares,requtil,validation}
internal/infra/<adaptador>/ queue, cloud, repositories, datasources, sockethub…
internal/shared/            utilidades sin lógica de negocio
```

## Nombres

- Ficheros `kebab-case.go`; paquetes en minúsculas, una palabra (`uploaddone`, `requtil`).
- `New<Tipo>` devuelve puntero (o `(*T, error)` si puede fallar). Casos de uso: `<X>UC` con `Execute(ctx, ...)`.
- Handlers: `<Recurso>Handler.Handle<Acción>`. Rutas: `Setup<Recurso>Routes`. DTOs: `<X>Request` / `<X>Response`.
- Puertos sin prefijo `I` (`CloudStorage`, `Queue`). Implementaciones: `<Tecnología><Puerto>` (`NatsQueue`, `SQLiteSourceRepository`). Mocks: `Mock<Puerto>`.
- Errores `Err<Causa>`; enums `<Tipo><Valor>`; constantes en camelCase/PascalCase, nunca `UPPER_SNAKE`.
- Acrónimos en mayúsculas en el código nuevo (`MD5`, `URL`), salvo `Id`: siempre `Id` / `xId` (`UserId`, `ParentId`, `GetByIdAndUserId`), nunca `ID`. Receptor de 1–2 letras, siempre puntero y el mismo en todo el tipo.
- Inglés en código, comentarios y mensajes de error. Revisa las erratas en nombres de paquetes y tipos.

## Dominio

- Entidades con `New<X>` que fija invariantes; la lógica de la entidad vive en sus métodos, no en los handlers.
- Enums: preferir `type X string` + constantes. Si hace falta orden, `int32`+`iota` con `String`/`MarshalJSON`/`UnmarshalJSON`/`IsValid`.
- Puerto con comentario: `// X is the port implemented by infra/<pkg> to ...`.
- El estado compartido entre goroutines se encapsula en tipos seguros (atomic, mutex); no se exponen los campos crudos.

## Casos de uso

- Las dependencias son campos privados con tipo de puerto. `ctx` siempre es el primer parámetro.
- Nada de `http.Request`, SQL ni SDKs. Un UC puede recibir otro UC por constructor.
- Flujo común con un paso variable: Template Method por embedding (`BaseTransactionUC` + `processor`).

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

## Errores, concurrencia, observabilidad

- Nada de `panic` salvo en el arranque desde `main` (`log.Fatal`). Envolver con `fmt.Errorf("ctx: %w", err)` y comparar con `errors.Is/As`.
- Mensajes de error en minúscula y sin punto final. Nunca ignorar errores en silencio: `_ = f()` con un comentario.
- Estado compartido con `atomic`, `sync.Map` o `RWMutex` + `defer Unlock()`. No llamar callbacks con un lock tomado. Los getters devuelven copias (`slices.Clone`).
- Las goroutines largas y los `select` sobre canales respetan `ctx.Done()`. Los procesos periódicos usan `Start(ctx, interval)` con `time.Ticker`.
- OTel: cada `Execute`/`Process` abre un span `"<Tipo>.<Método>"` con `defer span.End()`. El tracer name va en una constante del servicio. En caso de error: `span.RecordError(err)` + `span.SetStatus(codes.Error, err.Error())`.

## Tests

- testify: `require` para las precondiciones, `assert` para el resto. Nombre: `Test<Tipo>_<Método>_<Escenario>_<Resultado>`.
- Comentarios `// Arrange` `// Act` `// Assert`; `t.Parallel()`; table-driven con `t.Run`; `t.Helper()`; `t.TempDir()`.
- Paquete externo `<pkg>_test` por defecto.
- Mocks escritos a mano (sin generadores) con campos `XxxFn` y un comportamiento por defecto inocuo.
- Todo UC nuevo lleva tests. Los adaptadores se prueban con dependencias embebidas (SQLite, NATS embebido, ficheros reales).
- Antes de terminar: `gofmt -l .`, `go vet ./...` y `go test ./... -race` en el servicio tocado.

## Deuda conocida en `api/` (no copiar; corregir al tocar)

- `usecases` importa `infra/epub` (hay que crear un puerto).
- `panic` en `NewTransactionHandler`.
- Strings mágicos `"md5"/"epub"/"cbz"`.
- `otel.Tracer("inkomi-api")` repetido en cada UC.
- Errores ignorados.
- `fmt.Errorf` sin `%w`, con mensajes en mayúscula o en español.
- Receptores valor/puntero mezclados.
- `AtomicList.GetAll` expone el slice interno.
- `MAX_CHUNK_SIZE`, el alias de paquete `booksFilter` y el tipo `TransactionsV2Handler`.
- Structs inicializados por posición.
- Reintentos por recursión.

## Commits

Conventional Commits con el servicio como scope: `feat(epub-worker): ...`, `fix(transaction-manager): ...`.

## Comments

Only put comments when its necessary if the code itself isn't explanatory.
