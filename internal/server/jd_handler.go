package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"

	"suprie/application_tracker/internal/domain"
	"suprie/application_tracker/internal/dto"
	"suprie/application_tracker/internal/service"
)

// --- Synchronous JD endpoints ---

func (h *handlers) listJDs(c *gin.Context) {
	var filter string
	if q := c.Query("status"); q != "" {
		normalized, ok := domain.NormalizeStatus(q)
		if !ok {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("unknown status: %q", q)})
			return
		}
		filter = normalized
	}

	jds, err := h.deps.JDRepo.List(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	items := make([]dto.JDListItem, len(jds))
	for i := range jds {
		items[i] = dto.ToJDListItem(&jds[i])
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
}

func (h *handlers) getJD(c *gin.Context) {
	id, ok := parseIDParam(c)
	if !ok {
		return
	}
	jd, err := h.deps.JDRepo.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "job description not found"})
		return
	}
	c.JSON(http.StatusOK, dto.ToJDResponse(jd))
}

func (h *handlers) patchJD(c *gin.Context) {
	id, ok := parseIDParam(c)
	if !ok {
		return
	}
	var req dto.PatchJDRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Status != nil {
		normalized, ok := domain.NormalizeStatus(*req.Status)
		if !ok {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("unknown status: %q", *req.Status)})
			return
		}
		if err := h.deps.JDRepo.UpdateStatus(c.Request.Context(), id, normalized); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	if req.ApplyURL != nil {
		var u *string
		if *req.ApplyURL != "" {
			u = req.ApplyURL
		}
		if err := h.deps.JDRepo.UpdateApplyURL(c.Request.Context(), id, u); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	jd, err := h.deps.JDRepo.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, dto.ToJDResponse(jd))
}

func (h *handlers) applyJD(c *gin.Context) {
	id, ok := parseIDParam(c)
	if !ok {
		return
	}
	if _, _, err := service.Apply(c.Request.Context(), h.deps, id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"id": id, "status": domain.StatusApplied})
}

// --- Async LLM-backed endpoints (return 202 + task_id) ---

func (h *handlers) createJD(c *gin.Context) {
	var req struct {
		Text     string `json:"text" binding:"required"`
		ApplyURL string `json:"apply_url"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	text, applyURL := req.Text, req.ApplyURL
	deps := h.deps

	taskID := h.tasks.Submit("parse_jd", func(ctx context.Context) (any, error) {
		jd, err := service.ParseJDText(ctx, deps, text, applyURL)
		if err != nil {
			return nil, err
		}
		return gin.H{"jd_id": jd.ID}, nil
	})
	c.JSON(http.StatusAccepted, gin.H{"task_id": taskID})
}

func (h *handlers) matchJD(c *gin.Context) {
	id, ok := parseIDParam(c)
	if !ok {
		return
	}
	deps := h.deps

	taskID := h.tasks.Submit("match", func(ctx context.Context) (any, error) {
		if _, err := service.Match(ctx, deps, id); err != nil {
			return nil, err
		}
		return gin.H{"jd_id": id}, nil
	})
	c.JSON(http.StatusAccepted, gin.H{"task_id": taskID})
}

func (h *handlers) rankJD(c *gin.Context) {
	id, ok := parseIDParam(c)
	if !ok {
		return
	}
	deps := h.deps

	taskID := h.tasks.Submit("rank", func(ctx context.Context) (any, error) {
		if _, err := service.Rank(ctx, deps, id); err != nil {
			return nil, err
		}
		return gin.H{"jd_id": id}, nil
	})
	c.JSON(http.StatusAccepted, gin.H{"task_id": taskID})
}

func (h *handlers) coverLetterJD(c *gin.Context) {
	id, ok := parseIDParam(c)
	if !ok {
		return
	}
	deps := h.deps

	taskID := h.tasks.Submit("cover_letter", func(ctx context.Context) (any, error) {
		out, err := service.CoverLetter(ctx, deps, id)
		if err != nil {
			return nil, err
		}
		res := gin.H{"jd_id": id, "tex_path": out.TeXPath}
		if out.PDFPath != "" {
			res["pdf_url"] = fmt.Sprintf("/api/jds/%d/cover-letter.pdf", id)
		}
		return res, nil
	})
	c.JSON(http.StatusAccepted, gin.H{"task_id": taskID})
}

// coverLetterPDF serves the most recently compiled cover letter for the JD.
func (h *handlers) coverLetterPDF(c *gin.Context) {
	id, ok := parseIDParam(c)
	if !ok {
		return
	}
	path := filepath.Join(h.deps.GeneratedDir, fmt.Sprintf("cover_letter_%d.pdf", id))
	if _, err := os.Stat(path); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "cover letter not generated yet"})
		return
	}
	c.Header("Content-Disposition", fmt.Sprintf(`inline; filename="cover_letter_%d.pdf"`, id))
	c.File(path)
}
