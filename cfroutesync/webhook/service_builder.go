package webhook

import (
	"fmt"

	"code.cloudfoundry.org/cf-k8s-networking/cfroutesync/models"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ServiceBuilder struct{}

func (b *ServiceBuilder) Build(routes []models.Route, template Template) []K8sResource {
	resources := []K8sResource{}
	for _, route := range routes {
		for _, s := range routeToServices(route, template) {
			resources = append(resources, s)
		}
	}
	return resources
}

func routeToServices(route models.Route, template Template) []Service {
	const httpPortName = "http"
	const podLabelPrefix = "cloudfoundry.org/"
	services := []Service{}
	for _, dest := range route.Destinations {
		service := Service{
			ApiVersion: "v1",
			Kind:       "Service",
			ObjectMeta: metav1.ObjectMeta{
				Name:        serviceName(dest),
				Labels:      cloneLabels(template.ObjectMeta.Labels),
				Annotations: map[string]string{},
			},
			Spec: ServiceSpec{
				Selector: map[string]string{
					podLabelPrefix + "app_guid":     dest.App.Guid,
					podLabelPrefix + "process_type": dest.App.Process.Type,
				},
				Ports: []ServicePort{
					{
						Port: dest.Port,
						Name: httpPortName,
					}},
			},
		}
		service.ObjectMeta.Labels["cloudfoundry.org/app"] = dest.App.Guid
		service.ObjectMeta.Labels["cloudfoundry.org/process"] = dest.App.Process.Type
		service.ObjectMeta.Labels["cloudfoundry.org/route"] = route.Guid
		service.ObjectMeta.Annotations["cloudfoundry.org/route-fqdn"] = route.FQDN()
		services = append(services, service)
	}
	return services
}

// service names cannot start with numbers
func serviceName(dest models.Destination) string {
	return fmt.Sprintf("s-%s", dest.Guid)
}
