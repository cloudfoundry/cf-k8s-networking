package networking

import (
	"context"
	"fmt"

	networkingv1alpha1 "code.cloudfoundry.org/cf-k8s-networking/routecontroller/apis/networking/v1alpha1"
	"code.cloudfoundry.org/cf-k8s-networking/routecontroller/resourcebuilders"
	"github.com/go-logr/logr"
	hpv1 "github.com/projectcontour/contour/apis/projectcontour/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

type ContourIngressProvider struct {
	client.Client
	TLSSecretName string
	HTTPSOnly     bool
}

func (p *ContourIngressProvider) ReconcileIngressResources(ctx context.Context, log logr.Logger, routes *networkingv1alpha1.RouteList) error {
	hpb := resourcebuilders.HTTPProxyBuilder{
		TLSSecretName: p.TLSSecretName,
		HTTPSOnly:     p.HTTPSOnly,
	}
	desiredHTTPProxies, err := hpb.Build(routes)
	if err != nil {
		return err
	}

	for _, desiredHTTPProxy := range desiredHTTPProxies {
		httpProxy := &hpv1.HTTPProxy{
			ObjectMeta: metav1.ObjectMeta{
				Name:      desiredHTTPProxy.ObjectMeta.Name,
				Namespace: desiredHTTPProxy.ObjectMeta.Namespace,
			},
		}
		mutateFn := hpb.BuildMutateFunction(httpProxy, &desiredHTTPProxy)
		result, err := controllerutil.CreateOrUpdate(ctx, p.Client, httpProxy, mutateFn)
		if err != nil {
			return err
		}
		log.Info(fmt.Sprintf("HTTPProxy %s/%s has been %s", httpProxy.Namespace, httpProxy.Name, result))
	}

	return nil
}

func (p *ContourIngressProvider) DeleteIngressResource(ctx context.Context, log logr.Logger, route *networkingv1alpha1.Route, namespace string) error {
	httpProxy := &hpv1.HTTPProxy{}
	httpProxyName := resourcebuilders.HTTPProxyName(route.FQDN())
	namespacedHTTPProxyName := types.NamespacedName{Namespace: namespace, Name: httpProxyName}
	if err := p.Get(ctx, namespacedHTTPProxyName, httpProxy); err != nil {
		if apierrors.IsNotFound(err) {
			log.Info("HTTPProxy no longer exists")
		}
	} else {
		err := p.Delete(ctx, httpProxy)
		if err != nil {
			return err
		}
		log.Info(fmt.Sprintf("HTTPProxy %s/%s has been deleted", httpProxy.Namespace, httpProxy.Name))
	}

	return nil
}
