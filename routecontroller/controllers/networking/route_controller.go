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
	"time"

	"code.cloudfoundry.org/cf-k8s-networking/routecontroller/resourcebuilders"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	istionetworkingv1alpha3 "code.cloudfoundry.org/cf-k8s-networking/routecontroller/apis/istio/networking/v1alpha3"
	networkingv1alpha1 "code.cloudfoundry.org/cf-k8s-networking/routecontroller/apis/networking/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
)

// RouteReconciler reconciles a Route object
type RouteReconciler struct {
	client.Client
	Log            logr.Logger
	Scheme         *runtime.Scheme
	IstioGateway   string
	ResyncInterval time.Duration
}

const fqdnFieldKey string = "spec.fqdn"
const serviceOwnerKey string = "spec.owner"
const finalizerName string = "routes.networking.cloudfoundry.org"

// +kubebuilder:rbac:groups=networking.cloudfoundry.org,resources=routes,verbs=get;list;watch
// +kubebuilder:rbac:groups=networking.cloudfoundry.org,resources=routes/status,verbs=get;update;patch

func (r *RouteReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("route", req.NamespacedName)

	routes := &networkingv1alpha1.RouteList{}
	route := &networkingv1alpha1.Route{}

	if err := r.Get(ctx, req.NamespacedName, route); err != nil {
		if apierrors.IsNotFound(err) {
			log.Info("Route no longer exists")
		}
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	err := r.List(ctx, routes, client.InNamespace(req.Namespace), client.MatchingFields{fqdnFieldKey: route.FQDN()})
	if err != nil {
		return ctrl.Result{}, err
	}

	if route.ObjectMeta.DeletionTimestamp.IsZero() {
		if !hasFinalizer(route, finalizerName) {
			controllerutil.AddFinalizer(route, finalizerName)
			if err := r.Update(context.Background(), route); err != nil {
				return ctrl.Result{}, err
			}
		}
	} else {
		// if the DeletionTimestamp is set, we should run the finalizer
		if hasFinalizer(route, finalizerName) {
			err = r.finalizeRouteForDeletion(req, route, routes, log, ctx)
			if err != nil {
				return ctrl.Result{}, err
			}
		}

		return ctrl.Result{}, nil
	}

	err = r.reconcileServices(req, route, log, ctx)
	if err != nil {
		return ctrl.Result{}, err
	}

	err = r.reconcileVirtualServices(req, routes, log, ctx)
	if err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{RequeueAfter: r.ResyncInterval}, nil
}

func (r *RouteReconciler) reconcileServices(req ctrl.Request, route *networkingv1alpha1.Route, log logr.Logger, ctx context.Context) error {
	sb := resourcebuilders.ServiceBuilder{}
	desiredServices := sb.Build(route)

	actualServicesForRoute := &corev1.ServiceList{}
	// get services owned by that route
	err := r.List(ctx, actualServicesForRoute, client.InNamespace(req.Namespace), client.MatchingFields{serviceOwnerKey: string(route.ObjectMeta.UID)})
	if err != nil {
		return err
	}

	for _, desiredService := range desiredServices {
		service := &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      desiredService.ObjectMeta.Name,
				Namespace: desiredService.ObjectMeta.Namespace,
			},
		}
		mutateFn := sb.BuildMutateFunction(service, &desiredService)
		result, err := controllerutil.CreateOrUpdate(ctx, r.Client, service, mutateFn)
		if err != nil {
			return err
		}
		log.Info(fmt.Sprintf("Service %s/%s has been %s", service.Namespace, service.Name, result))
	}

	servicesToDelete := findServicesForDeletion(actualServicesForRoute.Items, desiredServices)
	err = r.deleteServiceList(servicesToDelete, log, ctx)

	return err
}

