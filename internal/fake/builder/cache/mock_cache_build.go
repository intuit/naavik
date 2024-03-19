package cache

import (
	"fmt"
)

func CreateClusterNamespaceNumResourcesMap(noOfClusters int, noOfNamespacesInEachCluster int, noOfResourcesInNamespace int) map[string]map[string]int {
	clusterNamespaceNumDepsMap := make(map[string]map[string]int)
	for i := 0; i < noOfClusters; i++ {
		clusterID := fmt.Sprintf("cluster-%d", i)
		clusterNamespaceNumDepsMap[clusterID] = make(map[string]int)
		for j := 0; j < noOfNamespacesInEachCluster; j++ {
			namespace := fmt.Sprintf("namespace-%d", j)
			clusterNamespaceNumDepsMap[clusterID][namespace] = noOfResourcesInNamespace
		}
	}
	return clusterNamespaceNumDepsMap
}
