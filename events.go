package main

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"sync"
	"time"
)

func processEvent(event Event) error {
	processorLock.Lock()
	defer processorLock.Unlock()
	log.Println(event.Type)
	log.Println(event)
	switch {
	case event.Type == "ADDED":
		return processRDS(&event.Object)
	case event.Type == "DELETED":
		return deleteRDS(&event.Object)
	case event.Type == "MODIFIED":
		return modifyRDS(&event.Object)
	}
	return nil
}

func watchEvents(done chan struct{}, wg *sync.WaitGroup) {
	events, watchErrs := monitorEvents()
	go func() {
		for {
			select {
			case event := <-events:
				log.Println("We have an event")
				log.Println(event)
				err := processEvent(event)
				if err != nil {
					log.Println(err)
				}
			case err := <-watchErrs:
				log.Println(err)
			case <-done:
				wg.Done()
				log.Println("Stopped RDS event watcher.")
				return
			}
		}
	}()
}

func monitorEvents() (<-chan Event, <-chan error) {
	events := make(chan Event)
	errc := make(chan error, 1)
	go func() {
		for {

			client := &http.Client{}
			if authConf.SkipVerify {
				tr := &http.Transport{
					TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
				}
				client = &http.Client{Transport: tr}
			}

			req, err := http.NewRequest("GET", authConf.Watch, nil)

			for k, v := range authConf.Headers {
				req.Header.Set(k, v)
			}

			resp, err := client.Do(req)
			if err != nil {
				log.Println(err)
				errc <- err
				time.Sleep(5 * time.Second)
				continue
			}
			if resp.StatusCode != 200 {
				errc <- errors.New("Invalid status code: " + resp.Status)
				time.Sleep(5 * time.Second)
				continue
			}

			decoder := json.NewDecoder(resp.Body)
			for {
				var event Event
				err = decoder.Decode(&event)
				if err != nil {
					errc <- err
					break
				}
				events <- event
			}
		}
	}()

	return events, errc
}
