---
title: "Prometheus Monitoring"
weight: 2
linkTitle: "Prometheus Monitoring"
description: >
    Monitoring of MongoDB standalone and replicaset cluster using Prometheus
---

In MongoDB Operator, we are using [mongodb-exporter](https://github.com/percona/mongodb_exporter) to collect the stats, metrics for MongoDB database. This exporter is capable to capture the stats for standalone and cluster mode of MongoDB.

## MongoDB Monitoring

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

## MongoDB Alerting

Since we are using MongoDB exporter to capture the metrics, we are using the queries available by that exporter to create alerts as well. The alerts are available inside the [alerting](https://github.com/OT-CONTAINER-KIT/mongodb-operator/blob/main/monitoring/alerting/alerts.yaml) directory.

Similar to `ServiceMonitor`, there is another CRD object is available through which we can create Prometheus rules inside Kubernetes cluster using Prometheus Operator.

Example:-

```yaml
---
apiVersion: monitoring.coreos.com/v1
kind: PrometheusRule
metadata:
  labels:
    app: prometheus-mongodb-rules
  name: prometheus-mongodb-rules
spec:
  groups:
    - name: mongodb
      rules:
      - alert: MongodbDown
        expr: mongodb_up == 0
        for: 0m
        labels:
          severity: critical
        annotations:
          summary: MongoDB Down (instance {{ $labels.instance }})
          description: "MongoDB instance is down\n  VALUE = {{ $value }}\n  LABELS = {{ $labels }}"
```

#### Alerts description:-

| **AlertName**                | **Description**                                                                                                       |
|------------------------------|-----------------------------------------------------------------------------------------------------------------------|
| MongoDB Down                 | MongoDB instance is down                                                                                              |
| MongoDB replication lag      | Mongodb replication lag is more than 10s                                                                              |
| MongoDB replication Status 3 | MongoDB Replication set member either perform startup self-checks, or transition from completing a rollback or resync |
| MongoDB replication Status 6 | MongoDB Replication set member as seen from another member of the set, is not yet known                               |
| MongoDB number cursors open  | Too many cursors opened by MongoDB for clients (> 10k)                                                                |
| MongoDB cursors timeouts     | Too many cursors are timing out                                                                                       |
| MongoDB too many connections | Too many connections (> 80%)                                                                                          |
| MongoDB virtual memory usage | High memory usage on MongoDB                                                                                          |

