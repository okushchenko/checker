package common

import "time"

type Response struct {
	IsUp    bool
	Latency time.Duration
	Time    time.Time
}

type Status struct {
	Uptime            float64 `json:"uptime"`            // Percents
	Mean              float64 `json:"mean"`              // Seconds
	StandardDeviation float64 `json:"standardDeviation"` // Seconds
	Percentile90      float64 `json:"percentile90"`      // Seconds
	Percentile95      float64 `json:"percentile95"`      // Seconds
	Percentile99      float64 `json:"percentile99"`      // Seconds
}
