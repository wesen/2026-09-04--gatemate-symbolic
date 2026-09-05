import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';
export default defineConfig({
  plugins: [react(), { name: 'lazy-entry', configureServer(server) { server.middlewares.use((request, _response, next) => { if (request.url === '/') request.url = '/lazy/index.html'; next(); }); } }], base: '/static/',
  server: { proxy: { '/api': 'http://127.0.0.1:18090' } },
  build: { outDir: 'dist-lazy', assetsDir: '', rollupOptions: {
    input: 'lazy/index.html',
    output: { entryFileNames: 'app.js', assetFileNames: 'app.[ext]' },
  } },
});
