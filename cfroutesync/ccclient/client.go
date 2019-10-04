package ccclient

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
)

type Client struct {
	JSONClient jsonClient
	BaseURL    string
}

//go:generate counterfeiter -o fakes/json_client.go --fake-name JSONClient . jsonClient
type jsonClient interface {
	MakeRequest(*http.Request, interface{}) error
}

type Route struct {
	Guid string
	Host string
	Path string
	Url  string
}

type Destination struct {
	Guid string
	App  struct {
		Guid    string
		Process struct {
			Type string
		}
	}
	Weight *int
	Port   int
}

// determined by CC API: https://v3-apidocs.cloudfoundry.org/version/3.76.0/index.html#get-a-route
const MaxResultsPerPage int = 5000

func (c *Client) ListRoutes(token string) ([]Route, error) {
	reqURL := fmt.Sprintf("%s/v3/routes?per_page=%d", c.BaseURL, MaxResultsPerPage)
	request, err := http.NewRequest("GET", reqURL, strings.NewReader(""))
	if err != nil {
		return nil, err
	}
	request.Header.Set("Authorization", "bearer "+token)

	type listRoutesResponse struct {
		Pagination struct {
			TotalPages int `json:"total_pages"`
		}
		Resources []Route
	}
	response := &listRoutesResponse{}

	err = c.JSONClient.MakeRequest(request, response)
	if err != nil {
		return nil, err
	}
	if response.Pagination.TotalPages > 1 {
		return nil, errors.New("too many results, paging not implemented")
	}

	return response.Resources, nil
}

func (c *Client) ListDestinationsForRoute(routeGUID, token string) ([]Destination, error) {
	reqURL := fmt.Sprintf("%s/v3/routes/%s/destinations", c.BaseURL, routeGUID)
	request, err := http.NewRequest("GET", reqURL, strings.NewReader(""))
	if err != nil {
		return nil, err
	}
	request.Header.Set("Authorization", "bearer "+token)

	type listDestinationsResponse struct {
		Destinations []Destination
	}
	response := &listDestinationsResponse{}

	err = c.JSONClient.MakeRequest(request, response)
	if err != nil {
		return nil, err
	}

	return response.Destinations, nil
}
