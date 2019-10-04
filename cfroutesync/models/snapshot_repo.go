package models

type SnapshotRepo struct {
	snapshot *RouteSnapshot
}

type RouteSnapshot struct {
	Routes []*Route
}

type Route struct {
	Guid         string
	Host         string
	Path         string
	Destinations []*Destination
}

type Destination struct {
	Guid   string
	App    DestinationApp
	Weight int
	Port   int
}

type DestinationApp struct {
	Guid    string
	Process string
}

func (r *SnapshotRepo) Get() (*RouteSnapshot, bool) {
	return r.snapshot, true
}

func (r *SnapshotRepo) Put(snapshot *RouteSnapshot) {
	r.snapshot = snapshot
}

// TODO: Expect: it is safe to call Get and Replace at the same time
