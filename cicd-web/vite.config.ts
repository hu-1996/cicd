import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [react()],
  server: {
    proxy: {
      '/api': {
        // target: 'http://8.209.209.240:8029',
        target: 'http://127.0.0.1:8029',
        changeOrigin: true,
        rewrite: (path) => path.replace(/^\/api/, 'api'),
      },
    },
  },
})
