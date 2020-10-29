package resourcebuilders_test

import (
	"fmt"
	"testing"

	networkingv1alpha1 "code.cloudfoundry.org/cf-k8s-networking/routecontroller/apis/networking/v1alpha1"
	"github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestResourcebuilders(t *testing.T) {
	logrus.SetOutput(GinkgoWriter)
	RegisterFailHandler(Fail)
	RunSpecs(t, "Resourcebuilders Suite")
}

type ingressResourceParams struct {
	fqdn          string
	internal      bool
	https         []httpParams
	owners        []ownerParams
	tlsSecretName string
	httpsOnly     bool
}

type httpParams struct {
	matchPrefix  string
	destinations []destParams
}

type ownerParams struct {
	routeName string
	routeUID  types.UID
}

type destParams struct {
	host         string
	appGUID      string
	spaceGUID    string
	orgGUID      string
	weight       int32
	noHeadersSet bool
}

type routeParams struct {
	name         string
	namespace    string
	host         string
	path         string
	domain       string
	internal     bool
	destinations []routeDestParams
}

type routeDestParams struct {
	destGUID string
	port     int
	weight   *int
	appGUID  string
}

func constructRoute(params routeParams) networkingv1alpha1.Route {
	if params.namespace == "" {
		params.namespace = "workload-namespace"
	}

	route := networkingv1alpha1.Route{
		ObjectMeta: metav1.ObjectMeta{
			Name:      params.name,
			Namespace: params.namespace,
			Labels: map[string]string{
				"cloudfoundry.org/space_guid": "space-guid-0",
				"cloudfoundry.org/org_guid":   "org-guid-0",
			},
			UID: types.UID(fmt.Sprintf("%s-k8s-uid", params.name)),
		},
		TypeMeta: metav1.TypeMeta{
			Kind: "Route",
		},
		Spec: networkingv1alpha1.RouteSpec{
			Host: params.host,
			Path: params.path,
			Url:  fmt.Sprintf("%s.%s%s", params.host, params.domain, params.path),
			Domain: networkingv1alpha1.RouteDomain{
				Name:     params.domain,
				Internal: params.internal,
			},
		},
	}

	destinations := []networkingv1alpha1.RouteDestination{}
	for _, destination := range params.destinations {
		destinations = append(destinations, networkingv1alpha1.RouteDestination{
			Guid:   destination.destGUID,
			Port:   intPtr(destination.port),
			Weight: destination.weight,
			App: networkingv1alpha1.DestinationApp{
				Guid:    destination.appGUID,
				Process: networkingv1alpha1.AppProcess{Type: "process-type-1"},
			},
			Selector: networkingv1alpha1.DestinationSelector{
				MatchLabels: map[string]string{
					"cloudfoundry.org/app_guid":     destination.appGUID,
					"cloudfoundry.org/process_type": "process-type-1",
				},
			},
		})
	}
	route.Spec.Destinations = destinations
	return route
}
func intPtr(x int) *int {
	return &x
}
