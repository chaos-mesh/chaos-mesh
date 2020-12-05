# chaosctl

Chaostl is a tool (currently) used to print debug info. Maintainers would make use of it to better provide their suggestions for issues.

## How to build
```shell
cd $CHAOSMESH_DIR
make chaosctl
./bin/chaosctl --help #to see if build succeed.
```
Chaoctl support shell autocompletion, which could save you some typing. Do `./bin/chaosctl completion -h` for detail.

## How to use
**Debug**
`chaoctl debug` is used to print debug info of certain chaos. Currently, chaosctl support networkchaos, stresschaos and iochaos.
```shell
#to print info of each networkchaos
./bin/chaosctl debug networkchaos
#to print info of certain chaos in default namespace
./bin/chaosctl debug networkchaos CHAOSNAME
#to print info of each networkchaos in certain namespace
./bin/chaosctl debug networkchaos -n NAMESPACE
```

**Logs**
`chaoctl log` is used to easily print log from all chaos-mesh components, including controller-manager, chaos-daemon and chaos-dashboard.
```shell
# Default print all log of all chaosmesh components
chaosctl logs

# to print 100 log lines for chaosmesh components in node NODENAME
chaosctl logs -t 100 -n NODENAME
```
