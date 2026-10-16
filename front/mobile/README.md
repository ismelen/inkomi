# Inkomi mobile

App de Expo para enviar libros y manga al e-reader a través de los microservicios de `back/`.

## Arranque

```bash
cp .env.template .env      # EXPO_PUBLIC_BACKEND_URL, EXPO_PUBLIC_DROPBOX_API_KEY, EXPO_PUBLIC_GOOGLE_WEB_CLIENT_ID
npm install
npx expo run:android       # dev build (Google Sign-In y FCM no funcionan en Expo Go)
npm start                  # Metro para el dev build ya instalado
```

`google-services.json` (Firebase) va en la raíz del proyecto y no se versiona.

## Comprobaciones

```bash
npm run typecheck
npx prettier --check .
```

Reglas del proyecto, contratos y arquitectura: [CLAUDE.md](CLAUDE.md). Diseño: frame "Mobile" de `docs/ideas.excalidraw`.
