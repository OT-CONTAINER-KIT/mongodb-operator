---
title: "Failover Testing"
weight: 5
linkTitle: "Failover Testing"
description: >
    Failover testing for the MongoDB replicated setup
---

In this section, we will deactivate/delete the specific nodes (mainly primary node) to force the replica set to make an election and select a new primary node.

Before deleting/deactivating the primary node, we will try to put some dummy data inside it.

```shell
# Get inside the primary node pod
$ kubectl exec -it mongodb-ex-cluster-cluster-0 -n ot-operators -- bash

# Login inside the MongoDB shell
$ mongo -u $MONGO_ROOT_USERNAME -p $MONGO_ROOT_PASSWORD
```

Once we are inside the MongoDB primary pod, we will create a db called `failtestdb` and create a collection inside it with some dummy data.

```shell
# Creation of the MongoDB database
$ use failtestdb

# Collection creation inside MongoDB database
$ db.user.insert({name: "Abhishek Dubey", age: 24})
...
WriteResult({ "nInserted" : 1 })
```

Let's check if the data is written properly or not inside the primary node.

```shell
# Find the data available inside the collection
$ db.user.find().pretty();
...
{
	"_id" : ObjectId("61fc0e19a63b3bfa4f30d5e2"),
	"name" : "Abhishek Dubey",
	"age" : 24
}
```

## Deactivating/Deleting MongoDB Primary and Secondary

Once we have established that the database and content is written properly inside the MongoDB database. Now let's try to delete the primary pod to see what is the impact of it.

```shell
# Delete the primary pod
$ kubectl delete pod mongodb-ex-cluster-cluster-0 -n ot-operators
...
pod "mongodb-ex-cluster-cluster-0" deleted
```

When the pod is up and running, get inside the same pod and see the role assigned to that MongoDB node. Also, the replication status of the cluster.

```shell
# Login again inside the primary pod
$ kubectl exec -it mongodb-ex-cluster-cluster-0 -n ot-operators -- bash

# Login inside the MongoDB shell
$ mongo -u $MONGO_ROOT_USERNAME -p $MONGO_ROOT_PASSWORD
```

```shell
# Check the mongo node is still primary or not
$ db.hello()
...
{
	"topologyVersion" : {
		"processId" : ObjectId("61fc125bbb845ecebf0b6c1e"),
		"counter" : NumberLong(4)
	},
	"hosts" : [
		"mongodb-ex-cluster-cluster-0.mongodb-ex-cluster-cluster.ot-operators:27017",
		"mongodb-ex-cluster-cluster-1.mongodb-ex-cluster-cluster.ot-operators:27017",
		"mongodb-ex-cluster-cluster-2.mongodb-ex-cluster-cluster.ot-operators:27017"
	],
	"setName" : "mongodb-ex-cluster",
	"setVersion" : 1,
	"isWritablePrimary" : false,
	"secondary" : true,
	"primary" : "mongodb-ex-cluster-cluster-1.mongodb-ex-cluster-cluster.ot-operators:27017",
	}
}
```

As we can see in the above command, output this pod has become the secondary pod and `mongodb-ex-cluster-cluster-1` has become the primary pod of MongoDB. Also, this pod is now non-writable state. Let's check if we can fetch out the collection data from MongoDB database.

```shell
# Change the database to failtestdb
$ use failtestdb

# Set the slave read flag of MongoDB
$ secondaryOk()

# Find the data available inside the collection
$ db.user.find().pretty();
...
{
	"_id" : ObjectId("61fc0e19a63b3bfa4f30d5e2"),
	"name" : "Abhishek Dubey",
	"age" : 24
}
```