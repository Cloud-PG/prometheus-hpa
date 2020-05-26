# prometheus-hpa

## The problem
This repository contains code to deploy an horizontal pod autoscaler on Kubernetes cluster that scales a Deployment according to a Custom Metric collected by Prometheus.

![Overview](hpa_.png)

## Requirements
This requires a Kubernetes cluster v1.18 with a [kube-eagle](kube-eagle/kube-eagle.yaml) installation and [process-exporter](process_exporter/process_exporter_deployment.yaml) installation. An [httpgo server](httpgo/httpgo.yaml) should also be installed.

## Quick Start
Prometheus manifest is [this](prometheus/prometheus.yaml), it is set to scrape kube-eagle service (static_config, subsititute default kube-eagle service cluster-IP value with the one of your cluster) and process_exporter pod (kubernetes_sd_configs). The prometheus webUI is available at ```http://<masternode-publicIP>:<PrometheusService-nodePort>```.

As an example, the [prometheus_adapter](prometheus/prometheus_adapter.yaml) is set to look for a particular metric (```process_exporter_load1``` renamed as ```process_exporter_test```) and exposes it through Custom Metrics API. This can be seen running the command
````
$kubectl get --raw /apis/custom.metrics.k8s.io/v1beta1/ | jq
{
  "kind": "APIResourceList",
  "apiVersion": "v1",
  "groupVersion": "custom.metrics.k8s.io/v1beta1",
  "resources": [
    {
      "name": "jobs.batch/process_exporter_test",
      "singularName": "",
      "namespaced": true,
      "kind": "MetricValueList",
      "verbs": [
        "get"
      ]
    }
  ]
}
````

Then, an [horizontal pod autoscaler](hpa/hpa.yaml) is set to scale httpgo deployment according to the exposed metric. To see if scaling is active:
````
$ kubectl describe hpa
Name:                                                             httpgo-hpa
Namespace:                                                        default
Labels:                                                           <none>
Annotations:                                                      CreationTimestamp:  Mon, 18 May 2020 18:31:22 +0200
Reference:                                                        Deployment/httpgo
Metrics:                                                          ( current / target )
  "process_exporter_test" on Job/kubernetes-pods (target value):  1490m / 1200m
Min replicas:                                                     1
Max replicas:                                                     10
Deployment pods:                                                  10 current / 10 desired
Conditions:
  Type            Status  Reason               Message
  ----            ------  ------               -------
  AbleToScale     True    ScaleDownStabilized  recent recommendations were higher than current one, applying the highest recent recommendation
  ScalingActive   True    ValidMetricFound     the HPA was able to successfully calculate a replica count from Job metric process_exporter_test
````
