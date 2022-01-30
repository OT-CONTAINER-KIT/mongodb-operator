---
title: "MongoDB Setup"
weight: 3
description: >
    A detailed guide for designing the setup of MongoDB architecture
---

MongoDB is a NoSQL document database system that scales well horizontally and implements data storage through a key-value system. MongoDB can be setup in multiple mode like:-

- Standalone Mode
- Cluster replicated mode
- Cluster sharded mode

## MongoDB Standalone Setup

Just like any database mongodb also supports the standalone setup in which a single standalone instance is created and we setup MongoDB software on top of it. For small data chunks and development environment this setup can be ideal but in production grade environment this setup is not recommended because of the scalability and failover issues.

![](https://github.com/OT-CONTAINER-KIT/mongodb-operator/raw/docs-update/static/mongodb-standalone.png)

## MongoDB Replicated Setup

A replica set in MongoDB is a group of mongod processes that maintain the same data set. Replica sets provide redundancy and high availability, and are the basis for all production deployments.
These Mongod processes usually run on different nodes(machines) which together form a Replica set cluster.

![](https://github.com/OT-CONTAINER-KIT/mongodb-operator/raw/docs-update/static/mongodb-replicated.png)

## MongoDB Sharded Setup

Sharding is a method for distributing data across multiple machines. MongoDB uses sharding to support deployments with very large data sets and high throughput operations.

- Shard: Each shard contains a subset of the sharded data. Each shard can be deployed as a replica set to provide redundancy and high availability. Together, the cluster’s shards hold the entire data set for the cluster.
- Mongos: The mongos acts as a query router, providing an interface between client applications and the sharded cluster.
- Config Servers: Config servers store metadata and configuration settings for the cluster. They are also deployed as a replica set.

![](https://github.com/OT-CONTAINER-KIT/mongodb-operator/raw/docs-update/static/mongodb-sharded.png)
