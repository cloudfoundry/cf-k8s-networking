package webhook

import (
	"errors"

	"code.cloudfoundry.org/cf-k8s-networking/cfroutesync/models"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type SyncResponse struct {
	Children []Route `json:"children"`
}
type SyncRequest struct {
	Parent BulkSync `json:"parent"`
}

//go:generate counterfeiter -o fakes/snapshot_repo.go --fake-name SnapshotRepo . snapshotRepo
type snapshotRepo interface {
	Get() (*models.RouteSnapshot, bool)
}

var UninitializedError = errors.New("uninitialized: have not yet synchronized with cloud controller")

type Lineage struct {
	RouteSnapshotRepo snapshotRepo
}

// Sync generates child resources for a metacontroller /sync request
func (m *Lineage) Sync(syncRequest SyncRequest) (*SyncResponse, error) {
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

func snapshotToCRDList(snapshot *models.RouteSnapshot, template *Template) []Route {
	crds := make([]Route, len(snapshot.Routes))

	for i, route := range snapshot.Routes {
		crds[i] = routeToCRD(route, template)
	}
	return crds
}

func routeToCRD(route models.Route, template *Template) Route {
	crd := Route{
		ApiVersion: "apps.cloudfoundry.org/v1alpha1",
		Kind:       "Route",
		ObjectMeta: metav1.ObjectMeta{
			Name:   route.Guid,
			Labels: template.ObjectMeta.Labels,
		},
		Spec: RouteSpec{
			Selector: Selector{
				MatchLabels: map[string]string{
					"cloudfoundry.org/route": route.Guid,
				},
			},
			Host: route.Host,
			Path: route.Path,
			Domain: Domain{
				Guid:     route.Domain.Guid,
				Name:     route.Domain.Name,
				Internal: route.Domain.Internal,
			},
		},
	}

	crd.Spec.Destinations = make([]Destination, len(route.Destinations))
	for i, routeDest := range route.Destinations {
		crd.Spec.Destinations[i] = Destination{
			Guid: routeDest.Guid,
			App: App{
				Guid:    routeDest.App.Guid,
				Process: Process{Type: routeDest.App.Process.Type},
			},
			Weight: routeDest.Weight,
			Port:   routeDest.Port,
		}
	}

	return crd
}
