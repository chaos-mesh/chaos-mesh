# IstioChaos

IstioChaos injects an Istio delay or HTTP abort into one named HTTP route of a
VirtualService. The controller clones the target route immediately before the
original route, preserves its match and forwarding configuration, and adds the
requested `fault` block. The clone has a deterministic name and the
VirtualService is annotated with the owning IstioChaos resource.

Apply and recovery are idempotent. Recovery removes only the owned clone and
ownership annotation; it never restores a snapshot of the complete
VirtualService. Only one IstioChaos can control a VirtualService at a time.
