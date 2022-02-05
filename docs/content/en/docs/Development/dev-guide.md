---
title: "Development Guide"
weight: 2
linkTitle: "Development Guide"
description: >
   Development guide for MongoDB Operator
---

## Pre-requisites

**Access to Kubernetes cluster**

First, you will need access to a Kubernetes cluster. The easiest way to start is minikube.

- [Virtualbox](https://www.virtualbox.org/wiki/Downloads)
- [Minikube](https://kubernetes.io/docs/setup/minikube/)
- [Kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/)

**Tools to build an Operator**

Apart from kubernetes cluster, there are some tools which are needed to build and test the MongoDB Operator.

- [Git](https://git-scm.com/downloads)
- [Go](https://golang.org/dl/)
- [Docker](https://docs.docker.com/install/)
- [Operator SDK](https://github.com/operator-framework/operator-sdk/blob/v0.8.1/doc/user/install-operator-sdk.md)
- [Make](https://www.gnu.org/software/make/manual/make.html)

## Building Operator

To build the operator on local system, we can use `make` command.

```shell
$ make manager
...
go build -o bin/manager main.go
```

MongoDB operator gets packaged as a container image for running on the Kubernetes cluster.

```shell
$ make docker-build
```

If you want to play it on Kubernetes. You can use a minikube.

```shell
$ minikube start --vm-driver virtualbox
...
😄  minikube v1.0.1 on linux (amd64)
🤹  Downloading Kubernetes v1.14.1 images in the background ...
🔥  Creating kvm2 VM (CPUs=2, Memory=2048MB, Disk=20000MB) ...
📶  "minikube" IP address is 192.168.39.240
🐳  Configuring Docker as the container runtime ...
🐳  Version of container runtime is 18.06.3-ce
⌛  Waiting for image downloads to complete ...
✨  Preparing Kubernetes environment ...
🚜  Pulling images required by Kubernetes v1.14.1 ...
🚀  Launching Kubernetes v1.14.1 using kubeadm ... 
⌛  Waiting for pods: apiserver proxy etcd scheduler controller dns
🔑  Configuring cluster permissions ...
🤔  Verifying component health .....
💗  kubectl is now configured to use "minikube"
🏄  Done! Thank you for using minikube!
```

```shell
$ make test
```