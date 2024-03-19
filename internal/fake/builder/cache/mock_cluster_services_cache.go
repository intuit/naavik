package cache

import (
	"fmt"

	"github.com/intuit/naavik/internal/cache"
	resourcebuilder "github.com/intuit/naavik/internal/fake/builder/resource"
)

func BuildFakeServiceCache(clusterNamespaceNumDepsMap map[string]map[string]int) {
	for clusterID, namespaceNumDeps := range clusterNamespaceNumDepsMap {
		for namespace, numDeps := range namespaceNumDeps {
			for i := 0; i < numDeps; i++ {
				name := fmt.Sprintf("svc-%d", i)
				assetAlias := fmt.Sprintf("app.%s.service", namespace)
				appName := fmt.Sprintf("app-%d", i)
				svc := resourcebuilder.BuildFakeService(name, assetAlias, appName, namespace)
				cache.Services.Add(clusterID, svc)
			}
		}
	}
}
