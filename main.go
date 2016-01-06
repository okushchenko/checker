package main

import (
	"flag"
	"log"
	"time"

	"github.com/alexgear/checker/api"
	"github.com/alexgear/checker/config"
	"github.com/alexgear/checker/datastore"
	"github.com/alexgear/checker/network"
	"github.com/alexgear/checker/process"
	"github.com/alexgear/checker/worker"
)

var err error

func main() {
	log.Println("Load flags...")
	var agent bool
	var server bool
	flag.BoolVar(&agent, "agent", false, "help")
	flag.BoolVar(&server, "server", false, "help")
	flag.Parse()
	log.Println("Load config...")
	err := config.InitConfig()
	if err != nil {
		log.Fatal(err)
	}
	if server {
		log.Println("Init DB...")
		db, err := datastore.InitDB()
		if err != nil {
			log.Fatal(err)
		}
		defer db.Close()
		log.Println("Dialing...")
		go func() {
			ticker := time.NewTicker(30 * time.Second)
			for _ = range ticker.C {
				for _, ief := range []string{"wifi", "lan"} {
					log.Printf("=============%s============", ief)
					status, latency, err := datastore.Read(ief)
					if err != nil {
						log.Fatal(err.Error())
					}
					err = process.Compute(status, latency)
					if err != nil {
						log.Fatal(err.Error())
					}
					err = process.Plot(ief, latency)
					if err != nil {
						log.Fatal(err.Error())
					}
				}
			}
		}()
		err = api.InitServer()
		if err != nil {
			log.Fatal(err)
		}
	} else if agent {
		log.Println("Init Network...")
		err = network.InitNetwork()
		if err != nil {
			log.Fatal(err)
		}
		log.Println("Dialing...")
		worker.InitWorker()
		for {
			time.Sleep(60 * time.Second)
		}
	} else {
		log.Fatal("Either -agent or -server must be used.")
	}
}
