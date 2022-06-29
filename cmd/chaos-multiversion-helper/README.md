# Create new API version

To create a new api version, we need to copy/paste a new directory containing
the new api, and modify the controller to use a new api. If we still take care
of the compatiblity issues, some functions to convert the api between new
version and old version are still needed.

This program is a helper to create a new version and generate the convert
functions.

## Steps

Take the `v1alpha2` as an example, following steps will show how to generate a
new version of api from the original `v1alpha1`.

1. `./bin/chaos-multiversion-helper create --from v1alpha1 --to v1alpha2
   --as-storage-version`. It will copy the v1alpha1 to v1alpha2, and mark the
   new one as the storage version
2. `./bin/chaos-multiversion-helper migrate --from v1alpha1 --to v1alpha2`.  It
   will modify every reference to v1alpha1 in the chaos mesh code base to become
   `v1alpha2`
3. `./bin/chaos-multiversion-helper autoconvert --version v1alpha1 --hub
   v1alpha2` It will automatically generate the convert file. You may need to
   look at the `api/v1alpha1/zz_generated.convert.chaosmesh.go` to make sure it
   works as expected, as this program is not guaranteed to work for all
   situations.
4. `./bin/chaos-multiversion-helper addoldobjs --version v1alpha1`. It will add
   the old version objects to the
   `cmd/chaos-controller-manager/provider/convert.go` to register the convert
   webhook for them.
5. Modify the graphql schema in `pkg/ctrl/server/schema.graphqls` from
   `v1alpha1` to `v1alpha2`, which can be achieved through `sed -i
   "s/v1alpha1/v1alpha2/g" pkg/ctrl/server/schema.graphqls`
6. Modify the version in `cmd/chaos-builder/version.go` from `v1alpha1` to
   `v1alpha2`
7. Modify the groupversion_info in api/v1alpha2 from `v1alpha1` to `v1alpha2`
8. Add old scheme to the `cmd/chaos-controller-manager/provider/controller.go`

The former commands should work with a GNU toolchain, and have been proved to
**NOT** work on Mac OS. If you are a developer with Mac OS environment, please 
execute these commands in dev-env docker image.

Or, you can use a single command, which will run the former tasks in the dev-env
docker environment:

```
OLD_VERSION="v1alpha1" NEW_VERSION="v1alpha2" make migrate-version
```

Before running this command, it's highly suggested to commit all your changes and start a new revision to migrate the version. If you faced any error during the process, you could clean the code base through `git add * && git reset --hard` and retry.

After running these steps, you'll need to modify the helm charts to run webhook for new versions:

1. Add new versions (e.g. "v1alpha2") to the `/helm/chaos-mesh/templates/mutating-admission-webhooks.yaml` and `/helm/chaos-mesh/templates/validating-admission-webhooks.yaml`