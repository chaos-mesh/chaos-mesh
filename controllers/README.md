# Controller Design of Chaos Mesh

This document describes the common controller specification in Chaos Mesh.
Although no "standard" should be considered as absolute requirements (and the
real world is full of trade-off and corner case), they should be carefully
considered when you are trying to add a new controller.

## One controller per field

One field should only be "controlled" by at most one controller. In this
chapter, multiple reasons will be listed for this design:

### Avoid the hidden bugs

Multiple controllers modifying a single object could lead to a conflict
situation (which is more like a global optimistic lock). The common way to solve
conflict is to adapt the modification and retry. However, if multiple
controllers want to modify a single field, how could they merge the conflict?
What's more, it always leads to a hidden bug under the logic. Here is an
example:

If you want to split "pause" and "duration" (the former common chaos) into two
standalone controllers, let's try to describe the logic of them:

For the "pause" controller, when the annotation is added, the chaos should enter
"not injected" mode, and when the annotation is removed, the chaos should enter
"injected" mode.

For the "duration" controller, when the time exceeds the duration, the chaos
should enter "not injected" mode.

Though these logics seem to be intuitive, there is a bug under the conflict
"mode" (or the `desiredPhase` in the current code). What will happen if the user
removes the annotation after the duration exceed? The chaos will enter
"injected" and then turn into "not injected" mode (with the help of "duration"
controller), which is dirty and confusing.

If we obey the "One field per controller" rule, then they should be combined
into one controller and can never be split.

### Handle the conflict in an easier way

After retrying the conflict error, we don't need to rerun the whole controller
logic (as there may be some side effects in the controller). Instead, we could
save the single field, and set the corresponding field after getting the new
object. Which will give us more confidence in the retry attempting.

## Controller should work standalone

The behavior of every controller should be defined carefully, and they should be
able to work without other controllers. The behavior of the controller should
also be simple and easy to understand. Try to conclude the action/logic of the
controller in one hundred words, if you failed, please reconsider whether it
should be "one" controller, but not two or more (or even split a new
CustomResource).

## Controller should be well documented

Every controller should be described with a "little"/"short" document.

## Error Handling

According to the source code of `controller-runtime`:

```go
// RunInformersAndControllers the syncHandler, passing it the namespace/Name string of the
// resource to be synced.
if result, err := c.Do.Reconcile(req); err != nil {
    c.Queue.AddRateLimited(req)
    log.Error(err, "Reconciler error", "controller", c.Name, "request", req)
    ctrlmetrics.ReconcileErrors.WithLabelValues(c.Name).Inc()
    ctrlmetrics.ReconcileTotal.WithLabelValues(c.Name, "error").Inc()
    return false
} else if result.RequeueAfter > 0 {
    // The result.RequeueAfter request will be lost, if it is returned
    // along with a non-nil error. But this is intended as
    // We need to drive to stable reconcile loops before queuing due
    // to result.RequestAfter
    c.Queue.Forget(obj)
    c.Queue.AddAfter(req, result.RequeueAfter)
    ctrlmetrics.ReconcileTotal.WithLabelValues(c.Name, "requeue_after").Inc()
    return true
} else if result.Requeue {
    c.Queue.AddRateLimited(req)
    ctrlmetrics.ReconcileTotal.WithLabelValues(c.Name, "requeue").Inc()
    return true
}
```

If the `Reconcile` return a `Requeue` without `RequeueAfter`, this request will
be added to the `RateLimitQueue`. The default `RateLimitQueue` is constructured
in this way:

```go
// DefaultControllerRateLimiter is a no-arg constructor for a default rate limiter for a workqueue.  It has
// both overall and per-item rate limiting.  The overall is a token bucket and the per-item is exponential
func DefaultControllerRateLimiter() RateLimiter {
    return NewMaxOfRateLimiter(
        NewItemExponentialFailureRateLimiter(5*time.Millisecond, 1000*time.Second),
        // 10 qps, 100 bucket size.  This is only for retry speed and its only the overall factor (not per item)
        &BucketRateLimiter{Limiter: rate.NewLimiter(rate.Limit(10), 100)},
    )
}
```

So it's a good enough error back off without stopping the worker. When a
controller meets a retriable error, the simplest way to handle it is returning a
`ctrl.Result{Requeue: true}, nil`