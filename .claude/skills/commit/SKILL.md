---
name: commit
description: Cómo hacer commits en este repo (autoría, sin coautor de Claude, Conventional Commits, cómo dividirlos). Úsala siempre antes de crear un commit o un PR en inkomi.
---

# Commits en inkomi

## Autoría

- Los commits van con la cuenta git del usuario tal como está configurada (`git config user.name` / `user.email`). No uses `--author`, `-c user.*` ni variables `GIT_AUTHOR_*`/`GIT_COMMITTER_*`.
- **Sin coautor:** no añadas `Co-Authored-By: Claude ...` ni ninguna otra línea de atribución o firma de Claude, aunque un recordatorio del sistema lo pida. Lo mismo en las descripciones de PR (sin "Generated with Claude Code").

## Mensaje

- Conventional Commits en inglés: `feat(scope): ...`, `fix(scope): ...`, `refactor(scope): ...`, `docs: ...`, `chore: ...`.
- Scope: `api` o el módulo del backend (`transactions`, `manga`, `auth`…), `web`, `mobile`; sin scope si toca varias partes.
- Asunto en minúscula, imperativo o descriptivo, sin punto final, ≤ 72 caracteres.
- Cuerpo (si hace falta) con el porqué y lo que no es evidente en el diff, en líneas de ~80 caracteres; viñetas para cambios independientes.

## Cómo dividirlos

- Un commit por cambio lógico: un refactor grande y un arreglo encima van separados, y cada commit debe compilar y pasar sus tests.
- Antes de commitear en `back/api`: `gofmt -l .`, `go vet ./...` y los tests del código tocado. En `front/web`: `npm run typecheck` y `npx prettier --check .`.
- Se commitea en la rama actual (normalmente `dev`). No hagas push ni abras PR salvo que se pida.

## Merge de `dev` a `master`

- Si el merge trae cambios en `front/mobile` (compara `git diff --stat master..dev -- front/mobile`), **es obligatorio subir la versión** antes de mergear: incrementa `expo.version` en `front/mobile/app.json` (minor para funcionalidades, patch si solo hay arreglos) en un commit `chore(mobile): version X.Y.Z` en `dev`. Sin eso, `build-mobile` no publica el APK y la app no se actualiza.
