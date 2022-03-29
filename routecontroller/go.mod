module code.cloudfoundry.org/cf-k8s-networking/routecontroller

go 1.15

require (
	github.com/go-logr/logr v0.3.0
	github.com/gogo/protobuf v1.3.2
	github.com/onsi/ginkgo v1.14.1
	github.com/onsi/gomega v1.10.2
	github.com/prometheus/client_model v0.2.0
	github.com/prometheus/prom2json v1.3.0
	github.com/sirupsen/logrus v1.6.0
	istio.io/api v0.0.0-20200410141105-715a3039a0b5
	k8s.io/api v0.20.4
	k8s.io/apimachinery v0.20.4
	k8s.io/client-go v0.20.4
	sigs.k8s.io/controller-runtime v0.8.3
	sigs.k8s.io/kind v0.10.0
)
