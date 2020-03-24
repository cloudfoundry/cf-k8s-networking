package resourcebuilders

import (
	appsv1alpha1 "code.cloudfoundry.org/cf-k8s-networking/routecontroller/apis/apps/v1alpha1"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ServiceBuilder struct{}

func (b *ServiceBuilder) Build(routes *appsv1alpha1.RouteList) []core.Service {
	resources := []core.Service{}
	for _, route := range routes.Items {
		resources = append(resources, routeToServices(route)...)
	}
	return resources
}

func routeToServices(route appsv1alpha1.Route) []core.Service {
	const httpPortName = "http"
	services := []core.Service{}
	for _, dest := range route.Spec.Destinations {
		service := core.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:        serviceName(dest),
				Namespace:   route.ObjectMeta.Namespace,
				Labels:      map[string]string{},
				Annotations: map[string]string{},
			},
			Spec: core.ServiceSpec{
				Selector: dest.Selector.MatchLabels,
				Ports: []core.ServicePort{
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
