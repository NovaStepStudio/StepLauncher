import { fileURLToPath, URL } from 'node:url'

import { defineConfig } from 'vite'
import path from 'path';
import vue from '@vitejs/plugin-vue'

// https://vite.dev/config/
export default defineConfig({
  root: './web',
  base: './',
  build: {
    outDir: path.join(__dirname, 'dist'),
    emptyOutDir: true,
    assetsInlineLimit: 0,
    rollupOptions: {
      output: {
        entryFileNames: 'assets/js/[name]-[hash].js',
        chunkFileNames: 'assets/js/[name]-[hash].js',
        assetFileNames: (assetInfo) => {
          const name = assetInfo.name ?? '';
          const ext = path.extname(name).toLowerCase();
          if (ext === '.css') return 'assets/css/[name]-[hash][extname]';
          if (['.woff', '.woff2', '.ttf', '.otf', '.eot'].includes(ext)) {
            return 'assets/fonts/[name]-[hash][extname]';
          }
          if (['.png', '.jpg', '.jpeg', '.webp', '.gif', '.svg', '.ico', '.avif'].includes(ext)) {
            return 'assets/img/[name]-[hash][extname]';
          }
          return 'assets/[name]-[hash][extname]';
        },
        manualChunks: (id) => {
          const norm = id.replace(/\\/g, '/');
          if (!norm.includes('node_modules')) {
            if (norm.includes('/Layouts/Sections/Settings/')) return 'settings';
            if (norm.includes('/Modals/')) return 'modals';
            if (norm.includes('/Composables/SkinPlayer/')) return 'skin-player';
            if (norm.includes('/stores/')) return 'stores';
            if (norm.includes('/Composables/')) return 'composables';
            return undefined;
          }
          if (norm.includes('@tabler/icons-vue')) return 'vendor-icons';
          if (norm.includes('/vue/') || norm.includes('/@vue/') || norm.includes('/@vuejs/')) return 'vendor-vue';
          if (norm.includes('/vue-router/')) return 'vendor-vue';
          return 'vendor';
        },
      },
    },
  },
  plugins: [
    vue(),
  ],
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./web/src', import.meta.url)),
      '@wailsjs': fileURLToPath(new URL('./wailsjs', import.meta.url))
    },
  },
})