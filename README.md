## Table of contents

* [Introduction](#introduction)
* [Features](#features)
* [Working Example](#working-example)
* [Technical Support](#technical-support)
* [Copyrights](#copyrights)

## Introduction

APIMatic CodeGen Operator simplifies the configuration and lifecycle management of the APIMatic CodeGen SDKs, Docs and DX Portal Generation solution on different Kubernetes distributions and OpenShift. The Operator encapsulates key operational knowledge on how to configure and upgrade the APIMatic CodeGen application, making it easy to get it up and running.


More information about the underlying APIMatic CodeGen API that is exposed
by this operator can be found [here](https://apimatic-core-v3-docs.netlify.app/#/http/getting-started/overview-apimatic-core).

## Features

APIMatic CodeGen Operator provides the following features:
- Deploys the APIMatic CodeGen Web API service within the Kubernetes or OpenShift cluster.
- Exposes the APIMatic CodeGen API external to the cluster, using Service type as [NodePort](https://kubernetes.io/docs/concepts/services-networking/service/#nodeport), [LoadBalancer](https://kubernetes.io/docs/concepts/services-networking/service/#loadbalancer).
- For exposing the service through an ingress resource, create an Ingress resource in the namespace of your APIMatic CR and set owned APIMatic service created by the operator as a backed service. More information can be found [here](https://kubernetes.io/docs/concepts/services-networking/ingress/).
- Manual horizontal scaling of pods.
  ```sh
  kubectl scale cgn codegen-sample --replicas=2
  ```
## Working Example

A complete example using APIMatic CodeGen Operator in your OpenShift environment can be found [here](https://github.com/apimatic/apimatic-codegen-openshift-pipelines-demo).

## Technical Support

- To request additional features in the future, or if you notice any discrepancy regarding this document, please drop an email to [support@apimatic.io](mailto:support@apimatic.io).

### Copyrights

&copy; 2022 APIMatic.io
