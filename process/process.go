package process

import (
	"fmt"
	"log"

	"github.com/gonum/plot"
	"github.com/gonum/plot/plotter"
	"github.com/gonum/plot/plotutil"
	"github.com/gonum/plot/vg"
	"github.com/montanaflynn/stats"
)

func Compute(status []bool, latency []float64) error {
	mean, _ := stats.Mean(latency)
	std, _ := stats.StandardDeviation(latency)
	log.Println("Mean =", mean*1000, "+-", std*1000, "ms")
	p90, _ := stats.Percentile(latency, 90.0)
	log.Println("Percentile 90 =", p90*1000, "ms")
	p99, _ := stats.Percentile(latency, 99.0)
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

func Plot(ief string, latency []float64) error {
	p, err := plot.New()
	if err != nil {
		return err
	}

	p.Title.Text = fmt.Sprintf("Latency")
	p.X.Label.Text = "time"
	p.Y.Label.Text = "latency"

	pts := make(plotter.XYs, len(latency))
	for i := range pts {
		pts[i].Y = latency[i]
		pts[i].X = float64(i)
	}

	err = plotutil.AddLinePoints(p, "First", pts)
	if err != nil {
		return err
	}

	// Save the plot to a PNG file.
	err = p.Save(12*vg.Inch, 12*vg.Inch, fmt.Sprintf("./network_%s.png", ief))
	if err != nil {
		return err
	}
	return nil
}
