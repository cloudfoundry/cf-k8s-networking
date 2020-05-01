/*

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package networking

import (
	"context"
	"fmt"

	"code.cloudfoundry.org/cf-k8s-networking/routecontroller/resourcebuilders"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	istionetworkingv1alpha3 "code.cloudfoundry.org/cf-k8s-networking/routecontroller/apis/istio/networking/v1alpha3"
	networkingv1alpha1 "code.cloudfoundry.org/cf-k8s-networking/routecontroller/apis/networking/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
)

// RouteReconciler reconciles a Route object
type RouteReconciler struct {
	client.Client
	Log          logr.Logger
	Scheme       *runtime.Scheme
	IstioGateway string
}

const fqdnFieldKey string = "spec.fqdn"
const serviceOwnerKey string = "spec.owner"

// +kubebuilder:rbac:groups=networking.cloudfoundry.org,resources=routes,verbs=get;list;watch
// +kubebuilder:rbac:groups=networking.cloudfoundry.org,resources=routes/status,verbs=get;update;patch

func (r *RouteReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("route", req.NamespacedName)

	routes := &networkingv1alpha1.RouteList{}
	route := &networkingv1alpha1.Route{}

	if err := r.Get(ctx, req.NamespacedName, route); err != nil {
		log.Error(err, "unable to fetch Route")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	err := r.List(ctx, routes, client.InNamespace(req.Namespace), client.MatchingFields{fqdnFieldKey: route.FQDN()})
	if err != nil {
		log.Error(err, "failed to list routes")
	}

	vsb := resourcebuilders.VirtualServiceBuilder{IstioGateways: []string{r.IstioGateway}}
	sb := resourcebuilders.ServiceBuilder{}

	virtualservices := vsb.Build(routes)
	services := sb.Build(route)

	servicesForRoute := &corev1.ServiceList{}
	// get services owned by that route
	err = r.List(ctx, servicesForRoute, client.InNamespace(req.Namespace), client.MatchingFields{serviceOwnerKey: string(route.ObjectMeta.UID)})
	if err != nil {
		log.Error(err, "failed to list services")
	}

	for _, desiredService := range services {
		service := &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      desiredService.ObjectMeta.Name,
				Namespace: desiredService.ObjectMeta.Namespace,
			},
		}
		mutateFn := sb.BuildMutateFunction(service, &desiredService)
		result, err := controllerutil.CreateOrUpdate(ctx, r.Client, service, mutateFn)
		if err != nil {
			log.Error(err, fmt.Sprintf("Service %s/%s could not be created or updated", service.Namespace, service.Name))
		} else {
			log.Info(fmt.Sprintf("Service %s/%s has been %s", service.Namespace, service.Name, result))
		}
	}

	servicesToDelete := []corev1.Service{}
	for _, existingService := range servicesForRoute.Items {
		log.Info(fmt.Sprintf("existing service: %+v", existingService))
		foundInDesired := false
		for _, desiredService := range services {
			if desiredService.Name == existingService.Name &&
				desiredService.Namespace == existingService.Namespace {
				foundInDesired = true
			}
		}
		if !foundInDesired {
			servicesToDelete = append(servicesToDelete, existingService)
		}
	}

	log.Info(fmt.Sprintf("Found %d services to delete", len(servicesToDelete)))
	for _, service := range servicesToDelete {
		err := r.Delete(ctx, &service)
		if err != nil {
			log.Error(err, fmt.Sprintf("Service %s/%s could not be deleted", service.Namespace, service.Name))
			return ctrl.Result{}, err
		}
		log.Info(fmt.Sprintf("Service %s/%s has been deleted", service.Namespace, service.Name))
	}

	for _, desiredVirtualService := range virtualservices {
		virtualService := &istionetworkingv1alpha3.VirtualService{
			ObjectMeta: metav1.ObjectMeta{
				Name:      desiredVirtualService.ObjectMeta.Name,
				Namespace: desiredVirtualService.ObjectMeta.Namespace,
			},
		}
		mutateFn := vsb.BuildMutateFunction(virtualService, &desiredVirtualService)
		result, err := controllerutil.CreateOrUpdate(ctx, r.Client, virtualService, mutateFn)
		if err != nil {
			return ctrl.Result{}, err
		}
		log.Info(fmt.Sprintf("VirtualService %s/%s has been %s", virtualService.Namespace, virtualService.Name, result))
	}

	return ctrl.Result{}, nil
}

func (r *RouteReconciler) SetupWithManager(mgr ctrl.Manager) error {
	err := mgr.GetFieldIndexer().IndexField(&networkingv1alpha1.Route{}, fqdnFieldKey, func(rawObj runtime.Object) []string {
		route := rawObj.(*networkingv1alpha1.Route)
		return []string{route.FQDN()}
	})
	if err != nil {
		return err
	}

	err = mgr.GetFieldIndexer().IndexField(&corev1.Service{}, serviceOwnerKey, func(rawObj runtime.Object) []string {
		service := rawObj.(*corev1.Service)
		if len(service.ObjectMeta.OwnerReferences) == 0 {
			return []string{}
		}
		return []string{string(service.ObjectMeta.OwnerReferences[0].UID)}
	})
	if err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&networkingv1alpha1.Route{}).
		Complete(r)
}
