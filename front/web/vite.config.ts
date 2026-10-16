import { defineConfig, loadEnv } from 'vite';
import react from '@vitejs/plugin-react';
import tailwindcss from '@tailwindcss/vite';
import pkg from './package.json';

// Default ports of the services run with `go run .` (like the ingress, by path prefix).
const localServices: Record<string, string> = {
  '/auth': 'http://localhost:3100',
  '/users': 'http://localhost:3100',
  '/books': 'http://localhost:3001',
  '/uploads': 'http://localhost:3000',
  '/ws': 'http://localhost:3000',
  '/app': 'http://localhost:3200',
};

export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), '');
  // DEV_BACKEND: one gateway for every route (a deployed domain); otherwise the local services.
  const gateway = env.DEV_BACKEND;
  return {
    plugins: [react(), tailwindcss()],
    define: { __APP_VERSION__: JSON.stringify(pkg.version) },
    server: {
      proxy: Object.fromEntries(
        Object.entries(localServices).map(([prefix, local]) => [
          prefix,
          { target: gateway || local, changeOrigin: true, ws: prefix === '/ws' },
        ])
      ),
    },
  };
});
