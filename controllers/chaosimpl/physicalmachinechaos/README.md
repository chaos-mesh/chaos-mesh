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

# Physical machine process lifecycle

A `process` experiment is a one-shot action when its signal is nonzero, differs
from Linux `SIGSTOP` (`19`), and `recoverCmd` is empty. A completed one-shot
experiment stays complete when paused, when its duration expires, or when the
controller restarts. `SIGSTOP` and experiments with a recovery command keep the
normal recovery lifecycle.

The controller uses each target record's persisted `InjectedCount` to avoid
sending another POST after a recorded success, including older records that
have returned to `NotInjected`. Targets without a recorded success still retry.
This does not guarantee exactly-once delivery when an HTTP response or the
status write is lost.

Deleting a one-shot experiment ends its management in Chaos Mesh. Cleanup does
not restart or restore the process, send a Chaosd DELETE request, or mark the
Chaosd experiment history as recovered. Cleanup also works after the target
`PhysicalMachine` is deleted. Mesh record events and counters continue to track
management transitions, including migration and cleanup, rather than count
HTTP requests or prove that a process was restored.
