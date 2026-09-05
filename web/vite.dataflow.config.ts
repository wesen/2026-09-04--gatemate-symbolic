import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';
export default defineConfig({
  plugins: [react(), { name: 'dataflow-entry', configureServer(server) { server.middlewares.use((request, _response, next) => { if (request.url === '/') request.url = '/dataflow/index.html'; next(); }); } }], base: '/static/',
  server: { proxy: { '/api': 'http://127.0.0.1:8087' } },
  build: { outDir: 'dist-dataflow', assetsDir: '', rollupOptions: {
    input: 'dataflow/index.html',
    output: { entryFileNames: 'app.js', assetFileNames: 'app.[ext]' },
  } },
});
