package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
)

const LABEL = "failovermanager=enabled"

// Get all deployments in the cluster matching a given label.
func getDeployments(
	cfg Config,
	client *http.Client,
	token string,
) []Deployment {

	//Build the url to the specific deployment resource
	url := fmt.Sprintf(
		"%s/apis/apps/v1/deployments?labelSelector=%s",
		cfg.APIServer,
		LABEL)

	req, err := http.NewRequest("GET", url, nil)
	bearer := "Bearer " + token
	req.Header.Add("Authorization", bearer)
	if err != nil {
		log.Print(err)
	}

	//log.Printf("Making request to %s\n", url)
	resp, err := client.Do(req)
	if err != nil {
		log.Print(err)
	}
	defer resp.Body.Close()
	//log.Printf("Received %s", resp.Status)

	var deploymentList DeploymentList

	err = json.NewDecoder(resp.Body).Decode(&deploymentList)
	if err != nil {
		log.Print(err)
	}
	return deploymentList.Items

}

// Get updates about deployments from the list

//func watchDeployments(
//	cfg Config,
//	client *http.Client,
//	token string,
//	resVer string, //resourceVersion for the watch
//) []Deployment {
//}

// Patch a deployment resource to have a given number of replicas
func patchDeployment(
	cfg Config,
	deploy Deployment,
	replicas int,
	client *http.Client,
	token string) {

	url := fmt.Sprintf(
		"%s/apis/apps/v1/namespaces/%s/deployments/%s",
		cfg.APIServer,
		deploy.Metadata.Namespace,
		deploy.Metadata.Name)

	var patch [1]DeployJSONPatch

	patch[0] = DeployJSONPatch{
		Op:    "replace",
		Path:  "/spec/replicas",
		Value: replicas,
	}

	jsonBytes, err := json.Marshal(patch)
	if err != nil {
		log.Printf(
			"Failed to marshal JSON patch for deployment %s: %s",
			deploy.Metadata.Name,
			err,
		)
		return
	}

	jsonBuf := bytes.NewBuffer(jsonBytes)

	req, err := http.NewRequest("PATCH", url, jsonBuf)
	bearer := "Bearer " + token
	req.Header.Add("Authorization", bearer)
	if err != nil {
		log.Printf(
			"Failed to create http PATCH request for deployment %s: %s",
			deploy.Metadata.Name,
			err,
		)
		return
	}
	req.Header.Set("Content-Type", "application/json-patch+json")

	log.Printf("Patching deployment: %s, in namespace: %s\n", deploy.Metadata.Name, deploy.Metadata.Namespace)
	log.Printf("patch: %s\n", jsonBuf)
	resp, err := client.Do(req)
	if err != nil {
		log.Print(err)
	}
	log.Printf("PATCH result status code: %d\n", resp.StatusCode)

}

// Set local cluster to be active. For deployments, set the current replicas
// to match the active replicas annotation. For cronjobs, set the suspend flag
// to match the active suspend annotation.
func setActive(
	cfg Config,
	managedDeploys []Deployment,
	managedCrons []Cronjob,
	client *http.Client,
	token string) {

	//log.Println("Internal and external IDs match. Setting cluster to active.")

	for _, deploy := range managedDeploys {
		replStr := deploy.Metadata.Annotations.ActiveMinReplicas
		activeReplicas, err := strconv.Atoi(replStr)
		if err != nil {
			log.Printf("Failed to read active replicas from deployment \"%s\".", deploy.Metadata.Name)
			continue
		}
		if deploy.Spec.Replicas < activeReplicas {
			patchDeployment(
				cfg,
				deploy,
				activeReplicas,
				client,
				token)
		}

	}

	for _, cron := range managedCrons {
		suspendStr := cron.Metadata.Annotations.ActiveSuspend

		var suspend bool

		switch suspendStr {
		case "false":
			suspend = false
		case "true":
			suspend = true
		default:
			log.Printf("Failed to read active suspend annotation from CronJob \"%s\".", cron.Metadata.Name)
			continue
		}
		if cron.Spec.Suspend != suspend {
			patchCronjob(
				cfg,
				cron,
				client,
				token,
				suspend,
			)
		}
	}
}

