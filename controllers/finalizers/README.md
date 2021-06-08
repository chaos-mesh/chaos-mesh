# Death Controller

Death controller controls the `.ObjectMeta.Finalizers` field:

1. If the object has been deleted, iterate over the `records`, and if all of them are "not injected", remove the finalizer
   and go to step 3
2. If the object don't have a finalizer, add one for it.
3. If the finalizer has been updated, upload them to kubernetes server.
