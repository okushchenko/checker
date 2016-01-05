package main

import (
	"bytes"
	"fmt"
	"log"
	"net"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/boltdb/bolt"
	"github.com/gonum/plot"
	"github.com/gonum/plot/plotter"
	"github.com/gonum/plot/plotutil"
	"github.com/gonum/plot/vg"
	"github.com/montanaflynn/stats"
)

var timeout = 5 * time.Second
var db *bolt.DB
var err error
var cfg config

type config struct {
	SSID     string // ssid of wifi network
	Password string // password of wifi network
	LanGw    string // lan network gateway
	WifiGw   string // wifi network gateway
	LanIef   string // lan interface name
	WifiIef  string // wifi interface name
}

type response struct {
	Status  bool
	Latency time.Duration
}

func producer(ief string, c chan response) {
	for {
		start := time.Now()
		status := ping(ief)
		latency := time.Since(start)
		r := response{Status: status, Latency: latency}
		c <- r
		time.Sleep(100 * time.Millisecond)
	}
}

func consumer(ief string, c chan response) {
	for {
		r := <-c
		err = write(ief, r.Status, r.Latency)
		if err != nil {
			log.Println("Failed to write to db:", err.Error())
		}
	}
}

func ping(ief string) bool {
	var target string
	if ief == "wifi" {
		target = "8.8.4.4:53"
	} else {
		target = "8.8.8.8:53"
	}
	d := net.Dialer{Timeout: timeout}
	conn, err := d.Dial("tcp", target)
	if err != nil {
		log.Printf("Failed to initiate tcp connection: %s\n", err.Error())
		return false
	}
	//log.Println(conn.LocalAddr(), conn.RemoteAddr())
	defer conn.Close()
	return true
}

func initDB() error {
	buckets := []string{"wifi", "lan"}
	subbuckets := []string{"latency", "status"}
	err = db.Update(func(tx *bolt.Tx) error {
		for _, bucket := range buckets {
			b, err := tx.CreateBucketIfNotExists([]byte(bucket))
			if err != nil {
				return fmt.Errorf("create bucket: %s", err.Error())
			}
			for _, subbucket := range subbuckets {
				_, err = b.CreateBucketIfNotExists([]byte(subbucket))
				if err != nil {
					return fmt.Errorf("create subbucket: %s", err.Error())
				}

			}
		}
		return nil
	})
	return err
}

func write(ief string, status bool, latency time.Duration) error {
	t := time.Now()
	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(ief)).Bucket([]byte("latency"))
		err = b.Put([]byte(t.Format(time.RFC3339)), []byte(latency.String()))
		if err != nil {
			return fmt.Errorf("update bucket: %s", err.Error())
		}
		b = tx.Bucket([]byte(ief)).Bucket([]byte("status"))
		err = b.Put([]byte(t.Format(time.RFC3339)), []byte(fmt.Sprint(status)))
		if err != nil {
			return fmt.Errorf("update bucket: %s", err.Error())
		}
		return nil
	})
	return err
}

func read(ief string) ([]bool, []float64, error) {
	min := []byte(time.Now().Add(-1 * time.Hour).String())
	max := []byte(time.Now().Format(time.RFC3339))
	var latency []float64
	var status []bool
	err = db.View(func(tx *bolt.Tx) error {
		c := tx.Bucket([]byte(ief)).Bucket([]byte("latency")).Cursor()
		for k, v := c.Seek(min); k != nil && bytes.Compare(k, max) <= 0; k, v = c.Next() {
			//fmt.Printf("time=%s, latency=%s", k, v)
			v1, err := time.ParseDuration(string(v))
			if err != nil {
				return fmt.Errorf("Failed to parse duration: %s", err.Error())
			}
			latency = append(latency, v1.Seconds())

			c := tx.Bucket([]byte(ief)).Bucket([]byte("status")).Cursor()
			_, v = c.Seek([]byte(k))
			v2, err := strconv.ParseBool(string(v))
			if err != nil {
				return fmt.Errorf("Failed to parse status: %s", err.Error())
			}
			status = append(status, v2)
			//fmt.Printf(", status=%s\n", v)
		}
		return nil
	})
	if err != nil {
		return status, latency, fmt.Errorf("Failed to query db: %s", err.Error())
	}
	return status, latency, nil
}

func process(status []bool, latency []float64) error {
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

func initNetwork() error {
	lines, err := exec.Command("nmcli", "dev", "wifi", "list", "ifname", "wlo1").Output()
	if err != nil {
		return fmt.Errorf("Failed to list connections: %s", err.Error())
	}
	var signalStrength string
	for i, line := range strings.Split(string(lines), "\n") {
		fields := strings.Fields(line)
		if i > 0 && len(fields) > 0 && fields[0] == "*" {
			signalStrength = fields[6]
			log.Printf("Signal strength = %s\n", signalStrength)
			break
		}
	}
	if signalStrength == "" {
		log.Println("Disconnected, trying to reconnect")
		exec.Command("nmcli", "connection", "delete", cfg.SSID).Output()
		_, err := exec.Command("nmcli",
			"dev", "wifi",
			"connect", cfg.SSID,
			"password", cfg.Password,
			"name", cfg.SSID,
			"ifname", cfg.WifiIef).Output()
		if err != nil {
			fmt.Errorf("Failed to reconnect: %s", err.Error())
		}
		exec.Command("sudo", "ip", "route", "add", "8.8.4.4", "via", cfg.WifiGw, "dev", cfg.WifiIef).Output()
	}
	exec.Command("sudo", "ip", "route", "add", "8.8.4.4", "via", cfg.WifiGw, "dev", cfg.WifiIef).Output()
	exec.Command("sudo", "ip", "route", "add", "8.8.8.8", "via", cfg.LanGw, "dev", cfg.LanIef).Output()
	return nil
}

func main() {
	log.Println("Load config...")
	_, err = toml.DecodeFile("./config.toml", &cfg)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Init Network...")
	err = initNetwork()
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Init DB...")
	db, err = bolt.Open("my.db", 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	err = initDB()
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Dialing...")
	cWifi := make(chan response)
	cLan := make(chan response)
	go consumer("wifi", cWifi)
	go consumer("lan", cLan)
	go producer("wifi", cWifi)
	go producer("lan", cLan)
	for {
		log.Println("=============WIFI============")
		status, latency, err := read("wifi")
		if err != nil {
			log.Fatal(err.Error())
		}
		err = process(status, latency)
		if err != nil {
			log.Fatal(err.Error())
		}
		err = Plot("wifi", latency)
		if err != nil {
			log.Fatal(err.Error())
		}
		log.Println("=============LAN=============")
		status, latency, err = read("lan")
		if err != nil {
			log.Fatal(err.Error())
		}
		err = process(status, latency)
		if err != nil {
			log.Fatal(err.Error())
		}
		err = Plot("lan", latency)
		if err != nil {
			log.Fatal(err.Error())
		}
		time.Sleep(60 * time.Second)
	}
}
