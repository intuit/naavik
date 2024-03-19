package server

import (
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/intuit/naavik/internal/server/api"
	"github.com/intuit/naavik/internal/server/api/clusters"
	"github.com/intuit/naavik/internal/server/api/dependency"
	trafficconfig "github.com/intuit/naavik/internal/server/api/trafficconfig"
	"github.com/intuit/naavik/internal/server/api/workload"
	"github.com/intuit/naavik/internal/server/swagger"
	"github.com/intuit/naavik/pkg/logger"
)

const HealthCheckPath = "/health/full"

func SetupRouter() *gin.Engine {
	gin.DefaultWriter = io.Discard
	r := gin.Default()

	r.Use(Logger(), gin.Recovery())
	r.GET(HealthCheckPath, api.HealthCheck)

	// Base Path
	group := r.Group("api/v1")

	// Health Check
	api.AddRoutes(group)
	swagger.AddRoutes(group)
	workload.AddRoutes(group)
	clusters.AddRoutes(group)
	dependency.AddRoutes(group)
	trafficconfig.AddRoutes(group)

	return r
}

func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path
		c.Next()
		// Skip logging health check
		if !strings.HasSuffix(path, api.HealthCheckPath) {
			return
		}
		statusCode := c.Writer.Status()
		clientIP := c.ClientIP()
		clientUserAgent := c.Request.UserAgent()
		referer := c.Request.Referer()

		log := logger.Log.Int("status", statusCode).Str("clientIp", clientIP).Str("method", c.Request.Method).
			Str("path", path).Str("referer", referer).Str("userAgent", clientUserAgent)
		if len(c.Errors) > 0 {
			log.Error(c.Errors.ByType(gin.ErrorTypePrivate).String())
		} else {
			if statusCode >= http.StatusInternalServerError {
				log.Error("")
			} else if statusCode >= http.StatusBadRequest {
				log.Warn("")
			} else {
				log.Info("")
			}
		}
	}
}
