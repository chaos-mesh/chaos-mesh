<!--
Copyright 2026 Chaos Mesh Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
-->

# EnvoyGatewayChaos

EnvoyGatewayChaos injects delay and abort faults by creating an Envoy Gateway
`BackendTrafficPolicy` for one `HTTPRoute` or `GRPCRoute`. It does not modify the
route or any existing policy.

The managed policy has a deterministic name and an ownership annotation. Apply
and recovery are idempotent, and recovery deletes only the owned policy. The
controller rejects an experiment when another direct reference or label selector
already targets the route at the same policy level.
