# Contributing to Chaos Mesh 

Thanks for your help improving the project! 

## Getting Help 

If you have a question about Chaos Mesh or have encountered problems using it, [raise an issue in Github](https://github.com/pingcap/chaos-mesh/issues) or join us the #sig-chaos-mesh channel in the [TiDB Community](https://pingcap.com/tidbslack) slack workspace..

## Submitting a Pull Request                                                                             

### Step 1: Fork Chaos Mesh on Github
1. Visit https://github.com/pingcap/chaos-mesh
2. Click `Fork` button(top) to establish a cloud-based fork. 

### Step 2: Get Chaos Mesh repo

Create your clone: 
```bash
$ export user={your github. profile name}
$ git clone https://github.com/${user}/chaos-mesh.git
```

Set your clone to track upstream repository:
```bash
$ cd chaos-mesh
$ git remote add upstream https://github.com/pingcap/chaos-mesh
```

### Step 3: Create a new branch

Get your local master up to date:
```bash
$ git fetch upstream
$ git checkout master 
$ git rebase upstream/master
```

Create branch: 
```bash
$ git checkout -b myfeature
```

### Step 4: Develop

You can new edit the code on the `myfeature` branch.

If you want to update the `crd.yam` according the the CRD structs, run the following commands: 

```bash
$ make generate
$ make yaml
```

### Step 5: Check the code

Run following command to check your code change: 
```bash
$ make check
```

This will show errors if your code change does not pass checks(eg: fmt, lint), Please fix them before submitting the PR.

#### Run unit tests

Before running your code in a real Kubernetes cluster, make sure it passes all unit tests: 
```bash
$ make ensure-kubebuilder # install some test dependencies
$ make test
```

#### Run e2e test

At first, you must have [Docker](https://www.docker.com/get-started/) installed and running.

Now you can run the following command to run all e2e test: 
```bash
$ ./hack/e2e.sh
```

It's possible to limit specs to run, for example:

```bash
$ ./hack/e2e.sh -- --ginkgo.focus='Basic'
```

Run the following command to see help:
```bash
$ ./hack/e2e.sh -h
```

#### Start Chaos Mesh locally and do manual tests

We uses [kind](https://kind.sigs.k8s.io/docs/user/quick-start/#installation) to start a Kubernetes custer locally and and [kubectl](https://kubernetes.io/docs/reference/kubectl/overview/) must be installed to access Kubernetes cluster.

You can refer to their official references to install them on your machine, and then you need to create a Kubernetes cluster with `kind`:
```bash
$ kind create cluster
```

You also can run the following command to install these dependencies by [`install.sh`](https://github.com/pingcap/chaos-mesh/blob/master/install.sh) in local binary directory `~/local/bin`:

```bash
# this command will install kind, kubectl, and a local Kubernetes cluster.
$ ./install.sh --local kind --dependency-only
```

Make sure they are installed correctly:

```bash
$ source ~/.bash_profile
$ kind --version 
...
$ kubectl version 
...
```

Build image: 
```bash
$ make image-chaos-mesh
$ make image-chaos-dahsboard
$ make image-chaos-daemon
# or build all images
$ make image
```

Load image into Kubernetes nodes: 
```bash
$ kind load docker-image pingcap/chaos-mesh:latest 
$ kind load docker-image pingcap/chaos-dashboard:latest 
$ kind load docker-image pingcap/chaos-daemon:latest 
```

Deploy Chaos Mesh:
```bash
$ ./install.sh --runtime containerd
```

### Step 6: Commit your changes

Run the following commands to keep your branch in sync: 

```bash
$ git fetch upstream
$ git rebase upstream/master
```

Commit your changes: 
```bash
$ git add -A
$ git commit
```
Likely you'll go back and edit/build/test some more than commit --amend in a few cycles.

### Step 7: Push your changes

When your commit is ready for review (or just to establish an offsite backup of your work), push your branch to your fork on `github.com`:
```bash
$ git push -f origin myfeature
```

### Step 8: Create a pull request

1. Visit your fork at https://github.com/$user/chaos-mesh (replace $user obviously).
2. Click the Compare & pull request button next to your `myfeature` branch.
3. Edit the description of the pull request to match your changes.

### Step 9: Get a code review

Once your pull request has been opened, it will be assigned to at least two reviewers. Those reviewers will do a thorough code review, looking for correctness, bugs, opportunities for improvement, documentation and comments, and style.

Commit changes made in response to review comments to the same branch on your fork.

Very small PRs are easy to review. Very large PRs are very difficult to review.


