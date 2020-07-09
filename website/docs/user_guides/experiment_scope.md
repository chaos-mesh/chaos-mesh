---
id: experiment_scope
title: Define the Scope of Chaos Experiment
sidebar_label: Define the Scope of Chaos Experiment
---

This document describes how to define the scope of a chaos experiment.

Chaos Mesh provides a variety of selectors, which you can use to define the scope of your chaos experiment. These selectors are defined in the `spec.selector` field of the chaos object.

## Namespace selectors

Namespace selectors filter the chaos experiment targets by the namespace. Defined as a set of strings. The default namespace selector for Chaos Mesh is the chaos experiment object. For example:

```yaml
spec: 
  selector:
    namespaces:
      - "app-ns"
```

## Label selectors

Label selectors filter chaos experiment targets by the label. Defined as a map of string keys and values. For example:

```yaml
spec: 
  selector:
    labelSelectors:
      "app.kubernetes.io/component": "tikv"
```

## Annotation selectors

Annotation selectors filter chaos experiment targets by the annotation. Defined as a map of string keys and values. For example:

```yaml
spec: 
  selector:
    annotationSelectors:
      "example-annotation": "group-a"
```

## Field selectors 

Field selectors filter chaos experiment targets by the resource field. Defined as a map of string keys and values. For example:

```yaml
spec: 
  selector:
    fieldSelectors:
      "metadata.name": "my-pod"
```

For more details about field selectors, refer to the [Kubernetes document](https://kubernetes.io/docs/concepts/overview/working-with-objects/field-selectors/).

## Pod phase selectors

Pod Phase selectors filter chaos experiment targets by the condition. Defined as a set of string. Supported conditions: `Pending`, `Running`, `Succeeded`, `Failed`, `Unknown`. For example:

```yaml
spec: 
  selector:
    podPhaseSelectors:
      - "Running"
```

## Pod selectors

Pod selectors filter chaos experiment targets by the pod. Defined as a map of string keys and values. The key in this map specifies the namespace which the pods belong to, and each value under the key is a pod. If this selector is not empty, these pod defined in this map are used directly and other defined selectors will be ignored. For example:

```yaml
spec: 
  selector:
    pods:
      tidb-cluster: # namespace of the target pods
        - basic-tidb-0
        - basic-pd-0
        - basic-tikv-0
        - basic-tikv-1
```
