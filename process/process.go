package process

import (
	"log"
	"time"

	"github.com/montanaflynn/stats"
)

func Compute(status map[time.Time]bool, latency map[time.Time]float64) error {
	latencyList := make([]float64, 0, len(latency))
	for _, value := range latency {
		latencyList = append(latencyList, value)
	}
	mean, _ := stats.Mean(latencyList)
	std, _ := stats.StandardDeviation(latencyList)
	log.Println("Mean =", mean*1000, "+-", std*1000, "ms")
	p90, _ := stats.Percentile(latencyList, 90.0)
	log.Println("Percentile 90 =", p90*1000, "ms")
	p99, _ := stats.Percentile(latencyList, 99.0)
	log.Println("Percentile 99 =", p99*1000, "ms")
	var up int
	for _, i := range status {
		if i {
			up += 1
		}
	}
	log.Println("Uptime =", float64(up)*100/float64(len(status)), "%", up, len(status))
	return nil
}
