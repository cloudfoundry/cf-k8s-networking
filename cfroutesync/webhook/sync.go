package webhook

import (
	"code.cloudfoundry.org/cf-k8s-networking/cfroutesync/models"
	"errors"
)

type K8sResource interface{}

type SyncResponse struct {
	Children []K8sResource `json:"children"`
}
type SyncRequest struct {
	Parent BulkSync `json:"parent"`
}

//go:generate counterfeiter -o fakes/k8s_resource_builder.go --fake-name K8sResourceBuilder . K8sResourceBuilder
type K8sResourceBuilder interface {
	Build([]models.Route, Template) []K8sResource
}

//go:generate counterfeiter -o fakes/snapshot_repo.go --fake-name SnapshotRepo . snapshotRepo
type snapshotRepo interface {
	Get() (*models.RouteSnapshot, bool)
}

var UninitializedError = errors.New("uninitialized: have not yet synchronized with cloud controller")

type Lineage struct {
	RouteSnapshotRepo   snapshotRepo
	K8sResourceBuilders []K8sResourceBuilder
}

// Sync generates child resources for a metacontroller /sync request
func (m *Lineage) Sync(syncRequest SyncRequest) (*SyncResponse, error) {
	snapshot, ok := m.RouteSnapshotRepo.Get()
	if !ok {
		return nil, UninitializedError
	}
	children := make([]K8sResource, 0)
	for _, builder := range m.K8sResourceBuilders {
		children = append(children, builder.Build(snapshot.Routes, syncRequest.Parent.Spec.Template)...)
	}

	response := &SyncResponse{
		Children: children,
	}

	return response, nil
}
