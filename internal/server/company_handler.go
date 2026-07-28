package server

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"suprie/application_tracker/internal/dto"
	"suprie/application_tracker/internal/service"
)

func (h *handlers) listCompanies(c *gin.Context) {
	companies, err := service.ListCompanies(c.Request.Context(), h.deps, c.Query("q"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	out := make([]dto.CompanyResponse, len(companies))
	for i := range companies {
		out[i] = dto.ToCompanyResponse(&companies[i])
	}
	c.JSON(http.StatusOK, gin.H{"items": out})
}

func (h *handlers) createCompany(c *gin.Context) {
	var req dto.CreateCompanyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	company, err := service.CreateCompany(c.Request.Context(), h.deps, req)
	if err != nil {
		if errors.Is(err, service.ErrCompanyExists) {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, dto.ToCompanyResponse(company))
}

func (h *handlers) getCompany(c *gin.Context) {
	id, ok := parseIDParam(c)
	if !ok {
		return
	}
	company, err := service.GetCompany(c.Request.Context(), h.deps, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "company not found"})
		return
	}
	c.JSON(http.StatusOK, dto.ToCompanyResponse(company))
}

func (h *handlers) updateCompany(c *gin.Context) {
	id, ok := parseIDParam(c)
	if !ok {
		return
	}
	var req dto.UpdateCompanyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	company, err := service.UpdateCompany(c.Request.Context(), h.deps, id, req)
	if err != nil {
		if errors.Is(err, service.ErrCompanyExists) {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, dto.ToCompanyResponse(company))
}
