import { defineConfig } from 'vite'
import { svelte } from '@sveltejs/vite-plugin-svelte'

// The built SPA is emitted into the Go embed package's dist/ so it is served
// from the single binary via internal/web/embed.go. During `vite dev`, /api
// requests are proxied to the Go server (ats serve) on :8080.
export default defineConfig({
  plugins: [svelte()],
  build: {
    outDir: '../internal/web/dist',
    emptyOutDir: true,
  },
  server: {
    port: 5173,
    proxy: {
      '/api': 'http://localhost:8080',
    },
  },
})
