---
title: "MongoDB Replicated Cluster"
weight: 3
linkTitle: "MongoDB Replicated Cluster"
description: >
    MongoDB replicated cluster configuration for CRD and helm chart
---

MongoDB cluster configuration is easily customizable using `helm` as well `kubectl`. Since all the configurations are in the form YAML file, it can be easily changed and customized.

The `values.yaml` file for MongoDB cluster setup can be found [here](https://github.com/OT-CONTAINER-KIT/helm-charts/tree/main/charts/mongodb-cluster). But if the setup is not done using `Helm`, in that case Kubernetes manifests needs to be customized.

## Parameters for Helm Chart

| **Name**                                  |        **Value**         | **Description**                                   |
|-------------------------------------------|:------------------------:|---------------------------------------------------|
| `clusterSize`                             |            3             | Size of the MongoDB cluster                       |
| `image.name`                              |  quay.io/opstree/mongo   | Name of the MongoDB image                         |
| `image.tag`                               |           v5.0           | Tag for the MongoDB image                         |
| `image.imagePullPolicy`                   |       IfNotPresent       | Image Pull Policy of the MongoDB                  |
| `resources`                               |            {}            | Request and limits for MongoDB statefulset        |
| `storage.enabled`                         |           true           | Storage is enabled for MongoDB or not             |
| `storage.accessModes`                     |    ["ReadWriteOnce"]     | AccessMode for storage provider                   |
| `storage.storageSize`                     |           1Gi            | Size of storage for MongoDB                       |
| `storage.storageClass`                    |           gp2            | Name of the storageClass to create storage        |
| `mongoDBMonitoring.enabled`               |           true           | MongoDB exporter should be deployed or not        |
| `mongoDBMonitoring.image.name`            | bitnami/mongodb-exporter | Name of the MongoDB exporter image                |
| `mongoDBMonitoring.image.tag`             |  0.11.2-debian-10-r382   | Tag of the MongoDB exporter image                 |
| `mongoDBMonitoring.image.imagePullPolicy` |       IfNotPresent       | Image Pull Policy of the MongoDB exporter image   |
| `serviceMonitor.enabled`                  |          false           | Servicemonitor to monitor MongoDB with Prometheus |
| `serviceMonitor.interval`                 |           30s            | Interval at which metrics should be scraped.      |
| `serviceMonitor.scrapeTimeout`            |           10s            | Timeout after which the scrape is ended           |
| `serviceMonitor.namespace`                |        monitoring        | Namespace in which Prometheus operator is running |

## Parameters for CRD Object Definition

These are the parameters that are currently supported by the MongoDB operator for the cluster MongoDB database setup:-

- clusterSize
- kubernetesConfig
- storage
- mongoDBSecurity
- mongoDBMonitoring

### clusterSize

`clusterSize` is the size of MongoDB replicated cluster. We have to provide the number of node count that we want to make part of MongoDB cluster. For example:- 1 primary and 2 secondary is 3 as pod count.

```yaml
  clusterSize: 3
```

### kubernetesConfig

`kubernetesConfig` is the general configuration paramater for MongoDB CRD in which we are defining the Kubernetes related configuration details like- image, tag, imagePullPolicy, and resources.

```yaml
  kubernetesConfig:
    image: quay.io/opstree/mongo:v5.0
    imagePullPolicy: IfNotPresent
    resources:
      requests:
        cpu: 1
        memory: 8Gi
      limits:
        cpu: 1
        memory: 8Gi
```

### storage

`storage` is the storage specific configuration for MongoDB CRD. With this parameter we can make enable persistence inside the MongoDB statefulset. In this parameter, we will provide inputs like- accessModes, size of the storage, and storageClass.

```yaml
  storage:
    accessModes: ["ReadWriteOnce"]
    storageSize: 1Gi
    storageClass: csi-cephfs-sc
```

### mongoDBSecurity

`mongoDBSecurity` is the security specification for MongoDB CRD. If we want to enable our MongoDB database authenticated, in that case, we can enable this configuration. To enable the authentication we need to provide paramaters like- admin username, secret reference in Kubernetes.

```yaml
  mongoDBSecurity:
    mongoDBAdminUser: admin
    secretRef:
      name: mongodb-secret
      key: password
```

### mongoDBMonitoring

`mongoDBMonitoring` is the monitoring feature for MongoDB CRD. By using this parameter we can enable the MongoDB monitoring using **[MongoDB Exporter](https://github.com/percona/mongodb_exporter)**. In this parameter, we need to provide image, imagePullPolicy and resources for mongodb exporter.

```yaml
  mongoDBMonitoring:
    enableExporter: true
    image: bitnami/mongodb-exporter:0.11.2-debian-10-r382
    imagePullPolicy: IfNotPresent
    resources: {}
```
