package workload

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/intuit/naavik/internal/cache"
)

func AddRoutes(routerGroup *gin.RouterGroup) *gin.RouterGroup {
	workloadRoutes := routerGroup.Group("/workload")
	workloadRoutes.GET("/clusters/:cluster/identities/:identity/env/:env", getByClusterIdentityEnv)
	workloadRoutes.GET("/clusters/:cluster/identities/:identity", getWorkloadsByClusterIdentity)
	workloadRoutes.GET("/clusters/identities/:identity", getWorkloadsByIdentity)
	workloadRoutes.GET("/clusters/:cluster/namespaces/:namespace/services", getServicesByClusterNamespace)
	return routerGroup
}

// getWorkloadsByIdentity godoc
//
//	@Summary		Workload By Identity
//	@Description	Get Workload by Identity
//	@Tags			Workloads
//	@Produce		json
//	@Param			identity	path		string	assetAlias	"Asset Alias"
//	@Success		200			{object}	map[string][]string
//	@Router			/workload/clusters/identities/{identity} [get].
func getWorkloadsByIdentity(c *gin.Context) {
	identity := c.Params.ByName("identity")
	deployments := cache.Deployments.GetByIdentity(identity)
	rollouts := cache.Rollouts.GetByIdentity(identity)

	workloadItems := Items{
		Identity:        identity,
		DeploymentItems: deployments,
		RolloutItems:    rollouts,
	}

	c.JSON(http.StatusOK, workloadItems)
}

// getWorkloadsByClusterIdentity godoc
//
//	@Summary		Workload by Cluster and Identity
//	@Description	Get Workload by Cluster and Identity
//	@Tags			Workloads
//	@Produce		json
//	@Param			identity	path		string	assetAlias	"Asset Alias"
//	@Param			cluster		path		string	assetAlias	"Cluster Name"
//	@Success		200			{object}	map[string][]string
//	@Router			/workload/clusters/{cluster}/identities/{identity} [get].
func getWorkloadsByClusterIdentity(c *gin.Context) {
	cluster := c.Params.ByName("cluster")
	identity := c.Params.ByName("identity")
	deploymentEntry := cache.Deployments.GetByClusterIdentity(cluster, identity)
	rolloutEntry := cache.Rollouts.GetByClusterIdentity(cluster, identity)

	clusterWorkloads := ClusterWorkload{
		Cluster:         cluster,
		DeploymentItems: deploymentEntry,
		RolloutItems:    rolloutEntry,
	}

	c.JSON(http.StatusOK, clusterWorkloads)
}

// getByClusterIdentityEnv godoc
//
//	@Summary		Workload by Cluster, Identity and Env
//	@Description	Get Workload by Cluster, Identity and Env
//	@Tags			Workloads
//	@Produce		json
//	@Param			identity	path		string	true	"Asset Alias"
//	@Param			cluster		path		string	true	"Cluster Name"
//	@Param			env			path		string	true	"Env"
//	@Success		200			{object}	map[string][]string
//	@Router			/workload/clusters/{cluster}/identities/{identity}/env/{env} [get].
func getByClusterIdentityEnv(c *gin.Context) {
	cluster := c.Params.ByName("cluster")
	identity := c.Params.ByName("identity")
	env := c.Params.ByName("env")
	deployment := cache.Deployments.GetByClusterIdentityEnv(cluster, identity, env)
	rollout := cache.Rollouts.GetByClusterIdentityEnv(cluster, identity, env)

	workload := Workload{
		Identity:   identity,
		Deployment: deployment,
		Rollout:    rollout,
	}

	c.JSON(http.StatusOK, workload)
}

// getServicesByClusterNamespace godoc
//
//	@Summary		Services by Cluster and Namespace
//	@Description	Get Services by Cluster and Namespace
//	@Tags			Workloads
//	@Produce		json
//	@Param			cluster		path		string	true	"Cluster name"
//	@Param			namespace	path		string	true	"Namespace name"
//	@Success		200			{object}	map[string][]string
//	@Router			/workload/clusters/{cluster}/namespaces/{namespace}/services [get].
func getServicesByClusterNamespace(c *gin.Context) {
	cluster := c.Params.ByName("cluster")
	namespace := c.Params.ByName("namespace")
	services := cache.Services.GetByClusterNamespace(cluster, namespace)

	c.JSON(http.StatusOK, services)
}
