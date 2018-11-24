Docker authentication controller for Kubernetes
===
Kubernetes custom controller creating/refreshing docker config secret for private docker repos such as ECR.

Requirements
---
- go v1.11

Features
---
- Observing namespace event such as create, update and delete
- Automatically create/update/delete private docker repo secret to namespaces
- Automatically refresh secret in a given period (default two hourly)

Docker Hub Image
---
[liangrog/kctlr-docker-auth](https://hub.docker.com/r/liangrog/kctlr-docker-auth/)

Version Mapping
---
|     Branch     |   Tag   | Docker Image | Kubernetes | 
|:--------------:|:-------:|:------------:|:----------:|
| Master         | HEAD    | latest       | 1.9.3      |
| release-1.9.3  | v1.9.3  | 1.9.3        | 1.9.3      |

How it works
---

Installation
---
1. Deploy to kubernetes cluster via Helm. [Details see here](https://github.com/liangrog/chart-kctlr-docker-auth)


2. Deploy outside kubernetes cluster via docker 


Configurations
---


