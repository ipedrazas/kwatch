package main

import (
	"flag"

	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

var (
	watchUrl          = "https://api.bootstrap.staging.houseseven.com/apis/infra.k8s.uk/v1/namespaces/default/rdses?watch=true"
	apiserverUser     = "admin"
	apiserverPassword = ""
)

var processorLock = &sync.Mutex{}

func main() {
	flag.StringVar(&watchUrl, "url", watchUrl, "Watch url, it has to be a valid APISERVER url.")
	flag.StringVar(&apiserverUser, "user", apiserverUser, "User for the apiserver using basic auth ")
	flag.StringVar(&apiserverPassword, "password", apiserverPassword, "Password for the apiserver using basic auth.")
	flag.Parse()

	go func() {
		log.Println(http.ListenAndServe("127.0.0.1:6060", nil))
	}()

	doneChan := make(chan struct{})
	var wg sync.WaitGroup

	log.Println("Watching for Events.")
	wg.Add(1)
	watchRdsEvents(doneChan, &wg)
	wg.Add(1)

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	for {
		select {
		case <-signalChan:
			log.Printf("Shutdown signal received, exiting...")
			close(doneChan)
			wg.Wait()
			os.Exit(0)
		}
	}

}
