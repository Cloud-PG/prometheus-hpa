# Horizontal Autoscaling via Prometheus

This repository contains code to deploy an horizontal pod autoscaler on Kubernetes cluster that scales a Deployment according to a Custom Metric collected by Prometheus.

![Overview](hpa_new.png)

<a name="quickstart"></a>
## Quick Start
This repository requires a Kubernetes 1.13 installation.

<a name="quickstart"></a>
### Brig up playground with script
Just type
```
sh deploy.sh
```
to deploy:
- three different apps:
  - a pod with httpgo server and a process exporter - exposed through NodePort service
  - a pod with httpd server and an apache exporter - exposed through NodePort service
  - a pod with CouchDB and a couchDB exporter - exposed through NodePort service
- a Prometheus server
- a Prometheus adapter pod
- three different horizontal pod autoscalers to scale each app (up to 10 replicas) accornding to a specific metric:
  - httpgo: ```number of open file descriptors```
  - httpd: ```number of accesses per second```
  - couchDB: ```number of reads```

*Here NodePort services are used in order to make apps reachable from outside the cluster.*

<a name="quickstart"></a>
### Step-by-step configuration
First of all we need to deploy the three apps:
- ```httpgo server```: in this case, our deployment will use ```ttedesch/httpgo_exporter:latest``` image, which is built inserting process_exporter process inside ```veknet/httpgo``` image. Dockerfile, entrypoint and scripts used can be found [here](httpgo_exporter). The deployment file is [this](manifests_no_configs/httpgo_and_exporter.yaml). A Nodeport service exposes port 31000 in order to make the httpgo reachable from outside.
  ```
  kubectl apply -f manifests_no_configs/httpgo_and_exporter.yaml
  ```

- ```httpd server```: in this case, our deployment will use ```ttedesch/httpd:latest``` and ```bitnami/apache-exporter``` images. The first one is built using ```apt-get install apache2``` and setting mod_status, Dockerfile and config file used can be found [here](apache_server). The deployment file is [this](manifests_no_configs/httpd_and_exporter.yaml). A Nodeport service is used to make the httpd server reachable from outside.
  ```
  kubectl apply -f manifests_no_configs/httpd_and_exporter.yaml
  ```

- ```couchDB```: in this case, our deployment will use ```couchdb:latest``` and ``` gesellix/couchdb-prometheus-exporter``` images. The deployment file is [this](manifests_no_configs/couchdb_and_exporter.yaml). A Nodeport service is used to make the couchDB server reachable from outside.
  ```
  kubectl apply -f manifests_no_configs/couchdb_and_exporter.yaml
  ```
  
Then, let's deploy the Prometheus Server. First of all we need to create the ConfigMap contaning its configuration. In the Prometheus Server section you can see how to write a proper [configuration file](configs/prometheus.yml).
```
kubectl create configmap prometheus-example-cm --from-file configs/prometheus.yml
```
Then, let's deploy Prometheus server itself, mounting that ConfigMap as volume ([manifest](manifests_no_configs/prometheus.yaml)
```
kubectl apply -f manifests_no_configs/prometheus.yaml
```
Analogously, we need to [configure](configs/prometheus_adapter.yml) and [deploy](manifests_no_configs/prometheus_adapter.yaml) the prometheus adapter deployment which will query prometheus and expose metrics through Custom Metrics API.
```
kubectl create configmap prometheus-example-cm --from-file configs/prometheus_adapter.yml
kubectl apply -f manifests_no_configs/prometheus_adapter.yaml
```

In the end, let's deploy the three Horizontal Pod Autoscalers ([httpgo](manifests_no_configs/hpa_hpptgo.yaml), [httpd](manifests_no_configs/hpa_httpd.yaml), [couchdb](manifests_no_configs/hpa_couchdb.yaml)) which will scale those three apps according to specific metrics:
- httpgo: ```number of open file descriptors```
- httpd: ```number of accesses per second```
- couchDB: ```number of reads```

```
kubectl apply -f manifests_no_configs/hpa_hpptgo.yaml
kubectl apply -f manifests_no_configs/hpa_httpd.yaml
kubectl apply -f manifests_no_configs/hpa_couchdb.yaml
```



<a name="quickstart"></a>
### How to test and debug
The prometheus webUI is available at ```http://<masternode-publicIP>:<PrometheusService-nodePort>```:

![webUI](prometheus_WebUI.png)



The exposed metrics can be seen running:
```
$ kubectl get --raw /apis/custom.metrics.k8s.io/v1beta1/ | jq
```

To see if scaling is active:
```
$ kubectl describe hpa
```
<a name="quickstart"></a>
## To expose additional metrics


