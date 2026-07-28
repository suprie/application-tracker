package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *handlers) getTask(c *gin.Context) {
	res, ok := h.tasks.Get(c.Param("id"))
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "task not found"})
		return
	}
	c.JSON(http.StatusOK, res)
}

func (h *handlers) cancelTask(c *gin.Context) {
	if !h.tasks.Cancel(c.Param("id")) {
		c.JSON(http.StatusNotFound, gin.H{"error": "task not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"cancelled": c.Param("id")})
}
