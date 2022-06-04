## Table of contents

* [Introduction](#introduction)
* [Features](#features)
* [Running the Sample](#running-the-sample)
  * [Prerequisites](#prerequisites)
  * [Steps for Direct Deployment](#steps-for-direct-deployment)
  * [Steps for OLM Deployment](#steps-for-olm-deployment)
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

## Running the Sample 

### Prerequisites

Please contact APIMatic at [support@apimatic.io](mailto:support@apimatic.io) to acquire a valid license to run the APIMatic CodeGen API. The sample used found [here](./config/samples/apicodegen_licenseblob_license.yaml) is a valid license that does not support any SDKs or Docs and is only used to verify that the CodeGen API is successfully started. 

Furthermore, the default image used for the CodeGen application is the latest version of the image found in the [RedHat Container Registry](https://catalog.redhat.com/software/containers/apimatic/apimatic-codegen-ubi8/6284c35ec7259ecf5a67ed99) so you will need RedHat credentials to pull the CodeGen image and the [CodeGen Operator Image](https://catalog.redhat.com/software/containers/apimatic/apimatic-codegen-operator-ubi8/6298c55a90316b739bb8ec87).

Further prerequisites for running the source code on your machine include:

- [go](https://golang.org/) v1.16+
- [git](https://git-scm.com/)
- [make](https://www.gnu.org/software/make/)
- [Operator SDK](https://sdk.operatorframework.io/docs/overview/)
- A running Kubernetes cluster with [kubectl](https://kubernetes.io/docs/reference/kubectl/overview/) on client. For testing purposes, you can use [Minikube](https://minikube.sigs.k8s.io/docs/) or [kind](https://kind.sigs.k8s.io/)
- For checking the service created by the APIMatic operator on-prem, you can use [MetalLB](https://metallb.org/)

### Steps for Direct Deployment

To run the sample for checking the APIMatic operator:

- Log in to `registry.connect.redhat.com`.
  ```sh
  podman login registry.connect.redhat.com
  Username: {REGISTRY-SERVICE-ACCOUNT-USERNAME}
  Password: {REGISTRY-SERVICE-ACCOUNT-PASSWORD}
  ```

- Clone the APIMatic repository into your working directory using the following command:
  ```sh  
  git clone https://github.com/apimatic/apimatic-codegen-operator.git  
  ```
- Run `make deploy IMG=<codegen-operator-image>` to set up the APIMatic operator resources. This will deploy the `apimatic-codegen-operator-system` namespace as well as the CRD and the RBAC manifests.

- Now run the sample using the following command:
  ```sh  
  kubectl apply -f config/samples/apicodegen_licenseblob_license.yaml
  ```
- You will now see a new [Deployment](https://kubernetes.io/docs/concepts/workloads/controllers/deployment/) with replica count of 1 and [Service](https://kubernetes.io/docs/concepts/services-networking/service/) of type `ClusterIP` created, both named ***codegen-sample***. 

- Once done, you can remove the APIMatic resources using the following command:
  ```sh
  make undeploy
  ```

### Steps for OLM Deployment

The following steps can be used to utilize [Operator LifeCycle Manager (OLM)](https://olm.operatorframework.io/docs/) to deploy the operator and run the sample. The steps are as follows:

- Log in to `registry.connect.redhat.com` and pull the [APIMatic CodeGen Bundle Image](https://catalog.redhat.com/software/containers/apimatic/apimatic-codegen-operator-bundle/629b41f3e9314a38a9a33855).
  ```sh
  podman login registry.connect.redhat.com
  Username: {REGISTRY-SERVICE-ACCOUNT-USERNAME}
  Password: {REGISTRY-SERVICE-ACCOUNT-PASSWORD}
  ```

- If not already done so, clone the APIMatic repository into your working directory:
  ```sh  
  git clone https://github.com/apimatic/apimatic-codegen-operator.git  
  ``` 
- [Install OLM in your Kubernetes cluster](https://olm.operatorframework.io/docs/getting-started/#installing-olm-in-your-cluster).

- Run the following script to install the resources required by OLM to deploy the APIMatic CodeGen operator in the Kubernetes cluster within the `apimatic-codegen-operator-system` namespace. Information about the different resources required can be found using the steps given [here](https://olm.operatorframework.io/docs/tasks/).
  ```sh
  operator-sdk run bundle <apimatic-codegen-operator-bundle-image>
  ```

- Further steps are the same as in the previous [section](#steps-for-direct-deployment)

## Technical Support

- To request additional features in the future, or if you notice any discrepancy regarding this document, please drop an email to [support@apimatic.io](mailto:support@apimatic.io).

### Copyrights

&copy; 2022 APIMatic.io
