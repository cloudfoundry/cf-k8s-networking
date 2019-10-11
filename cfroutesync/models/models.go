package models

type RouteSnapshot struct {
	Routes []Route
}

type Route struct {
	Guid         string
	Host         string
	Path         string
	Domain       Domain
	Destinations []Destination
}

type Domain struct {
	Guid     string
	Name     string
	Internal bool
}

type Destination struct {
	Guid   string
	App    App
	Weight *int
	Port   int
}

type App struct {
	Guid    string
	Process Process
}

type Process struct {
	Type string
}

func IntPtr(x int) *int {
	return &x
}
