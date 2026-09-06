import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';
export default defineConfig({
 plugins: [react(), {name:'language-entry',configureServer(server){server.middlewares.use((request,_response,next)=>{if(request.url==='/')request.url='/language/index.html';next();});}}],
 base:'/static/', server:{proxy:{'/api':'http://127.0.0.1:18091'}},
 build:{outDir:'dist-language',assetsDir:'',rollupOptions:{input:'language/index.html',output:{entryFileNames:'app.js',assetFileNames:'app.[ext]'}}},
});
