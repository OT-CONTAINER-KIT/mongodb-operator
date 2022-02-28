---
title: "Installation"
weight: 2
linkTitle: "Installation"
description: >
    MongoDB Operator installation, upgrade guide
---

MongoDB operator is based on the CRD framework of Kubernetes, for more information about the CRD framework please refer to the [official documentation](https://kubernetes.io/docs/concepts/extend-kubernetes/api-extension/custom-resources/). In a nutshell, CRD is a feature through which we can develop our own custom API's inside Kubernetes.

The API versions for MongoDB Operator available are:-

- MongoDB
- MongoDBCluster

MongoDB Operator requires a Kubernetes cluster of version >=1.16.0. If you have just started with the CRD and Operators, its highly recommended using the latest version of Kubernetes.

Setup of MongoDB operator can be easily done by using simple [helm](https://helm.sh) and [kubectl](https://kubernetes.io/docs/reference/kubectl/overview/) commands.

{{< alert title="Note" >}}The recommded of way of installation is helm.{{< /alert >}}

## Operator Setup by Helm

The setup can be done by using helm. The mongodb-operator can easily get installed using helm commands.

```shell
# Add the helm chart
$ helm repo add ot-helm https://ot-container-kit.github.io/helm-charts/
...
"ot-helm" has been added to your repositories
```

```shell
# Deploy the MongoDB Operator
$ helm install mongodb-operator ot-helm/mongodb-operator \
  --namespace ot-operators
...
Release "mongodb-operator" does not exist. Installing it now.
NAME: mongodb-operator
LAST DEPLOYED: Sun Jan  9 23:05:13 2022
NAMESPACE: ot-operators
STATUS: deployed
REVISION: 1
```

Once the helm chart is deployed, we can test the status of operator pod by using:

```shell
# Testing Operator
$ helm test mongodb-operator --namespace ot-operators
...
NAME:           mongodb-operator
LAST DEPLOYED:  Sun Jan  9 23:05:13 2022
NAMESPACE:      ot-operators
STATUS:         deployed
REVISION:       1
TEST SUITE:     mongodb-operator-test-connection
Last Started:   Sun Jan  9 23:05:54 2022
Last Completed: Sun Jan  9 23:06:01 2022
Phase:          Succeeded
```

Verify the deployment of MongoDB Operator using `kubectl` command.

```shell
# List the pod and status of mongodb-operator
$ kubectl get pods -n ot-operators -l name=mongodb-operator
...
NAME                               READY   STATUS    RESTARTS   AGE
mongodb-operator-fc88b45b5-8rmtj   1/1     Running   0          21d
```

## Operator Setup by Kubectl

In any case using helm chart is not a possiblity, the MongoDB operator can be installed by `kubectl` commands as well.

As a first step, we need to setup a namespace and then deploy the CRD definitions inside Kubernetes.

```shell
# Setup of CRDs
$ kubectl create namespace ot-operators
$ kubectl apply -f https://github.com/OT-CONTAINER-KIT/mongodb-operator/raw/main/config/crd/bases/opstreelabs.in_mongodbs.yaml
$ kubectl apply -f https://github.com/OT-CONTAINER-KIT/mongodb-operator/raw/main/config/crd/bases/opstreelabs.in_mongodbclusters.yaml
```

Once we have namespace in the place, we need to setup the RBAC related stuff like:- **ClusterRoleBindings**, **ClusterRole**, **Serviceaccount**.

```shell
# Setup of RBAC account
$ kubectl apply -f https://raw.githubusercontent.com/OT-CONTAINER-KIT/mongodb-operator/main/config/rbac/service_account.yaml
$ kubectl apply -f https://raw.githubusercontent.com/OT-CONTAINER-KIT/mongodb-operator/main/config/rbac/role.yaml
$ kubectl apply -f https://github.com/OT-CONTAINER-KIT/mongodb-operator/blob/main/config/rbac/role_binding.yaml
```

As last part of the setup, now we can deploy the MongoDB Operator as deployment of Kubernetes.

```shell
# Deployment for MongoDB Operator
$ kubectl apply -f https://github.com/OT-CONTAINER-KIT/mongodb-operator/raw/main/config/manager/manager.yaml
```

Verify the deployment of MongoDB Operator using `kubectl` command.

```shell
# List the pod and status of mongodb-operator
$ kubectl get pods -n ot-operators -l name=mongodb-operator
...
NAME                               READY   STATUS    RESTARTS   AGE
mongodb-operator-fc88b45b5-8rmtj   1/1     Running   0          21d
```
