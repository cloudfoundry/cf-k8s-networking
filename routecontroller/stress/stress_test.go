package stress_test

import (
	"bytes"
	"fmt"
	"path/filepath"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("Stress Tests", func() {
	var (
		numberOfRoutes = 1000
	)

	BeforeEach(func() {
		routes := buildRoutes(numberOfRoutes)
		session, err := kubectl.RunWithStdin(routes, "apply", "-f", "-")
		Expect(err).NotTo(HaveOccurred())
		Eventually(session).Should(gexec.Exit(0))

		Expect(kubectl.GetNumberOf("routes")).To(Equal(numberOfRoutes))
	})

	AfterEach(func() {
		session, err := kubectl.Run("delete", "deployment", "routecontroller")
		Expect(err).NotTo(HaveOccurred())
		Eventually(session).Should(gexec.Exit(0))

		Eventually(func() int { return kubectl.GetNumberOf("pods") }).Should(Equal(0))
	})

	Measure("routecontroller stress", func(b Benchmarker) {
		// Make sure we're starting from a blank slate
		Expect(kubectl.GetNumberOf("virtualservices")).To(Equal(0))

		yttSession, err := ytt.Run(
			"-f", filepath.Join("..", "..", "config", "routecontroller"),
			"-f", filepath.Join("..", "..", "config", "values.yaml"),
			"-v", "systemNamespace=default",
		)
		Expect(err).NotTo(HaveOccurred())
		Eventually(yttSession).Should(gexec.Exit(0))
		// TODO: why do we need to get Contents() ?
		yttContents := yttSession.Out.Contents()
		yttReader := bytes.NewReader(yttContents)

		session, err := kubectl.RunWithStdin(yttReader, "apply", "-f", "-")
		Expect(err).NotTo(HaveOccurred())
		Eventually(session).Should(gexec.Exit(0))

		session, err = kubectl.Run("rollout", "status", "deployment", "routecontroller")
		Expect(err).NotTo(HaveOccurred())
		Eventually(session).Should(gbytes.Say("successfully rolled out"))
		Eventually(session).Should(gexec.Exit(0))

		b.Time(fmt.Sprintf("Processing %d new routes at once", numberOfRoutes), func() {
			Eventually(func() int { return kubectl.GetNumberOf("virtualservices") }, 30*time.Minute, 500*time.Millisecond).Should(Equal(numberOfRoutes))
		})

		b.Time(fmt.Sprintf("Deleting %d routes at once", numberOfRoutes), func() {
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
	}, 2)
})
