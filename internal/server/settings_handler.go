package server

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"suprie/application_tracker/internal/domain"
	"suprie/application_tracker/internal/dto"
	"suprie/application_tracker/internal/service"
)

func (h *handlers) getLLMSettings(c *gin.Context) {
	s, err := service.GetLLMSettings(c.Request.Context(), h.deps)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if s == nil {
		s = &domain.LLMSettings{}
	}
	c.JSON(http.StatusOK, dto.ToLLMSettingsResponse(s))
}

func (h *handlers) updateLLMSettings(c *gin.Context) {
	var req dto.UpdateLLMSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	s, err := service.UpdateLLMSettings(c.Request.Context(), h.deps, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, dto.ToLLMSettingsResponse(s))
}
