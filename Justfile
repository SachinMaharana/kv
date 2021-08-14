cluster_name := "toggle"

docker:
    apt update
    apt install -y apt-transport-https ca-certificates curl gnupg-agent software-properties-common
    add-apt-repository "deb [arch=amd64] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable"
    curl -fsSL https://download.docker.com/linux/ubuntu/gpg | apt-key add -
    apt update
    apt install -y docker-ce docker-ce-cli containerd.io

cluster-up:
    curl -fsSLo /usr/share/keyrings/kubernetes-archive-keyring.gpg https://packages.cloud.google.com/apt/doc/apt-key.gpg
    echo "deb [signed-by=/usr/share/keyrings/kubernetes-archive-keyring.gpg] https://apt.kubernetes.io/ kubernetes-xenial main" | sudo tee /etc/apt/sources.list.d/kubernetes.list
    apt-get update
    apt-get install -y kubectl
    -kind create cluster --name {{cluster_name}} --image kindest/node:v1.19.1  --config ./kind-config.yaml
    sleep "10"
    kubectl wait --namespace kube-system --for=condition=ready pod --selector="tier=control-plane" --timeout=180s
    kubectl apply -f https://docs.projectcalico.org/v3.8/manifests/calico.yaml
    kubectl wait --namespace kube-system --for=condition=ready pod --selector="k8s-app=calico-node" --timeout=180s
    
helm:
    curl https://baltocdn.com/helm/signing.asc | sudo apt-key add -
    apt-get install apt-transport-https --yes
    echo "deb https://baltocdn.com/helm/stable/debian/ all main" | tee /etc/apt/sources.list.d/helm-stable-debian.list
    apt-get update
    apt-get install helm

    @echo Adding Repos
    helm repo add bitnami https://charts.bitnami.com/bitnami
    helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
    helm repo update

prom:
    -helm install prom prometheus-community/prometheus -f prometheus/values.yaml

grafana:
    -kubectl create secret generic datasource --from-file=datasource.yaml
    -kubectl create configmap mydashboard --from-file=grafana.json
    -helm install graff bitnami/grafana -f grafana/values.yaml
    echo "User: admin , Password: $(kubectl get secret graff-grafana-admin --namespace default -o jsonpath="{.data.GF_SECURITY_ADMIN_PASSWORD}" | base64 --decode)"
    @echo enabling port-forward
    -kubectl port-forward --address 0.0.0.0 svc/graff-grafana 8080:3000 > 2>&1 &


app:
    -helm install redis-db bitnami/redis
    kubectl wait --namespace default --for=condition=ready pod --selector="app.kubernetes.io/instance=redis-db" --timeout=180s
    -kubectl apply -f deployment.yaml
    kubectl wait --namespace default --for=condition=ready pod --selector="app=kv" --timeout=180s
    
ingress:
    kubectl apply -f https://raw.githubusercontent.com/containous/traefik/v1.7/examples/k8s/traefik-rbac.yaml
    kubectl apply -f https://raw.githubusercontent.com/containous/traefik/v1.7/examples/k8s/traefik-ds.yaml
    kubectl wait --namespace kube-system --for=condition=ready pod --selector="k8s-app=traefik-ingress-lb" --timeout=180s
    kubectl apply -f traefik-service.yaml
    kubectl apply -f ingress.yaml


seed:
   #!/bin/bash
   for i in {6..10}; do curl -s -H "Host: kv-api.com" -i -X POST -H "Content-Type: application/json" -d "{\"key\":\"xy-$i\", \"value\":\"val-$i\"}" http://localhost:30100/set; done

    
risk-it-all: docker cluster-up helm prom grafana app ingress seed




