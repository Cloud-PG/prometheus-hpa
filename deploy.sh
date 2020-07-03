kubectl apply -f httpgo_and_exporter.yaml
kubectl apply -f httpd_and_exporter.yaml
kubectl apply -f couchdb_and_exporter.yaml
kubectl apply -f prometheus.yaml
kubectl apply -f prometheus_adapter.yaml
kubectl apply -f hpa_httpgo.yaml
kubectl apply -f hpa_httpd.yaml
kubectl apply -f hpa_couchdb.yaml
