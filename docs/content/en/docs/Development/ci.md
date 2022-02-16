---
title: "Continous Integration Pipeline"
weight: 3
linkTitle: "Continous Integration Pipeline"
description: >
    Continous Integration Pipeline for MongoDB Operator
---

We are using Azure DevOps pipeline for the Continous Integration in the MongoDB Operator. It checks all the important checks for the corresponding Pull Request. Also, this pipeline is capable of making releases on Quay, Dockerhub, and GitHub.

The pipeline definition can be edited inside the [.azure-pipelines](https://github.com/OT-CONTAINER-KIT/mongodb-operator/tree/main/.azure-pipelines).

![](https://github.com/OT-CONTAINER-KIT/mongodb-operator/blob/main/static/mongodb-ci-pipeline.png?raw=true)

Tools used for CI process:-

- **Golang ---> https://go.dev/**
- **Golang CI Lint ---. https://github.com/golangci/golangci-lint**
- **Hadolint ---> https://github.com/hadolint/hadolint**
- **GoSec ---> https://github.com/securego/gosec**
- **Trivy ---> https://github.com/aquasecurity/trivy**


