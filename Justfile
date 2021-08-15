cluster_name := "toggle"

docker:
    sudo apt update
    sudo apt install -y apt-transport-https ca-certificates curl gnupg-agent software-properties-common
    curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo gpg --dearmor -o /usr/share/keyrings/docker-archive-keyring.gpg
    echo \
    "deb [arch=amd64 signed-by=/usr/share/keyrings/docker-archive-keyring.gpg] https://download.docker.com/linux/ubuntu \
    $(lsb_release -cs) stable" | sudo tee /etc/apt/sources.list.d/docker.list > /dev/null
    sudo apt update
    sudo apt install -y docker-ce docker-ce-cli containerd.io
    sudo usermod -aG docker ${USER}
    sudo systemctl restart docker
    newgrp docker

cluster:
    curl -Lo ./kind https://kind.sigs.k8s.io/dl/v0.11.1/kind-linux-amd64
    chmod +x ./kind
    sudo mv ./kind /usr/local/bin

cluster-up:
    sudo curl -fsSLo /usr/share/keyrings/kubernetes-archive-keyring.gpg https://packages.cloud.google.com/apt/doc/apt-key.gpg
    echo "deb [signed-by=/usr/share/keyrings/kubernetes-archive-keyring.gpg] https://apt.kubernetes.io/ kubernetes-xenial main" | sudo tee /etc/apt/sources.list.d/kubernetes.list
    sudo apt-get update
    sudo apt-get install -y kubectl
    -kind create cluster --name {{cluster_name}} --image kindest/node:v1.19.1  --config ./infra/kind-config.yaml
    sleep "10"
    kubectl wait --namespace kube-system --for=condition=ready pod --selector="tier=control-plane" --timeout=180s
    kubectl apply -f https://docs.projectcalico.org/v3.8/manifests/calico.yaml
    sleep "10"
    kubectl wait --namespace kube-system --for=condition=ready pod --selector="k8s-app=calico-node" --timeout=180s
    
helm:
    curl https://baltocdn.com/helm/signing.asc | sudo apt-key add -
    sudo apt-get install apt-transport-https --yes
    echo "deb https://baltocdn.com/helm/stable/debian/ all main" | sudo tee /etc/apt/sources.list.d/helm-stable-debian.list
    sudo apt-get update
    sudo apt-get install helm

    @echo Adding Required Repos
    helm repo add bitnami https://charts.bitnami.com/bitnami
    helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
    helm repo update

prom:
    helm install prom prometheus-community/prometheus -f infra/prometheus/values.yaml

grafana:
    -kubectl create secret generic datasource --from-file=infra/datasource.yaml
    -kubectl create configmap mydashboard --from-file=infra/grafana.json
    -helm install graff bitnami/grafana -f infra/grafana/values.yaml

app:
    -helm install --set replica.replicaCount=1 redis-db bitnami/redis 
    sleep "5"
    kubectl wait --namespace default --for=condition=ready pod --selector="app.kubernetes.io/instance=redis-db" --timeout=180s
    -kubectl apply -f infra/deployment.yaml
    sleep "5"
    kubectl wait --namespace default --for=condition=ready pod --selector="app=kv" --timeout=180s
    
ingress:
    kubectl apply -f https://raw.githubusercontent.com/containous/traefik/v1.7/examples/k8s/traefik-rbac.yaml
    kubectl apply -f https://raw.githubusercontent.com/containous/traefik/v1.7/examples/k8s/traefik-ds.yaml
    sleep "5"
    kubectl wait --namespace kube-system --for=condition=ready pod --selector="k8s-app=traefik-ingress-lb" --timeout=180s
    kubectl apply -f infra/traefik-service.yaml
    kubectl apply -f infra/ingress.yaml

seed:
   #!/bin/bash
   for i in {1..50}; do curl -s -H "Host: kv-api.com" -i -X POST -H "Content-Type: application/json" -d "{\"key\":\"xyz-$i\", \"value\":\"val-$i\"}" http://localhost:30100/set; done
#  curl -s -H "Host: kv-api.com" http://localhost:30100/get/xyz-2  
#  curl -s -H "Host: kv-api.com" 'http://localhost:30100/search?suffix=-2'

    
grafana-access:
    #!/bin/sh
    kubectl port-forward --address 0.0.0.0 svc/graff-grafana 9000:3000 > /dev/null 2>&1 &
    echo "Grafana is at:PUBLIC_IP:9000"
    curl icanhazip.com
    echo "User: admin , Password: $(kubectl get secret graff-grafana-admin --namespace default -o jsonpath="{.data.GF_SECURITY_ADMIN_PASSWORD}" | base64 --decode)"

# run docker before as it requires relogin
all: cluster cluster-up helm prom grafana app ingress seed grafana-access



# hey -n 10 -c 2 -m POST -T "application/json" -d "{\"key\":\"xy-$i\", \"value\":\"val-$i\"}" http://localhost:30100/set
# hey -n 10 -c 2 -m POST -T "application/x-www-form-urlencoded" -d 'username=1&message=hello' http://your-rest-url/resource
# loadtest -c 10 --rps 4 -H "Host: kv-api.com" http://localhost:30100/get/xyz-6
# loadtest -c 10 --rps 200 -T "application/json" -m POST -d '{\"key\":\"xy-$i\", \"value\":\"val-$i\"}' -H "Host: kv-api.com" http://localhost:30100/set