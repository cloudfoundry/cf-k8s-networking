package collector

import (
	"net/http"
	"sync"
	"time"

	"code.cloudfoundry.org/cf-k8s-networking/test/uptime/internal/uptime"
)

type Request struct {
	DataPlaneSLOMaxRequestLatency                  time.Duration
	ControlPlaneSLODataPlaneAvailabilityPercentage float64
	Client                                         http.Client

	results   *uptime.ControlPlaneResults
	waitGroup sync.WaitGroup
	mutex     sync.Mutex
}

func (r *Request) Request(route string, delay time.Duration, sampleTime time.Duration) {
	r.waitGroup.Add(1)
	go func() {
		time.Sleep(delay)

		startTime := time.Now()
		var result uptime.DataPlaneResults
		for {
			if time.Since(startTime) > sampleTime {
				r.addResult(result)
				r.waitGroup.Done()
				return
			}

			resp, err, requestLatency := r.timeGetRequest(route)
			if err != nil {
				result.RecordError(err)
				continue
			}

			hasStatusOK := resp.StatusCode == http.StatusOK
			hasMetRequestLatencySLO := requestLatency < r.DataPlaneSLOMaxRequestLatency
			hasPassedSLI := hasStatusOK && hasMetRequestLatencySLO

			result.Record(hasPassedSLI,
				hasStatusOK,
				hasMetRequestLatencySLO,
				requestLatency)
		}
	}()
}

func (r *Request) Wait() {
	r.waitGroup.Wait()
}

func (r *Request) GetResults() *uptime.ControlPlaneResults {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	return r.results
}

func (r *Request) addResult(result uptime.DataPlaneResults) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if r.results == nil {
		r.results = &uptime.ControlPlaneResults{
			ControlPlaneSLODataPlaneAvailabilityPercentage: r.ControlPlaneSLODataPlaneAvailabilityPercentage,
		}
	}

	r.results.AddResult(&result)
}

func (r *Request) timeGetRequest(requestURL string) (*http.Response, error, time.Duration) {
	start := time.Now()
	resp, err := r.Client.Get(requestURL)
	requestLatency := time.Since(start)
	return resp, err, requestLatency
}
