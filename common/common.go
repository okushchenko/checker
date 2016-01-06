package common

import "time"

type Response struct {
	Status  bool
	Latency time.Duration
	Time    time.Time
}
