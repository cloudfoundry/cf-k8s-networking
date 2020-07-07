package stress_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"time"

	dto "github.com/prometheus/client_model/go"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
	"github.com/prometheus/prom2json"
)

type Results struct {
	Time       int32     `json:"time"`
	Add1000P95 []float64 `json:"bulk_add_1000_p95s"`
}

var _ = Describe("Stress Tests", func() {
	var (
		numberOfRoutes = 1000
		numSamples     = 3

		results = Results{
			Time:       int32(time.Now().Unix()),
			Add1000P95: []float64{},
		}
	)

	It("does not get worse", func() {
		fmt.Printf("Stress test starting...\n")
		for i := 0; i < numSamples; i++ {
			fmt.Printf("Performing stress test %d of %d\n", i, numSamples)
			results = stressRouteController(numberOfRoutes, results)
		}

		for _, res := range results.Add1000P95 {
			Expect(res).To(BeNumerically("<", 0.5))
		}

		writeResults(results)
	})
})

func stressRouteController(numberOfRoutes int, results Results) Results {
	setupRoutes(numberOfRoutes, "initial")

	Expect(kubectl.GetNumberOf("virtualservices")).To(Equal(0))

	yttSession, err := ytt.Run(
		"-f", filepath.Join("..", "..", "config", "routecontroller"),
		"-f", filepath.Join("..", "..", "config", "values.yaml"),
		"-v", "systemNamespace=default",
	)
	Expect(err).NotTo(HaveOccurred())
	Eventually(yttSession).Should(gexec.Exit(0))
	yttContents := yttSession.Out.Contents()
	yttReader := bytes.NewReader(yttContents)

	fmt.Printf("Adding %d routes all at once\n", numberOfRoutes)
	session, err := kubectl.RunWithStdin(yttReader, "apply", "-f", "-")
	Expect(err).NotTo(HaveOccurred())
	Eventually(session).Should(gexec.Exit(0))

	session, err = kubectl.Run("rollout", "status", "deployment", "routecontroller")
	Expect(err).NotTo(HaveOccurred())
	Eventually(session).Should(gbytes.Say("successfully rolled out"))
	Eventually(session).Should(gexec.Exit(0))

	// Wait for all the virtualservices to be created
	Eventually(func() int { return kubectl.GetNumberOf("virtualservices") }, 30*time.Minute, 500*time.Millisecond).Should(Equal(numberOfRoutes))

	resp, err := http.Get("http://localhost:30080/metrics")
	Expect(err).NotTo(HaveOccurred())
	metricsChan := make(chan *dto.MetricFamily, 1024)
	err = prom2json.ParseResponse(resp, metricsChan)
	Expect(err).NotTo(HaveOccurred())

	var reconcileTime *dto.MetricFamily

	for metric := range metricsChan {
		if metric.GetName() == "controller_runtime_reconcile_time_seconds" {
			reconcileTime = metric
		}
	}

	histogram := reconcileTime.Metric[0].Histogram

	p95count := float32(histogram.GetSampleCount()) * float32(0.95)

	var p95Add float64
	for _, bucket := range histogram.Bucket {
		if float32(bucket.GetCumulativeCount()) > p95count {
			p95Add = bucket.GetUpperBound()
			break
		}
	}

	results.Add1000P95 = append(results.Add1000P95, p95Add)

	// TODO -- leaving these comments to help with the subsequent stories, make sure everything is
	// gone before all the stress stories are complete

	//Expect(gbytes.BufferReader(resp.Body)).To(gbytes.Say("controller_runtime_reconcile"))

	// addTime := timer(func() {
	// 	Eventually(func() int { return kubectl.GetNumberOf("virtualservices") }, 30*time.Minute, 500*time.Millisecond).Should(Equal(numberOfRoutes))
	// })

	// Expect(addTime.Seconds()).Should(BeNumerically("<", 120), "Should handle 1000 added routes in under 120 second")
	// results.AddTimes = append(results.AddTimes, addTime.Seconds())

	// fmt.Println("Adding 100 routes one at a time")
	// times := []float64{}
	// for i := 0; i < 100; i++ {
	// 	route := buildSingleRoute(i, "the100")
	// 	t := timer(func() {
	// 		session, err := kubectl.RunWithStdin(route, "apply", "-f", "-")
	// 		Expect(err).NotTo(HaveOccurred())
	// 		Eventually(session).Should(gexec.Exit(0))

	// 		expectedNum := numberOfRoutes + i + 1
	// 		Expect(kubectl.GetNumberOf("routes")).To(Equal(expectedNum))
	// 		Eventually(func() int { return kubectl.GetNumberOf("virtualservices") }, 30*time.Minute, 500*time.Millisecond).Should(Equal(expectedNum))
	// 	})

	// 	times = append(times, t.Seconds())
	// }
	// mean, err := stats.Mean(times)
	// Expect(err).NotTo(HaveOccurred())
	// p95, err := stats.Percentile(times, 95)
	// Expect(err).NotTo(HaveOccurred())
	// results.Add100Means = append(results.Add100Means, mean)
	// results.Add100P95s = append(results.Add100P95s, p95)
	// currentNumberOfRoutes := kubectl.GetNumberOf("routes")

	// fmt.Printf("Deleting 100 routes one at a time from the current %d routes\n", currentNumberOfRoutes)
	// times = []float64{}
	// for i := 0; i < 100; i++ {
	// 	route := buildSingleRoute(i, "the100")
	// 	t := timer(func() {
	// 		session, err := kubectl.RunWithStdin(route, "delete", "-f", "-")
	// 		Expect(err).NotTo(HaveOccurred())
	// 		Eventually(session).Should(gexec.Exit(0))

	// 		expectedNum := currentNumberOfRoutes - i - 1
	// 		Expect(kubectl.GetNumberOf("routes")).To(Equal(expectedNum))
	// 		Eventually(func() int { return kubectl.GetNumberOf("virtualservices") }, 30*time.Minute, 500*time.Millisecond).Should(Equal(expectedNum))
	// 	})

	// 	times = append(times, t.Seconds())
	// }
	// mean, err = stats.Mean(times)
	// Expect(err).NotTo(HaveOccurred())
	// p95, err = stats.Percentile(times, 95)
	// Expect(err).NotTo(HaveOccurred())
	// results.Delete100Means = append(results.Delete100Means, mean)
	// results.Delete100P95s = append(results.Delete100P95s, p95)

	// fmt.Printf("Deleting %d routes all at once\n", numberOfRoutes)
	// deleteTime := timer(func() {
	session, err = kubectl.Run("delete", "routes", "-l", "tag=initial", "--wait=false")
	Expect(err).NotTo(HaveOccurred())
	Eventually(session).Should(gexec.Exit(0))

	Eventually(func() int {
		return kubectl.GetNumberOf("routes")
	}, 30*time.Minute, 500*time.Millisecond).Should(Equal(0))

	Eventually(func() int {
		return kubectl.GetNumberOf("virtualservices")
	}, 30*time.Minute, 500*time.Millisecond).Should(Equal(0))
	// })

	// Expect(deleteTime.Seconds()).Should(BeNumerically("<", 120), "Should handle 1000 removed routes in under 120 seconds")
	// results.DeleteTimes = append(results.DeleteTimes, deleteTime.Seconds())

	fmt.Println("Stress test complete, cleaning up...")
	deleteRoutecontroller()
	return results
}

func setupRoutes(numberOfRoutes int, tag string) {
	routes := buildRoutes(numberOfRoutes, tag)
	session, err := kubectl.RunWithStdin(routes, "apply", "-f", "-")
	Expect(err).NotTo(HaveOccurred())
	Eventually(session).Should(gexec.Exit(0))

	Expect(kubectl.GetNumberOf("routes")).To(Equal(numberOfRoutes))
}

func deleteRoutecontroller() {
	session, err := kubectl.Run("delete", "deployment", "routecontroller")
	Expect(err).NotTo(HaveOccurred())
	Eventually(session).Should(gexec.Exit(0))

	Eventually(func() int { return kubectl.GetNumberOf("pods") }).Should(Equal(0))
}

func writeResults(results Results) {
	file, err := json.MarshalIndent(results, "", " ")
	Expect(err).NotTo(HaveOccurred())
	err = ioutil.WriteFile(resultsPath, file, 0644)
	Expect(err).NotTo(HaveOccurred())
}

