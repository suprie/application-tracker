package server

import (
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"

	"suprie/application_tracker/internal/service"
	"suprie/application_tracker/internal/task"
)

// handlers holds the shared dependencies for all HTTP handlers.
type handlers struct {
	deps  service.Deps
	tasks *task.Runner
}

// New builds the Gin engine with all routes registered.
func New(deps service.Deps) *gin.Engine {
	h := &handlers{deps: deps, tasks: task.New(taskConcurrency())}
	r := gin.Default()

	api := r.Group("/api")
	{
		// Job descriptions (applications).
		api.GET("/jds", h.listJDs)
		api.POST("/jds", h.createJD) // async: parse pasted text
		api.GET("/jds/:id", h.getJD)
		api.PATCH("/jds/:id", h.patchJD)
		api.POST("/jds/:id/apply", h.applyJD)
		api.POST("/jds/:id/match", h.matchJD)
		api.POST("/jds/:id/rank", h.rankJD)
		api.POST("/jds/:id/cover-letter", h.coverLetterJD)
		api.GET("/jds/:id/cover-letter.pdf", h.coverLetterPDF)

		// Companies.
		api.GET("/companies", h.listCompanies)
		api.POST("/companies", h.createCompany)
		api.GET("/companies/:id", h.getCompany)
		api.PUT("/companies/:id", h.updateCompany)

		// Async tasks.
		api.GET("/tasks/:id", h.getTask)
		api.DELETE("/tasks/:id", h.cancelTask)
	}

	h.registerSPA(r)
	return r
}

// taskConcurrency reads $ATS_TASK_CONCURRENCY (default 2) so a local LLM is
// not flooded with concurrent requests.
func taskConcurrency() int {
	if v := os.Getenv("ATS_TASK_CONCURRENCY"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			return n
		}
	}
	return 2
}

// parseIDParam reads the :id route param, responding 400 on invalid input.
func parseIDParam(c *gin.Context) (int, bool) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return 0, false
	}
	return id, true
}
