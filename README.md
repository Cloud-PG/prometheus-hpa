# prometheus-hpa

## The problem
This repository contains code to deploy an horizontal pod autoscaler on Kubernetes cluster that scales a Deployment according to a Custom Metric collected by Prometheus.

![Overview](hpa_.png)

## Requirements
This requires a Kubernetes cluster v1.14 or higher. 

## Quick Start
First of all, let's deploy two Prometheus exporters which make a certain list of metrics available to Prometheus:
- the first one is [kube-eagle](https://github.com/cloudworkz/kube-eagle) which is a standard Prometheus exporter:
    ```
    $ kubectl apply -f manifests/kube-eagle.yaml
    ```
- the second one is a custom process exporter whose image is built starting from this [go script](process_exporter/process_exporter.go) and with this [Dockerfile](process_exporter/Dockerfile)
    ```
    $ kubectl apply -f manifests/process_exporter_deployment.yaml
    ```

Then, let's deploy an [httpgo server](httpgo/httpgo.yaml) and an ingress:
    ```
    $ kubectl apply -f manifests/httpgo.yaml
    $ kubectl apply -f manifests/ingress.yaml
    ```

Now we have to deploy a Prometheus server (https://prometheus.io/docs/introduction/overview/).
In order to make it scrape kube-eagle service and process_exporter pod, in ```manifests/prometheus.yaml``` subsititute ```kube-eagle-service-cluster-IP``` with the real value in your cluster, which could be obtained with 
```
$ kubectl describe service kube-eagle -n monitoring
```
Then, let's deploy the Prometheus server
```
$ kubectl apply -f manifests/prometheus.yaml
```

The prometheus webUI is available at ```http://<masternode-publicIP>:<PrometheusService-nodePort>```.

As an example, let's deploy a [prometheus_adapter](https://github.com/DirectXMan12/k8s-prometheus-adapter) which is set to look for a particular metric (```process_exporter_load1``` renamed as ```process_exporter_test```) and exposes it through Custom Metrics API. 
So, run
````
$ kubectl apply -f manifests/prometheus_adapter.yaml
````

The exposed metrics can be seen running:
````
$ kubectl get --raw /apis/custom.metrics.k8s.io/v1beta1/ | jq
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

Finally, let's deploy an [horizontal pod autoscaler](https://kubernetes.io/docs/tasks/run-application/horizontal-pod-autoscale/) which is set to scale httpgo deployment according to the exposed metric. 
````
$ kubectl apply -f manifests/hpa.yaml
````

Now, let's generate some load on our system:
```
$ go get -u github.com/rakyll/hey
$ export GOPATH="$HOME/go"
$ PATH="$GOPATH/bin:$PATH"
$ hey -q 10 -c 1 -z 1m http://<name_of_your_node>/http
```
where the name of the node can be obtained with 

```
$ kubectl get nodes
```

This should make our metric rise above the previously-set threshold and the horizontal pod autoscaler should get in action scaling httpgo deployment. 

To see if scaling is active:
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
