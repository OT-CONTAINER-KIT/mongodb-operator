---
title: "Standalone Setup"
weight: 3
linkTitle: "Standalone Setup"
description: >
    MongoDB database standalone setup guide
---

The MongoDB operator is capable for setting up MongoDB in the standalone mode with alot of additional power-ups like monitoring.

In this guide, we will see how we can set up MongoDB standalone with the help MongoDB operator and custom CRDS. We are going to use Kubernetes deployment tools like:- **[helm](https://helm.sh)** and **[kubectl](https://kubernetes.io/docs/reference/kubectl/overview/)**.

Standalone architecture:

![](https://github.com/OT-CONTAINER-KIT/mongodb-operator/raw/main/static/mongo-k8s-standalone.png)

## Setup using Helm Chart

Add the helm repository, so that MongoDB chart can be available for the installation. The repository can be added by:-

```shell
# Adding helm repository
$ helm repo add ot-helm https://ot-container-kit.github.io/helm-charts/
...
"ot-helm" has been added to your repositories
```

If the repository is added make sure you have updated it with the latest information.

```shell
# Updating ot-helm repository
$ helm repo update
```

Once all these things have completed, we can install MongoDB database by using:-

```shell
# Install the helm chart of MongoDB
$ helm install mongodb-ex --namespace ot-operators ot-helm/mongodb
...
NAME:          mongodb-ex
LAST DEPLOYED: Mon Jan 31 20:29:54 2022
NAMESPACE:     ot-operators
STATUS:        deployed
REVISION:      1
TEST SUITE:    None
NOTES:
  CHART NAME:    mongodb
  CHART VERSION: 0.1.0
  APP VERSION:   0.1.0

The helm chart for MongoDB standalone setup has been deployed.

Get the list of pods by executing:
    kubectl get pods --namespace ot-operators -l app=mongodb-ex-standalone

For getting the credential for admin user:
    kubectl get secrets -n ot-operators mongodb-ex-secret -o jsonpath="{.data.password}" | base64 -d
```

Verify the pod status and secret value by using:-

```shell
# Verify the status of the pods
$ kubectl get pods --namespace ot-operators -l app=mongodb-ex-standalone
...
NAME                      READY   STATUS    RESTARTS   AGE
mongodb-ex-standalone-0   2/2     Running   0          2m10s
```

```shell
# Verify the secret value
$ export PASSWORD=$(kubectl get secrets -n ot-operators mongodb-ex-secret -o jsonpath="{.data.password}" | base64 -d)
$ echo ${PASSWORD}
...
G7orFUuIrajGDK1iQzoD
```

## Setup using Kubectl Commands

It is not a recommended way for setting for MongoDB database, it can be used for the POC and learning of MongoDB operator deployment.

All the **kubectl** related manifest are located inside the **[examples]()** folder which can be applied using `kubectl apply -f`.

For an example:-

```shell
$ kubectl apply -f examples/basic/standalone.yaml -n ot-operators
```

## Validation of MongoDB Database

To validate the state of MongoDB database, we can take the shell access of the MongoDB pod.

```shell
# For getting the MongoDB container
$ kubectl exec -it mongodb-ex-standalone-0 -c mongo -n ot-operators -- bash
```

Execute mongodb ping command to check the health of MongoDB.

```shell
# MongoDB ping command
$ mongosh --eval "db.adminCommand('ping')"
...
Current Mongosh Log ID:	61f81de07cacf368e9a25322
Connecting to:		mongodb://127.0.0.1:27017/?directConnection=true&serverSelectionTimeoutMS=2000
Using MongoDB:		5.0.5
Using Mongosh:		1.1.7

{ ok: 1 }
```

We can also check the state of MongoDB by listing out the stats from admin database, but it would require username and password for same.

```shell
# MongoDB command for checking db stats
$ mongosh -u $MONGO_ROOT_USERNAME -p $MONGO_ROOT_PASSWORD --eval "db.stats()"
...
Current Mongosh Log ID:	61f820ab99bd3271e034d3b6
Connecting to:		mongodb://127.0.0.1:27017/?directConnection=true&serverSelectionTimeoutMS=2000
Using MongoDB:		5.0.5
Using Mongosh:		1.1.7
{
  db: 'test',
  collections: 0,
  views: 0,
  objects: 0,
  avgObjSize: 0,
  dataSize: 0,
  storageSize: 0,
  totalSize: 0,
  indexes: 0,
  indexSize: 0,
  scaleFactor: 1,
  fileSize: 0,
  fsUsedSize: 0,
  fsTotalSize: 0,
  ok: 1
}
```

Create a database inside MongoDB database and try to insert some data inside it.

```shell
# create database inside MongoDB
$ use validatedb

$ db.user.insert({name: "Abhishek Dubey", age: 24})
$ db.user.insert({name: "Sajal Jain", age: 32})
...
WriteResult({ "nInserted" : 1 })
```

Let's try to get out information from `validatedb` to see if write operation is successful or not.

```shell
$ db.user.find().pretty();
...
{
	"_id" : ObjectId("61fbe07d052a532fa6842a00"),
	"name" : "Abhishek Dubey",
	"age" : 24
}
{
	"_id" : ObjectId("61fbe1b7052a532fa6842a01"),
	"name" : "Sajal Jain",
	"age" : 32
}
```