func (r *RouteReconciler) reconcileVirtualServices(req ctrl.Request, routes *networkingv1alpha1.RouteList, log logr.Logger, ctx context.Context) error {
	vsb := resourcebuilders.VirtualServiceBuilder{IstioGateways: []string{r.IstioGateway}}
	desiredVirtualServices, err := vsb.Build(routes)
	if err != nil {
		return err
	}

	for _, desiredVirtualService := range desiredVirtualServices {
		virtualService := &istionetworkingv1alpha3.VirtualService{
			ObjectMeta: metav1.ObjectMeta{
				Name:      desiredVirtualService.ObjectMeta.Name,
				Namespace: desiredVirtualService.ObjectMeta.Namespace,
			},
		}
		mutateFn := vsb.BuildMutateFunction(virtualService, &desiredVirtualService)
		result, err := controllerutil.CreateOrUpdate(ctx, r.Client, virtualService, mutateFn)
		if err != nil {
			return err
		}
		log.Info(fmt.Sprintf("VirtualService %s/%s has been %s", virtualService.Namespace, virtualService.Name, result))
	}

	return nil
}

func (r *RouteReconciler) finalizeRouteForDeletion(req ctrl.Request, route *networkingv1alpha1.Route, routes *networkingv1alpha1.RouteList, log logr.Logger, ctx context.Context) error {
	actualServicesForRoute := &corev1.ServiceList{}
	err := r.List(ctx, actualServicesForRoute, client.InNamespace(req.Namespace), client.MatchingFields{serviceOwnerKey: string(route.ObjectMeta.UID)})
	if err != nil {
		return err
	}

	err = r.deleteServiceList(actualServicesForRoute.Items, log, ctx)
	if err != nil {
		return err
	}

	routes.Items = removeRouteFromRouteList(route, routes)
	if len(routes.Items) == 0 {
		vs := &istionetworkingv1alpha3.VirtualService{}
		vsName := resourcebuilders.VirtualServiceName(route.FQDN())
		namespacedVSName := types.NamespacedName{Namespace: req.Namespace, Name: vsName}
		if err := r.Get(ctx, namespacedVSName, vs); err != nil {
			if apierrors.IsNotFound(err) {
				log.Info("VirtualService no longer exists")
			}
		} else {
			err := r.Delete(ctx, vs)
			if err != nil {
				return err
			}
			log.Info(fmt.Sprintf("VirtualService %s/%s has been deleted", vs.Namespace, vs.Name))
		}
	} else {
		if err := r.reconcileVirtualServices(req, routes, log, ctx); err != nil {
			return err
		}
	}

	controllerutil.RemoveFinalizer(route, finalizerName)
	if err := r.Update(context.Background(), route); err != nil {
		return err
	}

	return nil
}

func findServicesForDeletion(actualServices, desiredServices []corev1.Service) []corev1.Service {
	servicesToDelete := []corev1.Service{}
	for _, existingService := range actualServices {
		foundInDesired := false
		for _, desiredService := range desiredServices {
			if desiredService.Name == existingService.Name &&
				desiredService.Namespace == existingService.Namespace {
				foundInDesired = true
			}
		}

		if !foundInDesired {
			servicesToDelete = append(servicesToDelete, existingService)
		}
	}

	return servicesToDelete
}

func (r *RouteReconciler) deleteServiceList(services []corev1.Service, log logr.Logger, ctx context.Context) error {
	for _, service := range services {
		err := r.Delete(ctx, &service)
		if err != nil {
			return err
		}
		log.Info(fmt.Sprintf("Service %s/%s has been deleted", service.Namespace, service.Name))
	}
	return nil
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

func hasFinalizer(o metav1.Object, finalizerName string) bool {
	for _, f := range o.GetFinalizers() {
		if f == finalizerName {
			return true
		}
	}
	return false
}

func removeRouteFromRouteList(routeToRemove *networkingv1alpha1.Route, routes *networkingv1alpha1.RouteList) []networkingv1alpha1.Route {
	for idx, route := range routes.Items {
		if route.ObjectMeta.Name == routeToRemove.ObjectMeta.Name {
			return append(routes.Items[:idx], routes.Items[idx+1:]...)
		}
	}

	return routes.Items
}
