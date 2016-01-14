package process

import (
	"fmt"
	"time"

	"github.com/alexgear/checker/common"
	"github.com/montanaflynn/stats"
)

var err error

func Compute(status map[time.Time]common.Status) (common.Status, error) {
	var mean []float64
	var percentile90 []float64
	var percentile99 []float64
	var uptime []float64
	for _, s := range status {
		mean = append(mean, s.Mean)
		percentile90 = append(percentile90, s.Percentile90)
		percentile99 = append(percentile99, s.Percentile99)
		uptime = append(uptime, s.Uptime)
	}
	// Calculating statistics
	var s common.Status
	s.Mean, err = stats.Mean(mean)
	if err != nil {
		return s, fmt.Errorf("Failed to calculate mean: %s", err.Error())
	}
	s.Percentile90, err = stats.Mean(percentile90)
	if err != nil {
		return s, fmt.Errorf("Failed to calculate mean: %s", err.Error())
	}
	s.Percentile99, err = stats.Mean(percentile99)
	if err != nil {
		return s, fmt.Errorf("Failed to calculate mean: %s", err.Error())
	}
	s.Uptime, err = stats.Mean(uptime)
	if err != nil {
		return s, fmt.Errorf("Failed to calculate mean: %s", err.Error())
	}
	return s, nil
}
