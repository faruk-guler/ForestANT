import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [react()],
  server: {
    port: 3000,
    proxy: {
      '/api': {
        target: 'http://localhost:80',
        changeOrigin: true
      }
    }
  },
  build: {
    rollupOptions: {
      output: {
        manualChunks: {
          'vendor-ui': ['recharts', 'lucide-react', 'date-fns'],
          'vendor-core': ['react', 'react-dom', 'axios'],
        }
      }
    },
    chunkSizeWarningLimit: 1000
  }
})
