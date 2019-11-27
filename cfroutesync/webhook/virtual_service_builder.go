package webhook

import (
	"code.cloudfoundry.org/cf-k8s-networking/cfroutesync/models"
	"crypto/sha256"
	"errors"
	"fmt"
	"sort"

	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// "mesh" is a special reserved word on Istio VirtualServices
// https://istio.io/docs/reference/config/networking/v1alpha3/virtual-service/#VirtualService
const MeshInternalGateway = "mesh"

// Istio destination weights are percentage based and must sum to 100%
// https://istio.io/docs/concepts/traffic-management/
const IstioExpectedWeight = int(100)

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
			virtualService, err := b.fqdnToVirtualService(fqdn, routesForFQDN[fqdn], template)
			if err == nil {
				resources = append(resources, virtualService)
			} else {
				log.WithError(err).Errorf("unable to create VirtualService for fqdn '%s'", fqdn)
			}
		}
	}

	return resources
}

func (b *VirtualServiceBuilder) fqdnToVirtualService(fqdn string, routes []models.Route, template Template) (VirtualService, error) {
	vs := VirtualService{
		ApiVersion: "networking.istio.io/v1alpha3",
		Kind:       "VirtualService",
		ObjectMeta: metav1.ObjectMeta{
			Name:   VirtualServiceName(fqdn),
			Labels: cloneLabels(template.ObjectMeta.Labels),
			Annotations: map[string]string{
				"cloudfoundry.org/fqdn": fqdn,
			},
		},
		Spec: VirtualServiceSpec{Hosts: []string{fqdn}},
	}

	err := validateRoutesForFQDN(routes)
	if err != nil {
		return VirtualService{}, err
	}

	if routes[0].Domain.Internal {
		vs.Spec.Gateways = []string{MeshInternalGateway}
	} else {
		vs.Spec.Gateways = b.IstioGateways
	}

	sortRoutes(routes)

	for _, route := range routes {
		if len(route.Destinations) != 0 {
			istioDestinations, err := destinationsToHttpRouteDestinations(route, route.Destinations)
			if err != nil {
				return VirtualService{}, err
			}

			istioRoute := HTTPRoute{
				Route: istioDestinations,
			}
			if route.Path != "" {
				istioRoute.Match = []HTTPMatchRequest{{Uri: HTTPPrefixMatch{Prefix: route.Path}}}
			}
			vs.Spec.Http = append(vs.Spec.Http, istioRoute)
		}
	}

	return vs, nil
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
	// Sorting so that the results are stable
	sort.Strings(fqdnSlice)
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

func destinationsToHttpRouteDestinations(route models.Route, destinations []models.Destination) ([]HTTPRouteDestination, error) {
	err := validateWeights(route, destinations)
	if err != nil {
		return nil, err
	}
	httpDestinations := make([]HTTPRouteDestination, 0)
	for _, destination := range destinations {
		httpDestination := HTTPRouteDestination{
			Destination: VirtualServiceDestination{
				Host: serviceName(destination),
			},
			Headers: VirtualServiceHeaders{
				Request: VirtualServiceHeaderOperations{
					Set: map[string]string{
						"App-Id":           destination.App.Guid,
						"App-Process-Type": destination.App.Process.Type,
					},
				},
			},
		}
		if destination.Weight != nil {
			httpDestination.Weight = destination.Weight
		}
		httpDestinations = append(httpDestinations, httpDestination)
	}
	if len(destinations) > 1 && destinations[0].Weight == nil {
		n := len(destinations)
		for i, _ := range httpDestinations {
			weight := int(IstioExpectedWeight / n)
			if i == 0 {
				// pad the first destination's weight to ensure all weights sum to 100
				remainder := IstioExpectedWeight - n*weight
				weight += remainder
			}
			httpDestinations[i].Weight = models.IntPtr(weight)
		}
	}
	return httpDestinations, nil
}

func validateWeights(route models.Route, destinations []models.Destination) error {
	// Cloud Controller validates these scenarios
	//
	weightSum := 0
	for _, d := range destinations {
		if (d.Weight == nil) != (destinations[0].Weight == nil) {
			msg := fmt.Sprintf(
				"invalid destinations for route %s: weights must be set on all or none",
				route.Guid)
			return errors.New(msg)
		}

		if d.Weight != nil {
			weightSum += *d.Weight
		}
	}

	weightsHaveBeenSet := destinations[0].Weight != nil
	if weightsHaveBeenSet && weightSum != IstioExpectedWeight {
		msg := fmt.Sprintf(
			"invalid destinations for route %s: weights must sum up to 100",
			route.Guid)
		return errors.New(msg)
	}
	return nil
}

func validateRoutesForFQDN(routes []models.Route) error {
	// We are assuming that internal and external routes cannot share an fqdn
	// Cloud Controller should validate and prevent this scenario
	for _, route := range routes {
		if routes[0].Domain.Internal != route.Domain.Internal {
			msg := fmt.Sprintf(
				"route guid %s and route guid %s disagree on whether or not the domain is internal",
				routes[0].Guid,
				route.Guid)
			return errors.New(msg)
		}
	}

	return nil
}

// virtual service names cannot contain special characters
func VirtualServiceName(fqdn string) string {
	sum := sha256.Sum256([]byte(fqdn))
	return fmt.Sprintf("vs-%x", sum)
}
