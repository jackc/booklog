import { defineConfig } from "vite"
import FullReload from "vite-plugin-full-reload"
import viteCompression from 'vite-plugin-compression'

// https://vitejs.dev/config/
export default defineConfig({
  server: {
    proxy: {
      "^/(?!@vite|@fs|src|__vite|node_modules|assets|css).*": {
        target: "http://127.0.0.1:3000"
      }
    }
  },
  plugins: [
    FullReload(["html/*.html"]),
    viteCompression(),
  ],
  build: {
    outDir: "build/frontend",
    manifest: true,
    rollupOptions: {
      input: {
        css: "css/main.css",
      },
    },
  },
})
