# KV

A task runner is provided to manage this repo.

### Install Just task Runner

```
{
    curl --proto '=https' --tlsv1.2 -sSf https://just.systems/install.sh | bash -s -- --to ~/
    sudo mv ~/just /usr/local/bin
    just --version
}
```

### Steps

_This was tested in an ubuntu box(18.04) ec2 instance with port 9000 exposed for grafana._

```
# installs docker
just docker
```

```
# istalls a local k8s cluster
just cluster

just cluster-up

# installs helm repos required
just helm

# setups prometheus
just prom

# setups grafana with preconfigured datasource and dashboards
just grafana

# setups the api
just app

# installs and setups the ingress
just ingress

# adds key-value into db
just seed

# access grafana using public ip
just grafana-access
```
