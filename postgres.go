package main

import (
	"bytes"
	"fmt"
	"log"
	"math/rand"
	"os/exec"
	"strings"
	"sync"
	"time"
)

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

func random(min, max int) int {
	rand.Seed(time.Now().UTC().UnixNano())
	return rand.Intn(max-min) + min
}
