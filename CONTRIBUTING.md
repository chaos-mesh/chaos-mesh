# Contributing to Chaos Mesh 

Thanks for your interest in improving the project! This document provides a step-by-step guide for general contributions to Chaos Mesh. 

## Communications

Before starting work on something major, please reach out to us via GitHub, Slack, email, etc. We will make sure no one else is already working on it and ask you to open a GitHub issue. Also, we will provide necessary guidance should you need it.

Specifically, if you want to develop a specific chaos type, you may also find [Development Guide](https://chaos-mesh.org/docs/development_guides/development_overview) useful.

## Submitting a PR

If you have a specific idea of a fix or update, follow these steps below to submit a PR:

### Step 1: Make the change

1. Fork the Chaos Mesh repo, and then clone it to your local:
  
```bash
$ export user={your github. profile name}
$ git clone https://github.com/${user}/chaos-mesh.git
```

2. Set your cloned local to track the upstream repository:

```bash
$ cd chaos-mesh
$ git remote add upstream https://github.com/pingcap/chaos-mesh
```

3. Get your local master up-to-date and create your working branch:

```bash
$ git fetch upstream
$ git checkout master 
$ git rebase upstream/master
$ git checkout -b myfeature
```

4. Make the change on the code 

You can new edit the code on the `myfeature` branch.

If you want to update the `crd.yam` according the the CRD structs, run the following commands: 

```bash
$ make generate
$ make yaml
```

5. Check the code change by running the following command:

```bash
$ make check
```

This will show errors if your code change does not pass the check. (eg: fmt, lint). Please fix them before submitting the PR.

### Step 2: Run unit tests

Before running your code in a real Kubernetes cluster, make sure it passes all unit tests:

```bash
$ make ensure-kubebuilder # install some test dependencies
$ make test
```

### Step 3: Run e2e test

Chaos Mesh code must pass e2e test before being submitted. Before started, you must have [Docker](https://www.docker.com/get-started/) installed and running.

Run the following command to run all e2e test:

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

### Step 4: Start Chaos Mesh locally and do manual tests

1. Start a Kubernetes cluster locally. There are two options:

  - Use [kind](https://kind.sigs.k8s.io/docs/user/quick-start/#installation) to start a Kubernetes cluster locally and and [kubectl](https://kubernetes.io/docs/reference/kubectl/overview/) to access the cluster. If you install these manually, run `kind`: `kind create cluster`to start the cluster.
  
  -  Install the above dependencies in `~/local/bin` using [`install.sh`](https://github.com/pingcap/chaos-mesh/blob/master/install.sh):
  
  $ ./install.sh --local kind --dependency-only

2. Make sure the installation in step 1 is successful:

```bash
$ source ~/.bash_profile
$ kind --version 
...
$ kubectl version 
...
```

3. Build the image:

```bash
$ make image-chaos-mesh
$ make image-chaos-dahsboard
$ make image-chaos-daemon
# or build all images
$ make image
```

4. Load image into Kubernetes nodes:

```bash
$ kind load docker-image pingcap/chaos-mesh:latest 
$ kind load docker-image pingcap/chaos-dashboard:latest 
$ kind load docker-image pingcap/chaos-daemon:latest 
```

5. Deploy Chaos Mesh:

```bash
$ ./install.sh --runtime containerd
```

Now you can test your code update on the deployed cluster.

### Step 5: Commit and push your changes

Congratulations! Now you have finished all tests and are ready to commit your code. 

1. Run the following commands to keep your branch in sync: 

```bash
$ git fetch upstream
$ git rebase upstream/master
```

2. Commit your changes: 
```bash
$ git add -A
$ git commit
```
 
3. Push your changes to the remote branch:

```bash
$ git push -f origin myfeature
```

### Step 6: Create a pull request

1. Visit your fork at https://github.com/$user/chaos-mesh (replace $user with your name).
2. Click the Compare & pull request button next to your `myfeature` branch.
3. Edit the description of the pull request to match your changes.

### Step 7: Get a code review

Once your pull request has been opened, it will be assigned to at least two reviewers. The reviewers will do a thorough code review, looking for correctness, bugs, opportunities for improvement, documentation and comments, and style.

Commit changes made in response to review comments to the same branch on your fork.



