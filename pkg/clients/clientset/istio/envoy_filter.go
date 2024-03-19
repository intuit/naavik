package istio

import (
	"errors"
	"strings"
	"time"

	"github.com/intuit/naavik/internal/types/context"
	"github.com/intuit/naavik/pkg/logger"
	"github.com/intuit/naavik/pkg/types"
	"istio.io/client-go/pkg/apis/networking/v1alpha3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (i *istioClientData) GetEnvoyFilter(_ context.Context, name string, namespace string, options metav1.GetOptions) (*v1alpha3.EnvoyFilter, error) {
	return i.istioClient.NetworkingV1alpha3().EnvoyFilters(namespace).Get(context.Background(), name, options)
}

func (i *istioClientData) CreateEnvoyFilter(ctx context.Context, envoyFilter *v1alpha3.EnvoyFilter, options metav1.CreateOptions) (*v1alpha3.EnvoyFilter, error) {
	envoyFilter.Annotations[types.LastUpdatedTimestampKey] = time.Now().UTC().Format(time.RFC3339)
	ef, err := i.istioClient.NetworkingV1alpha3().EnvoyFilters(envoyFilter.Namespace).Create(context.Background(), envoyFilter, options)
	if err != nil {
		ctx.Log.Str(logger.ClusterKey, i.clusterID).Str(logger.OperationKey, "Create").Str(logger.NameKey, envoyFilter.Name).Str(logger.NamespaceKey, envoyFilter.Namespace).Str(logger.ErrorKey, err.Error()).Error("error creating envoy filter")
		return nil, err
	}
	ctx.Log.Str(logger.ClusterKey, i.clusterID).Str(logger.OperationKey, "Create").Str(logger.NameKey, ef.Name).Str(logger.NamespaceKey, ef.Namespace).Info("envoy filter created")
	return ef, nil
}

func (i *istioClientData) UpdateEnvoyFilter(ctx context.Context, envoyFilter *v1alpha3.EnvoyFilter, options metav1.UpdateOptions) (*v1alpha3.EnvoyFilter, error) {
	envoyFilter.Annotations[types.LastUpdatedTimestampKey] = time.Now().UTC().Format(time.RFC3339)
	ef, err := i.istioClient.NetworkingV1alpha3().EnvoyFilters(envoyFilter.Namespace).Update(context.Background(), envoyFilter, options)
	if err != nil {
		ctx.Log.Str(logger.ClusterKey, i.clusterID).Str(logger.OperationKey, "Update").Str(logger.NameKey, envoyFilter.Name).Str(logger.NamespaceKey, envoyFilter.Namespace).Str(logger.ErrorKey, err.Error()).Error("error updating envoy filter")
		return nil, err
	}
	ctx.Log.Str(logger.ClusterKey, i.clusterID).Str(logger.OperationKey, "Update").Str(logger.NameKey, ef.Name).Str(logger.NamespaceKey, ef.Namespace).Info("envoy filter updated")
	return ef, nil
}

func (i *istioClientData) DeleteEnvoyFilter(ctx context.Context, name string, namespace string, options metav1.DeleteOptions) error {
	err := i.istioClient.NetworkingV1alpha3().EnvoyFilters(namespace).Delete(context.Background(), name, options)
	if err != nil {
		ctx.Log.Str(logger.ClusterKey, i.clusterID).Str(logger.OperationKey, "Delete").Str(logger.NameKey, name).Str(logger.NamespaceKey, namespace).Str(logger.ErrorKey, err.Error()).Error("error deleting envoy filter")
		return err
	}
	ctx.Log.Str(logger.ClusterKey, i.clusterID).Str(logger.OperationKey, "Delete").Str(logger.NameKey, name).Str(logger.NamespaceKey, namespace).Info("envoy filter deleted")
	return nil
}

func (i *istioClientData) ListEnvoyFilters(_ context.Context, namespace string, options metav1.ListOptions) (*v1alpha3.EnvoyFilterList, error) {
	return i.istioClient.NetworkingV1alpha3().EnvoyFilters(namespace).List(context.Background(), options)
}

func checkIfPresent(filter *v1alpha3.EnvoyFilter, filterList []*v1alpha3.EnvoyFilter) bool {
	for _, f := range filterList {
		if strings.EqualFold(f.Name, filter.Name) {
			return true
		}
	}
	return false
}

func (i *istioClientData) CreateEnvoyFilters(ctx context.Context, filterList []*v1alpha3.EnvoyFilter) error {
	var errorList error
	for _, f := range filterList {
		_, err := i.CreateEnvoyFilter(ctx, f, metav1.CreateOptions{})
		if err != nil {
			errors.Join(errorList, err)
		}
	}
	return errorList
}

func (i *istioClientData) DeleteEnvoyFilters(ctx context.Context, filterList []*v1alpha3.EnvoyFilter) error {
	var errorList error
	for _, f := range filterList {
		err := i.DeleteEnvoyFilter(ctx, f.Name, f.Namespace, metav1.DeleteOptions{})
		if err != nil {
			errors.Join(errorList, err)
		}
	}
	return errorList
}

func (i *istioClientData) UpdateEnvoyFilters(ctx context.Context, filterList []*v1alpha3.EnvoyFilter) error {
	var errorList error
	for _, f := range filterList {
		_, err := i.UpdateEnvoyFilter(ctx, f, metav1.UpdateOptions{})
		if err != nil {
			errors.Join(errorList, err)
		}
	}
	return errorList
}

func (i *istioClientData) ApplyEnvoyFilters(ctx context.Context,
	requestedEnvoyFilterList []*v1alpha3.EnvoyFilter, existingEnvoyFilterList *v1alpha3.EnvoyFilterList,
) error {
	filtersToBeDeleted := make([]*v1alpha3.EnvoyFilter, 0)
	filtersToBeCreated := make([]*v1alpha3.EnvoyFilter, 0)
	filtersToBeUpdated := make([]*v1alpha3.EnvoyFilter, 0)
	if len(requestedEnvoyFilterList) > 0 {
		for _, requestedFilter := range requestedEnvoyFilterList {
			if checkIfPresent(requestedFilter, existingEnvoyFilterList.Items) {
				filtersToBeUpdated = append(filtersToBeUpdated, requestedFilter)
			} else {
				filtersToBeCreated = append(filtersToBeCreated, requestedFilter)
			}
		}
		for _, existingFilter := range existingEnvoyFilterList.Items {
			if !checkIfPresent(existingFilter, requestedEnvoyFilterList) {
				filtersToBeDeleted = append(filtersToBeDeleted, existingFilter)
			}
		}
	} else {
		filtersToBeDeleted = append(filtersToBeDeleted, existingEnvoyFilterList.Items...)
	}

	var err error

	if len(filtersToBeCreated) > 0 {
		createerr := i.CreateEnvoyFilters(ctx, filtersToBeCreated)
		errors.Join(err, createerr)
	}
	if len(filtersToBeUpdated) > 0 {
		updateerr := i.UpdateEnvoyFilters(ctx, filtersToBeUpdated)
		if err != nil {
			errors.Join(err, updateerr)
		}
	}

	if len(filtersToBeDeleted) > 0 {
		deleteerr := i.DeleteEnvoyFilters(ctx, filtersToBeDeleted)
		if err != nil {
			errors.Join(err, deleteerr)
		}
	}
	return err
}
