package resourcebuilders

import (
	networkingv1alpha1 "code.cloudfoundry.org/cf-k8s-networking/routecontroller/apis/networking/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

type ServiceBuilder struct{}

func (b *ServiceBuilder) BuildMutateFunction(actualService, desiredService *corev1.Service) controllerutil.MutateFn {
	return func() error {
		actualService.ObjectMeta.Labels = desiredService.ObjectMeta.Labels
		actualService.ObjectMeta.Annotations = desiredService.ObjectMeta.Annotations
		actualService.Spec.Selector = desiredService.Spec.Selector
		actualService.Spec.Ports = desiredService.Spec.Ports
		return nil
	}
}

func (b *ServiceBuilder) Build(route *networkingv1alpha1.Route) []corev1.Service {
	const httpPortName = "http"
	services := []corev1.Service{}
	for _, dest := range route.Spec.Destinations {
		service := corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:        serviceName(dest),
				Namespace:   route.ObjectMeta.Namespace,
				Labels:      map[string]string{},
				Annotations: map[string]string{},
			},
			Spec: corev1.ServiceSpec{
				Selector: dest.Selector.MatchLabels,
				Ports: []corev1.ServicePort{
					{
						Port: int32(*dest.Port),
						Name: httpPortName,
					}},
			},
		}
		service.ObjectMeta.Labels["cloudfoundry.org/app_guid"] = dest.App.Guid
		service.ObjectMeta.Labels["cloudfoundry.org/process_type"] = dest.App.Process.Type
		service.ObjectMeta.Labels["cloudfoundry.org/route_guid"] = route.ObjectMeta.Name
		service.ObjectMeta.Annotations["cloudfoundry.org/route-fqdn"] = route.FQDN()
		services = append(services, service)
	}
	return services
}
