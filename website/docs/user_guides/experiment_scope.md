---
id: experiment_scope
title: Define the Scope of Chaos Experiment
sidebar_label: Define the Scope of Chaos Experiment
---

This document describes how to define the chaos experiment scope.

Chaos Mesh provides a lot of selectors for users and you can use them to define the scope of our chaos experiment. 
These selectors are defined in the `spec.selector` field of the chaos object.

## Namespace selectors

Namespace selectors are used to filtering the chaos experiment targets by the namespaces and defined as a set of string. If this selector is empty, 
Chaos Mesh will use the namespace of the chaos experiment object as the default namespace selector. For example:

```yaml
spec: 
  selector:
    namespaces:
      - "app-ns"
```

## Label selectors

Label selectors are used to filtering the chaos experiment targets by the labels and defined as a map of string keys and values. For example:

```yaml
spec: 
  selector:
    labelSelectors:
      "app.kubernetes.io/component": "tikv"
```

## Annotation selectors

Annotation selectors are used to filtering the chaos experiment targets by the annotations and defined as a map of string keys and values. For example:

```yaml
spec: 
  selector:
    annotationSelectors:
      "example-annotation": "group-a"
```

## Field selectors 

Field selectors are used to filtering the chaos experiment targets by the resource fields and defined as a map of string keys and values. 
This selector lets you select Kubernetes resources based on the value of one or more resource fields. For example: 

```yaml
spec: 
  selector:
    fieldSelectors:
      "metadata.name": "my-pod"
```

More details about field selectors, you can refer to the [Kubernetes document](https://kubernetes.io/docs/concepts/overview/working-with-objects/field-selectors/).


## Pod phase selectors

Pod Phase selectors are used to filtering the chaos experiment targets by the condition and defined as a set of string. 
Supported condition: `Pending`, `Running`, `Succeeded`, `Failed`, `Unknown`. For example: 

```yaml
spec: 
  selector:
    podPhaseSelectors:
      - "Running"
```

## Pod selectors

Pod selectors specify specific pods and defined as a map of string keys and a set values. The key of this map defines the namespace which pods belong to, 
and each value is a set of pod names. If this selector is not empty, these pod defined in this map are used directly and ignore the other selectors. For example: 

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

