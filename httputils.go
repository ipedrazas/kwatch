package main

import (
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
)

func basicAuth(username string, password string) string {
	auth := username + ":" + password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}

func doRequest(METHOD string, url string, data io.Reader, skipVerify bool, headers map[string]string) ([]byte, error) {
	client := &http.Client{}
	if skipVerify {
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		client = &http.Client{Transport: tr}
	}

	if debug {
		log.Println("URL: " + url)
	}

	req, err := http.NewRequest(METHOD, url, data)

	for k, v := range headers {
		log.Println("Adding headers " + k + " - " + v)
		req.Header.Set(k, v)
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	if debug {
		log.Println("Response: ")
		log.Println("Status Code: " + strconv.Itoa(resp.StatusCode))
		log.Print(string(body))
	}
	return body, err

}

func updatePostgresObject(rds *RdsDB, status string) ([]byte, error) {
	authConf.Headers["Content-Type"] = "application/merge-patch+json"
	postgresURL := authConf.APIServer + "/apis/infra.sohohouse.com/v1/namespaces/" + namespace + "/postgreses/" + rds.Metadata["name"]

	data := "{\"spec\": {\"status\": \"" + status + "\"}}"
	return doRequest("PATCH", postgresURL, bytes.NewBufferString(data), authConf.SkipVerify, authConf.Headers)

}
