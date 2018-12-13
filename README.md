Docker authentication controller for Kubernetes
===
Custom controller for Kubernetes to manage docker config secret for private docker repos ECR.

Requirements
---
1.9.3 < Kubernetes version < 1.13.0

Features
---
- Observing namespace event such as create, update and delete
- Automatically create/update/delete private docker repo secret to namespaces
- Automatically refresh secret in a given period (default two hourly)

Docker Hub Image
---
[liangrog/kctlr-docker-auth](https://hub.docker.com/r/liangrog/kctlr-docker-auth/)

Installation
---
Deploy to kubernetes cluster via Helm. [Details see here](https://github.com/liangrog/chart-kctlr-docker-auth)

Development
---
The codes work in go 1.11.2. Package management is using go modules.

