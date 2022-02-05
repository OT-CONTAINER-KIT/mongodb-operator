---
title: "Prometheus Monitoring"
weight: 2
linkTitle: "Prometheus Monitoring"
description: >
    Monitoring of MongoDB standalone and replicaset cluster using Prometheus
---

In MongoDB Operator, we are using [mongodb-exporter](https://github.com/percona/mongodb_exporter) to collect the stats, metrics for MongoDB database. This exporter is capable to capture the stats for standalone and cluster mode of MongoDB.

If we are using the `helm` chart for installation purpose, we can simply enable this configuration inside the [values.yaml](https://github.com/OT-CONTAINER-KIT/helm-charts/blob/main/charts/mongodb-cluster/values.yaml). 

```yaml
mongoDBMonitoring:
  enabled: true
  image:
    name: bitnami/mongodb-exporter
    tag: 0.11.2-debian-10-r382
    imagePullPolicy: IfNotPresent
  resources: {}
```

In case of `kubectl` installation, we can add a code snippet in yaml manifest like this:-

```yaml
  mongoDBMonitoring:
    enableExporter: true
    image: bitnami/mongodb-exporter:0.11.2-debian-10-r382
    imagePullPolicy: IfNotPresent
    resources: {}
```

## ServiceMonitor for Prometheus Operator

Once the exporter is configured, the next aligned task would be to ask [Prometheus](https://prometheus.io) to monitor it. For [Prometheus Operator](https://github.com/prometheus-operator/prometheus-operator), we have to create CRD object in Kubernetes called "ServiceMonitor". We can update this using the `helm` as well.

```yaml
serviceMonitor:
  enabled: false
  interval: 30s
  scrapeTimeout: 10s
  namespace: monitoring
```

For kubectl related configuration, we may have to create `ServiceMonitor` definition in a yaml file and apply it using kubectl command. A `ServiceMonitor` definition looks like this:-

```yaml
---
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: mongodb-prometheus-monitoring
  labels:
    app.kubernetes.io/name: mongodb
    app.kubernetes.io/managed-by: mongodb
    app.kubernetes.io/instance: mongodb
    app.kubernetes.io/version: v0.1.0
    app.kubernetes.io/component: middleware
spec:
  selector:
    matchLabels:
      app: mongodb
      mongodb_setup: standalone
      role: standalone
  endpoints:
  - port: metrics
    interval: 30s
    scrapeTimeout: 30s
  namespaceSelector:
    matchNames:
      - middleware-production
```


