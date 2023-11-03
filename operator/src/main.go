package main

import (
	"crypto/tls"
	"crypto/x509"
	"log"
	"net/http"
	"os"
	"time"
)

//TODO(alb): Add configurability around the number of failed health checks

//URL of the API server
//const APISERVER = "https://kubernetes.default.svc"
//const APISERVER = "http://127.0.0.1:8080"

// Path to the service account
const SERVACCTPATH = "/var/run/secrets/kubernetes.io/serviceaccount"

// Path to the service account token
const TOKENPATH = SERVACCTPATH + "/token"

// Path to the CA certificate
const CAPATH = SERVACCTPATH + "/ca.crt"

func main() {
	log.Println("Starting Failover Manager...")

	var token string
	tokenBytes, err := os.ReadFile(TOKENPATH)
	if err != nil {
		log.Println(err)
		log.Printf("Failed to read token at path: %s\n", TOKENPATH)
		token = ""
	}
	token = string(tokenBytes)
	log.Println("Read token")

	//TODO(alb): handle SIGTERM and SIGHUP gracefully

	cfg := readConfig()
	log.Println("Read config")

	rootCAs, err := x509.SystemCertPool()
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Read CA certififcate")
	if !cfg.Debug {
		cert, err := os.ReadFile(CAPATH)
		if err != nil {
			log.Fatal(err)
		}
		if ok := rootCAs.AppendCertsFromPEM(cert); !ok {
			log.Fatalf("Failed to append CA to local cert chain")
		}
	}

	tlsConfig := &tls.Config{RootCAs: rootCAs}

	transport := &http.Transport{TLSClientConfig: tlsConfig}

	client := &http.Client{Transport: transport}

	log.Println("Managing workloads...")
	//Run the main loop
	for true {
		failure := false

		managedDeploys := getDeployments(cfg, client, token)
		managedCrons := getCronjobs(cfg, client, token)

		//Get the ID of the local (internal) cluster.
		intID, err := getID(cfg.IntURL)
		if err != nil {
			log.Print(err)
			failure = true
		}
		if !failure {
			//log.Printf("Found internal ID: %s", intID.ID)
		}

		//Get the ID of the external (according to DNS) cluster.
		extID, err := getID(cfg.ExtURL)
		if err != nil {
			log.Print(err)
			failure = true
		}
		if !failure {
			//log.Printf("Found external ID: %s", extID.ID)
		}

		if !failure {
			if intID.ID != extID.ID {
				setInactive(cfg, managedDeploys, managedCrons, client, token)
			}
			if intID.ID == extID.ID {
				setActive(cfg, managedDeploys, managedCrons, client, token)
			}
		}
		time.Sleep(time.Duration(cfg.PollPeriodSeconds) * time.Second)
	}
}
