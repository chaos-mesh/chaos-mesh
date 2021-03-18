# Common Controller

Common controller controls the `.Status.Experiment.Records` field with the steps below:

1. if the `records` are nil, try to select new objects and save to the `records`.
2. iterate over `records`, for every `record`, if the `Phase` of it doesn't match the `DesiredPhase`, try to sync them
through `Apply` or `Recover`, and update the `Phase` accordingly.
3. if the `records` has changed, upload them to the kubernetes server.

## Design Discussion

### The implementation of chaos should be simple

The definition of every chaos should be simple. It needs a configuration, one or more selectors, and an implementation
for a single object. For example, the implementation of PodKill should only kill one pod, but not "select", "iterate" and
delete all of them. These things (like "select", "iterate") should be done in the powerful controller.

### Selector should be powerful

1. "Pod" is not the only target of chaos. We could inject chaos on volume, container, pod, ec2 machine... The selector
    abstraction should be able to handle the diversity.

2. One chaos definition should be able to have a lot of selector, e.g. the `NetworkChaos`.
