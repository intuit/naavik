package cache

import (
	"fmt"

	"github.com/intuit/naavik/internal/cache"
	resourcebuilder "github.com/intuit/naavik/internal/fake/builder/resource"
)

func BuildFakeRolloutsCache(clusterNamespaceNumDepsMap map[string]map[string]int) {
	for clusterID, namespaceNumDeps := range clusterNamespaceNumDepsMap {
		for namespace, numRollouts := range namespaceNumDeps {
			for i := 0; i < numRollouts; i++ {
				name := fmt.Sprintf("rollout-%d", i)
				assetAlias := fmt.Sprintf("intuit.%s.rollout", namespace)
				appName := fmt.Sprintf("app-%d", i)
				env := fmt.Sprintf("env-%d", i)
				rollout := resourcebuilder.BuildFakeRollout(name, assetAlias, appName, env, namespace)
				cache.Rollouts.Add(clusterID, rollout)
			}
		}
	}
}
