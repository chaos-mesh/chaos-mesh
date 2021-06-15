## Issues concerning Chaos-mesh 2.0 upgrade
Due to controller refactor and CRD modification, old chaos resource (1.x) won't work with chaos-mesh 2.0, and you have to manually upgrade the CRD since helm won't do it for you. There are several steps for upgrading.

1. Export old chaos resource (If you don't care for them, you can skip this step.)
2. Delete the old CRDs and apply new ones.
3. Upgrade chaos-mesh to 2.0
4. Upgrade the old CRDs and reapply them.

We provide a bash script to do step 1,2 and 4. Check it out.

``` bash
./migrate.sh -h
```

You can also upgrade old yaml manually.

``` bash
go build main.go
./main
```
