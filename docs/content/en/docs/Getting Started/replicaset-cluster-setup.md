---
title: "Replication Setup"
weight: 4
linkTitle: "Replication Setup"
description: >
    MongoDB database replication setup guide
---

If we are running our application inside the production environment, in that case we should always go with HA architecture. MongoDB operator can also set up MongoDB database in the replication mode where there can be one primary instance and many secondary instances.

Architecture:-

![](https://github.com/OT-CONTAINER-KIT/mongodb-operator/raw/main/static/mongodb-k8s-cluster.png)

In this guide, we will see how we can set up MongoDB replicated cluster using MongoDB operator and custom CRDS.

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

Once all these things have completed, we can install MongoDB cluster database by using:-

```shell
# Installation of MongoDB replication cluster
$ helm install mongodb-ex-cluster --namespace ot-operators ot-helm/mongodb-cluster
...
NAME:          mongodb-ex-cluster
LAST DEPLOYED: Tue Feb  1 23:18:36 2022
NAMESPACE:     ot-operators
STATUS:        deployed
REVISION:      1
TEST SUITE:    None
NOTES:
  CHART NAME:    mongodb-cluster
  CHART VERSION: 0.1.0
  APP VERSION:   0.1.0

The helm chart for MongoDB standalone setup has been deployed.

Get the list of pods by executing:
    kubectl get pods --namespace ot-operators -l app=mongodb-ex-cluster-cluster

For getting the credential for admin user:
    kubectl get secrets -n ot-operators mongodb-ex-secret -o jsonpath="{.data.password}" | base64 -d
```

Verify the pod status of mongodb database cluster and secret value by using:-

```shell
# Verify the status of the mongodb cluster pods
$ kubectl get pods --namespace ot-operators -l app=mongodb-ex-cluster-cluster
...
NAME                           READY   STATUS    RESTARTS   AGE
mongodb-ex-cluster-cluster-0   2/2     Running   0          5m57s
mongodb-ex-cluster-cluster-1   2/2     Running   0          5m28s
mongodb-ex-cluster-cluster-2   2/2     Running   0          4m48s
```

```shell
# Verify the secret value
$ export PASSWORD=$(kubectl get secrets -n ot-operators mongodb-ex-cluster-secret -o jsonpath="{.data.password}" | base64 -d)
$ echo ${PASSWORD}
...
fEr9FScI9ojh6LSh2meK
```

## Setup using Kubectl Commands

It is not a recommended way for setting for MongoDB database, it can be used for the POC and learning of MongoDB operator deployment.

All the **kubectl** related manifest are located inside the **[examples]()** folder which can be applied using `kubectl apply -f`.

For an example:-

```shell
$ kubectl apply -f examples/basic/clusterd.yaml -n ot-operators
```

## Validation of MongoDB Cluster

Once the cluster is created and in `running` state, we should verify the health of the MongoDB cluster.

```shell
# Verifying the health of the cluster
$ kubectl exec -it mongodb-ex-cluster-cluster-0 -n ot-operators -- bash

$ mongo -u $MONGO_ROOT_USERNAME -p $MONGO_ROOT_PASSWORD --eval "db.adminCommand( { replSetGetStatus: 1 } )"
...
{
	"set" : "mongodb-ex-cluster",
	"date" : ISODate("2022-02-03T08:52:09.257Z"),
	"myState" : 1,
	"term" : NumberLong(1),
	"syncSourceHost" : "",
	"syncSourceId" : -1,
	"heartbeatIntervalMillis" : NumberLong(2000),
	"majorityVoteCount" : 2,
	"writeMajorityCount" : 2,
	"votingMembersCount" : 3,
	"writableVotingMembersCount" : 3,
	"optimes" : {
		"lastCommittedOpTime" : {
			"ts" : Timestamp(1643878322, 1),
			"t" : NumberLong(1)
		},
		"lastCommittedWallTime" : ISODate("2022-02-03T08:52:02.111Z"),
		"readConcernMajorityOpTime" : {
			"ts" : Timestamp(1643878322, 1),
			"t" : NumberLong(1)
		},
		"appliedOpTime" : {
			"ts" : Timestamp(1643878322, 1),
			"t" : NumberLong(1)
		},
		"durableOpTime" : {
			"ts" : Timestamp(1643878322, 1),
			"t" : NumberLong(1)
		},
		"lastAppliedWallTime" : ISODate("2022-02-03T08:52:02.111Z"),
		"lastDurableWallTime" : ISODate("2022-02-03T08:52:02.111Z")
	},
	"lastStableRecoveryTimestamp" : Timestamp(1643878312, 1),
	"electionCandidateMetrics" : {
		"lastElectionReason" : "electionTimeout",
		"lastElectionDate" : ISODate("2022-02-01T17:50:49.975Z"),
		"electionTerm" : NumberLong(1),
		"lastCommittedOpTimeAtElection" : {
			"ts" : Timestamp(1643737839, 1),
			"t" : NumberLong(-1)
		},
		"lastSeenOpTimeAtElection" : {
			"ts" : Timestamp(1643737839, 1),
			"t" : NumberLong(-1)
		},
		"numVotesNeeded" : 2,
		"priorityAtElection" : 1,
		"electionTimeoutMillis" : NumberLong(10000),
		"numCatchUpOps" : NumberLong(0),
		"newTermStartDate" : ISODate("2022-02-01T17:50:50.072Z"),
		"wMajorityWriteAvailabilityDate" : ISODate("2022-02-01T17:50:50.926Z")
	},
	"members" : [
		{
			"_id" : 0,
			"name" : "mongodb-ex-cluster-cluster-0.mongodb-ex-cluster-cluster.ot-operators:27017",
			"health" : 1,
			"state" : 1,
			"stateStr" : "PRIMARY",
			"uptime" : 140603,
			"optime" : {
				"ts" : Timestamp(1643878322, 1),
				"t" : NumberLong(1)
			},
			"optimeDate" : ISODate("2022-02-03T08:52:02Z"),
			"lastAppliedWallTime" : ISODate("2022-02-03T08:52:02.111Z"),
			"lastDurableWallTime" : ISODate("2022-02-03T08:52:02.111Z"),
			"syncSourceHost" : "",
			"syncSourceId" : -1,
			"infoMessage" : "",
			"electionTime" : Timestamp(1643737849, 1),
			"electionDate" : ISODate("2022-02-01T17:50:49Z"),
			"configVersion" : 1,
			"configTerm" : 1,
			"self" : true,
			"lastHeartbeatMessage" : ""
		},
		{
			"_id" : 1,
			"name" : "mongodb-ex-cluster-cluster-1.mongodb-ex-cluster-cluster.ot-operators:27017",
			"health" : 1,
			"state" : 2,
			"stateStr" : "SECONDARY",
			"uptime" : 140490,
			"optime" : {
				"ts" : Timestamp(1643878322, 1),
				"t" : NumberLong(1)
			},
			"optimeDurable" : {
				"ts" : Timestamp(1643878322, 1),
				"t" : NumberLong(1)
			},
			"optimeDate" : ISODate("2022-02-03T08:52:02Z"),
			"optimeDurableDate" : ISODate("2022-02-03T08:52:02Z"),
			"lastAppliedWallTime" : ISODate("2022-02-03T08:52:02.111Z"),
			"lastDurableWallTime" : ISODate("2022-02-03T08:52:02.111Z"),
			"lastHeartbeat" : ISODate("2022-02-03T08:52:08.021Z"),
			"lastHeartbeatRecv" : ISODate("2022-02-03T08:52:07.926Z"),
			"pingMs" : NumberLong(65),
			"lastHeartbeatMessage" : "",
			"syncSourceHost" : "mongodb-ex-cluster-cluster-2.mongodb-ex-cluster-cluster.ot-operators:27017",
			"syncSourceId" : 2,
			"infoMessage" : "",
			"configVersion" : 1,
			"configTerm" : 1
		},
		{
			"_id" : 2,
			"name" : "mongodb-ex-cluster-cluster-2.mongodb-ex-cluster-cluster.ot-operators:27017",
			"health" : 1,
			"state" : 2,
			"stateStr" : "SECONDARY",
			"uptime" : 140490,
			"optime" : {
				"ts" : Timestamp(1643878322, 1),
				"t" : NumberLong(1)
			},
			"optimeDurable" : {
				"ts" : Timestamp(1643878322, 1),
				"t" : NumberLong(1)
			},
			"optimeDate" : ISODate("2022-02-03T08:52:02Z"),
			"optimeDurableDate" : ISODate("2022-02-03T08:52:02Z"),
			"lastAppliedWallTime" : ISODate("2022-02-03T08:52:02.111Z"),
			"lastDurableWallTime" : ISODate("2022-02-03T08:52:02.111Z"),
			"lastHeartbeat" : ISODate("2022-02-03T08:52:09.022Z"),
			"lastHeartbeatRecv" : ISODate("2022-02-03T08:52:07.475Z"),
			"pingMs" : NumberLong(1),
			"lastHeartbeatMessage" : "",
			"syncSourceHost" : "mongodb-ex-cluster-cluster-0.mongodb-ex-cluster-cluster.ot-operators:27017",
			"syncSourceId" : 0,
			"infoMessage" : "",
			"configVersion" : 1,
			"configTerm" : 1
		}
	],
	"ok" : 1,
	"$clusterTime" : {
		"clusterTime" : Timestamp(1643878322, 1),
		"signature" : {
			"hash" : BinData(0,"M4Mj+BFC8//nI2QIT+Hic/N6N4g="),
			"keyId" : NumberLong("7059800308947353604")
		}
	},
	"operationTime" : Timestamp(1643878322, 1)
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

{{< alert title="Note" >}}Make sure you perform this operation on Secondary node{{< /alert >}}

Let's try to get out information from `validatedb` to see if write operation is successful or not. 

```shell
# Get inside the secondary pod
$ kubectl exec -it mongodb-ex-cluster-cluster-1 -n ot-operators -- bash
```

```shell
# Login inside the MongoDB shell
$ mongo -u $MONGO_ROOT_USERNAME -p $MONGO_ROOT_PASSWORD
```

List out the collection data on MongoDB secondary node by using mongo shell commands.

```shell
$ secondaryOk()
$ db.user.find().pretty();
...
{
	"_id" : ObjectId("61fbe23d5b403e9fb63bc374"),
	"name" : "Abhishek Dubey",
	"age" : 24
}
{
	"_id" : ObjectId("61fbe2465b403e9fb63bc375"),
	"name" : "Sajal Jain",
	"age" : 32
}
```

As we can see, we are able to list out the data which have been written inside the primary node is also queryable from the secondary nodes as well.
