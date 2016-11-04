package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

func processRDSEvent(rds RdsEvent) error {
	processorLock.Lock()
	defer processorLock.Unlock()
	log.Println(rds.Type)
	log.Println(rds)
	switch {
	case rds.Type == "ADDED":
		return processRDS(rds.Object)
	case rds.Type == "DELETED":
		return deleteRDS(rds.Object)
	}
	return nil
}

func watchRdsEvents(done chan struct{}, wg *sync.WaitGroup) {
	events, watchErrs := monitorRdsEvents()
	go func() {
		for {
			select {
			case event := <-events:
				err := processRDSEvent(event)
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

func monitorRdsEvents() (<-chan RdsEvent, <-chan error) {
	events := make(chan RdsEvent)
	errc := make(chan error, 1)
	go func() {
		for {

			tr := &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			}
			client := &http.Client{Transport: tr}
			req, err := http.NewRequest("GET", watchUrl, nil)
			req.Header.Add("Authorization", "Basic "+basicAuth(apiserverUser, apiserverPassword))

			resp, err := client.Do(req)
			if err != nil {
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
				var event RdsEvent
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

func processRDS(rds RdsDB) error {

	var buffer bytes.Buffer

	buffer.WriteString("aws rds create-db-instance --db-instance-identifier ")
	buffer.WriteString(rds.Spec.Name)
	buffer.WriteString(" --allocated-storage ")
	buffer.WriteString(rds.Spec.Storage)
	buffer.WriteString(" --db-instance-class  ")
	buffer.WriteString(rds.Spec.InstanceClass)
	buffer.WriteString(" --engine  ")
	buffer.WriteString(rds.Spec.Engine)
	buffer.WriteString(" --master-username ")
	buffer.WriteString(rds.Spec.Username)
	buffer.WriteString(" --master-user-password  ")
	buffer.WriteString(rds.Spec.Password)
	fmt.Println(buffer.String())
	return nil

}

func deleteRDS(rds RdsDB) error {

	var buffer bytes.Buffer

	buffer.WriteString("aws rds create-db-instance --db-instance-identifier ")
	buffer.WriteString(rds.Spec.Name)
	buffer.WriteString(" --allocated-storage ")
	buffer.WriteString(rds.Spec.Storage)
	buffer.WriteString(" --db-instance-class  ")
	buffer.WriteString(rds.Spec.InstanceClass)
	buffer.WriteString(" --engine  ")
	buffer.WriteString(rds.Spec.Engine)
	buffer.WriteString(" --master-username ")
	buffer.WriteString(rds.Spec.Username)
	buffer.WriteString(" --master-user-password  ")
	buffer.WriteString(rds.Spec.Password)
	fmt.Println(buffer.String())

	return nil

}
