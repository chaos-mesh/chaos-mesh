# DesiredPhase Controller

This controller will control the `.Status.Experiment.DesiredPhase` field with the steps below:

1. if the `desiredPhase` is empty, set it to "running" and go to step 4
2. if duration exceeded, set `desiredPhase` to "stopped" and go the step 4
3. if it has been paused, set `desiredPhase` to "stopped"; if not, set it to "running".
4. if the `desiredPhase` has been updated， sync the difference to the kubernetes server.
