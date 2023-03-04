import { defineConfig } from "vite"
import FullReload from "vite-plugin-full-reload"

// https://vitejs.dev/config/
export default defineConfig({
  server: {
    proxy: {
      "^/(?!@vite|@fs|src|__vite|node_modules).*": {
        target: "http://127.0.0.1:3000"
      }
    }
  },
  plugins: [
    FullReload(["view/*.html"]),
  ],
})
