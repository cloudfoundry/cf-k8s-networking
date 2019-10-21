package webhook

import (
	"errors"
	"fmt"
	"path"
	"sort"

	"code.cloudfoundry.org/cf-k8s-networking/cfroutesync/models"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type K8sResource interface{}

type SyncResponse struct {
	Children []K8sResource `json:"children"`
}
type SyncRequest struct {
	Parent BulkSync `json:"parent"`
}

//go:generate counterfeiter -o fakes/snapshot_repo.go --fake-name SnapshotRepo . snapshotRepo
type snapshotRepo interface {
	Get() (*models.RouteSnapshot, bool)
}

var UninitializedError = errors.New("uninitialized: have not yet synchronized with cloud controller")

type Lineage struct {
	RouteSnapshotRepo snapshotRepo
	IstioGateways     []string
}

// Sync generates child resources for a metacontroller /sync request
func (m *Lineage) Sync(syncRequest SyncRequest) (*SyncResponse, error) {
	snapshot, ok := m.RouteSnapshotRepo.Get()
	if !ok {
		return nil, UninitializedError
	}
	children := m.snapshotToK8sResources(snapshot, syncRequest.Parent.Spec.Template)
	response := &SyncResponse{
		Children: children,
	}

	return response, nil
}

func (m *Lineage) snapshotToK8sResources(snapshot *models.RouteSnapshot, template Template) []K8sResource {
	resources := make([]K8sResource, 0)
	for _, route := range snapshot.Routes {
		for _, s := range routeToServices(route, template) {
			resources = append(resources, s)
		}
	}

	routesForFQDN := groupByFQDN(snapshot.Routes)
	sortedFQDNs := sortFQDNs(routesForFQDN)

	for _, fqdn := range sortedFQDNs {
		destinations := destinationsForFQDN(fqdn, routesForFQDN)
		if len(destinations) != 0 {
			resources = append(resources, m.fqdnToVirtualService(fqdn, routesForFQDN[fqdn], template))
		}
	}
	return resources
}

func (m *Lineage) fqdnToVirtualService(fqdn string, routes []models.Route, template Template) VirtualService {
	vs := VirtualService{
		ApiVersion: "networking.istio.io/v1alpha3",
		Kind:       "VirtualService",
		ObjectMeta: metav1.ObjectMeta{
			Name:   fqdn,
			Labels: cloneLabels(template.ObjectMeta.Labels),
		},
		Spec: VirtualServiceSpec{Hosts: []string{fqdn}},
	}

	// we are assuming that internal and external routes cannot share an fqdn
	if routes[0].Domain.Internal {
		vs.Spec.Gateways = []string{"mesh"}
	} else {
		vs.Spec.Gateways = m.IstioGateways
	}

	sortRoutes(routes)

	for _, route := range routes {
		istioRoute := HTTPRoute{
			Route: destinationsToHttpRouteDestinations(route.Destinations),
		}
		if route.Path != "" {
			istioRoute.Match = []HTTPMatchRequest{{Uri: HTTPPrefixMatch{Prefix: route.Path}}}
		}
		vs.Spec.Http = append(vs.Spec.Http, istioRoute)
	}

	return vs
}

func destinationsForFQDN(fqdn string, routesByFQDN map[string][]models.Route) []models.Destination {
	destinations := make([]models.Destination, 0)
	routes := routesByFQDN[fqdn]
	for _, route := range routes {
		destinations = append(destinations, route.Destinations...)
	}
	return destinations
}

func groupByFQDN(routes []models.Route) map[string][]models.Route {
	fqdns := make(map[string][]models.Route)
	for _, route := range routes {
		n := fqdn(route)
		fqdns[n] = append(fqdns[n], route)
	}
	return fqdns
}

func sortFQDNs(fqdns map[string][]models.Route) []string {
	var fqdnSlice []string
	for fqdn, _ := range fqdns {
		fqdnSlice = append(fqdnSlice, fqdn)
	}
	sort.Strings(fqdnSlice)
	// so that the results are stable
	return fqdnSlice
}

func sortRoutes(routes []models.Route) {
	sort.Slice(routes, func(i, j int) bool {
		return url(routes[i]) > url(routes[j])
	})
}

func url(route models.Route) string {
	return path.Join(fqdn(route), route.Path)
}

// service names cannot start with numbers
func serviceName(dest models.Destination) string {
	return fmt.Sprintf("s-%s", dest.Guid)
}

func cloneLabels(template map[string]string) map[string]string {
	labels := make(map[string]string)
	for k, v := range template {
		labels[k] = v
	}
	return labels
}

func fqdn(route models.Route) string {
	if route.Host == "" {
		return route.Domain.Name
	}
	return fmt.Sprintf("%s.%s", route.Host, route.Domain.Name)
}

func routeToServices(route models.Route, template Template) []Service {
	services := []Service{}
	for _, dest := range route.Destinations {
		service := Service{
			ApiVersion: "v1",
			Kind:       "Service",
			ObjectMeta: metav1.ObjectMeta{
				Name:   serviceName(dest),
				Labels: cloneLabels(template.ObjectMeta.Labels),
			},
			Spec: ServiceSpec{
				Selector: map[string]string{
					"app_guid":     dest.App.Guid,
					"process_type": dest.App.Process.Type,
				},
				Ports: []ServicePort{{Port: dest.Port}},
			},
		}
		service.ObjectMeta.Labels["cloudfoundry.org/app"] = dest.App.Guid
		service.ObjectMeta.Labels["cloudfoundry.org/process"] = dest.App.Process.Type
		service.ObjectMeta.Labels["cloudfoundry.org/route"] = route.Guid
		service.ObjectMeta.Labels["cloudfoundry.org/route-fqdn"] = fqdn(route)
		services = append(services, service)
	}
	return services
}

func destinationsToHttpRouteDestinations(destinations []models.Destination) []HTTPRouteDestination {
	httpDestinations := make([]HTTPRouteDestination, 0)
	for _, destination := range destinations {
		httpDestination := HTTPRouteDestination{
			Destination: VirtualServiceDestination{Host: serviceName(destination)},
		}
		if destination.Weight != nil {
			httpDestination.Weight = destination.Weight
		}
		httpDestinations = append(httpDestinations, httpDestination)
	}
	return httpDestinations
}
