# EnvoyGatewayChaos

EnvoyGatewayChaos injects delay and abort faults by creating an Envoy Gateway
`BackendTrafficPolicy` for one `HTTPRoute` or `GRPCRoute`. It does not modify the
route or any existing policy.

The managed policy has a deterministic name and an ownership annotation. Apply
and recovery are idempotent, and recovery deletes only the owned policy. The
controller rejects an experiment when another direct reference or label selector
already targets the route at the same policy level.
