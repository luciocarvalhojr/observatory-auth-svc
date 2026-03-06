package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// Health handles liveness and readiness probes.
type Health struct {
	redis *redis.Client
}

// NewHealth creates a new Health handler.
func NewHealth(rdb *redis.Client) *Health {
	return &Health{redis: rdb}
}

// Register wires health routes.
func (h *Health) Register(rg *gin.RouterGroup) {
	rg.GET("/healthz", h.Liveness)
	rg.GET("/readyz", h.Readiness)
}

// Liveness godoc
// @Summary  Liveness probe
// @Tags     health
// @Success  200  {object}  map[string]string
// @Router   /healthz [get]
func (h *Health) Liveness(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// Readiness godoc
// @Summary  Readiness probe
// @Tags     health
// @Success  200  {object}  map[string]string
// @Failure  503  {object}  map[string]string
// @Router   /readyz [get]
func (h *Health) Readiness(c *gin.Context) {
	if err := h.redis.Ping(c.Request.Context()).Err(); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "unavailable",
			"redis":  "unreachable",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok", "redis": "ok"})
}
