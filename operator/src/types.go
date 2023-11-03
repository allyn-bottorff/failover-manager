package main

import (
	"log"
)

// ID from the failover manager API which uniquely identifies a cluster
type ID struct {
	ID string `json:"id"`
}

// Configuration information read from a ConfigMap (or config file)
type Config struct {
	IntURL            string `json:"intURL"`
	ExtURL            string `json:"extURL"`
	PollPeriodSeconds int    `json:"pollPeriodSeconds"`
	APIServer         string `json:"apiServer"`
	Debug             bool   `json:"debug,omitempty"`
	//Label             string `json:"label,omitempty:"`
	// label: vu.com/failover-manger: enabled
}

// DeploymentList type from reading multiple deployments
type DeploymentList struct {
	Items    []Deployment       `json:"items"`
	Metadata DeploymentListMeta `json:"metadata"`
}

// DeploymentList metadata for resource version
type DeploymentListMeta struct {
	ResourceVersion string `json:"resourceVersion"`
}

// Holds kubenetes deployment manifest information, but only what is relevant
// to this project.
type Deployment struct {
	Kind       string             `json:"kind"`
	ApiVersion string             `json:"apiVersion"`
	Metadata   DeploymentMetadata `json:"metadata"`
	Spec       DeploymentSpec     `json:"spec"`
}

// Print the Deployment out with manual formatting.
func (d *Deployment) fPrint() {
	log.Printf("Kind: %s\n", d.Kind)
	log.Printf("ApiVersion: %s\n", d.ApiVersion)
	log.Printf("Metadata: \n")
	log.Printf("\t Name: %s\n", d.Metadata.Name)
	log.Printf("\t Namespace: %s\n", d.Metadata.Namespace)
	log.Printf("Spec: \n")
	log.Printf("\t Replicas: %d\n", d.Spec.Replicas)
	log.Printf("Annotations: \n")
	log.Printf("\t Active Replicas: %s\n", d.Metadata.Annotations.ActiveMinReplicas)
	log.Printf("\t Inactive Replicas: %s\n", d.Metadata.Annotations.InactiveMaxReplicas)
}

// Top level deployment manifest metadata.
type ManifestMeta struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

// Deployment spec partial content.
type DeploymentSpec struct {
	Replicas int `json:"replicas"`
}

// Deployment metadata partial content.
type DeploymentMetadata struct {
	Annotations DeploymentAnnotations `json:"annotations"`
	Name        string                `json:"name"`
	Namespace   string                `json:"namespace"`
}

// Deployment annotations partial content. Contains primary and secondary
// replica details
type DeploymentAnnotations struct {
	ActiveMinReplicas   string `json:"failovermanager/active-min-replicas"`
	InactiveMaxReplicas string `json:"failovermanager/inactive-max-replicas"`
}

// Cronjob list spec
type CronjobList struct {
	Items []Cronjob `json:"items"`
}

type Cronjob struct {
	ApiVersion string          `json:"apiVersion"`
	Metadata   CronjobMetadata `json:"metadata"`
	Spec       CronjobSpec     `json:"spec"`
}

type CronjobMetadata struct {
	Annotations CronjobAnnotations `json:"annotations"`
	Name        string             `json:"name"`
	Namespace   string             `json:"namespace"`
}

type CronjobSpec struct {
	Suspend bool `json:"suspend"`
}

type CronjobAnnotations struct {
	ActiveSuspend   string `json:"failovermanager/active-suspend"`
	InactiveSuspend string `json:"failovermanager/inactive-suspend"`
}

// JSON Patch struct that can be used to patch an integer JSON value.
type DeployJSONPatch struct {
	Op    string `json:"op"`
	Path  string `json:"path"`
	Value int    `json:"value"`
}

type CronJSONPatch struct {
	Op    string `json:"op"`
	Path  string `json:"path"`
	Value bool   `json:"value"`
}
