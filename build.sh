#!/bin/bash


echo "Running the tests"
go test -v 

echo "building the app"
CGO_ENABLED=0 GOOS=linux go build -a -tags netgo -ldflags '-w' .

echo "fetching charts"
cp -R /Users/ipedrazas/workspace/sohohouse/charts .

echo "docker build"
docker build -t ipedrazas/postgres-operator:$1 .

rm -rf charts

echo "docker push"
docker push ipedrazas/postgres-operator:$1
