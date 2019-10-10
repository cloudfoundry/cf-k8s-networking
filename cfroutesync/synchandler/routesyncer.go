package synchandler

import (
	"errors"

	"code.cloudfoundry.org/cf-k8s-networking/cfroutesync/models"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type SyncResponse struct {
	Children []*RouteCRD `json:"children"`
}
type SyncRequest struct {
	Parent BulkSync `json:"parent"`
}

type BulkSync struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              BulkSyncSpec `json:"spec"`
}

type BulkSyncSpec struct {
	Selector ParentSelector `json:"selector"`
	Template ParentTemplate `json:"template"`
}

type ParentSelector struct {
	MatchLabels map[string]string `json:"matchLabels"`
}

type ParentTemplate struct {
	metav1.ObjectMeta `json:"metadata"`
}

type RouteCRD struct {
	ApiVersion        string `json:"apiVersion"`
	Kind              string `json:"kind"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              RouteCRDSpec `json:"spec"`
}

type RouteCRDSpec struct {
	Host         string                `json:"host"`
	Path         string                `json:"path"`
	Destinations []RouteCRDDestination `json:"destinations"`
}

type RouteCRDDestination struct {
	Guid   string                 `json:"guid"`
	Port   int                    `json:"port"`
	Weight *int                   `json:"weight,omitempty"`
	App    RouteCRDDestinationApp `json:"app"`
}

type RouteCRDDestinationApp struct {
	Guid    string `json:"guid"`
	Process string `json:"process"`
}

//go:generate counterfeiter -o fakes/snapshot_repo.go --fake-name SnapshotRepo . snapshotRepo
type snapshotRepo interface {
	Get() (*models.RouteSnapshot, bool)
}

var UninitializedError = errors.New("uninitialized: have not yet synchronized with cloud controller")

type RouteSyncer struct {
	RouteSnapshotRepo snapshotRepo
}

func (m *RouteSyncer) Sync(syncRequest SyncRequest) (*SyncResponse, error) {
	snapshot, ok := m.RouteSnapshotRepo.Get()
	if !ok {
		return nil, UninitializedError
	}
	crds := snapshotToCRDList(snapshot, &syncRequest.Parent.Spec.Template)
	response := &SyncResponse{
		Children: crds,
	}

	return response, nil
}

func snapshotToCRDList(snapshot *models.RouteSnapshot, template *ParentTemplate) []*RouteCRD {
	crds := make([]*RouteCRD, len(snapshot.Routes))

	for i, route := range snapshot.Routes {
		crds[i] = routeToCRD(route, template)
	}
	return crds
}

func routeToCRD(route *models.Route, template *ParentTemplate) *RouteCRD {
	crd := RouteCRD{
		ApiVersion: "apps.cloudfoundry.org/v1alpha1",
		Kind:       "Route",
		ObjectMeta: metav1.ObjectMeta{
			Name:   route.Guid,
			Labels: template.ObjectMeta.Labels,
		},
		Spec: RouteCRDSpec{
			Host: route.Host,
			Path: route.Path,
		},
	}

	crd.Spec.Destinations = make([]RouteCRDDestination, len(route.Destinations))
	for i, routeDest := range route.Destinations {
		crd.Spec.Destinations[i] = RouteCRDDestination{
			Guid: routeDest.Guid,
			App: RouteCRDDestinationApp{
				Guid:    routeDest.App.Guid,
				Process: routeDest.App.Process,
			},
			Weight: routeDest.Weight,
			Port:   routeDest.Port,
		}
	}

	return &crd
}
