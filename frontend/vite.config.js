import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react-swc'
import tailwindcss from '@tailwindcss/vite'

// https://vite.dev/config/
export default defineConfig({
  server: {
    port: 80,
    proxy: {
      "/api": {
        target: "http://number-backend:8080",
        secure: false,
        changeOrigin: true,
        rewrite: path => path.replace(/^\/api/, ''),
      }
    }
  },
  plugins: [tailwindcss(), react()],
})
