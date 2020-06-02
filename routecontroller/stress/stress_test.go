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
	Time        int32     `json:"time"`
	AddTimes    []float64 `json:"add_times"`
	DeleteTimes []float64 `json:"delete_times"`
}

var _ = Describe("Stress Tests", func() {
	var (
		numberOfRoutes        = 1000
		numSamples            = 5
		allowableDeltaPercent = 10

		results = Results{
			Time:        int32(time.Now().Unix()),
			AddTimes:    []float64{},
			DeleteTimes: []float64{},
		}
	)

	It(fmt.Sprintf("does not get more than %d%% worse", allowableDeltaPercent), func() {
		for i := 0; i < numSamples; i++ {
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
			fmt.Sprintf("add %d routes", numberOfRoutes))
		compareAverages(previousResults.DeleteTimes,
			results.DeleteTimes,
			allowableDeltaPercent,
			fmt.Sprintf("delete %d routes", numberOfRoutes))

		writeResults(results)
	})
})

func stressRouteController(numberOfRoutes int, results Results) Results {
	setupRoutes(numberOfRoutes)

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

	Expect(addTime.Seconds()).Should(BeNumerically("<", 90), "Should handle 1000 added routes in under 90 second")

	results.AddTimes = append(results.AddTimes, addTime.Seconds())

	deleteTime := timer(func() {
		session, err := kubectl.Run("delete", "routes", "--all", "--wait=false")
		Expect(err).NotTo(HaveOccurred())
		Eventually(session).Should(gexec.Exit(0))

		Eventually(func() int {
			return kubectl.GetNumberOf("routes")
		}, 30*time.Minute, 500*time.Millisecond).Should(Equal(0))

		Eventually(func() int {
			return kubectl.GetNumberOf("virtualservices")
		}, 30*time.Minute, 500*time.Millisecond).Should(Equal(0))
	})

	Expect(deleteTime.Seconds()).Should(BeNumerically("<", 90), "Should handle 1000 removed routes in under 90 seconds")
	results.DeleteTimes = append(results.DeleteTimes, addTime.Seconds())

	deleteRoutecontroller()
	return results
}

func setupRoutes(numberOfRoutes int) {
	routes := buildRoutes(numberOfRoutes)
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

	fmt.Fprintf(GinkgoWriter, "It took %f seconds on average to %s.\n", curmean, logStr)

	change := percentageChange(prevmean, curmean)
	Expect(change).To(BeNumerically("<", allowableDeltaPercent))
}

func percentageChange(old, new float64) (delta float64) {
	diff := new - old
	delta = (diff / old) * 100
	return delta
}
