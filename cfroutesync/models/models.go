package models

import "fmt"

type RouteSnapshot struct {
	Routes []Route
}

type Route struct {
	Guid         string
	Host         string
	Path         string
	Url          string
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

func (r Route) FQDN() string {
	if r.Host == "" {
		return r.Domain.Name
	}
	return fmt.Sprintf("%s.%s", r.Host, r.Domain.Name)
}

func IntPtr(x int) *int {
	return &x
}
