package stress_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"time"

	"github.com/montanaflynn/stats"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
)

type Results struct {
	Time           int32     `json:"time"`
	AddTimes       []float64 `json:"add_times"`
	Add100Means    []float64 `json:"add_100_means"`
	Add100P95s     []float64 `json:"add_100_p95s"`
	Delete100Means []float64 `json:"delete_100_means"`
	Delete100P95s  []float64 `json:"delete_100_p95s"`
	DeleteTimes    []float64 `json:"delete_times"`
}

var _ = Describe("Stress Tests", func() {
	var (
		numberOfRoutes        = 1000
		numSamples            = 3
		allowableDeltaPercent = 200

		results = Results{
			Time:        int32(time.Now().Unix()),
			AddTimes:    []float64{},
			DeleteTimes: []float64{},
		}
	)

	It(fmt.Sprintf("does not get more than %d%% worse", allowableDeltaPercent), func() {
		fmt.Printf("Stress test starting...\n")
		for i := 0; i < numSamples; i++ {
			fmt.Printf("Performing stress test %d of %d\n", i, numSamples)
			results = stressRouteController(numberOfRoutes, results)
		}

		previousResults := Results{}
		bytes, err := ioutil.ReadFile(resultsPath)
		if err != nil {
			writeResults(results)
			Fail("You must have older results to compare to. If running locally, please run again. If running remote, check your concourse configuration")
		}

		err = json.Unmarshal(bytes, &previousResults)
		Expect(err).NotTo(HaveOccurred())

		// make comparison to previous run
		compareAverages(
			previousResults.AddTimes,
			results.AddTimes,
			allowableDeltaPercent,
			fmt.Sprintf("add %d routes in bulk", numberOfRoutes))
		compareAverages(
			previousResults.Add100Means,
			results.Add100Means,
			allowableDeltaPercent,
			"add 100 routes one at a time (mean of attempts)")
		compareAverages(
			previousResults.Add100P95s,
			results.Add100P95s,
			allowableDeltaPercent,
			"add 100 routes one at a time (95th percentile of attempts)")
		compareAverages(
			previousResults.Delete100Means,
			results.Delete100Means,
			allowableDeltaPercent,
			"delete 100 routes one at a time (mean of attempts)")
		compareAverages(
			previousResults.Delete100P95s,
			results.Delete100P95s,
			allowableDeltaPercent,
			"delete 100 routes one at a time (95th percentile of attempts)")
		compareAverages(previousResults.DeleteTimes,
			results.DeleteTimes,
			allowableDeltaPercent,
			fmt.Sprintf("delete %d routes in bulk", numberOfRoutes))

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

	addTime := timer(func() {
		Eventually(func() int { return kubectl.GetNumberOf("virtualservices") }, 30*time.Minute, 500*time.Millisecond).Should(Equal(numberOfRoutes))
	})

	Expect(addTime.Seconds()).Should(BeNumerically("<", 120), "Should handle 1000 added routes in under 120 second")
	results.AddTimes = append(results.AddTimes, addTime.Seconds())

	fmt.Println("Adding 100 routes one at a time")
	times := []float64{}
	for i := 0; i < 100; i++ {
		route := buildSingleRoute(i, "the100")
		t := timer(func() {
			session, err := kubectl.RunWithStdin(route, "apply", "-f", "-")
			Expect(err).NotTo(HaveOccurred())
			Eventually(session).Should(gexec.Exit(0))

			expectedNum := numberOfRoutes + i + 1
			Expect(kubectl.GetNumberOf("routes")).To(Equal(expectedNum))
			Eventually(func() int { return kubectl.GetNumberOf("virtualservices") }, 30*time.Minute, 500*time.Millisecond).Should(Equal(expectedNum))
		})

		times = append(times, t.Seconds())
	}
	mean, err := stats.Mean(times)
	Expect(err).NotTo(HaveOccurred())
	p95, err := stats.Percentile(times, 95)
	Expect(err).NotTo(HaveOccurred())
	results.Add100Means = append(results.Add100Means, mean)
	results.Add100P95s = append(results.Add100P95s, p95)
	currentNumberOfRoutes := kubectl.GetNumberOf("routes")

	fmt.Printf("Deleting 100 routes one at a time from the current %d routes\n", currentNumberOfRoutes)
	times = []float64{}
	for i := 0; i < 100; i++ {
		route := buildSingleRoute(i, "the100")
		t := timer(func() {
			session, err := kubectl.RunWithStdin(route, "delete", "-f", "-")
			Expect(err).NotTo(HaveOccurred())
			Eventually(session).Should(gexec.Exit(0))

			expectedNum := currentNumberOfRoutes - i - 1
			Expect(kubectl.GetNumberOf("routes")).To(Equal(expectedNum))
			Eventually(func() int { return kubectl.GetNumberOf("virtualservices") }, 30*time.Minute, 500*time.Millisecond).Should(Equal(expectedNum))
		})

		times = append(times, t.Seconds())
	}
	mean, err = stats.Mean(times)
	Expect(err).NotTo(HaveOccurred())
	p95, err = stats.Percentile(times, 95)
	Expect(err).NotTo(HaveOccurred())
	results.Delete100Means = append(results.Delete100Means, mean)
	results.Delete100P95s = append(results.Delete100P95s, p95)

	fmt.Printf("Deleting %d routes all at once\n", numberOfRoutes)
	deleteTime := timer(func() {
		session, err := kubectl.Run("delete", "routes", "-l", "tag=initial", "--wait=false")
		Expect(err).NotTo(HaveOccurred())
		Eventually(session).Should(gexec.Exit(0))

		Eventually(func() int {
			return kubectl.GetNumberOf("routes")
		}, 30*time.Minute, 500*time.Millisecond).Should(Equal(0))

		Eventually(func() int {
			return kubectl.GetNumberOf("virtualservices")
		}, 30*time.Minute, 500*time.Millisecond).Should(Equal(0))
	})

	Expect(deleteTime.Seconds()).Should(BeNumerically("<", 120), "Should handle 1000 removed routes in under 120 seconds")
	results.DeleteTimes = append(results.DeleteTimes, deleteTime.Seconds())

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

func timer(body func()) time.Duration {
	start := time.Now()
	body()
	return time.Since(start)
}

func writeResults(results Results) {
	file, err := json.MarshalIndent(results, "", " ")
	Expect(err).NotTo(HaveOccurred())
	err = ioutil.WriteFile(resultsPath, file, 0644)
	Expect(err).NotTo(HaveOccurred())
}

func compareAverages(previous, current []float64, allowableDeltaPercent int, logStr string) {
	prevmean, err := stats.Mean(previous)
	Expect(err).NotTo(HaveOccurred())

	curmean, err := stats.Mean(current)
	Expect(err).NotTo(HaveOccurred())

	fmt.Printf("It took %f seconds on average to %s.\n", curmean, logStr)

	change := percentageChange(prevmean, curmean)
	Expect(change).To(BeNumerically("<", allowableDeltaPercent), fmt.Sprintf("Took too long to %s", logStr))
}

func percentageChange(old, new float64) (delta float64) {
	diff := new - old
	delta = (diff / old) * 100
	return delta
}
