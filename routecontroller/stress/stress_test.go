package stress_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
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
	Add100P95  []float64 `json:"individually_add_100_p95s"`
	Del100P95  []float64 `json:"individually_delete_100_p95s"`
	Del1000P95 []float64 `json:"bulk_delete_1000_p95s"`
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

		fmt.Printf("Results: %+v", results)

		for _, res := range results.Add1000P95 {
			Expect(res).To(BeNumerically("<", 0.5))
		}

		for _, res := range results.Add100P95 {
			Expect(res).To(BeNumerically("<", 0.5))
		}

		for _, res := range results.Del100P95 {
			Expect(res).To(BeNumerically("<", 0.5))
		}

		for _, res := range results.Del1000P95 {
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

	add1000hist := getReconcileTime().Metric[0].Histogram

	p95count := float32(add1000hist.GetSampleCount()) * float32(0.95)
	for _, bucket := range add1000hist.Bucket {
		if float32(bucket.GetCumulativeCount()) > p95count {
			results.Add1000P95 = append(results.Add1000P95, bucket.GetUpperBound())
			break
		}
	}

	fmt.Println("Adding 100 routes one at a time")
	for i := 0; i < 100; i++ {
		route := buildSingleRoute(i, "the100")
		session, err := kubectl.RunWithStdin(route, "apply", "-f", "-")
		Expect(err).NotTo(HaveOccurred())
		Eventually(session).Should(gexec.Exit(0))

		expectedNum := numberOfRoutes + i + 1
		Expect(kubectl.GetNumberOf("routes")).To(Equal(expectedNum))
		Eventually(func() int { return kubectl.GetNumberOf("virtualservices") }, 30*time.Minute, 500*time.Millisecond).Should(Equal(expectedNum))
	}

	add100hist := getReconcileTime().Metric[0].Histogram
	results.Add100P95 = append(results.Add100P95, findBucket(add100hist, add1000hist))

	currentNumberOfRoutes := kubectl.GetNumberOf("routes")
	fmt.Printf("Deleting 100 routes one at a time from the current %d routes\n", currentNumberOfRoutes)
	for i := 0; i < 100; i++ {
		route := buildSingleRoute(i, "the100")
		session, err := kubectl.RunWithStdin(route, "delete", "-f", "-")
		Expect(err).NotTo(HaveOccurred())
		Eventually(session).Should(gexec.Exit(0))

		expectedNum := currentNumberOfRoutes - i - 1
		Expect(kubectl.GetNumberOf("routes")).To(Equal(expectedNum))
		Eventually(func() int { return kubectl.GetNumberOf("virtualservices") }, 30*time.Minute, 500*time.Millisecond).Should(Equal(expectedNum))
	}

	del100hist := getReconcileTime().Metric[0].Histogram
	results.Del100P95 = append(results.Del100P95, findBucket(del100hist, add100hist))

	fmt.Printf("Deleting %d routes all at once\n", numberOfRoutes)
	session, err = kubectl.Run("delete", "routes", "-l", "tag=initial", "--wait=false")
	Expect(err).NotTo(HaveOccurred())
	Eventually(session).Should(gexec.Exit(0))

	Eventually(func() int {
		return kubectl.GetNumberOf("routes")
	}, 30*time.Minute, 500*time.Millisecond).Should(Equal(0))

	Eventually(func() int {
		return kubectl.GetNumberOf("virtualservices")
	}, 30*time.Minute, 500*time.Millisecond).Should(Equal(0))

	del1000hist := getReconcileTime().Metric[0].Histogram
	results.Del1000P95 = append(results.Del1000P95, findBucket(del1000hist, del100hist))

	fmt.Println("Stress test complete, cleaning up...")
	deleteRoutecontroller()
	return results
}

func findBucket(current, previous *dto.Histogram) float64 {
	p95count := float32(current.GetSampleCount()-previous.GetSampleCount()) * float32(0.95)
	for i, bucket := range current.Bucket {
		if float32(bucket.GetCumulativeCount()-previous.Bucket[i].GetCumulativeCount()) > p95count {
			return bucket.GetUpperBound()
		}
	}
	return math.Inf(1)
}

func getReconcileTime() *dto.MetricFamily {
	resp, err := http.Get("http://localhost:30080/metrics")
	Expect(err).NotTo(HaveOccurred())
	metricsChan := make(chan *dto.MetricFamily, 1024)
	err = prom2json.ParseResponse(resp, metricsChan)
	Expect(err).NotTo(HaveOccurred())

	for metric := range metricsChan {
		if metric.GetName() == "controller_runtime_reconcile_time_seconds" {
			return metric
		}
	}
	return nil
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

