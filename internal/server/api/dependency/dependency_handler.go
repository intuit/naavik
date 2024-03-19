package dependency

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/intuit/naavik/internal/cache"
)

type TotalDependencies struct {
	Total int `json:"total"`
}

func AddRoutes(routerGroup *gin.RouterGroup) *gin.RouterGroup {
	workloadRoutes := routerGroup.Group("/dependency")
	workloadRoutes.GET("/identities/:identity", getDependenciesForIdentity)
	workloadRoutes.GET("/total", getTotalDependencies)

	workloadRoutes = routerGroup.Group("/dependents")
	workloadRoutes.GET("/identities/:identity", getDependentsForIdentity)

	return routerGroup
}

// getDependenciesForIdentity godoc
//
//	@Summary		Dependency for Identity
//	@Description	Get Dependency for Identity
//	@Tags			Dependency
//	@Produce		json
//	@Param			identity	path		string	true	"Asset Alias"
//	@Success		200			{object}	map[string][]string
//	@Router			/dependency/identities/{identity} [get].
func getDependenciesForIdentity(c *gin.Context) {
	identity := c.Params.ByName("identity")
	dependencies := cache.IdentityDependency.GetDependenciesForIdentity(identity)

	c.JSON(http.StatusOK, dependencies)
}

// getDependentsForIdentity godoc
//
//	@Summary		Dependents for Identity
//	@Description	Get Dependents for Identity
//	@Tags			Dependency
//	@Produce		json
//	@Param			identity	path		string	true	"Asset Alias"
//	@Success		200			{object}	map[string][]string
//	@Router			/dependents/identities/{identity} [get].
func getDependentsForIdentity(c *gin.Context) {
	identity := c.Params.ByName("identity")
	dependents := cache.IdentityDependency.GetDependentsForIdentity(identity)

	c.JSON(http.StatusOK, dependents)
}

// getTotalDependencies godoc
//
//	@Summary		Total Dependencies
//	@Description	Get Total Dependencies
//	@Tags			Dependency
//	@Produce		json
//	@Success		200	{object}	map[string][]string
//	@Router			/dependents/total [get].
func getTotalDependencies(c *gin.Context) {
	totalDependencies := cache.IdentityDependency.GetTotalDependencies()

	c.JSON(http.StatusOK, TotalDependencies{Total: totalDependencies})
}
