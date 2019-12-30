# Network Loss / Delay / Duplicate / Corrupt

* **action** defines the specific pod chaos action, supported action: loss / delay / duplicate / corrupt.
* **mode** defines the mode to run chaos action.
* **selector** is used to select pods that are used to inject chaos action.
* **loss** represents the percentage of packet loss.
* **latency** Latency indicates the delay time in sending packets. jitter represents the jitter of the delay time.
* **duplicate** represents the percentage of duplicate packets.
* **corrupt** represents the percentage of corrupt packets.
* **correlation** Network chaos variation isn't purely random, so to emulate that there is a correlation value as well.
* **duration** define the duration time for each chaos experiment.
* **scheduler** defines some scheduler rules to the running time of the chaos experiment about pods.

# Network Partition

* **action** defines the specific pod chaos action. In this case, it means network partition, represents the chaos action of network partition of pods.
* **mode** defines the mode to run chaos action.
* **selector** is used to select pods that are used to inject chaos action.
* **direction** represents the partition direction, supported direction: from / to / both.
* **target** represents network partition target.
* **duration** define the duration time for each chaos experiment.
* **scheduler** defines some scheduler rules to the running time of the chaos experiment about pods.
