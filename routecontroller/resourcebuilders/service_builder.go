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
		actualService.ObjectMeta.OwnerReferences = desiredService.ObjectMeta.OwnerReferences
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
				OwnerReferences: []metav1.OwnerReference{routeToOwnerRef(route)},
				Name:            serviceName(dest),
				Namespace:       route.ObjectMeta.Namespace,
				Labels:          map[string]string{},
				Annotations:     map[string]string{},
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

func routeToOwnerRef(r *networkingv1alpha1.Route) metav1.OwnerReference {
	return metav1.OwnerReference{
		APIVersion: networkingv1alpha1.SchemeBuilder.GroupVersion.String(),
		Kind:       r.TypeMeta.Kind,
		Name:       r.ObjectMeta.Name,
		UID:        r.ObjectMeta.UID,
	}
}

func boolPtr(x bool) *bool {
	return &x
}
