package swagger

import (
	"github.com/gin-gonic/gin"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func AddRoutes(routerGroup *gin.RouterGroup) *gin.RouterGroup {
	// use ginSwagger middleware to serve the API docs
	routerGroup.GET("/docs/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))
	return routerGroup
}

func Initialize() {
	SwaggerInfo.BasePath = "/api/v1"
	SwaggerInfo.Title = "Mesh Naavik REST API"
	SwaggerInfo.Description = "Mesh Naavik REST APIs to watch state of Naavik."
	SwaggerInfo.Version = "1.0"
}
