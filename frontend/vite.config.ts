import { fileURLToPath, URL } from 'node:url'

import { defineConfig } from 'vite'
import path from 'path';
import vue from '@vitejs/plugin-vue'

// https://vite.dev/config/
export default defineConfig({
  root: './web',
  base: './',
  build: {
    outDir: fileURLToPath(new URL('./dist', import.meta.url)),
    emptyOutDir: true,
    minify: true,
    cssMinify: true,
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
            if (norm.includes('/Instances/')) return 'instances';
            if (norm.includes('/Screenshots/')) return 'screenshots';
            if (norm.includes('/Accounts/')) return 'accounts';
            if (norm.includes('/Versions/')) return 'versions';
            if (norm.includes('/Settings/')) return 'settings';
            if (norm.includes('/Launcher/')) return 'launcher';
            if (norm.includes('/Downloads/')) return 'downloads';
            if (norm.includes('/News/')) return 'news';
            if (norm.includes('/Welcome/')) return 'welcome';
            if (norm.includes('/Updates/')) return 'updates';
            if (norm.includes('/Crash/')) return 'crash';
            if (norm.includes('/Login/')) return 'login';
            if (norm.includes('/Common/Composables/SkinPlayer/')) return 'skin-player';
            if (norm.includes('/Common/')) return 'common';
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