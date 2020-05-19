package uptime

import (
	"fmt"
	"time"
)

type DataPlaneResults struct {
	SliPassCount                        int
	SliFailCount                        int
	NumberOfUnexpectedResponseCodes     int
	NumberOfExceededSLORequestLatencies int
	Errors                              []error
	RequestLatencies                    []time.Duration
}

func (u *DataPlaneResults) PrintResults() {
	fmt.Println("Data Plane Uptime SLI:")
	fmt.Printf("\tPass: %d\n", u.SliPassCount)
	fmt.Printf("\tFail: %d\n", u.SliFailCount)
	fmt.Printf("\tSuccess Percentage: %.2f%%\n", u.SuccessPercentage()*100)
	fmt.Printf("Number Of Request Errors: %d\n", len(u.Errors))
	fmt.Printf("Number Of Unexpected Response Codes: %d\n", u.NumberOfUnexpectedResponseCodes)
	fmt.Printf("Number Of Exceeded SLO Request Latencies: %d\n", u.NumberOfExceededSLORequestLatencies)
	fmt.Println("Response Time:")
	low, high, avg := u.calculateRequestLatencyStats()
	fmt.Printf("\tLowest: %s\n", low.String())
	fmt.Printf("\tHighest: %s\n", high.String())
	fmt.Printf("\tAverage: %s\n", avg.String())
	if len(u.Errors) > 0 {
		fmt.Println("Errors:")
		for _, err := range u.uniqErrors() {
			fmt.Printf("\t%s\n", err)
		}
	}
}

func (u *DataPlaneResults) SuccessPercentage() float64 {
	return float64(u.SliPassCount) / float64(u.SliFailCount+u.SliPassCount)
}

func (u *DataPlaneResults) RecordError(err error) {
	u.SliFailCount++
	u.Errors = append(u.Errors, err)
}

func (u *DataPlaneResults) Record(hasPassedSLI, hasStatusOK, hasMetRequestLatencySLO bool, requestLatency time.Duration) {
	if hasPassedSLI {
		u.SliPassCount++
	} else {
		u.SliFailCount++
	}

	if !hasStatusOK {
		u.NumberOfUnexpectedResponseCodes++
	}

	if !hasMetRequestLatencySLO {
		u.NumberOfExceededSLORequestLatencies++
	}

	u.RequestLatencies = append(u.RequestLatencies, requestLatency)
}

func (u *DataPlaneResults) uniqErrors() []error {
	errMap := map[string]error{}
	for _, err := range u.Errors {
		errMap[err.Error()] = err
	}

	uniqErrs := []error{}
	for _, v := range errMap {
		uniqErrs = append(uniqErrs, v)
	}

	return uniqErrs
}

func (u *DataPlaneResults) calculateRequestLatencyStats() (time.Duration, time.Duration, time.Duration) {
	var low, high, avg, accumulator time.Duration
	if len(u.RequestLatencies) == 0 {
		return low, high, avg
	}

	low = u.RequestLatencies[0]
	for _, time := range u.RequestLatencies {
		if low > time {
			low = time
		}
		if high < time {
			high = time
		}

		accumulator += time
	}

	avg = time.Duration(int64(accumulator) / int64(len(u.RequestLatencies)))

	return low, high, avg
}