<a name="quickstart"></a>
# Exporter 
This component retrives metrics coming from third-party's applications and make them available to Prometheus server. As an example we will see three different types of exporters associated to different kinds of applications.

<a name="quickstart"></a>
## Process Exporter
https://github.com/ncabatoff/process-exporter

Prometheus exporter that mines /proc to report on selected processes.

<a name="quickstart"></a>
## Apache Exporter
https://github.com/Lusitaniae/apache_exporter

Apache Exporter is a Prometheus exporter for Apache metrics that exports Apache server status reports generated by ```mod_status``` with the URL of ```http://127.0.0.1/server-status/?auto```.

<a name="quickstart"></a>
## CouchDB Exporter
https://github.com/gesellix/couchdb-prometheus-exporter

The CouchDB metrics exporter requests the CouchDB stats from the /_stats and /_active_tasks endpoints and exposes them for Prometheus consumption.

<a name="quickstart"></a>
# Prometheus
https://prometheus.io/docs/prometheus/latest/configuration/configuration/

This is the Prometheus server itself, which collects all metrics from various exporters in the form of time series. Those can be accessed and visualized through a WebUI.

A generic Prometheus configuration has to be written in YAML format.
For our purposes, the only sections we will use are:

- ```global```: This configuration specifies parameters that are valid in all other configuration contexts. They also serve as  defaults for other configuration sections.
- ```scrape_config```: This section specifies a set of targets and parameters describing how to scrape them. In the general case, one scrape configuration specifies a single job. In advanced configurations, this may change. Targets may be statically configured via the ```static_configs``` parameter or dynamically discovered using one of the supported service-discovery mechanisms:
  - ```static_configs```: configure targets statically
  - ```kubernetes_sd_configs```: Kubernetes SD configurations allow retrieving scrape targets from Kubernetes' REST API and always staying synchronized with the cluster state. One of the following role types can be configured to discover targets:
    - ```node```: The node role discovers one target per cluster node with the address defaulting to the Kubelet's HTTP port. The target address defaults to the first existing address of the Kubernetes node object in the address type order of NodeInternalIP, NodeExternalIP, NodeLegacyHostIP, and NodeHostName.
    - ```service```: The service role discovers a target for each service port for each service. This is generally useful for blackbox monitoring of a service. The address will be set to the Kubernetes DNS name of the service and respective service port.
    -  ```pod```: The pod role discovers all pods and exposes their containers as targets. For each declared port of a container, a single target is generated. If a container has no specified ports, a port-free target per container is created for manually adding a port via relabeling.
    - ```endpoint```: The endpoints role discovers targets from listed endpoints of a service. For each endpoint address one target is discovered per port. If the endpoint is backed by a pod, all additional container ports of the pod, not bound to an endpoint port, are discovered as targets as well.
    - ```ingress```: The ingress role discovers a target for each path of each ingress. This is generally useful for blackbox monitoring of an ingress. The address will be set to the host specified in the ingress spec.

<a name="quickstart"></a>
## Example
```
    global:                                                         # How frequently to scrape targets and evaluate rules by default
      scrape_interval: 10s
      evaluation_interval: 10s
    scrape_configs:                                                 
      
      - job_name: 'kube-eagle'                                      # Scrape an exporter statically using its service IP address and port, name this scraping job as 'kube-eagle'
        static_configs:                                             
            - targets: ['kube-eagle-service-cluster-IP:8080']
          
      - job_name: 'httpgo-pod'                                      # Scrape an exporter dynamically, looking for a pod with label app 'httpgo', name this scraping job as 'httpgo-pod'                                    
        kubernetes_sd_configs:                                           
        - role: pod
        relabel_configs:
        - source_labels: [__meta_kubernetes_pod_label_app]
          action: keep
          regex: httpgo
```
<a name="quickstart"></a>
# Prometheus Adapter
https://github.com/DirectXMan12/k8s-prometheus-adapter/blob/master/docs/config.md

The Prometheus Adapter application selects (and manipulates) certain time series from Prometheus Server and exposes them through Custom Metrics API in order to make them available to the Horizontal Pod Autoscaler.
The adapter takes the standard Kubernetes generic API server arguments (including those for authentication and authorization). By default, it will attempt to using Kubernetes in-cluster config to connect to the cluster.

It takes ```--config=<yaml-file>``` (```-c```) as an argument: this configures how the adapter discovers available Prometheus metrics and the associated Kubernetes resources, and how it presents those metrics in the custom metrics API.
The adapter determines which metrics to expose, and how to expose them, through a set of "discovery" rules which are made of four parts:
- ```Discovery```, which specifies how the adapter should find all Prometheus metrics for this rule. You can use two fields: 
  - ```seriesQuery```: specifies Prometheus series query (as passed to the /api/v1/series endpoint in Prometheus) to use to find some set of Prometheus series 
  - ```seriesFilters```: to do additional filtering on metric names
