import React from 'react'

const features = [
  {
    title: <>Easy to Use</>,
    imgUrl: 'img/undraw_server_down_s4lk.svg',
    description: (
      <>
        <p>
          No special dependencies, Chaos Mesh can be easily deployed directly on Kubernetes clusters, including{' '}
          <a href="https://github.com/kubernetes/minikube">Minikube</a> and{' '}
          <a href="https://kind.sigs.k8s.io/docs/user/quick-start/">Kind</a>.
        </p>
        <ul>
          <li>Require no modification to the deployment logic of the system under test (SUT)</li>
          <li>Easily orchestrate fault injection behaviors in chaos experiments</li>
          <li>Hide underlying implementation details so that users can focus on orchestrating the chaos experiments</li>
        </ul>
      </>
    ),
  },
  {
    title: <>Design for Kubernetes</>,
    imgUrl: 'img/logos/kubernetes.svg',
    description: (
      <>
        <p>
          Chaos Mesh uses{' '}
          <a
            href="https://kubernetes.io/docs/concepts/extend-kubernetes/api-extension/custom-resources/"
            target="_blank"
          >
            CustomResourceDefinitions
          </a>{' '}
          (CRD) to define chaos objects.
        </p>
        <p>
          In the Kubernetes realm, CRD is a mature solution for implementing custom resources, with abundant
          implementation cases and toolsets available. Using CRD makes Chaos Mesh naturally integrate with the
          Kubernetes ecosystem.
        </p>
      </>
    ),
    reverse: true,
  },
]

export default features
