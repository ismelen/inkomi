import { defineConfig, loadEnv } from 'vite';
import react from '@vitejs/plugin-react';
import tailwindcss from '@tailwindcss/vite';
import pkg from './package.json';

// Routes of the api (back/api) run locally with `go run .` on its default port.
const localAPI = 'http://localhost:3000';
const apiRoutes = ['/auth', '/users', '/books', '/uploads', '/ws', '/app'];

export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), '');
  // DEV_BACKEND: another backend for every route (a deployed domain); otherwise the local api.
  const target = env.DEV_BACKEND || localAPI;
  return {
    plugins: [react(), tailwindcss()],
    define: { __APP_VERSION__: JSON.stringify(pkg.version) },
    server: {
      proxy: Object.fromEntries(
        apiRoutes.map((prefix) => [prefix, { target, changeOrigin: true, ws: prefix === '/ws' }])
      ),
    },
  };
});
