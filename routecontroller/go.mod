module code.cloudfoundry.org/cf-k8s-networking/routecontroller

go 1.13

require (
	github.com/go-logr/logr v0.1.0
	github.com/gogo/protobuf v1.3.1
	github.com/onsi/ginkgo v1.12.1
	github.com/onsi/gomega v1.10.1
	github.com/projectcontour/contour v1.8.1
	github.com/prometheus/client_model v0.2.0
	github.com/prometheus/prom2json v1.3.0
	github.com/sirupsen/logrus v1.4.2
	google.golang.org/appengine v1.6.5 // indirect
	istio.io/api v0.0.0-20200410141105-715a3039a0b5
	k8s.io/api v0.18.6
	k8s.io/apimachinery v0.18.6
	k8s.io/client-go v0.18.6
	sigs.k8s.io/controller-runtime v0.6.2
	sigs.k8s.io/kind v0.7.0
)
