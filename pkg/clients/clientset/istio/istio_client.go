package istio

import (
	"github.com/intuit/naavik/internal/types/context"
	"istio.io/client-go/pkg/apis/networking/v1alpha3"
	istioclientset "istio.io/client-go/pkg/clientset/versioned"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ClientInterface interface {
	// GetEnvoyFilter returns the envoy filter for the given name and namespace
	GetEnvoyFilter(ctx context.Context, name string, namespace string, options metav1.GetOptions) (*v1alpha3.EnvoyFilter, error)
	// CreateEnvoyFilter creates the envoy filter
	CreateEnvoyFilter(ctx context.Context, envoyfilter *v1alpha3.EnvoyFilter, options metav1.CreateOptions) (*v1alpha3.EnvoyFilter, error)
	// UpdateEnvoyFilter updates the envoy filter
	UpdateEnvoyFilter(ctx context.Context, envoyfilter *v1alpha3.EnvoyFilter, options metav1.UpdateOptions) (*v1alpha3.EnvoyFilter, error)
	// DeleteEnvoyFilter deletes the envoy filter
	DeleteEnvoyFilter(ctx context.Context, name string, namespace string, option metav1.DeleteOptions) error
	// ListEnvoyFilters lists the envoy filters
	ListEnvoyFilters(ctx context.Context, namespace string, options metav1.ListOptions) (*v1alpha3.EnvoyFilterList, error)
	// CreateEnvoyFilters creates the envoy filters
	CreateEnvoyFilters(ctx context.Context, envoyfilterList []*v1alpha3.EnvoyFilter) error
	// DeleteEnvoyFilters deletes the envoy filters
	DeleteEnvoyFilters(ctx context.Context, envoyfilterList []*v1alpha3.EnvoyFilter) error
	// UpdateEnvoyFilters updates the envoy filters
	UpdateEnvoyFilters(ctx context.Context, envoyfilterList []*v1alpha3.EnvoyFilter) error
	// ApplyEnvoyFilters adds, updates and deletes the envoy filters based on the requested and existing envoy filters
	ApplyEnvoyFilters(ctx context.Context, requestedEnvoyFilterList []*v1alpha3.EnvoyFilter, existingEnvoyFilterList *v1alpha3.EnvoyFilterList) error

	// GetVirtualService returns the virtual service for the given name and namespace
	GetVirtualService(ctx context.Context, name string, namespace string, options metav1.GetOptions) (*v1alpha3.VirtualService, error)
	// CreateVirtualService creates the virtual service
	CreateVirtualService(ctx context.Context, virtualService *v1alpha3.VirtualService, options metav1.CreateOptions) (*v1alpha3.VirtualService, error)
	// UpdateVirtualService updates the virtual service
	UpdateVirtualService(ctx context.Context, virtualService *v1alpha3.VirtualService, options metav1.UpdateOptions) (*v1alpha3.VirtualService, error)
	// DeleteVirtualService deletes the virtual service
	DeleteVirtualService(ctx context.Context, name string, namespace string, options metav1.DeleteOptions) error
	// ListVirtualServices lists the virtual services
	ListVirtualServices(ctx context.Context, namespace string, options metav1.ListOptions) (*v1alpha3.VirtualServiceList, error)
}

type istioClientData struct {
	clusterID   string
	istioClient istioclientset.Interface
}

func NewIstioClient(clusterID string, istioClient istioclientset.Interface) ClientInterface {
	return &istioClientData{
		clusterID:   clusterID,
		istioClient: istioClient,
	}
}
