package workload

import (
	argov1alpha1 "github.com/argoproj/argo-rollouts/pkg/apis/rollouts/v1alpha1"
	"github.com/intuit/naavik/internal/cache"
	k8sAppsV1 "k8s.io/api/apps/v1"
)

type Workload struct {
	Identity   string                `json:"identity"`
	Deployment *k8sAppsV1.Deployment `json:"deployment"`
	Rollout    *argov1alpha1.Rollout `json:"rollout"`
}

type Items struct {
	Identity        string                  `json:"identity"`
	DeploymentItems []*cache.DeploymentItem `json:"deploymentsItems"`
	RolloutItems    []*cache.RolloutItem    `json:"rolloutItems"`
}

type ClusterWorkload struct {
	Cluster         string                 `json:"cluster"`
	DeploymentItems *cache.DeploymentEntry `json:"deploymentsItems"`
	RolloutItems    *cache.RolloutEntry    `json:"rolloutItems"`
}
