package server

import (
	"context"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"

	"suprie/application_tracker/internal/service"
)

// uploadCV accepts a multipart CV PDF, runs parse-cv, and overwrites the
// master profile. Async: returns 202 + task_id, matching the other
// LLM-backed endpoints.
func (h *handlers) uploadCV(c *gin.Context) {
	file, err := c.FormFile("cv")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing \"cv\" file field"})
		return
	}
	if !strings.EqualFold(filepath.Ext(file.Filename), ".pdf") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "only .pdf files are supported"})
		return
	}

	uploadDir, err := os.MkdirTemp("", "ats-cv-upload-*")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	uploadPath := filepath.Join(uploadDir, filepath.Base(file.Filename))
	if err := c.SaveUploadedFile(file, uploadPath); err != nil {
		os.RemoveAll(uploadDir)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	deps := h.deps
	taskID := h.tasks.Submit("parse_cv", func(ctx context.Context) (any, error) {
		defer os.RemoveAll(uploadDir)
		if err := service.ParseCV(ctx, deps, uploadPath); err != nil {
			return nil, err
		}
		return gin.H{"profile_path": deps.ProfilePath}, nil
	})
	c.JSON(http.StatusAccepted, gin.H{"task_id": taskID})
}
