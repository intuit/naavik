package clusters

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/intuit/naavik/cmd/options"
	"github.com/intuit/naavik/internal/cache"
	"github.com/intuit/naavik/internal/server/api"
	"github.com/intuit/naavik/internal/types"
	"github.com/intuit/naavik/internal/types/context"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func AddRoutes(routerGroup *gin.RouterGroup) *gin.RouterGroup {
	workloadRoutes := routerGroup.Group("/clusters")
	workloadRoutes.GET("/", getRemoteClusters)
	workloadRoutes.GET("/identities/:identity", getClustersForIdentity)
	workloadRoutes.GET("/:clusterId/envoyfilters/identities/:identity", getEnvoyFiltersForClusterAndIdentity)
	workloadRoutes.GET("/:clusterId/envoyfilters", getEnvoyFiltersForCluster)
	workloadRoutes.GET("/:clusterId/virtualservices/identities/:identity", getVirtualServicesForClusterAndIdentity)
	workloadRoutes.GET("/:clusterId/virtualservices", getVirtualServicesForCluster)

	return routerGroup
}

// getRemoteClusters godoc
//
//	@Summary		Remote Clusters
//	@Description	Get List of Remote Clusters Naavik is aware of
//	@Tags			Clusters
//	@Produce		json
//	@Success		200	{object}	map[string][]string
//	@Router			/clusters [get].
func getRemoteClusters(c *gin.Context) {
	clusterNames := []Cluster{}
	clusters := cache.RemoteCluster.ListClusters()
	for _, cluster := range clusters {
		clusterNames = append(clusterNames, Cluster{
			Name: cluster.GetClusterID(),
			Host: cluster.GetHost(),
		})
	}
	c.JSON(http.StatusOK, clusterNames)
}

// getClustersForIdentity godoc
//
//	@Summary		Remote Clusters for Identity
//	@Description	Get List of Remote Clusters Naavik is aware of for a given Identity
//	@Tags			Clusters
//	@Produce		json
//	@Param			identity	path		string	true	"Asset Alias"
//	@Success		200			{object}	map[string][]string
//	@Router			/clusters/identities/{identity} [get].
func getClustersForIdentity(c *gin.Context) {
	identity := c.Params.ByName("identity")
	clusters := cache.IdentityCluster.GetClustersForIdentity(identity)

	c.JSON(http.StatusOK, clusters)
}

// getEnvoyFiltersForCluster godoc
//
//	@Summary		List EnvoyFilters for Cluster
//	@Description	Get List of EnvoyFilters for a given Cluster managed by Naavik
//	@Tags			Clusters
//	@Produce		json
//	@Param			clusterId	path		string	true	"Cluster Name"
//	@Success		200			{object}	map[string][]string
//	@Router			/clusters/{clusterId}/envoyfilters [get].
func getEnvoyFiltersForCluster(c *gin.Context) {
	clusterID := c.Params.ByName("clusterId")
	rc, ok := cache.RemoteCluster.GetCluster(clusterID)
	if !ok {
		c.JSON(http.StatusNotFound, api.ErrorResponse{Message: fmt.Sprintf("Cluster %s not found", clusterID)})
		return
	}
	envoyfilters, err := rc.IstioClient().ListEnvoyFilters(context.NewContextWithLogger(), types.NamespaceIstioSystem, metav1.ListOptions{
		LabelSelector: fmt.Sprintf("%s=%s", types.CreatedByKey, types.NaavikName),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, api.ErrorResponse{Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, envoyfilters)
}

// getVirtualServicesForCluster godoc
//
//	@Summary		List VirtualServices for Cluster
//	@Description	Get List of VirtualServices for a given Cluster managed by Naavik
//	@Tags			Clusters
//	@Produce		json
//	@Param			clusterId	path		string	true	"Cluster Name"
//	@Success		200			{object}	map[string][]string
//	@Router			/clusters/{clusterId}/virtualservices [get].
func getVirtualServicesForCluster(c *gin.Context) {
	clusterID := c.Params.ByName("clusterId")
	rc, ok := cache.RemoteCluster.GetCluster(clusterID)
	if !ok {
		c.JSON(http.StatusNotFound, api.ErrorResponse{Message: fmt.Sprintf("Cluster %s not found", clusterID)})
		return
	}
	virtualservices, err := rc.IstioClient().ListVirtualServices(context.NewContextWithLogger(), options.GetSyncNamespace(), metav1.ListOptions{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, api.ErrorResponse{Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, virtualservices)
}

// getEnvoyFiltersForClusterAndIdentity godoc
//
//	@Summary		List EnvoyFilters for Cluster and Identity
//	@Description	Get List of EnvoyFilters for a given Cluster and Identity managed by Naavik
//	@Tags			Clusters
//	@Produce		json
//	@Param			clusterId	path		string	true	"Cluster Name"
//	@Param			identity	path		string	true	"Asset Alias"
//	@Success		200			{object}	map[string][]string
//	@Router			/clusters/{clusterId}/envoyfilters/identities/{identity} [get].
func getEnvoyFiltersForClusterAndIdentity(c *gin.Context) {
	clusterID := c.Params.ByName("clusterId")
	identity := c.Params.ByName("identity")
	rc, ok := cache.RemoteCluster.GetCluster(clusterID)
	if !ok {
		c.JSON(http.StatusNotFound, api.ErrorResponse{Message: fmt.Sprintf("Cluster %s not found", clusterID)})
		return
	}
	envoyfilters, err := rc.IstioClient().ListEnvoyFilters(context.NewContextWithLogger(), types.NamespaceIstioSystem, metav1.ListOptions{
		LabelSelector: fmt.Sprintf("%s=%s", types.CreatedForKey, identity),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, api.ErrorResponse{Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, envoyfilters)
}

// getVirtualServicesForClusterAndIdentity godoc
//
//	@Summary		List VirtualServices for Cluster and Identity
//	@Description	Get List of VirtualServices for a given Cluster and Identity managed by Naavik
//	@Tags			Clusters
//	@Produce		json
//	@Param			clusterId	path		string	true	"Cluster Name"
//	@Param			identity	path		string	true	"Asset Alias"
//	@Success		200			{object}	map[string][]string
//	@Router			/clusters/{clusterId}/virtualservices/identities/{identity} [get].
func getVirtualServicesForClusterAndIdentity(c *gin.Context) {
	clusterID := c.Params.ByName("clusterId")
	identity := c.Params.ByName("identity")
	rc, ok := cache.RemoteCluster.GetCluster(clusterID)
	if !ok {
		c.JSON(http.StatusNotFound, api.ErrorResponse{Message: fmt.Sprintf("Cluster %s not found", clusterID)})
		return
	}
	virtualservices, err := rc.IstioClient().ListVirtualServices(context.NewContextWithLogger(), options.GetSyncNamespace(), metav1.ListOptions{
		LabelSelector: fmt.Sprintf("%s=%s", types.CreatedForKey, identity),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, api.ErrorResponse{Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, virtualservices)
}
