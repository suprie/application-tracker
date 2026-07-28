package web

import "embed"

// Assets holds the built frontend under dist/. The dist/ directory always
// contains at least a placeholder index.html, so the Go binary compiles even
// before `npm run build` has produced the real assets. Vite is configured to
// output its build into this directory (see web/vite.config.js).
//
//go:embed all:dist
var Assets embed.FS
