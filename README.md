# Horizontal Autoscaling via Prometheus

This repository contains code to deploy an horizontal pod autoscaler on Kubernetes cluster that scales a Deployment according to a Custom Metric collected by Prometheus.

![Overview](hpa__.png)

## Exporter 
This component retrives metrics coming from third-party's applications and make them available to Prometheus server. As an example we will see three different types of exporters associated to different kinds of applications.

### Process Exporter
https://github.com/ncabatoff/process-exporter

Prometheus exporter that mines /proc to report on selected processes.
### Apache Exporter
https://github.com/Lusitaniae/apache_exporter

Apache Exporter is a Prometheus exporter for Apache metrics that exports Apache server status reports generated by ```mod_status``` with the URL of ```http://127.0.0.1/server-status/?auto```.

### CouchDB Exporter
https://github.com/gesellix/couchdb-prometheus-exporter

The CouchDB metrics exporter requests the CouchDB stats from the /_stats and /_active_tasks endpoints and exposes them for Prometheus consumption.

## Prometheus
https://prometheus.io/docs/prometheus/latest/configuration/configuration/

This is the Prometheus server itself, which collects all metrics from various exporters in the form of time series. Those can be accessed and visualized through a WebUI.

A generic Prometheus configurations has to be written in YAML format.
For our purposes, the only sections we will use are:

- ```global```: This configuration specifies parameters that are valid in all other configuration contexts. They also serve as  defaults for other configuration sections.``
- ```scrape_config```: This section specifies a set of targets and parameters describing how to scrape them. In the general case, one scrape configuration specifies a single job. In advanced configurations, this may change. Targets may be statically configured via the ```static_configs``` parameter or dynamically discovered using one of the supported service-discovery mechanisms:
  - ```static_configs```:
  - ```kubernetes_sd_configs```: Kubernetes SD configurations allow retrieving scrape targets from Kubernetes' REST API and always staying synchronized with the cluster state. One of the following role types can be configured to discover targets:
    - ```node```: The node role discovers one target per cluster node with the address defaulting to the Kubelet's HTTP port. The target address defaults to the first existing address of the Kubernetes node object in the address type order of NodeInternalIP, NodeExternalIP, NodeLegacyHostIP, and NodeHostName.
    - ```service```: The service role discovers a target for each service port for each service. This is generally useful for blackbox monitoring of a service. The address will be set to the Kubernetes DNS name of the service and respective service port.
    -  ```pod```: The pod role discovers all pods and exposes their containers as targets. For each declared port of a container, a single target is generated. If a container has no specified ports, a port-free target per container is created for manually adding a port via relabeling.
    - ```endpoint```: The endpoints role discovers targets from listed endpoints of a service. For each endpoint address one target is discovered per port. If the endpoint is backed by a pod, all additional container ports of the pod, not bound to an endpoint port, are discovered as targets as well.
    - ```ingress```: The ingress role discovers a target for each path of each ingress. This is generally useful for blackbox monitoring of an ingress. The address will be set to the host specified in the ingress spec.

### Example
```
    global:
      scrape_interval: 10s
      evaluation_interval: 10s
    scrape_configs:      
      - job_name: 'kube-eagle'
        static_configs:
            - targets: ['kube-eagle-service-cluster-IP:8080']
          
      - job_name: 'httpgo-pod'
        kubernetes_sd_configs:
        - role: pod
        relabel_configs:
        - source_labels: [__meta_kubernetes_pod_label_app]
          action: keep
          regex: httpgo
```

## Prometheus Adapter
https://github.com/DirectXMan12/k8s-prometheus-adapter/blob/master/docs/config.md

The Prometheus Adapter application selects (and manipulates) certain time series from Prometheus Server and exposes them through Custom Metrics API in order to make them available to the Horizontal Pod Autoscaler.
The adapter takes the standard Kubernetes generic API server arguments (including those for authentication and authorization). By default, it will attempt to using Kubernetes in-cluster config to connect to the cluster.

It takes ```--config=<yaml-file>``` (```-c```) as an argument: this configures how the adapter discovers available Prometheus metrics and the associated Kubernetes resources, and how it presents those metrics in the custom metrics API.
The adapter determines which metrics to expose, and how to expose them, through a set of "discovery" rules which are made of four parts:
- ```Discovery```, which specifies how the adapter should find all Prometheus metrics for this rule. You can use two fields: 
  - ```seriesQuery```: specifies Prometheus series query (as passed to the /api/v1/series endpoint in Prometheus) to use to find some set of Prometheus series 
  - ```seriesFilters```:
- ```Association```, which specifies how the adapter should determine which Kubernetes resources a particular metric is associated with. You can use the ```resources``` field.
There are two ways to associate resources with a particular metric, using two different sub-fields:
  - ```template```: specify that any label name that matches some particular pattern refers to some group-resource based on the label name. The pattern is specified as a Go template, with the ```Group``` and ```Resource``` fields representing group and resource. 
  - ```overrides```: specify that some particular label represents some particular Kubernetes resource. Each override maps a Prometheus label to a Kubernetes group-resource. 
- ```Naming```, which specifies how the adapter should expose the metric in the custom metrics API. You can use the ```name``` field, specifying a pattern to extract an API name from a Prometheus name, and potentially a transformation on that extracted value:
  - ```matches```: a regular expression that specifies pattern. If not specified, it defaults to ```.*```. 
  - ```as```: specifies the transformation. You can use any capture groups defined in the ```matches``` field. If the matches field doesn't contain capture groups, the ```as``` field defaults to ```$0```. If it contains a single capture group, the ```as``` field defaults to ```$1```. Otherwise, it's an error not to specify the ```as``` field.
- ```Querying```, which specifies how a request for a particular metric on one or more Kubernetes objects should be turned into a query to Prometheus. You can use the ```metricsQuery``` field, specifying a Go template that gets turned into a Prometheus query, using input from a particular call to the custom metrics API. A given call to the custom metrics API is distilled down to a metric name, a group-resource, and one or more objects of that group-resource. These get turned into the following fields in the template:
  - ```Series```: the metric name
  - ```LabelMatchers```: a comma-separated list of label matchers matching the given objects. Currently, this is the label for the particular group-resource, plus the label for namespace, if the group-resource is namespaced.
  - ```GroupBy```: a comma-separated list of labels to group by. Currently, this contains the group-resource label used in ```LabelMatchers```.

### Example
```
  rules:
  - seriesQuery: 'process_exporter_load1{instance="10.100.1.135:18000",job="kubernetes-pods"}'  
    resources:
      template: "<<.Resource>>"
    name:
      matches: "^(.*)_load1"
      as: "${1}_test"
    metricsQuery: <<.Series>>
```
