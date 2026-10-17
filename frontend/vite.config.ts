import react from '@vitejs/plugin-react'
import tailwindcss from '@tailwindcss/vite'
import { fileURLToPath } from 'node:url'
import { defineConfig } from 'vite'

// https://vite.dev/config/
export default defineConfig({
  plugins: [react(), tailwindcss()],
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url)),
    },
  },
  server: {
    host: '0.0.0.0',
    watch: {
      usePolling: true,
    },
    proxy: {
      '/health': {
        target: process.env.API_PROXY_TARGET ?? 'http://localhost:8080',
      },
      '/payments': {
        target: process.env.API_PROXY_TARGET ?? 'http://localhost:8080',
      },
    },
  },
})
