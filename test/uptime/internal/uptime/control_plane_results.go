package uptime

import "fmt"

type ControlPlaneResults struct {
	ControlPlaneSLODataPlaneAvailabilityPercentage float64

	collectorResults []*DataPlaneResults

	dataReady         bool
	sliPassCount      int
	sliFailCount      int
	successPercentage float64
}

func (c *ControlPlaneResults) AddResult(result *DataPlaneResults) {
	c.collectorResults = append(c.collectorResults, result)
	c.dataReady = false
}

func (c *ControlPlaneResults) PrintResults() {
	c.prepareResults()

	fmt.Println("Control Plane Uptime SLI:")
	fmt.Printf("\tPass: %d\n", c.sliPassCount)
	fmt.Printf("\tFail: %d\n", c.sliFailCount)
	fmt.Printf("\tSuccess Percentage: %.2f%%\n", c.successPercentage*100)
	fmt.Println("Request Samples:")
	for i, result := range c.collectorResults {
		fmt.Printf("\tSample %d: Total Samples: %d\tErrors: %d\t Non-200 Codes: %d\t Exceeded Request Latency: %d\n",
			i,
			c.sliPassCount+c.sliFailCount,
			len(result.Errors),
			result.NumberOfUnexpectedResponseCodes,
			result.NumberOfExceededSLORequestLatencies,
		)
	}
}

func (c *ControlPlaneResults) SuccessPercentage() float64 {
	c.prepareResults()

	return c.successPercentage
}

func (c *ControlPlaneResults) prepareResults() {
	if !c.dataReady {
		for _, result := range c.collectorResults {
			if result.SuccessPercentage() > c.ControlPlaneSLODataPlaneAvailabilityPercentage {
				c.sliPassCount++
			} else {
				c.sliFailCount++
			}
		}

		c.successPercentage = float64(c.sliPassCount) / float64((c.sliPassCount + c.sliFailCount))
		c.dataReady = true
	}
}
