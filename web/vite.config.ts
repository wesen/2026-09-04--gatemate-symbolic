import { defineConfig } from 'vitest/config';
import react from '@vitejs/plugin-react';

export default defineConfig({
  plugins: [react()],
  base: '/static/',
  server: { proxy: { '/api': 'http://127.0.0.1:8086' } },
  build: {
    assetsDir: '',
    rollupOptions: { output: { entryFileNames: 'app.js', assetFileNames: 'app.[ext]' } },
  },
  test: { environment: 'jsdom', setupFiles: ['./src/test-setup.ts'], restoreMocks: true },
});
