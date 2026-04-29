# Chaos Mesh BDD E2E Tests

This directory provides a [Godog](https://github.com/cucumber/godog) +
[Gherkin](https://cucumber.io/docs/gherkin/) BDD layer for the Chaos Mesh E2E
test suite. It lives alongside the existing Ginkgo-based tests, which remain
the primary CI path until the BDD layer is proven stable.

## Structure

```
e2e/bdd/
├── features/
│   ├── podchaos/
│   │   ├── container_kill.feature
│   │   ├── pod_failure.feature
│   │   └── pod_kill.feature
│   └── networkchaos/
│       ├── network_delay.feature
│       ├── network_partition.feature
│       └── peers_crossover.feature
├── steps/
│   ├── context.go          – shared ScenarioContext struct
│   ├── network_probe.go    – low-level network probing helpers
│   ├── networkchaos_steps.go
│   └── podchaos_steps.go
└── suite_test.go           – Godog test runner
```

## Running

Prerequisites: a running Kubernetes cluster with Chaos Mesh installed and a
kubeconfig pointing at it.

```bash
cd e2e-test
go test ./e2e/bdd/... -v \
    --kubeconfig=$HOME/.kube/config \
    --namespace=chaos-testing
```

Flags:

| Flag            | Default          | Description                          |
|-----------------|------------------|--------------------------------------|
| `--kubeconfig`  | `$KUBECONFIG`    | Path to kubeconfig                   |
| `--namespace`   | `chaos-testing`  | Namespace for test resources         |

## Writing new scenarios

1. Add or edit a `.feature` file under `features/`.
2. Run the tests – Godog will print `TODO` for every unimplemented step.
3. Implement the missing step in the appropriate `steps/*_steps.go` file by
   calling `sc.Step(regex, handler)`.

### Step definition conventions

- **Given** steps set up Kubernetes workloads and snapshot initial state.
- **When** steps create, delete, pause, or unpause Chaos Mesh custom resources.
- **Then** steps assert the expected cluster state using
  `wait.PollUntilContextTimeout`.
- Reuse helpers from `e2e-test/pkg/fixture` and `e2e-test/e2e/util` instead
  of duplicating test infrastructure.
- Use the callback context (not `context.TODO()`) for all Kubernetes API calls
  inside `PollUntilContextTimeout` to ensure cancellation propagates correctly.

### NetworkChaos scenarios

NetworkChaos scenarios require network peer pods and port-forwarding set up by
the outer test environment. Populate `ScenarioContext.NetworkPeers` and
`ScenarioContext.Ports` before the scenario runs (typically done in a
`BeforeScenario` hook in the suite initialiser).

## Relationship to existing Ginkgo tests

The BDD scenarios cover the same behaviors as the Ginkgo tests in:

- `e2e/chaos/podchaos/`
- `e2e/chaos/networkchaos/`

Both test paths can run in parallel. The BDD layer is intended to make test
intent more readable and lower the barrier for new contributors.
