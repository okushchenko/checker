package process

import (
	"fmt"
	"time"

	"github.com/montanaflynn/stats"
)

var err error

type Status struct {
	Uptime            float64 `json:"uptime"`            // Percents
	Mean              float64 `json:"mean"`              // Seconds
	StandardDeviation float64 `json:"standardDeviation"` // Seconds
	Percentile90      float64 `json:"percentile90"`      // Seconds
	Percentile95      float64 `json:"percentile95"`      // Seconds
	Percentile99      float64 `json:"percentile99"`      // Seconds
}

func Compute(uptime map[time.Time]bool, latency map[time.Time]float64) (Status, error) {
	var s Status
	// Map to list translation
	latencyList := make([]float64, 0, len(latency))
	for _, value := range latency {
		latencyList = append(latencyList, value)
	}
	// Calculating statistics
	s.Mean, err = stats.Mean(latencyList)
	if err != nil {
		return s, fmt.Errorf("Failed to calculate mean: %s", err.Error())
	}
	s.StandardDeviation, err = stats.StandardDeviation(latencyList)
	if err != nil {
		return s, fmt.Errorf("Failed to calculate standard deviation: %s", err.Error())
	}
	s.Percentile90, err = stats.Percentile(latencyList, 90.0)
	if err != nil {
		return s, fmt.Errorf("Failed to calculate 90th percentile: %s", err.Error())
	}
	s.Percentile90, err = stats.Percentile(latencyList, 95.0)
	if err != nil {
		return s, fmt.Errorf("Failed to calculate 95th percentile: %s", err.Error())
	}
	s.Percentile99, err = stats.Percentile(latencyList, 99.0)
	if err != nil {
		return s, fmt.Errorf("Failed to calculate 99th percentile: %s", err.Error())
	}
	var uptimeUp int
	for _, i := range uptime {
		if i {
			uptimeUp += 1
		}
	}
	s.Uptime = float64(uptimeUp) * 100 / float64(len(uptime))
	return s, nil
}
