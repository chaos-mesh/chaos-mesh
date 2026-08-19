<img src="static/logo.svg#gh-light-mode-only" alt="Chaos Mesh Logo" width="450" />
<img src="static/logo-white.svg#gh-dark-mode-only" alt="Chaos Mesh Logo" width="450" />

---

<!-- markdown-link-check-disable -->

<!-- prettier-ignore -->
[![LICENSE](https://img.shields.io/github/license/chaos-mesh/chaos-mesh.svg)](https://github.com/chaos-mesh/chaos-mesh/blob/master/LICENSE)
[![Upload Image](https://github.com/chaos-mesh/chaos-mesh/actions/workflows/upload_image.yml/badge.svg?event=schedule)](https://github.com/chaos-mesh/chaos-mesh/actions/workflows/upload_image.yml)
[![codecov](https://codecov.io/gh/chaos-mesh/chaos-mesh/branch/master/graph/badge.svg)](https://codecov.io/gh/chaos-mesh/chaos-mesh)
[![GoDoc](https://img.shields.io/badge/Godoc-reference-blue.svg)](https://godoc.org/github.com/chaos-mesh/chaos-mesh)
[![Artifact Hub](https://img.shields.io/endpoint?url=https%3A%2F%2Fartifacthub.io%2Fbadge%2Frepository%2Fchaos-mesh)](https://artifacthub.io/packages/helm/chaos-mesh/chaos-mesh)
[![OpenSSF Best Practices](https://www.bestpractices.dev/projects/3680/badge)](https://www.bestpractices.dev/projects/3680)

<!-- markdown-link-check-enable -->

Chaos Mesh is an open source, cloud-native Chaos Engineering platform for Kubernetes. It uses Kubernetes custom resources to define, orchestrate, and observe controlled fault injection against workloads, infrastructure, cloud services, and applications.

<!-- prettier-ignore -->
![cncf_logo](./static/cncf.png#gh-light-mode-only)
![cncf_logo](./static/cncf-white.png#gh-dark-mode-only)

Chaos Mesh is a [Cloud Native Computing Foundation (CNCF)](https://www.cncf.io/) incubating project. If you are an organization that wants to help shape the evolution of technologies that are container-packaged, dynamically-scheduled and microservices-oriented, consider joining the CNCF.

## Features

- **Broad fault coverage:** Pod, network, DNS, HTTP, I/O, time, stress, kernel, block device, JVM, physical machine, AWS, Azure, and GCP faults.
- **Kubernetes-native API:** Chaos experiments are defined as custom resources and managed through the Kubernetes API using standard Kubernetes tooling and RBAC.
- **Experiment orchestration:** `Schedule`, `Workflow`, and `StatusCheck` resources support recurring experiments, serial or parallel workflows, and application health checks.
- **Dashboard:** create, manage, and inspect experiments, schedules, and workflows through a web UI and API.
- **Multi-cluster execution:** manage remote clusters and dispatch supported chaos experiments from a management cluster.

See the [Chaos Mesh documentation](https://chaos-mesh.org/docs/) for the behavior and configuration of each fault type.

## Architecture

Chaos Mesh has three main runtime components:

![Chaos Mesh architecture](./static/architecture.svg)

- **Chaos Controller Manager** watches Chaos Mesh resources, validates requests through admission webhooks, schedules workflows and experiments, and coordinates injection and recovery.
- **Chaos Daemon** runs on Kubernetes nodes as a DaemonSet. It performs privileged node- and container-level operations for faults involving runtimes, processes, networks, filesystems, clocks, and kernels.
- **Chaos Dashboard** provides the HTTP API and web interface for managing and observing experiments. It is optional when experiments are managed directly through the Kubernetes API.

Users create or update Chaos Mesh resources through the Kubernetes API, either directly or through the Dashboard. The Controller Manager reconciles the desired state and delegates node-level operations to Chaos Daemon when required.

For implementation details, start with the [controller architecture guide](controllers/README.md), [command entry points](cmd/README.md), or [Helm chart guide](helm/chaos-mesh/README.md).

## Get started

- [Install Chaos Mesh using Helm](https://chaos-mesh.org/docs/production-installation-using-helm/)
- [Run a Chaos experiment](https://chaos-mesh.org/docs/run-a-chaos-experiment/)
- [Supported releases and environments](https://chaos-mesh.org/supported-releases/)

Prefer to try it without setting up a cluster? Run the [interactive Killercoda playground](https://killercoda.com/saiyampathak/scenario/chaos-mesh) in your browser: install Chaos Mesh on a real 2-node Kubernetes cluster, run PodChaos and NetworkChaos experiments, explore the Dashboard, and chain a Workflow - in about 20 minutes.

## Contributing

Read [CONTRIBUTING.md](CONTRIBUTING.md) for the basic workflow and links to area-specific development guides. All contributors must follow the [Code of Conduct](CODE_OF_CONDUCT.md).

Report security issues according to [SECURITY.md](SECURITY.md).

## Adopters

See [ADOPTERS](ADOPTERS.md).

## Blogs

Blogs on Chaos Mesh design & implementation, features, chaos engineering, community updates, etc. See [Chaos Mesh Blogs](https://chaos-mesh.org/blog). Here are some recommended ones for you to start with:

- [Chaos Mesh 2.0: To a Chaos Engineering Ecology](https://chaos-mesh.org/blog/chaos-mesh-2.0-to-a-chaos-engineering-ecology/)
- [Chaos Mesh - Your Chaos Engineering Solution for System Resiliency on Kubernetes](https://chaos-mesh.org/blog/chaos_mesh_your_chaos_engineering_solution/)
- [Run Your First Chaos Experiment in 10 Minutes](https://chaos-mesh.org/blog/run_your_first_chaos_experiment/)
- [How to Simulate I/O Faults at Runtime](https://chaos-mesh.org/blog/how-to-simulate-io-faults-at-runtime/)
- [Simulating Clock Skew in K8s Without Affecting Other Containers on the Node](https://chaos-mesh.org/blog/simulating-clock-skew-in-k8s-without-affecting-other-containers-on-node/)
- [Building an Automated Testing Framework Based on Chaos Mesh and Argo](https://chaos-mesh.org/blog/building_automated_testing_framework)

## Community

Please reach out for bugs, feature requests, and other issues via:

- Following us on Twitter [@chaos_mesh](https://twitter.com/chaos_mesh).

- Joining the `#project-chaos-mesh` channel in the [CNCF Slack](https://slack.cncf.io/) workspace.

- Filling an issue or opening a PR against this repository.

### Community meetings

- Chaos Mesh Community Monthly (Community and project-level updates, community sharing/demo, office hours)
  - Time: on the fourth Thursday of every month (unless otherwise specified)
  - [RSVP here](https://community.cncf.io/chaos-mesh-community/)
  - [Meeting minutes](https://docs.google.com/document/d/1H8IfmhIJiJ1ltg-XLjqR_P_RaMHUGrl1CzvHnKM_9Sc/edit?usp=sharing)

- Chaos Mesh Development Meeting (Releases, roadmap/features/RFC planning and discussion, issue triage/discussion, etc)
  - Time: Every other Tuesday (unless otherwise specified)
  - [RSVP here](https://community.cncf.io/chaos-mesh-community/)
  - [Meeting minutes](https://docs.google.com/document/d/1s9X6tTOy3OGZaLDZQesGw1BNOrxQfWExjBFIn5irpPE/edit)

### Community blogs

- Grant Tarrant-Fisher: [Integrate your Reliability Toolkit with Your World](https://medium.com/search?q=Integrate+your+Reliability+Toolkit+with+Your+World)
- Yoshinori Teraoka: [Streake: Chaos Mesh によるカオスエンジニアリング](https://medium.com/sreake-jp/chaos-mesh-%E3%81%AB%E3%82%88%E3%82%8B%E3%82%AB%E3%82%AA%E3%82%B9%E3%82%A8%E3%83%B3%E3%82%B8%E3%83%8B%E3%82%A2%E3%83%AA%E3%83%B3%E3%82%B0-46fa2897c742)
- Sébastien Prud'homme: [Chaos Mesh : un générateur de chaos pour Kubernetes](https://www.cowboysysop.com/post/chaos-mesh-un-generateur-de-chaos-pour-kubernetes/)
- Craig Morten
  - [K8s Chaos Dive: Chaos-Mesh Part 1](https://dev.to/craigmorten/k8s-chaos-dive-2-chaos-mesh-part-1-2i96)
  - [K8s Chaos Dive: Chaos-Mesh Part 2](https://dev.to/craigmorten/k8s-chaos-dive-chaos-mesh-part-2-536m)
- Ronak Banka: [Getting Started with Chaos Mesh and Kubernetes](https://itnext.io/getting-started-with-chaos-mesh-and-kubernetes-bfd98d25d481)
- kondoumh: [​Kubernetes ネイティブなカオスエンジニアリングツール Chaos Mesh を使ってみる](https://blog.kondoumh.com/entry/2020/10/23/123431)
- Vadim Tkachenko: [ChaosMesh to Create Chaos in Kubernetes](https://www.percona.com/blog/2020/11/05/chaosmesh-to-create-chaos-in-kubernetes/)
- Hui Zhang: [How a Top Game Company Uses Chaos Engineering to Improve Testing](https://chaos-mesh.org/blog/how-a-top-game-company-uses-chaos-engineering-to-improve-testing)
- Anurag Paliwal
  - [Securing tenant services while using chaos mesh using OPA](https://anuragpaliwal-93749.medium.com/securing-tenant-services-while-using-chaos-mesh-using-opa-3ae80c7f4b85)
  - [Securing namespaces using restrict authorization feature in chaos mesh](https://anuragpaliwal-93749.medium.com/securing-namespaces-using-restrict-authorization-feature-in-chaos-mesh-2e110c3e0fb7)
- Pavan Kumar: [Chaos Engineering in Kubernetes using Chaos Mesh](https://link.medium.com/1V90dEknugb)
- Jessica Cherry: [Test your Kubernetes experiments with an open source web interface](https://opensource.com/article/21/6/chaos-mesh-kubernetes)
- λ.eranga: [Chaos Engineering with Chaos Mesh](https://medium.com/rahasak/chaos-engineering-with-chaos-mesh-b040169b51bd)
- Tomáš Kubica: [Kubernetes prakticky: zlounství s Chaos Mesh a Azure Chaos Studio](https://www.tomaskubica.cz/post/2021/kubernetes-prakticky-zlounstvi-s-chaos-mesh-a-azure-chaos-studio2/)
- mend: [Chaos Meshで何ができるのか見てみた](https://qiita.com/mend/items/dcdfab5e980467bf58e9)

### Community talks

- Twain Taylor: [Chaos Mesh Simplifies & Organizes Chaos Engineering For Kubernetes](https://youtu.be/shbrjAY86ZQ)
- Saiyam Pathak
  - [Let's explore chaos mesh](https://youtu.be/kMbTYItsTTI)
  - [Chaos Mesh - Chaos Engineering for Kubernetes](https://youtu.be/HAU_cjW1bMw)
  - [Chaos Mesh 2.0](https://youtu.be/HmQ9cFwxF7g)

## Media coverage

- CodeZine: [オープンソースのカオステストツール「Chaos Mesh 1.0」、一般提供を開始](https://codezine.jp/article/detail/12996)
- @IT atmarkit: [Kubernetes 向けカオスエンジニアリングプラットフォーム「Chaos Mesh 1.0」が公開](https://www.atmarkit.co.jp/ait/articles/2010/09/news108.html)
- Publickey: [Kubernetes の Pod やネットワークをわざと落としまくってカオスエンジニアリングのテストができる「Chaos Mesh」がバージョン 1.0 に到達](https://www.publickey1.jp/blog/20/kubernetespodchaos_mesh10.html)
- InfoQ: [Chaos Engineering on Kubernetes : Chaos Mesh Generally Available with v1.0](https://www.infoq.com/news/2020/10/kubernetes-chaos-mesh-ga/)
- TechGenix: [Chaos Mesh Promises to Bring Order to Chaos Engineering](http://techgenix.com/chaos-mesh-chaos-engineering/)

## License

Chaos Mesh is licensed under the Apache License 2.0. See [LICENSE](LICENSE).

## Trademark

Chaos Mesh is a trademark of The Linux Foundation. All rights reserved.
