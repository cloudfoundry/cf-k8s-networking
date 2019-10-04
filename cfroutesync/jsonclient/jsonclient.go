package jsonclient

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

//go:generate counterfeiter -o fakes/http_client.go --fake-name HTTPClient . HttpClient
type HttpClient interface {
	Do(*http.Request) (*http.Response, error)
}

type JSONClient struct {
	HTTPClient HttpClient
}

func (c *JSONClient) MakeRequest(request *http.Request, response interface{}) error {
	resp, err := c.HTTPClient.Do(request)
	if err != nil {
		return fmt.Errorf("http client: %w", err)
	}
	defer resp.Body.Close()

	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read body: %w", err)
	}

	if resp.StatusCode != 200 {
		return fmt.Errorf("bad response, code %d: %s", resp.StatusCode, string(respBytes))
	}

	err = json.Unmarshal(respBytes, &response)
	if err != nil {
		return fmt.Errorf("unmarshal json: %w", err)
	}
	return nil
}
