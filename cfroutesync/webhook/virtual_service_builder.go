package webhook

import (
	"code.cloudfoundry.org/cf-k8s-networking/cfroutesync/models"
	"fmt"
	"sort"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type VirtualServiceBuilder struct {
	IstioGateways []string
}

func (b *VirtualServiceBuilder) Build(routes []models.Route, template Template) []K8sResource {
	resources := []K8sResource{}

	routesForFQDN := groupByFQDN(routes)
	sortedFQDNs := sortFQDNs(routesForFQDN)

	for _, fqdn := range sortedFQDNs {
		destinations := destinationsForFQDN(fqdn, routesForFQDN)
		if len(destinations) != 0 {
			resources = append(resources, b.fqdnToVirtualService(fqdn, routesForFQDN[fqdn], template))
		}
	}

	return resources
}

// "mesh" is a special reserved word on Istio VirtualServices
// https://istio.io/docs/reference/config/networking/v1alpha3/virtual-service/#VirtualService
const MeshInternalGateway = "mesh"

func (b *VirtualServiceBuilder) fqdnToVirtualService(fqdn string, routes []models.Route, template Template) VirtualService {
	vs := VirtualService{
		ApiVersion: "networking.istio.io/v1alpha3",
		Kind:       "VirtualService",
		ObjectMeta: metav1.ObjectMeta{
			Name:   fqdn,
			Labels: cloneLabels(template.ObjectMeta.Labels),
		},
		Spec: VirtualServiceSpec{Hosts: []string{fqdn}},
	}

	validateRoutesForFQDN(routes)

	if routes[0].Domain.Internal {
		vs.Spec.Gateways = []string{MeshInternalGateway}
	} else {
		vs.Spec.Gateways = b.IstioGateways
	}

	sortRoutes(routes)

	for _, route := range routes {
		if len(route.Destinations) != 0 {
			istioRoute := HTTPRoute{
				Route: destinationsToHttpRouteDestinations(route.Destinations),
			}
			if route.Path != "" {
				istioRoute.Match = []HTTPMatchRequest{{Uri: HTTPPrefixMatch{Prefix: route.Path}}}
			}
			vs.Spec.Http = append(vs.Spec.Http, istioRoute)
		}
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
		n := route.FQDN()
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
		return routes[i].Url > routes[j].Url
	})
}

func cloneLabels(template map[string]string) map[string]string {
	labels := make(map[string]string)
	for k, v := range template {
		labels[k] = v
	}
	return labels
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

func validateRoutesForFQDN(routes []models.Route) {
	// we are assuming that internal and external routes cannot share an fqdn
	for _, route := range routes {
		if routes[0].Domain.Internal != route.Domain.Internal {
			msg := fmt.Sprintf(
				"route guid %s and route guid %s disagree on whether or not the domain is internal",
				routes[0].Guid,
				route.Guid)
			panic(msg)
		}
	}
}
