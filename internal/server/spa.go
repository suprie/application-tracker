package server

import (
	"io/fs"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"suprie/application_tracker/internal/web"
)

// registerSPA serves the embedded frontend and falls back to index.html for
// client-side routes, so deep links and hard refreshes work.
func (h *handlers) registerSPA(r *gin.Engine) {
	dist, err := fs.Sub(web.Assets, "dist")
	if err != nil {
		// Should not happen given the committed placeholder.
		return
	}
	fileServer := http.FileServer(http.FS(dist))

	r.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path

		// Unmatched API route → JSON 404.
		if strings.HasPrefix(path, "/api/") {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}

		// Serve a real static asset if it exists under dist.
		if f, err := dist.Open(strings.TrimPrefix(path, "/")); err == nil {
			f.Close()
			fileServer.ServeHTTP(c.Writer, c.Request)
			return
		}

		// SPA fallback: rewrite to index.html.
		c.Request.URL.Path = "/"
		fileServer.ServeHTTP(c.Writer, c.Request)
	})
}
