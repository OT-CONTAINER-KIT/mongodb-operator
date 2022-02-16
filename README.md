<p align="center">
  <img src="./static/mongodb-operator-logo.svg" height="220" width="220">
</p>

<p align="center">
  <a href="https://dev.azure.com/opstreedevops/DevOps/_build/latest?definitionId=7&repoName=OT-CONTAINER-KIT%2Fmongodb-operator&branchName=main">
    <img src="https://dev.azure.com/opstreedevops/DevOps/_apis/build/status/OT-CONTAINER-KIT.mongodb-operator?repoName=OT-CONTAINER-KIT%2Fmongodb-operator&branchName=main" alt="Azure Pipelines">
  </a>
  <a href="https://goreportcard.com/report/github.com/OT-CONTAINER-KIT/mongodb-operator">
    <img src="https://goreportcard.com/badge/github.com/OT-CONTAINER-KIT/mongodb-operator" alt="GoReportCard">
  </a>
  <a href="http://golang.org">
    <img src="https://img.shields.io/github/go-mod/go-version/OT-CONTAINER-KIT/mongodb-operator" alt="GitHub go.mod Go version (subdirectory of monorepo)">
  </a>
  <a href="https://quay.io/repository/opstree/mongodb-operator">
    <img src="https://img.shields.io/badge/container-ready-green" alt="Docker">
  </a>
  <a href="https://github.com/OT-CONTAINER-KIT/mongodb-operator/master/LICENSE">
    <img src="https://img.shields.io/badge/License-Apache%202.0-blue.svg" alt="License">
  </a>
</p>

MongoDB Operator is an operator created in Golang to create, update, and manage MongoDB standalone, replicated, and arbiter replicated setup on Kubernetes and Openshift clusters. This operator is capable of doing the setup for MongoDB with all the required best practices.

For documentation, please refer to https://ot-mongodb-operator.netlify.app/

## Architecture

Architecture for MongoDB operator looks like this:-

<div align="center">
    <img src="./static/mongodb-operator-arc.png">
</div>

## Purpose 

The aim and purpose of creating this MongoDB operator are to provide an easy and extensible way of deploying a Production grade MongoDB setup on Kubernetes. It helps in designing the different types of MongoDB setup like - standalone, replicated, etc with security and monitoring best practices.

## Supported Features

- MongoDB replicated cluster setup
- MongoDB standalone setup
- MongoDB replicated cluster failover and recovery
- Monitoring support with MongoDB Exporter
- Password based authentication for MongoDB
- Kubernetes's resources for MongoDB standalone and cluster

## Prerequisites

MongoDB Operator requires a Kubernetes cluster of version `>=1.16.0`. If you have just started with the CRD and Operators, its highly recommended using the latest version of Kubernetes.

## Getting Started

If you want to start using the mongodb-operator in a quickstart mode, you can begin with the [documentation](https://ot-container-kit.github.io/mongodb-operator/). It will help you and guide you through the setup of MongoDB step-by-step.

The configuration for MongoDB and MongoDB cluster setup is defined inside the CRD manifests. But all the examples manifests can be found in the [examples](./examples) directory.

## Quickstart

### MongoDB Operator Setup

The setup can be done by using helm. The mongodb-operator can easily get installed using helm commands.

```shell
# Add the helm chart
$ helm repo add ot-helm https://ot-container-kit.github.io/helm-charts/
...
"ot-helm" has been added to your repositories
````

```shell
# Deploy the MongoDB Operator
$ helm upgrade mongodb-operator ot-helm/mongdb-operator \
  --install --namespace ot-operators --install
...
Release "mongodb-operator" does not exist. Installing it now.
NAME: mongodb-operator
LAST DEPLOYED: Sun Jan  9 23:05:13 2022
NAMESPACE: ot-operators
STATUS: deployed
REVISION: 1
```

After the deployment, verify the installation of operator

```shell
# Testing Operator
$ helm test mongodb-operator --namespace ot-operators
...
NAME: mongodb-operator
LAST DEPLOYED: Sun Jan  9 23:05:13 2022
NAMESPACE: ot-operators
STATUS: deployed
REVISION: 1
TEST SUITE:     mongodb-operator-test-connection
Last Started:   Sun Jan  9 23:05:54 2022
Last Completed: Sun Jan  9 23:06:01 2022
Phase:          Succeeded
```

## Upcoming Features

- MongoDB sharded cluster setup
- Customizable configuration changes in MongoDB
- TLS security support
- Backup and restore support
- DB and user creation 
- Insightful Grafana dashboards
