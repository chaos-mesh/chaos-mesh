# Chaos Mesh® Adopters

Chaos Mesh now has over 40 adopters, and here we have listed some of them. Some have already gone into production, and others are at various stages of testing.

If you are a Chaos Mesh adopter, and are willing to share your Chaos Mesh story, feel free to raise a PR!

- [Apache APISIX](https://github.com/apache/apisix)
  - Apache APISIX is a dynamic, real-time, high-performance open source API gateway, which provides rich traffic management features, such as load balancing, dynamic upstream and more.
  - APISIX integrates Chaos Mesh into open-source CI pipelines, to further enhance APISIX's resiliency and robustness. See [Use case](https://chaos-mesh.org/blog/How-Chaos-Mesh-Helps-Apache-APISIX-Improve-System-Stability)
- [ByteDance](https://bytedance.com/en/)
  - ByteDance is a technology company operating a range of content platforms including TikTok, Lark, Helo, Vigo Video, Douyin, and Huoshan in over 150 markets and 75 languages.
  - ByteDance's self-developed chaos engineering platform is mainly used by the company's own technology system. As there are some cloud-native deployment services involved, they integrated Chaos Mesh as the underlying fault injection engine, which is a key supplement to ByteDance’s chaos engineering platform.
- [Celo](https://celo.org/)
- [Dailymotion](https://www.dailymotion.com/)
- [DataStax Fallout](https://github.com/datastax/fallout)
- [NetEase Fuxi Lab](https://fuxi.163.com/en/about.html)
  - NetEase Fuxi AI Lab is China’s first professional game AI research institution. Researchers use their Kubernetes-based Danlu platform for algorithm development, training and tuning, and online publishing.
  - They use Chaos Mesh to improve the stability of their internal hybrid cloud. In addition, their users with cloud platforms also access Chaos Mesh to test the stability of user services. See [Use case](https://chaos-mesh.org/blog/how-a-top-game-company-uses-chaos-engineering-to-improve-testing).
- [JuiceFS](https://juicefs.com/?hl=en)
- [KingNet](https://www.kingnet.com/)
  - KingNet’s main business includes the development, operation and distribution of premium entertainment content.
  - KingNet mainly uses Chaos Mesh for testing the availability of multiple data centers and microservice links. Chaos Mesh also helps them with mocking service unavailability or abnormal network conditions.
- [Meituan Dianping](https://about.meituan.com/en)
- [PingCAP](https://en.pingcap.com/)
- [Apache Pulsar](https://pulsar.apache.org/)
- [Qihoo 360](https://360.cn/)
- [Qiniu Cloud](https://qiniu.com/en)
  - Qiniu Cloud is a distributed cloud system that carries massive amounts of data, and is one that requires high data consistency and high availability, with the data quantity level at 1 trillion+.
  - To ensure the reliability of cloud storage products, they use Chaos Mesh to perform chaos tests on metadata and the underlying storage system under conditions such as: single point of failure of services, network abnormality, abnormal resource consumption (CPU, memory, I/O), etc.
- [S.J. Distributors](https://www.sjfood.com/)
- [Tencent](https://www.tencent.com/en-us)
  - After Tencent Interactive Entertainment migrated their online operations to the Tencent Cloud Kubernetes engine, they wished to provide users with a more stable and reliable experience, which is why they introduced Chaos Mesh. Tencent mainly use Chaos Mesh to simulate the following types of failures:
    - Fault isolation, such as simulating pod abnormality, and checking whether the system can automatically isolate fault instances;
      Service degradation, such as simulating a downstream recommended service failure through network failure, and verifying whether the local cache is effective;
    - Verifying if the alarm works, for example, purposefully burning the CPU to 90%, and checking whether the alarm is timely issued in time.
- [Vald](https://vald.vdaas.org/)
- [WeBank](https://www.webank.com/)
- [Xpeng](https://en.xiaopeng.com/)
  - Xpeng Motors is China's leading smart electric vehicle designer and manufacturer, as well as a technology company integrating cutting-edge Internet and AI innovation. They use Chaos Mesh in the following scenarios:
    - Rolling updates of microservices and lossless verification of traffic;
    - Microservices, multi-registries, multi-party synchronization, and traffic lossless verification;
    - MQTT cluster two-way subscription verification;
    - Exactly-once consumer business verification for message queues;
    - Simulation of weak 4G network for in-vehicle systems, saving drive test costs;
    - AIOPS anomaly detection dataset generation.
