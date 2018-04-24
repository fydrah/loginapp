# Loginapp

[![Docker Repository on Quay](https://quay.io/repository/fydrah/loginapp/status "Docker Repository on Quay")](https://quay.io/repository/fydrah/loginapp)

**Simple login application for Kubernetes & Dex**

Original source code from [coreos/dex repository](https://github.com/coreos/dex/tree/master/cmd/example-app)

## Dockerfiles

* From scratch: ([scratch/Dockerfile](./dockerfiles/scratch/Dockerfile))
* Alpine 3.6: ([alpine/Dockerfile](./dockerfiles/alpine/Dockerfile))

The default image available [here](https://quay.io/fydrah/loginapp) is built from scratch.

## Usage

```shell
Usage:
  loginapp <config file> [flags]

Flags:
  -h, --help   help for loginapp
```

* Test:

```shell
# Update example/config.yml
# Test on local
docker run --rm -v $(pwd)/example/:/config/ quay.io/fydrah/loginapp:<tag> /config/config.yml
```

* Kubernetes:

A full example is available on [ObjectifLibre/k8s-ldap repository](https://github.com/ObjectifLibre/k8s-ldap)

* Helm:

[Helm chart](https://github.com/ObjectifLibre/k8s-ldap/tree/master/charts/k8s-ldap) is also available on ObjectifLibre/k8s-ldap repository.
