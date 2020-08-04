/*
Copyright (c) 2019 StackRox Inc.

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

package main

import (
	"fmt"
	"log"
	"net/http"
	"path/filepath"

	"k8s.io/api/admission/v1beta1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

const (
	tlsDir      = `/run/secrets/tls`
	tlsCertFile = `tls.crt`
	tlsKeyFile  = `tls.key`
)

const (
	workloadsNs             = "cf-workloads"
	serviceAccountNameLabel = "cloudfoundry.org/app_guid"
)

var (
	podResource = metav1.GroupVersionResource{Version: "v1", Resource: "pods"}
)

type Injector struct {
	server     *http.Server
	kubeclient *kubernetes.Clientset
}

// const (
// 	admissionWebhookAnnotationInjectKey = "code.cloudfoundry.org/inject-sa/inject"
// 	admissionWebhookAnnotationStatusKey = "code.cloudfoundry.org/inject-sa/status"
// )
func main() {
	certPath := filepath.Join(tlsDir, tlsCertFile)
	keyPath := filepath.Join(tlsDir, tlsKeyFile)
	injector := &Injector{}

	// creates the in-cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		log.Fatal(err)
	}
	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatal(err)
	}
	injector.kubeclient = clientset

	mux := http.NewServeMux()
	mux.Handle("/mutate", admitFuncHandler(injector.updateServiceAccount))
	server := &http.Server{
		// We listen on port 8443 such that we do not need root privileges or extra capabilities for this server.
		// The Service object will take care of mapping this port to the HTTPS port 443.
		Addr:    ":8443",
		Handler: mux,
	}

	injector.server = server

	log.Fatal(server.ListenAndServeTLS(certPath, keyPath))
}

func (inj *Injector) updateServiceAccount(req *v1beta1.AdmissionRequest) ([]patchOperation, error) {
	// This handler should only get called on Pod objects as per the MutatingWebhookConfiguration in the YAML file.
	// However, if (for whatever reason) this gets invoked on an object of a different kind, issue a log message but
	// let the object request pass through otherwise.
	if req.Namespace != workloadsNs || req.Resource != podResource {
		log.Printf("expect resource to be %s", podResource)
		return nil, nil
	}

	// Parse the Pod object.
	raw := req.Object.Raw
	pod := corev1.Pod{}
	if _, _, err := universalDeserializer.Decode(raw, nil, &pod); err != nil {
		return nil, fmt.Errorf("could not deserialize pod object: %v", err)
	}

	actualServiceAccountName := pod.Spec.ServiceAccountName
	desiredServiceAccountName := inj.getServiceAccountName(&pod)
	if err := inj.checkAndCreateServiceAccount(desiredServiceAccountName, pod.ObjectMeta.OwnerReferences); err != nil {
		return nil, fmt.Errorf("could not create service account object: %v", err)
	}

	// Create patch operations to apply sensible defaults, if those options are not set explicitly.
	var patches []patchOperation
	if actualServiceAccountName != desiredServiceAccountName {
		patches = append(patches, patchOperation{
			Op:   "replace",
			Path: "/spec/serviceAccountName",
			// The value must not be true if runAsUser is set to 0, as otherwise we would create a conflicting
			// configuration ourselves.
			Value: desiredServiceAccountName,
		})
	}

	return patches, nil
}

func (inj *Injector) checkAndCreateServiceAccount(name string, ownerReferences []metav1.OwnerReference) error {
	log.Printf("checking if service account for pod %s exist", name)
	_, err := inj.kubeclient.CoreV1().ServiceAccounts(workloadsNs).Get(name, metav1.GetOptions{})
	if err != nil && !errors.IsNotFound(err) {
		return err
	}

	if err == nil {
		return nil
	}

	log.Printf("service account for %s doesnt't exist, creating it...", name)
	serviceAccount := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:            name,
			OwnerReferences: ownerReferences,
		},
	}
	_, err = inj.kubeclient.CoreV1().ServiceAccounts(workloadsNs).Create(serviceAccount)

	return err
}

func (ink *Injector) getServiceAccountName(pod *corev1.Pod) string {
	return pod.ObjectMeta.Labels[serviceAccountNameLabel]
}
