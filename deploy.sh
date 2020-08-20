#!/bin/bash
##H Usage: deploy.sh ACTION DEPLOYMENT 
##H
##H Script actions:
##H   help       show this help
##H   clean      cleanup all services (no argument needed)
##H   create     create provided deployment (all if no argument provided)
##H   scale      scale given deployment
##H
##H Deployments:
##H   namespaces   
##H   configmaps   
##H   services   
##H   ingress    
##H   monitoring 

if [ "$1" == "-h" ] || [ "$1" == "-help" ] || [ "$1" == "--help" ] || [ "$1" == "help" ] || [ "$1" == "" ]; then
    perl -ne '/^##H/ && do { s/^##H ?//; print }' < $0
    exit 1
fi

deploy_ns()
{
    # deploy all appropriate namespaces
    kubectl create namespace http
    kubectl create namespace frontend
    kubectl create namespace database
    kubectl create namespace monitoring
}

deploy_configmaps()
{
    kubectl create configmap prometheus-example-cm --from-file configs/prometheus.yml -n monitoring
    kubectl create configmap prometheus-adapter-example-cm --from-file configs/prometheus_adapter.yml
}

deploy_services()
{
    kubectl apply -f manifests_no_configs/httpgo_and_exporter.yaml --validate=false -n http
    kubectl apply -f manifests_no_configs/httpd_and_exporter.yaml --validate=false -n frontend
    kubectl apply -f manifests_no_configs/couchdb_and_exporter.yaml --validate=false -n database
}

deploy_monitoring()
{
    kubectl apply -f manifests_no_configs/prometheus.yaml
    kubectl apply -f manifests_no_configs/prometheus_adapter.yaml
}

deploy_ingress()
{
    kubectl apply -f manifests_no_configs/ingress.yaml
}

scale(){
    kubectl apply -f manifests_no_configs/hpa_httpgo.yaml
    kubectl apply -f manifests_no_configs/hpa_httpd.yaml
    kubectl apply -f manifests_no_configs/hpa_couchdb.yaml
}

cleanup(){

    kubectl delete configmap prometheus-example-cm -n monitoring
    kubectl delete configmap prometheus-adapter-example-cm 
    kubectl delete -f manifests_no_configs/httpgo_and_exporter.yaml  -n http
    kubectl delete -f manifests_no_configs/httpd_and_exporter.yaml -n frontend
    kubectl delete -f manifests_no_configs/couchdb_and_exporter.yaml -n database
    kubectl delete -f manifests_no_configs/prometheus.yaml
    kubectl delete -f manifests_no_configs/prometheus_adapter.yaml
    kubectl delete -f manifests_no_configs/ingress.yaml
    kubectl delete -f manifests_no_configs/hpa_httpgo.yaml
    kubectl delete -f manifests_no_configs/hpa_httpd.yaml
    kubectl delete -f manifests_no_configs/hpa_couchdb.yaml
    kubectl delete namespace http
    kubectl delete namespace frontend
    kubectl delete namespace database
    kubectl delete namespace monitoring

}

action=$1
deployment=$2

if [ "$action" == "create" ]; then
    if [ -z "$deployment" ]; then
        deploy_ns
        deploy_configmaps
        deploy_services
        deploy_monitoring
        deploy_ingress
    elif [ "$deployment" == "namespaces" ]; then
        deploy_ns
    elif [ "$deployment" == "configmaps" ]; then
        deploy_configmaps
    elif [ "$deployment" == "services" ]; then
        deploy_services
    elif [ "$deployment" == "monitoring" ]; then
        deploy_monitoring
    elif [ "$deployment" == "ingress" ]; then
        deploy_ingress
    fi

elif [ "$action" == "clean" ]; then
    cleanup
elif [ "$action" == "scale" ]; then
    scale

fi
