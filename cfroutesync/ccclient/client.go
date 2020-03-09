package ccclient

import (
	"errors"
	"fmt"
	"net/http"
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
	Guid          string
	Host          string
	Path          string
	Url           string
	Destinations  []Destination
	Relationships struct {
		Domain struct {
			Data struct {
				Guid string
			}
		}
		Space struct {
			Data struct {
				Guid string
			}
		}
	}
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

type Domain struct {
	Guid     string
	Name     string
	Internal bool
}

type Space struct {
	Guid          string
	Relationships struct {
		Organization struct {
			Data struct {
				Guid string
			}
		}
	}
}

// determined by CC API: https://v3-apidocs.cloudfoundry.org/version/3.76.0/index.html#get-a-route
const MaxResultsPerPage int = 5000

func (c *Client) ListRoutes(token string) ([]Route, error) {
	pathAndQuery := fmt.Sprintf("v3/routes?per_page=%d", MaxResultsPerPage)

	var response struct {
		Pagination struct {
			TotalPages int `json:"total_pages"`
		}
		Resources []Route
	}

	err := c.getList(pathAndQuery, token, &response)
	if err != nil {
		return nil, err
	}
	if response.Pagination.TotalPages > 1 {
		return nil, errors.New("too many results, paging not implemented")
	}

	return response.Resources, nil
}

func (c *Client) ListDomains(token string) ([]Domain, error) {
	pathAndQuery := fmt.Sprintf("v3/domains?per_page=%d", MaxResultsPerPage)

	var response struct {
		Pagination struct {
			TotalPages int `json:"total_pages"`
		}
		Resources []Domain
	}

	err := c.getList(pathAndQuery, token, &response)
	if err != nil {
		return nil, err
	}
	if response.Pagination.TotalPages > 1 {
		return nil, errors.New("too many results, paging not implemented")
	}

	return response.Resources, nil
}

func (c *Client) ListSpaces(token string) ([]Space, error) {
	pathAndQuery := fmt.Sprintf("v3/spaces?per_page=%d", MaxResultsPerPage)

	var response struct {
		Pagination struct {
			TotalPages int `json:"total_pages"`
		}
		Resources []Space
	}

	err := c.getList(pathAndQuery, token, &response)
	if err != nil {
		return nil, err
	}
	if response.Pagination.TotalPages > 1 {
		return nil, errors.New("too many results, paging not implemented")
	}

	return response.Resources, nil
}

func (c *Client) getList(pathAndQuery string, token string, response interface{}) error {
	reqURL := fmt.Sprintf("%s/%s", c.BaseURL, pathAndQuery)
	request, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return err
	}
	request.Header.Set("Authorization", "bearer "+token)

	err = c.JSONClient.MakeRequest(request, response)
	if err != nil {
		return err
	}
	return nil
}
