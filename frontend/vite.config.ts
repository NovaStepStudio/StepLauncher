import { fileURLToPath, URL } from 'node:url'

import { defineConfig } from 'vite'
import path from 'path';
import vue from '@vitejs/plugin-vue'

// https://vite.dev/config/
export default defineConfig({
  root: './web',
  base: './',
  build: {
    outDir: path.join(__dirname,'dist')

  },
  plugins: [
    vue(),
  ],
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url)),
      '@wailsjs': fileURLToPath(new URL('./wailsjs', import.meta.url))
    },
  },
})
