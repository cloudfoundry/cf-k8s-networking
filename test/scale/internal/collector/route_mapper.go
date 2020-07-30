package collector

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/cloudfoundry-incubator/cf-test-helpers/cf"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	. "github.com/onsi/gomega/gexec"
)

type RouteMapper struct {
	Client http.Client

	results   []float64
	waitGroup sync.WaitGroup
	mutex     sync.Mutex
}

func (r *RouteMapper) MapRoute(appName, domain, routeToDelete, routeToMap string) {
	r.waitGroup.Add(1)

	go func() {
		defer r.waitGroup.Done()
		defer GinkgoRecover()

		session := cfWithRetry("delete-route", domain, "--hostname", routeToDelete, "-f")
		Eventually(session, "30s").Should(Exit(0))

		session = cfWithRetry("map-route", appName, domain, "--hostname", routeToMap)
		Eventually(session, "30s").Should(Exit(0))

		startTime := time.Now().Unix()
		lastFailure := time.Now().Unix()
		for j := 0; j < 40; j++ {
			time.Sleep(1 * time.Second)

			url := fmt.Sprintf("http://%s.%s/status/200", routeToMap, domain)
			resp, err := r.Client.Get(url)
			if err != nil {
				continue
			}

			if resp.StatusCode != http.StatusOK {
				lastFailure = time.Now().Unix()
			}
		}

		r.addResult(float64(lastFailure - startTime))
	}()
}

func cfWithRetry(args ...string) *gexec.Session {
	for i := 0; i < 3; i++ {
		session := cf.Cf(args...)
		session.Wait(5 * time.Second)
		if session.ExitCode() == 0 {
			return session
		}
	}
	Fail("Never successfully ran cf command")
	panic("How did you get here?")
}

func (r *RouteMapper) Wait() {
	r.waitGroup.Wait()
}

func (r *RouteMapper) GetResults() []float64 {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	return r.results
}

func (r *RouteMapper) addResult(result float64) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	r.results = append(r.results, result)
}
