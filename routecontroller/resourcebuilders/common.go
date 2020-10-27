package resourcebuilders

import (
	"errors"
	"fmt"
	"sort"

	networkingv1alpha1 "code.cloudfoundry.org/cf-k8s-networking/routecontroller/apis/networking/v1alpha1"
)

func validateRoutesForFQDN(routes []networkingv1alpha1.Route) error {
	for _, route := range routes {
		// We are assuming that internal and external routes cannot share an fqdn
		// Cloud Controller should validate and prevent this scenario
		if routes[0].Spec.Domain.Internal != route.Spec.Domain.Internal {
			msg := fmt.Sprintf(
				"route guid %s and route guid %s disagree on whether or not the domain is internal",
				routes[0].ObjectMeta.Name,
				route.ObjectMeta.Name)
			return errors.New(msg)
		}

		// Guard against two Routes for the same fqdn belonging to different namespaces
		if routes[0].ObjectMeta.Namespace != route.ObjectMeta.Namespace {
			msg := fmt.Sprintf(
				"route guid %s and route guid %s share the same FQDN but have different namespaces",
				routes[0].ObjectMeta.Name,
				route.ObjectMeta.Name)
			return errors.New(msg)
		}
	}

	return nil
}
func destinationsForFQDN(fqdn string, routesByFQDN map[string][]networkingv1alpha1.Route) []networkingv1alpha1.RouteDestination {
	destinations := make([]networkingv1alpha1.RouteDestination, 0)
	routes := routesByFQDN[fqdn]
	for _, route := range routes {
		destinations = append(destinations, route.Spec.Destinations...)
	}
	return destinations
}

func groupByFQDN(routes *networkingv1alpha1.RouteList) map[string][]networkingv1alpha1.Route {
	fqdns := make(map[string][]networkingv1alpha1.Route)
	for _, route := range routes.Items {
		n := route.FQDN()
		fqdns[n] = append(fqdns[n], route)
	}
	return fqdns
}

func sortFQDNs(fqdns map[string][]networkingv1alpha1.Route) []string {
	var fqdnSlice []string
	for fqdn := range fqdns {
		fqdnSlice = append(fqdnSlice, fqdn)
	}
	// Sorting so that the results are stable
	sort.Strings(fqdnSlice)
	return fqdnSlice
}

func sortRoutes(routes []networkingv1alpha1.Route) {
	sort.Slice(routes, func(i, j int) bool {
		return routes[i].Spec.Url > routes[j].Spec.Url
	})
}

func cloneLabels(template map[string]string) map[string]string {
	labels := make(map[string]string)
	for k, v := range template {
		labels[k] = v
	}
	return labels
}

func intPtr(x int) *int {
	return &x
}

// service names cannot start with numbers
func serviceName(dest networkingv1alpha1.RouteDestination) string {
	return fmt.Sprintf("s-%s", dest.Guid)
}
