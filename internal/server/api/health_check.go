package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

const HealthCheckPath = "/health/full"

func AddRoutes(routerGroup *gin.RouterGroup) *gin.RouterGroup {
	return routerGroup
}

// HealthCheck godoc
//
//	@Summary		Health Status of Naavik
//	@Description	Health Status of Naavik
//	@Tags			Health Check
//	@Produce		json
//	@Success		200	string	Health	full	ok
//	@Router			/health/full [get].
func HealthCheck(c *gin.Context) {
	c.String(http.StatusOK, "Health full ok")
}
