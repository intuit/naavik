package trafficconfig

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/intuit/naavik/internal/cache"
)

func AddRoutes(routerGroup *gin.RouterGroup) *gin.RouterGroup {
	workloadRoutes := routerGroup.Group("/trafficonfig")
	workloadRoutes.GET("/resources/identities/:identity", getResourcesRelatedToIdentity)
	workloadRoutes.GET("/resources/identities/:identity/dependents/:dependent", getResourcesRelatedToIdentity)
	workloadRoutes.GET("/resources/identities/:identity/dependents/:dependent/env/:env", getResourcesRelatedToIdentityAndDependentAndEnv)
	workloadRoutes.GET("/identities/:identity", getByIdentity)
	workloadRoutes.GET("/identities/:identity/env/:env", getByIdentityEnv)
	return routerGroup
}

// getByIdentity godoc
//
//	@Summary		Traffic Config By Identity
//	@Description	Get Traffic Config by Identity
//	@Tags			Traffic Config
//	@Produce		json
//	@Param			identity	path		string	true	"Asset Alias"
//	@Success		200			{object}	map[string][]string
//	@Router			/trafficonfig/identities/{identity} [get].
func getByIdentity(c *gin.Context) {
	identity := c.Params.ByName("identity")
	trafficConfigEntry := cache.TrafficConfigCache.GetTrafficConfigEntry(identity)

	c.JSON(200, trafficConfigEntry)
}

// getByIdentityEnv godoc
//
//	@Summary		Traffic Config By Identity and Env
//	@Description	Get Traffic Config by Identity and Env
//	@Tags			Traffic Config
//	@Produce		json
//	@Param			identity	path		string	true	"Asset Alias"
//	@Param			env			path		string	true	"Environment"
//	@Success		200			{object}	map[string][]string
//	@Router			/trafficonfig/identities/{identity}/env/{env} [get].
func getByIdentityEnv(c *gin.Context) {
	identity := c.Params.ByName("identity")
	env := c.Params.ByName("env")
	trafficConfig := cache.TrafficConfigCache.Get(identity, env)

	c.JSON(http.StatusOK, trafficConfig)
}

// getResourcesRelatedToIdentity godoc
//
//	@Summary		Resources Related to Traffic Config Identity
//	@Description	Get Resources Related to Traffic Config Identity
//	@Tags			Traffic Config
//	@Produce		json
//	@Param			identity	path		string	true	"Asset Alias"
//	@Success		200			{object}	map[string][]string
//	@Router			/trafficonfig/resources/identities/{identity} [get].
func getResourcesRelatedToIdentity(c *gin.Context) {
	identity := c.Params.ByName("identity")
	sourceClusters := cache.IdentityCluster.GetClustersForIdentity(identity)
	trafficConfigResources := Resources{
		ClusterResources: make(map[string]map[string]interface{}),
	}
	// Get Throttle EnvoyFilters from source clusters
	for _, cluster := range sourceClusters {
		rc, found := cache.RemoteCluster.GetCluster(cluster)
		if found {
			trafficConfigResources.updateThrottleEnvoyFilters(rc, identity, "")
		}
	}

	dependents := cache.IdentityDependency.GetDependentsForIdentity(identity)
	for _, dependent := range dependents {
		dependentClusters := cache.IdentityCluster.GetClustersForIdentity(dependent)
		for _, cluster := range dependentClusters {
			rc, ok := cache.RemoteCluster.GetCluster(cluster)
			if ok {
				// Update VirtualServices from dependent clusters
				trafficConfigResources.updateVirtualServices(rc, identity, "")
			}
		}
	}

	c.JSON(http.StatusOK, trafficConfigResources)
}

// getResourcesRelatedToIdentity godoc
//
//	@Summary		Resources Related to Traffic Config Identity and Dependent
//	@Description	Get Resources Related to Traffic Config Identity and Dependent
//	@Tags			Traffic Config
//	@Produce		json
//	@Param			identity	path		string	true	"Asset Alias"
//	@Param			dependent	path		string	true	"Asset Alias"
//	@Success		200			{object}	map[string][]string
//	@Router			/trafficonfig/resources/identities/{identity}/dependents/{dependent} [get].
func getResourcesRelatedToIdentityAndDependent(c *gin.Context) {
	identity := c.Params.ByName("identity")
	dependent := c.Params.ByName("dependent")
	sourceClusters := cache.IdentityCluster.GetClustersForIdentity(identity)
	trafficConfigResources := Resources{
		ClusterResources: make(map[string]map[string]interface{}),
	}
	// Get Throttle EnvoyFilters from source clusters
	for _, cluster := range sourceClusters {
		rc, found := cache.RemoteCluster.GetCluster(cluster)
		if found {
			trafficConfigResources.updateThrottleEnvoyFilters(rc, identity, "")
		}
	}

	dependentClusters := cache.IdentityCluster.GetClustersForIdentity(dependent)
	for _, cluster := range dependentClusters {
		rc, ok := cache.RemoteCluster.GetCluster(cluster)
		if ok {
			// Update VirtualServices from dependent clusters
			trafficConfigResources.updateVirtualServices(rc, identity, "")
		}
	}

	c.JSON(http.StatusOK, trafficConfigResources)
}

// getResourcesRelatedToIdentityAndDependentAndEnv godoc
//
//	@Summary		Resources Related to Traffic Config Identity Dependent and Env
//	@Description	Get Resources Related to Traffic Config Identity and Dependent and Env
//	@Tags			Traffic Config
//	@Produce		json
//	@Param			identity	path		string	true	"Asset Alias"
//	@Param			dependent	path		string	true	"Asset Alias"
//	@Param			env			path		string	true	"Asset Alias"
//	@Success		200			{object}	map[string][]string
//	@Router			/trafficonfig/resources/identities/{identity}/dependents/{dependent}/env/{env} [get].
func getResourcesRelatedToIdentityAndDependentAndEnv(c *gin.Context) {
	identity := c.Params.ByName("identity")
	dependent := c.Params.ByName("dependent")
	env := c.Params.ByName("env")
	sourceClusters := cache.IdentityCluster.GetClustersForIdentity(identity)
	trafficConfigResources := Resources{
		ClusterResources: make(map[string]map[string]interface{}),
	}
	// Get Throttle EnvoyFilters from source clusters
	for _, cluster := range sourceClusters {
		rc, found := cache.RemoteCluster.GetCluster(cluster)
		if found {
			trafficConfigResources.updateThrottleEnvoyFilters(rc, identity, env)
		}
	}

	dependentClusters := cache.IdentityCluster.GetClustersForIdentity(dependent)
	for _, cluster := range dependentClusters {
		rc, ok := cache.RemoteCluster.GetCluster(cluster)
		if ok {
			// Update VirtualServices from dependent clusters
			trafficConfigResources.updateVirtualServices(rc, identity, env)
		}
	}

	c.JSON(http.StatusOK, trafficConfigResources)
}
