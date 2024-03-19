package istio

import (
	"time"

	"github.com/intuit/naavik/internal/types/context"
	"github.com/intuit/naavik/pkg/logger"
	"github.com/intuit/naavik/pkg/types"
	"istio.io/client-go/pkg/apis/networking/v1alpha3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Create Virtual Service.
func (i *istioClientData) CreateVirtualService(ctx context.Context, virtualService *v1alpha3.VirtualService, options metav1.CreateOptions) (*v1alpha3.VirtualService, error) {
	virtualService.Annotations[types.LastUpdatedTimestampKey] = time.Now().UTC().Format(time.RFC3339)
	vs, err := i.istioClient.NetworkingV1alpha3().VirtualServices(virtualService.Namespace).Create(context.Background(), virtualService, options)
	if err != nil {
		ctx.Log.Str(logger.ClusterKey, i.clusterID).Str(logger.OperationKey, "Create").Str(logger.NameKey, virtualService.Name).Str(logger.NamespaceKey, virtualService.Namespace).Str(logger.ErrorKey, err.Error()).Error("error creating virtual service")
		return nil, err
	}
	ctx.Log.Str(logger.ClusterKey, i.clusterID).Str(logger.OperationKey, "Create").Str(logger.NameKey, vs.Name).Str(logger.NamespaceKey, vs.Namespace).Info("virtual service created")
	return vs, nil
}

// Update Virtual Service.
func (i *istioClientData) UpdateVirtualService(ctx context.Context, virtualService *v1alpha3.VirtualService, options metav1.UpdateOptions) (*v1alpha3.VirtualService, error) {
	virtualService.Annotations[types.LastUpdatedTimestampKey] = time.Now().UTC().Format(time.RFC3339)
	vs, err := i.istioClient.NetworkingV1alpha3().VirtualServices(virtualService.Namespace).Update(context.Background(), virtualService, options)
	if err != nil {
		ctx.Log.Str(logger.ClusterKey, i.clusterID).Str(logger.OperationKey, "Update").Str(logger.NameKey, virtualService.Name).Str(logger.NamespaceKey, virtualService.Namespace).Str(logger.ErrorKey, err.Error()).Error("error updating virtual service")
		return nil, err
	}
	ctx.Log.Str(logger.ClusterKey, i.clusterID).Str(logger.OperationKey, "Update").Str(logger.NameKey, vs.Name).Str(logger.NamespaceKey, vs.Namespace).Info("virtual service updated")
	return vs, nil
}

// Delete Virtual Service.
func (i *istioClientData) DeleteVirtualService(ctx context.Context, name string, namespace string, options metav1.DeleteOptions) error {
	err := i.istioClient.NetworkingV1alpha3().VirtualServices(namespace).Delete(context.Background(), name, options)
	if err != nil {
		ctx.Log.Str(logger.ClusterKey, i.clusterID).Str(logger.OperationKey, "Delete").Str(logger.NameKey, name).Str(logger.NamespaceKey, namespace).Str(logger.ErrorKey, err.Error()).Error("error deleting virtual service")
		return err
	}
	ctx.Log.Str(logger.ClusterKey, i.clusterID).Str(logger.OperationKey, "Delete").Str(logger.NameKey, name).Str(logger.NamespaceKey, namespace).Info("virtual service deleted")
	return nil
}

// Get Virtual Service.
func (i *istioClientData) GetVirtualService(_ context.Context, name string, namespace string, options metav1.GetOptions) (*v1alpha3.VirtualService, error) {
	return i.istioClient.NetworkingV1alpha3().VirtualServices(namespace).Get(context.Background(), name, options)
}

// List Virtual Services.
func (i *istioClientData) ListVirtualServices(_ context.Context, namespace string, options metav1.ListOptions) (*v1alpha3.VirtualServiceList, error) {
	return i.istioClient.NetworkingV1alpha3().VirtualServices(namespace).List(context.Background(), options)
}
