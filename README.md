# KV

## Install Just

```
{
    curl --proto '=https' --tlsv1.2 -sSf https://just.systems/install.sh | bash -s -- --to ~/
    mv ./just /usr/local/bin
    just --verision
}
```

```
kind create cluster --config kind-config.yaml --name toggle

kubectl apply -f https://docs.projectcalico.org/v3.8/manifests/calico.yaml

helm install prom prometheus-community/prometheus -f prometheus/values.yaml
```

```
kubectl create secret generic datasource --from-file=datasource.yaml
kubectl create configmap mydashboard --from-file=grafana.json
helm install graff bitnami/grafana -f grafana/values.yaml
kubectl port-forward svc/graff-grafana 8080:3000 &
echo "Password: $(kubectl get secret graff-grafana-admin --namespace default -o jsonpath="{.data.GF_SECURITY_ADMIN_PASSWORD}" | base64 --decode)"

```

```
helm install redis-db bitnami/redis

k apply -f deployment.yaml
```

```
kubectl apply -f https://raw.githubusercontent.com/containous/traefik/v1.7/examples/k8s/traefik-rbac.yaml

kubectl apply -f https://raw.githubusercontent.com/containous/traefik/v1.7/examples/k8s/traefik-ds.yaml

kubectl apply -f traefik-service.yaml

kubectl apply -f ingress.yaml
```

```
for i in {6..10}; do curl -s -H "Host: kv-api.com" -i -X POST -H "Content-Type: application/json" -d "{\"key\":\"xy-$i\", \"value\":\"val-$i\"}" http://localhost:30100/set; done
```
