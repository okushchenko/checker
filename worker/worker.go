package worker

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/alexgear/checker/common"
	"github.com/alexgear/checker/config"
	"github.com/alexgear/checker/network"
)

var err error

func producer(ief string, c chan common.Response) {
	start := time.Now()
	status := network.Ping(ief)
	latency := time.Since(start)
	r := common.Response{Status: status, Latency: latency, Time: start}
	c <- r
}

func consumer(ief string, c chan common.Response) {
	for {
		r := <-c
		err = send(ief, r)
		if err != nil {
			log.Println("Failed to send payload:", err.Error())
		}
	}
}

func send(ief string, r common.Response) error {
	client := http.Client{Timeout: 5 * time.Second}
	u, err := url.Parse(config.C.Server)
	if err != nil {
		return fmt.Errorf("Failed to parse url: %s", err.Error())
	}
	u.Path = fmt.Sprintf("/v1/%s", ief)
	form := url.Values{}
	form.Set("status", fmt.Sprint(r.Status))
	form.Set("latency", r.Latency.String())
	form.Set("time", r.Time.Format(time.RFC3339Nano))
	resp, err := client.PostForm(fmt.Sprint(u), form)
	if err != nil {
		return fmt.Errorf("Failed to send payload: %s", err.Error())
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("Payload sent, but got %d error", resp.StatusCode)
	}
	return nil
}

func InitWorker() {
	cWifi := make(chan common.Response)
	cLan := make(chan common.Response)
	go consumer("wifi", cWifi)
	go consumer("lan", cLan)
	ticker := time.NewTicker(200 * time.Millisecond)
	go func() {
		for _ = range ticker.C {
			go producer("wifi", cWifi)
			go producer("lan", cLan)
		}
	}()
}
