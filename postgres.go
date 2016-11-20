package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"log"
	"math/rand"
	"net/http"
	"os/exec"
	"strconv"
	"strings"
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
		return processRDS(&rds.Object)
	case rds.Type == "DELETED":
		return deleteRDS(&rds.Object)
	case rds.Type == "MODIFIED":
		return modifyRDS(&rds.Object)
	}
	return nil
}

func watchRdsEvents(done chan struct{}, wg *sync.WaitGroup) {
	events, watchErrs := monitorRdsEvents()
	go func() {
		for {
			select {
			case event := <-events:
				log.Println("We have an event")
				log.Println(event)
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

func processRDS(rds *RdsDB) error {
	wg := new(sync.WaitGroup)
	wg.Add(1)
	runHelm(rds, wg)
	wg.Wait()

	return nil
}

func deleteRDS(rds *RdsDB) error {

	var buffer bytes.Buffer
	buffer.WriteString("aws rds create-db-instance --db-instance-identifier ")
	buffer.WriteString(rds.Spec.Name)
	fmt.Println(buffer.String())

	return nil
}

func modifyRDS(rds *RdsDB) error {

	var buffer bytes.Buffer
	buffer.WriteString("aws rds create-db-instance --db-instance-identifier ")
	buffer.WriteString(rds.Spec.Name)
	fmt.Println(buffer.String())
	return nil
}

// Investigate how to interact with tiller directly
func runHelm(rds *RdsDB, wg *sync.WaitGroup) {

	if rds.Spec.Status == "" {
		cmd := getHelmCommand(rds)
		if debug {
			log.Println("getHelmCommand:")
			log.Println(cmd)
		}
		out, err := exec.Command("/bin/helm", cmd...).Output()
		if err != nil {
			fmt.Println("error occured")
			fmt.Println(err)
		}
		fmt.Printf("%s", out)
		wg.Done()
	}
}

func getHelmCommand(rds *RdsDB) []string {
	cmd := make([]string, 6)
	if strings.ToLower(rds.Spec.Provider) == "aws" {
		cmd[0] = "upgrade"
		cmd[1] = "--install"
		cmd[2] = "postgres-" + rds.Metadata["name"]
		cmd[3] = "/charts/alpha/postgres-rds"
		cmd[4] = "--set"
		cmd[5] = "Rds.Name=" + rds.Spec.Name + ",Rds.User=" + rds.Spec.User
	}
	if strings.ToLower(rds.Spec.Provider) == "kubernetes" {
		cmd[0] = "upgrade"
		cmd[1] = "--install"
		cmd[2] = "postgres-" + rds.Metadata["name"]
		cmd[3] = "/charts/stable/postgresql"
		cmd[4] = "--set"
		cmd[5] = "postgresDatabase=" + rds.Spec.Name + ",postgresUser=" + rds.Spec.User
	}
	return cmd
}

func populateTemplate(templatePath string, values Helm) []byte {
	var doc bytes.Buffer
	t, _ := template.ParseFiles(templatePath)
	t.Execute(&doc, values)
	return doc.Bytes()
}

func createRDSJob(rds *RdsDB) {
	log.Println(rds)
	authConf.Headers["Content-Type"] = "application/yaml"
	// we want to process the job only once
	if rds.Spec.Status == "" {
		ret, err := updatePostgresObject(rds, "provision")
		if err != nil {
			log.Println("Error while updating postgres object " + rds.Metadata["name"])
			log.Println(err)
		}
		if debug {
			log.Println(ret)
		}
		job := Job{
			ID: strconv.Itoa(random(1, 10000)),
		}
		values := Values{
			Rds: *rds,
			Job: job,
		}
		helm := Helm{
			Values: values,
		}
		log.Println("Helm object:")
		log.Println(helm)
		dataDB := populateTemplate(chartsLocation+"/alpha/postgres-rds/templates/rds-db-job.yaml", helm)
		log.Println(string(dataDB))
		authConf.Headers["Content-Type"] = "application/yaml"
		respDB, err := doRequest("POST", apiserver+jobsURL, bytes.NewBuffer(dataDB), authConf.SkipVerify, authConf.Headers)
		if err != nil {
			log.Println("Error while creating rds-db-job")
			log.Println(err)
		}
		fmt.Println("response Body:", string(respDB))

		// We create a job to create a service because by the time we create this job the rds has not
		// assigned an endpoint to the RDS, so, we create a job that checks if the endpoint is ready (readinessProbe)
		// and creates the service once AWS assigns the endpoint
		dataSvc := populateTemplate(chartsLocation+"/alpha/postgres-rds/templates/rds-svc-job.yaml", helm)
		log.Println(string(dataSvc))
		authConf.Headers["Content-Type"] = "application/yaml"
		respSvc, err := doRequest("POST", apiserver+jobsURL, bytes.NewBuffer(dataSvc), authConf.SkipVerify, authConf.Headers)
		if err != nil {
			log.Println("Error while creating rds-svc-job")
			log.Println(err)
		}
		fmt.Println("response Body:", string(respSvc))
	} else {
		log.Println("Status is set - Ignore")
	}
}

func random(min, max int) int {
	rand.Seed(time.Now().UTC().UnixNano())
	return rand.Intn(max-min) + min
}
