package server

import (
	"errors"
	"log"
	"net"
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"

	"suprie/application_tracker/internal/domain"
	"suprie/application_tracker/internal/dto"
	"suprie/application_tracker/internal/service"
)

func (h *handlers) getLLMSettings(c *gin.Context) {
	s, err := service.GetLLMSettings(c.Request.Context(), h.deps)
	if err != nil {
		log.Printf("load LLM settings: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not load LLM settings"})
		return
	}
	if s == nil {
		s = &domain.LLMSettings{}
	}
	c.JSON(http.StatusOK, dto.ToLLMSettingsResponse(s))
}

func (h *handlers) updateLLMSettings(c *gin.Context) {
	if !localSettingsRequest(c.Request) {
		c.JSON(http.StatusForbidden, gin.H{"error": "LLM settings can only be changed from localhost"})
		return
	}
	if c.ContentType() != "application/json" {
		c.JSON(http.StatusUnsupportedMediaType, gin.H{"error": "Expected application/json"})
		return
	}
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 16<<10)
	var req domain.LLMSettingsUpdate
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	s, err := service.UpdateLLMSettings(c.Request.Context(), h.deps, req)
	if errors.Is(err, service.ErrInvalidLLMSettings) {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err != nil {
		log.Printf("save LLM settings: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not save LLM settings"})
		return
	}
	c.JSON(http.StatusOK, dto.ToLLMSettingsResponse(s))
}

func localSettingsRequest(req *http.Request) bool {
	peer, _, err := net.SplitHostPort(req.RemoteAddr)
	if err != nil || !net.ParseIP(peer).IsLoopback() {
		return false
	}
	host := req.Host
	if name, _, err := net.SplitHostPort(host); err == nil {
		host = name
	}
	if !strings.EqualFold(host, "localhost") && !net.ParseIP(strings.Trim(host, "[]")).IsLoopback() {
		return false
	}
	if origin := req.Header.Get("Origin"); origin != "" {
		u, err := url.Parse(origin)
		if err != nil || !strings.EqualFold(u.Host, req.Host) {
			return false
		}
	}
	return true
}
