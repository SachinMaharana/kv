cluster-up:
    kind create cluster --name {{cluster_name}} --image kindest/node:v1.19.1  --config ./kind-config.yaml
    sleep "10"
    kubectl wait --namespace kube-system --for=condition=ready pod --selector="tier=control-plane" --timeout=180s