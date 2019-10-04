package models

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

func (r *RouteSnapshot) Get() *RouteSnapshot {
	return &RouteSnapshot{}
}