// Set local cluster to be inactive. For deployments, set the current replicas
// to match the inactive replicas annotation. For cronjobs, set the suspend
// flag to match the inactive suspend annotation.
func setInactive(
	cfg Config,
	managedDeploys []Deployment,
	managedCrons []Cronjob,
	client *http.Client,
	token string) {

	//log.Println("Internal and external IDs do not match. Setting cluster to inactive.")

	for _, deploy := range managedDeploys {
		replStr := deploy.Metadata.Annotations.InactiveMaxReplicas
		inactiveReplicas, err := strconv.Atoi(replStr)
		if err != nil {
			log.Printf("Failed to read inactive replicas from deployment \"%s\".", deploy.Metadata.Name)
			continue
		}
		if deploy.Spec.Replicas > inactiveReplicas {
			patchDeployment(
				cfg,
				deploy,
				inactiveReplicas,
				client,
				token)
		}
	}

	for _, cron := range managedCrons {
		suspendStr := cron.Metadata.Annotations.InactiveSuspend

		var suspend bool

		switch suspendStr {
		case "false":
			suspend = false
		case "true":
			suspend = true
		default:
			log.Printf("Failed to read inactive suspend annotation from CronJob \"%s\".", cron.Metadata.Name)
			continue
		}
		if cron.Spec.Suspend != suspend {
			patchCronjob(
				cfg,
				cron,
				client,
				token,
				suspend,
			)
		}
	}
}

// Get all cronjobs in the cluster matching a given label
func getCronjobs(
	cfg Config,
	client *http.Client,
	token string,
) []Cronjob {

	//Build the url to the specific deployment resource
	url := fmt.Sprintf(
		"%s/apis/batch/v1/cronjobs?labelSelector=%s",
		cfg.APIServer,
		LABEL)

	req, err := http.NewRequest("GET", url, nil)
	bearer := "Bearer " + token
	req.Header.Add("Authorization", bearer)
	if err != nil {
		log.Print(err)
	}

	//log.Printf("Making request to %s\n", url)
	resp, err := client.Do(req)
	if err != nil {
		log.Print(err)
	}
	defer resp.Body.Close()
	//log.Printf("Received %s", resp.Status)

	var cronjobList CronjobList

	err = json.NewDecoder(resp.Body).Decode(&cronjobList)
	if err != nil {
		log.Print(err)
	}
	return cronjobList.Items

}

func patchCronjob(
	cfg Config,
	cron Cronjob,
	client *http.Client,
	token string,
	suspend bool) {
	url := fmt.Sprintf(
		"%s/apis/batch/v1/namespaces/%s/cronjobs/%s",
		cfg.APIServer,
		cron.Metadata.Namespace,
		cron.Metadata.Name)

	var patch [1]CronJSONPatch

	patch[0] = CronJSONPatch{
		Op:    "replace",
		Path:  "/spec/suspend",
		Value: suspend,
	}

	jsonBytes, err := json.Marshal(patch)
	if err != nil {
		log.Printf(
			"Failed to marshal JSON patch for cronjob %s: %s",
			cron.Metadata.Name,
			err,
		)
		return
	}

	jsonBuf := bytes.NewBuffer(jsonBytes)

	req, err := http.NewRequest("PATCH", url, jsonBuf)
	bearer := "Bearer " + token
	req.Header.Add("Authorization", bearer)
	if err != nil {
		log.Printf(
			"Failed to create http PATCH request for cronjob %s: %s",
			cron.Metadata.Name,
			err,
		)
		return
	}
	req.Header.Set("Content-Type", "application/json-patch+json")

	log.Printf("Patching CronJob: %s, in namespace: %s\n", cron.Metadata.Name, cron.Metadata.Namespace)
	log.Printf("patch: %s\n", jsonBuf)
	resp, err := client.Do(req)
	if err != nil {
		log.Print(err)
	}
	log.Printf("PATCH result status code: %d\n", resp.StatusCode)
}

// Get the cluster ID of a server
func getID(url string) (ID, error) {
	var id ID
	resp, err := http.Get(url)
	if err != nil {
		return id, err
	}

	body, err := io.ReadAll(resp.Body)
	resp.Body.Close()

	if resp.StatusCode != 200 {
		log.Printf("Request to %s failed with status code: %d", url, resp.StatusCode)
		return id, err
	}

	json.Unmarshal(body, &id)

	return id, nil
}

// Read configuration from ./config.json
func readConfig() Config {
	var cfg Config
	dat, err := os.ReadFile("./config/config.json")
	if err != nil {
		log.Fatalf("Failed to read config file: %s", err)
	}

	err = json.Unmarshal(dat, &cfg)
	if err != nil {
		log.Fatalf("Failed to parse config json: %s", err)
	}

	return cfg
}
