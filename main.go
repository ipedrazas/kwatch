package main

import (
	"flag"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

var (
	namespace      = "default"
	watchURL       = "/apis/infra.sohohouse.com/v1/namespaces/" + namespace + "/postgreses?watch=true"
	apiserver      = "http://127.0.0.1:8001"
	token          = ""
	skipVerify     = false
	jobsURL        = "/apis/batch/v1/namespaces/" + namespace + "/jobs"
	chartsLocation = "/charts"
	sa             = false
	debug          = true
)

var processorLock = &sync.Mutex{}
var authConf = AuthConfig{}

func main() {

	flag.StringVar(&watchURL, "url", watchURL, "Watch url, it has to be a valid APISERVER url.")
	flag.StringVar(&apiserver, "apiserver", apiserver, "Apiserver endpoint")
	flag.StringVar(&token, "token", token, "Token to auth against the apiserver.")
	flag.BoolVar(&skipVerify, "skipVerify", skipVerify, "Skip TLS verification for self signed certs.")
	flag.BoolVar(&sa, "sa", sa, "Use the serviceaccount.")
	flag.BoolVar(&debug, "debug", debug, "Debug.")
	flag.StringVar(&chartsLocation, "chartsLocation", chartsLocation, "Base directory where the charts are located.")
	flag.Parse()

	if sa {
		if _, err := os.Stat("/var/run/secrets/kubernetes.io/serviceaccount/token"); err == nil {
			bToken, err := ioutil.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/token")
			if err == nil {
				token = string(bToken)
			}
		}
	}

	authConf.APIServer = apiserver
	authConf.Token = token
	authConf.Watch = apiserver + watchURL

	headers := make(map[string]string)
	if len(token) > 0 {
		headers["Authorization"] = "Bearer " + token
	}
	authConf.Headers = headers
	if debug {
		log.Println(authConf)
	}

	go func() {
		log.Println(http.ListenAndServe("127.0.0.1:6060", nil))
	}()

	doneChan := make(chan struct{})
	var wg sync.WaitGroup

	log.Println("Watching for Events.")
	wg.Add(1)
	watchEvents(doneChan, &wg)
	wg.Add(1)

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	for {
		select {
		case <-signalChan:
			log.Printf("Shutdown signal received, exiting...")
			close(doneChan)
			os.Exit(0)
		}
	}

}
