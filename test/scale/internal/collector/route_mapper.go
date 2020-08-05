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
	failures  int
	waitGroup sync.WaitGroup
	mutex     sync.Mutex
}

func (r *RouteMapper) MapRoute(appName, domain, routeToDelete, routeToMap string) {
	r.waitGroup.Add(1)

	go func() {
		defer r.waitGroup.Done()
		defer GinkgoRecover()

		fmt.Println("Deleting:", routeToDelete)
		session := cfWithRetry("delete-route", domain, "--hostname", routeToDelete, "-f")
		Eventually(session, "10s").Should(Exit(0))

		fmt.Println("Route to Map:", routeToMap)
		session = cfWithRetry("map-route", appName, domain, "--hostname", routeToMap)
		Eventually(session, "10s").Should(Exit(0))

		startTime := time.Now().Unix()
		lastFailure := time.Now().Unix()
		succeeded := false
		for j := 0; j < 60; j++ {
			time.Sleep(1 * time.Second)

			url := fmt.Sprintf("http://%s.%s/", routeToMap, domain)
			resp, err := r.Client.Get(url)
			if err != nil {
				continue
			}

			if resp.StatusCode != http.StatusOK {
				lastFailure = time.Now().Unix()
			} else {
				if !succeeded {
					fmt.Println("Success for number", j, "route:", routeToMap)
					succeeded = true
				}
			}
		}

		if !succeeded {
			fmt.Println(routeToMap, "never became healthy this is a problem/failure")
			r.addFailure()
			return
		}

		r.addResult(float64(lastFailure - startTime))
	}()
}

func cfWithRetry(args ...string) *gexec.Session {
	for i := 0; i < 3; i++ {
		session := cf.Cf(args...)
		time.Sleep(2 * time.Second)
		// session.Wait(5 * time.Second)
		if session.ExitCode() == 0 {
			return session
		}
		time.Sleep(10 * time.Second)
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

func (r *RouteMapper) GetFailures() int {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	return r.failures
}

func (r *RouteMapper) addFailure() {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	r.failures = r.failures + 1
}
