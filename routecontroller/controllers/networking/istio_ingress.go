package networking

import (
	"context"
	"fmt"

	istionetworkingv1alpha3 "code.cloudfoundry.org/cf-k8s-networking/routecontroller/apis/istio/networking/v1alpha3"
	networkingv1alpha1 "code.cloudfoundry.org/cf-k8s-networking/routecontroller/apis/networking/v1alpha1"
	"code.cloudfoundry.org/cf-k8s-networking/routecontroller/resourcebuilders"
	"github.com/go-logr/logr"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

type IstioIngressProvider struct {
	client.Client
	IngressGateway string
}

func (p *IstioIngressProvider) ReconcileIngressResources(ctx context.Context, log logr.Logger, routes *networkingv1alpha1.RouteList) error {
	vsb := resourcebuilders.VirtualServiceBuilder{IstioGateways: []string{p.IngressGateway}}
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
		result, err := controllerutil.CreateOrUpdate(ctx, p.Client, virtualService, mutateFn)
		if err != nil {
			return err
		}
		log.Info(fmt.Sprintf("VirtualService %s/%s has been %s", virtualService.Namespace, virtualService.Name, result))
	}

	return nil
}

func (p *IstioIngressProvider) DeleteIngressResource(ctx context.Context, log logr.Logger, route *networkingv1alpha1.Route, namespace string) error {
	vs := &istionetworkingv1alpha3.VirtualService{}
	vsName := resourcebuilders.VirtualServiceName(route.FQDN())
	namespacedVSName := types.NamespacedName{Namespace: namespace, Name: vsName}
	if err := p.Get(ctx, namespacedVSName, vs); err != nil {
		if apierrors.IsNotFound(err) {
			log.Info("VirtualService no longer exists")
		}
	} else {
		err := p.Delete(ctx, vs)
		if err != nil {
			return err
		}
		log.Info(fmt.Sprintf("VirtualService %s/%s has been deleted", vs.Namespace, vs.Name))
	}

	return nil
}
