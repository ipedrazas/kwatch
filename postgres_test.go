package main

import (
	"fmt"
	"strconv"
	"testing"
)

func TestPopulateTemplate(t *testing.T) {

	job := Job{
		ID: strconv.Itoa(random(1, 10000)),
	}
	rdsSpec := RdsSpec{
		Name:     "short name",
		Provider: "aws",
		User:     "sqladmin",
		Service:  "postgres-svc",
	}
	meta := make(map[string]string)
	meta["name"] = "myname"
	rds := RdsDB{
		Metadata: meta,
		Spec:     rdsSpec,
	}
	values := Values{
		Rds: rds,
		Job: job,
	}

	helm := Helm{
		Values: values,
	}

	chartsLocation := "template-test.yaml"

	result := populateTemplate(chartsLocation, helm)

	fmt.Println(string(result))
}

func TestCreateRDSJob(t *testing.T) {
	spec := &RdsSpec{
		Name: "my name",
	}

	if spec.Status == "" {
		fmt.Println("Status is empty")
	} else {
		fmt.Println(spec.Name)
	}
}
