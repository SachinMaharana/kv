cluster_name := "toggle"

install-docker:
    apt install && apt-transport-https ca-certificates curl gnupg-agent software-properties-common
    add-apt-repository "deb [arch=amd64] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable"
    curl -fsSL https://download.docker.com/linux/ubuntu/gpg | apt-key add -
    apt update
    apt install docker-ce docker-ce-cli containerd.io

cluster-up:
    kind create cluster --name {{cluster_name}} --image kindest/node:v1.19.1  --config ./kind-config.yaml
    sleep "10"
    kubectl wait --namespace kube-system --for=condition=ready pod --selector="tier=control-plane" --timeout=180s