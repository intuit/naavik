package cache

import (
	"fmt"

	"github.com/intuit/naavik/internal/cache"
	resourcebuilder "github.com/intuit/naavik/internal/fake/builder/resource"
)

func BuildFakeDeploymentCache(clusterNamespaceNumDepsMap map[string]map[string]int) {
	for clusterID, namespaceNumDeps := range clusterNamespaceNumDepsMap {
		for namespace, numDeps := range namespaceNumDeps {
			for i := 0; i < numDeps; i++ {
				name := fmt.Sprintf("depoyment-%d", i)
				assetAlias := fmt.Sprintf("intuit.%s.deploy", namespace)
				appName := fmt.Sprintf("app-%d", i)
				env := fmt.Sprintf("env-%d", i)
				dep := resourcebuilder.BuildFakeDeployment(name, assetAlias, appName, env, namespace)
				cache.Deployments.Add(clusterID, dep)
			}
		}
	}
}