- ```Association```, which specifies how the adapter should determine which Kubernetes resources a particular metric is associated with. You can use the ```resources``` field.
There are two ways to associate resources with a particular metric, using two different sub-fields:
  - ```template```: specify that any label name that matches some particular pattern refers to some group-resource based on the label name. The pattern is specified as a Go template, with the ```Group``` and ```Resource``` fields representing group and resource. 
  - ```overrides```: specify that some particular label represents some particular Kubernetes resource. Each override maps a Prometheus label to a Kubernetes group-resource. 
- ```Naming```, which specifies how the adapter should expose the metric in the custom metrics API. You can use the ```name``` field, specifying a pattern to extract an API name from a Prometheus name, and potentially a transformation on that extracted value:
  - ```matches```: a regular expression (https://docs.python.org/3/library/re.html) that specifies pattern. If not specified, it defaults to ```.*```. 
  - ```as```: specifies the transformation. You can use any capture groups defined in the ```matches``` field. If the matches field doesn't contain capture groups, the ```as``` field defaults to ```$0```. If it contains a single capture group, the ```as``` field defaults to ```$1```. Otherwise, it's an error not to specify the ```as``` field.
- ```Querying```, which specifies how a request for a particular metric on one or more Kubernetes objects should be turned into a query to Prometheus. You can use the ```metricsQuery``` field, specifying a Go template that gets turned into a Prometheus query, using input from a particular call to the custom metrics API. A given call to the custom metrics API is distilled down to a metric name, a group-resource, and one or more objects of that group-resource. These get turned into the following fields in the template:
  - ```Series```: the metric name
  - ```LabelMatchers```: a comma-separated list of label matchers matching the given objects. Currently, this is the label for the particular group-resource, plus the label for namespace, if the group-resource is namespaced.
  - ```GroupBy```: a comma-separated list of labels to group by. Currently, this contains the group-resource label used in ```LabelMatchers```.

<a name="quickstart"></a>
## Example
```
  rules:
  - seriesQuery: 'testmetric_total{instance="10.100.1.135:18000",job="kubernetes-pods"}'          # DISCOVERY of process_exporter_load1 time series with certain instance and job labels
    resources:
      template: "<<.Resource>>"                                                                   # ASSOCIATION of that metric to the resource which is present in the labels (job.batch)
    name:
      matches: "^(.*)_total"                                                                      # NAMING of the metric: modify the last part of the name, from testmetric_total to test_metric_per_second
      as: "${1}_per_second"
    metricsQuery: 'sum(rate(<<.Series>>{<<.LabelMatchers>>}[2m])) by (<<.GroupBy>>)'              # QUERYING of the metric, calculating a rate value averaged every 2 minutes and summing up                                                                    
```

<a name="quickstart"></a>
# Horizontal Pod Autoscaler
https://kubernetes.io/docs/tasks/run-application/horizontal-pod-autoscale-walkthrough/#autoscaling-on-multiple-metrics-and-custom-metrics

By making use of the autoscaling/v2beta2 API version you can introduce metrics to use when autoscaling a deployment. The Horizontal Pod Autoscaler is implemented as a control loop that periodically queries a metrics API. There are three types of metrics:

- ```Resource metrics```: based on CPU or memory usage of a pod, exposed through ```metrics.k8s.io``` API (```Resource type```)

- ```Custom metrics```: based on any metric reported by a Kubernetes object in a cluster, exposed through ```custom.metrics.k8s.io``` API 
  - ```Pod type```: describe pods, and are averaged together across pods and compared with a target value to determine the replica count. 
  - ```Object type```: describe a different object in the same namespace, instead of describing pods. The metrics are not necessarily fetched from the object; they only describe it.
  
- ```External metrics```: based on a metric from an application or service external to your cluster, exposed through ```external.metrics.k8s.io``` API (```External type```).

<a name="quickstart"></a>
## Example
```
apiVersion: autoscaling/v2beta2     
kind: HorizontalPodAutoscaler
    metadata:                               # metadata of the autoscaler
  name: httpgo-hpa 
  namespace: default
spec:
  scaleTargetRef:                           # spec of the Kubernetes resource to be scaled (in this case a Deployment)
    apiVersion: apps/v1   
    kind: Deployment
    name: httpgo
  minReplicas: 1                            # max and min number of replicas of that resource
  maxReplicas: 10
  metrics:                                  # spec of the metric used for scaling (in this case a Job Object)
  - type: Object
    object:                               
      metric:
        name: myapphttp_process_open_fds
      describedObject:
        apiVersion: batch/v1
        kind: Job
        name: httpgo-pod
      target:                               # threshold value
        type: Value
        value: 0.5
 ```
 

