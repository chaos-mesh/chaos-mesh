---
id: features
title: Features
sidebar_label: Features
---

# Easy to Use

No special dependencies, Chaos Mesh can be easily deployed directly on Kubernetes clusters, including [Minikube](https://github.com/kubernetes/minikube) and [Kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

Require no modification to the deployment logic of the system under test (SUT)
Easily orchestrate fault injection behaviors in chaos experiments
Hide underlying implementation details so that users can focus on orchestrating the chaos experiments


# Design for Kubernetes

Chaos Mesh uses [CustomResourceDefinitions](https://kubernetes.io/docs/concepts/extend-kubernetes/api-extension/custom-resources/) (CRD) to define chaos objects.

In the Kubernetes realm, CRD is a mature solution for implementing custom resources, with abundant implementation cases and toolsets available. Using CRD makes Chaos Mesh naturally integrate with the Kubernetes ecosystem.