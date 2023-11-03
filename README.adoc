# Failover Manager

A Kubernetes controller for use in multi-cluster configurations which use DNS
for failover. 

## Background

When multiple Kubernetes clusters are configured for DR purposes, one could use
DNS with health checks to determine which cluster would receive API traffic.
This works well for APIs which initiate work upon request, but other types of
workloads like CronJobs or consumers, run regardless of the state of the DNS
health check.

Failover Manager sets up a tiny API which allows the Failover Manager
controller to become aware of where a DNS failover record is pointed. It then
uses this information to re-configure workloads based on whether the cluster is
active or not.


## Requirements

- Kubernetes versions 1.19 - 1.21 (other versions may work, but the changes to
  service account tokens in v1.22+ may break the controller authorization)
- Helm 3

## Installation

1. Configure the `values.yaml` file for the API to expose the service outside
  the cluster. 
-- Possible options for presenting the API externally include nginx ingress, an
   API Gateway, or possibly a service type of LoadBalancer. Take note of the
   URL by which this will be reached.
2. Configure the `values.yaml` file for the controller.

-- `intUrl` - the URL of API using the Kubernetes DNS name: i.e http://failovermanager-api.failovermanager.svc.cluster.local
-- `extUrl` - the URL of the API which is externally accessible
-- `pollPeriodSeconds` - the number of seconds between polling the internal and
   external APIs
--x `apiServer` - the internal url of the Kubernetes API. Typically
   `https://kubernetes.default.svc`
3. At the root of the helm directory run `helm install failovermanager ./
  --create-namespace --namespace failovermanager` 



## Usage

Failover Manager currently supports `Deployments` and `CronJob` workloads. Add
the following annotations and labels for workloads to be managed by Failover
Manager.

- Label: `failovermanager: enabled`
- Annotations:
-- Deployments:
--- `failovermanager/active-replicas: "3"`
---- Describes the minimum number of replicas set when the cluster is "active".
--- `failovermanager/inactive-replicas: "0"`
---- Describes the maximum number of replicas set when the cluster is "inactive".
-- CronJobs:
--- `failovermanager/active-suspend: "false"`
---- When the cluster is "active" the CronJob `suspend` flag will be `false`,
	 enabling normal job scheduling.
--- `failovermanager/inactive-suspend: "true"`
---- When the cluster is "inactive" the CronJob `suspend` flag will be `true`,
	 preventing the job from being scheduled.




