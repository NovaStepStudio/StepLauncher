import { fileURLToPath, URL } from 'node:url'

import { defineConfig } from 'vite'
import path from 'path';
import vue from '@vitejs/plugin-vue'
import vueDevtools from 'vite-plugin-vue-devtools';
import wails from '@wailsio/runtime/plugins/vite'

// https://vite.dev/config/
export default defineConfig({
  root: './web',
  base: './',
  server: {
    port: 9245,
    host: '127.0.0.1',
  },
  build: {
    outDir: fileURLToPath(new URL('./dist', import.meta.url)),
    emptyOutDir: true,
    minify: true,
    cssMinify: true,
    assetsInlineLimit: 0,
    rollupOptions: {
      // Suprime avisos INEFFECTIVE_DYNAMIC_IMPORT: SystemService y MusicService
      // se importan tanto estática (Welcome, MusicPanel) como dinámicamente
      // (Logger, CoverCache...). El módulo ya está en el chunk principal por el
      // import estático, el dinámico no puede moverlo -> warning informativo.
      // No se cambia la estructura de chunks sin evidencia (ver AGENTS.md).
      onwarn(warning, defaultHandler) {
        if ((warning as any).code === 'INEFFECTIVE_DYNAMIC_IMPORT') return
        defaultHandler(warning as any)
      },
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
    vueDevtools(),
    wails(fileURLToPath(new URL('./bindings', import.meta.url))),
  ],
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./web/src', import.meta.url)),
      '@wailsjs': fileURLToPath(new URL('./bindings', import.meta.url)),
      // Aliases cortos para la nueva API por dominio (spec: @wailsjs/StepLauncher/SystemService etc.)
      // Permiten importar con la ruta corta del spec además de la verbosa generada por Wails.
      '@wailsjs/StepLauncher/SystemService': fileURLToPath(new URL('./bindings/StepLauncher/internal/Services/System/systemservice.js', import.meta.url)),
      '@wailsjs/StepLauncher/ConfigService': fileURLToPath(new URL('./bindings/StepLauncher/internal/Services/Config/configservice.js', import.meta.url)),
      '@wailsjs/StepLauncher/InstanceService': fileURLToPath(new URL('./bindings/StepLauncher/internal/Services/Instance/instanceservice.js', import.meta.url)),
      '@wailsjs/StepLauncher/GameService': fileURLToPath(new URL('./bindings/StepLauncher/internal/Services/Game/gameservice.js', import.meta.url)),
      '@wailsjs/StepLauncher/DownloadService': fileURLToPath(new URL('./bindings/StepLauncher/internal/Services/Download/downloadservice.js', import.meta.url)),
      '@wailsjs/StepLauncher/AccountService': fileURLToPath(new URL('./bindings/StepLauncher/internal/Services/Account/accountservice.js', import.meta.url)),
      '@wailsjs/StepLauncher/ModLoaderService': fileURLToPath(new URL('./bindings/StepLauncher/internal/Services/ModLoader/modloaderservice.js', import.meta.url)),
      '@wailsjs/StepLauncher/MusicService': fileURLToPath(new URL('./bindings/StepLauncher/internal/Services/Music/musicservice.js', import.meta.url)),
      '@wailsjs/StepLauncher/AppearanceService': fileURLToPath(new URL('./bindings/StepLauncher/internal/Services/Appearance/appearanceservice.js', import.meta.url)),
    },
  },
})