# Loginapp

**Simple login application for Kubernetes & Dex**

Original source code from https://github.com/coreos/dex/tree/master/cmd/example-app

## Docker tags

* From scratch tags: **latest**, **VERSION** ([scratch/Dockerfile](./dockerfiles/scratch/Dockerfile))
* Alpine 3.6 tags: **alpine**, **VERSION-alpine** ([alpine/Dockerfile](./dockerfiles/alpine/Dockerfile))

## Usage

```
Usage:
  loginapp <config file> [flags]

Flags:
  -h, --help   help for loginapp
```

* Test:

```
# Update example/config.yml
docker run --rm -v $(pwd)/example/:/config/ fhardy/loginapp /config/config.yml
```

* Kubernetes:

A full example is available on [ObjectifLibre/k8s-ldap repository](https://github.com/ObjectifLibre/k8s-ldap)

* Helm:

[Helm chart](https://github.com/ObjectifLibre/k8s-ldap/tree/master/charts/k8s-ldap) is also available on ObjectifLibre/k8s-ldap repository.
