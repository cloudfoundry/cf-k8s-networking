package resourcebuilders

import (
	"crypto/sha256"
	"errors"
	"fmt"

	networkingv1alpha1 "code.cloudfoundry.org/cf-k8s-networking/routecontroller/apis/networking/v1alpha1"
	hpv1 "github.com/projectcontour/contour/apis/projectcontour/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

type HTTPProxyBuilder struct {
}

func HTTPProxyName(fqdn string) string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(fqdn)))
}

func (b *HTTPProxyBuilder) BuildMutateFunction(actual, desired *hpv1.HTTPProxy) controllerutil.MutateFn {
	return func() error {
		actual.ObjectMeta.Labels = desired.ObjectMeta.Labels
		actual.ObjectMeta.Annotations = desired.ObjectMeta.Annotations
		actual.ObjectMeta.OwnerReferences = desired.ObjectMeta.OwnerReferences
		actual.Spec = desired.Spec
		return nil
	}
}

func (b *HTTPProxyBuilder) Build(routes *networkingv1alpha1.RouteList) ([]hpv1.HTTPProxy, error) {
	resources := []hpv1.HTTPProxy{}

	routesForFQDN := groupByFQDN(routes)
	sortedFQDNs := sortFQDNs(routesForFQDN)

	for _, fqdn := range sortedFQDNs {
		httpProxy, err := b.fqdnToHTTPProxy(fqdn, routesForFQDN[fqdn])
		if err != nil {
			return []hpv1.HTTPProxy{}, err
		}

		resources = append(resources, httpProxy)
	}

	return resources, nil
}

func (b *HTTPProxyBuilder) fqdnToHTTPProxy(fqdn string, routes []networkingv1alpha1.Route) (hpv1.HTTPProxy, error) {
	name := HTTPProxyName(fqdn)
	hp := hpv1.HTTPProxy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: routes[0].ObjectMeta.Namespace,
			Labels:    map[string]string{},
			Annotations: map[string]string{
				"cloudfoundry.org/fqdn": fqdn,
			},
			OwnerReferences: []metav1.OwnerReference{},
		},
		Spec: hpv1.HTTPProxySpec{
			VirtualHost: &hpv1.VirtualHost{
				Fqdn: fqdn,
			},
		},
	}

	err := validateRoutesForFQDN(routes)
	if err != nil {
		return hpv1.HTTPProxy{}, err
	}

	sortRoutes(routes)

	for _, route := range routes {
		hp.ObjectMeta.OwnerReferences = append(hp.ObjectMeta.OwnerReferences, routeToOwnerRef(&route))
		var routeServices []hpv1.Service
		if len(route.Spec.Destinations) != 0 {
			routeServices, err = destinationsToServices(route, route.Spec.Destinations)
			if err != nil {
				return hpv1.HTTPProxy{}, err
			}
		} else if len(routes) > 1 {
			continue
		} else {
			routeServices = httpProxyNoDestinationsPlaceholder()
		}

		hpRoute := hpv1.Route{
			Services: routeServices,
		}
		if route.Spec.Path != "" {
			hpRoute.Conditions = []hpv1.MatchCondition{{
				Prefix: route.Spec.Path,
			}}
		}
		hp.Spec.Routes = append(hp.Spec.Routes, hpRoute)
	}

	return hp, nil
}

func destinationsToServices(route networkingv1alpha1.Route, destinations []networkingv1alpha1.RouteDestination) ([]hpv1.Service, error) {
	err := validateHTTPProxyWeights(route, destinations)
	if err != nil {
		return nil, err
	}
	routeServices := make([]hpv1.Service, 0)
	for _, destination := range destinations {
		routeService := hpv1.Service{
			Name: serviceName(destination), // comes from service_builder, will add later
			Port: 8080,
			RequestHeadersPolicy: &hpv1.HeadersPolicy{
				Set: []hpv1.HeaderValue{{
					Name:  "CF-App-Id",
					Value: destination.App.Guid,
				}, {
					Name:  "CF-Space-Id",
					Value: route.ObjectMeta.Labels["cloudfoundry.org/space_guid"],
				}, {
					Name:  "CF-Organization-Id",
					Value: route.ObjectMeta.Labels["cloudfoundry.org/org_guid"],
				}, {
					Name:  "CF-App-Process-Type",
					Value: destination.App.Process.Type,
				}},
			},
		}
		if destination.Weight != nil {
			routeService.Weight = int64(*destination.Weight)
		}
		routeServices = append(routeServices, routeService)
	}
	return routeServices, nil
}

func validateHTTPProxyWeights(route networkingv1alpha1.Route, destinations []networkingv1alpha1.RouteDestination) error {
	// Cloud Controller validates these scenarios
	//
	for _, d := range destinations {
		if (d.Weight == nil) != (destinations[0].Weight == nil) {
			msg := fmt.Sprintf(
				"invalid destinations for route %s: weights must be set on all or none",
				route.ObjectMeta.Name)
			return errors.New(msg)
		}
	}

	return nil
}

func httpProxyNoDestinationsPlaceholder() []hpv1.Service {
	const PLACEHOLDER_NON_EXISTING_DESTINATION = "no-destinations"

	return []hpv1.Service{{
		Name: PLACEHOLDER_NON_EXISTING_DESTINATION,
		Port: 8080,
	}}
}
