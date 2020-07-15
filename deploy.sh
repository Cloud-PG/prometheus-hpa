kubectl apply -f manifests_no_configs/httpgo_and_exporter.yaml --validate=false
kubectl apply -f manifests_no_configs/httpd_and_exporter.yaml --validate=false
kubectl apply -f manifests_no_configs/couchdb_and_exporter.yaml --validate=false
kubectl create namespace monitoring
kubectl create configmap prometheus-example-cm --from-file configs/prometheus.yml -n monitoring
kubectl create namespace monitoring
kubectl apply -f manifests_no_configs/prometheus.yaml
kubectl create configmap prometheus-adapter-example-cm --from-file configs/prometheus_adapter.yml
kubectl apply -f manifests_no_configs/prometheus_adapter.yaml
kubectl apply -f manifests_no_configs/hpa_httpgo.yaml
kubectl apply -f manifests_no_configs/hpa_httpd.yaml
kubectl apply -f manifests_no_configs/hpa_couchdb.yaml
