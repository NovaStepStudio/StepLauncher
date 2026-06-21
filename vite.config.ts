import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import vueDevTools from 'vite-plugin-vue-devtools'
import { viteStaticCopy } from 'vite-plugin-static-copy'
import { resolve } from 'path'

export default defineConfig({
  root: './renderer',
  base: './',
  build: {
    outDir: resolve(__dirname,'./dist/Renderer'),
    cssMinify: true,
    minify: true,
    cssCodeSplit: true,
    rollupOptions: {
      output: {
        entryFileNames: 'scripts/[name].js',
        chunkFileNames: 'scripts/[name].js',
        manualChunks(id) {
          if (id.includes('node_modules/vue') || id.includes('node_modules/pinia') || id.includes('node_modules/@vue')) {
            return 'vue'
          }
          if (id.includes('node_modules/three') || id.includes('node_modules/skinview3d') || id.includes('node_modules/skinview-utils')) {
            return 'skinview3d'
          }
        },

        assetFileNames: (assetInfo) => {
          const ext = assetInfo.name!.split('.').pop()

          if (ext === 'css') {
            return 'css/[name].[ext]'
          } else if (ext === 'woff2') {
            return 'ui/fonts/[name].[ext]'
          } else if (ext === 'png') {
            return 'ui/img/[name].[ext]'
          }

          return 'ui/[name].[ext]'
        }
      }
    }
  },
  plugins: [
    vue(),
    // vueDevTools(),
    viteStaticCopy({
      targets: [
        {
          src: 'assets/svg/*.svg',
          dest: '.'
        },
        {
          src: 'assets/loaders/*.png',
          dest: '.'
        },
        {
          src: 'assets/loaders/Grass.webp',
          dest: '.'
        },
        {
          src: 'assets/background/**/*',
          dest: '.'
        },
        {
          src: 'assets/icon.png',
          dest: '.'
        },
        {
          src: 'assets/defaults/*',
          dest: '.'
        },
        {
          src: 'assets/locales/**/*',
          dest: '.'
        }
      ]
    })
  ],
})
