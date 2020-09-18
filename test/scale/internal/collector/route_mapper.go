package collector

import (
	"fmt"
	"io/ioutil"
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

	results      []float64
	failures     int
	postFailures int
	waitGroup    sync.WaitGroup
	mutex        sync.Mutex
}

func (r *RouteMapper) MapRoute(appName, domain, routeToDelete, routeToMap string) {
	r.waitGroup.Add(1)

	go func() {
		defer r.waitGroup.Done()
		defer GinkgoRecover()

		fmt.Fprintln(GinkgoWriter, "Deleting:", routeToDelete)
		session := cfWithRetry("delete-route", domain, "--hostname", routeToDelete, "-f")
		Eventually(session, "10s").Should(Exit(0))

		fmt.Fprintln(GinkgoWriter, "Route to Map:", routeToMap)
		session = cfWithRetry("map-route", appName, domain, "--hostname", routeToMap)
		Eventually(session, "10s").Should(Exit(0))

		startTime := time.Now().Unix()
		lastFailure := time.Now().Unix()
		succeeded := false
		postFailed := false
		for j := 0; j < 60; j++ {
			url := fmt.Sprintf("https://%s.%s/", routeToMap, domain)
			resp, err := r.Client.Get(url)
			if err != nil {
				continue
			}

			if resp.StatusCode != http.StatusOK {
				lastFailure = time.Now().Unix()
				if succeeded && !postFailed {
					postFailed = true
					body, err := ioutil.ReadAll(resp.Body)
					if err != nil {
						body = []byte("Could not read body!!!")
					}

					fmt.Fprintln(GinkgoWriter,
						"Got post-success error at", j,
						"seconds on route:", routeToMap,
						"status code:", resp.StatusCode,
						"response body:", string(body),
					)
				}
			} else { // got a 200
				if !succeeded {
					fmt.Fprintln(GinkgoWriter, "First success after", j, "seconds on route:", routeToMap)
					succeeded = true
					postFailed = false
				} else if postFailed {
					fmt.Fprintln(GinkgoWriter,
						"Got post-error success at", j,
						"seconds on route:", routeToMap,
					)
					postFailed = false // ignore transient errors
				}
			}
			time.Sleep(1 * time.Second)
		}

		if !succeeded {
			fmt.Fprintln(GinkgoWriter, routeToMap, "never became healthy this is a problem/failure")
			r.addFailure()
			return
		}

		if postFailed {
			fmt.Fprintln(GinkgoWriter, routeToMap, "became healthy but then became unhealthy and did not recover problem/failure")
			r.addPostFailure()
		}

		r.addResult(float64(lastFailure - startTime))
	}()
}

func cfWithRetry(args ...string) *gexec.Session {
	for i := 0; i < 3; i++ {
		session := cf.Cf(args...)
		session.Wait(30 * time.Second)
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

func (r *RouteMapper) GetPostFailures() int {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	return r.postFailures
}

func (r *RouteMapper) addPostFailure() {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	r.postFailures = r.postFailures + 1
}
