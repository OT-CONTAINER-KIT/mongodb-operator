---
title: "Overview"
linkTitle: "Overview"
weight: 1
description: >
  Overview of the MongoDB Operator
---

MongoDB Operator is an operator created in Golang to create, update, and manage MongoDB standalone, replicated, and arbiter replicated setup on Kubernetes and Openshift clusters. This operator is capable of doing the setup for MongoDB with all the required best practices.

## Architecture

Architecture for MongoDB operator looks like this:-

![](https://github.com/OT-CONTAINER-KIT/mongodb-operator/raw/docs-update/static/mongodb-operator-arc.png)

## Purpose

The aim and purpose of creating this MongoDB operator are to provide an easy and extensible way of deploying a Production grade MongoDB setup on Kubernetes. It helps in designing the different types of MongoDB setup like - standalone, replicated, etc with security and monitoring best practices.

## Supported Features

- MongoDB replicated cluster setup
- MongoDB standalone setup
- MongoDB replicated cluster failover and recovery
- Monitoring support with MongoDB Exporter
- Password based authentication for MongoDB
- Kubernetes's resources for MongoDB standalone and cluster

