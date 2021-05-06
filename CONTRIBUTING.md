# Contributing to Chaos Mesh

Thanks for your interest in improving the project! This document provides a step-by-step guide for general contributions to Chaos Mesh.

## Communications

Before starting work on something major, please reach out to us via GitHub, Slack, email, etc. We will make sure no one else is already working on it and ask you to open a GitHub issue. Also, we will provide necessary guidance should you need it.

Specifically, if you want to develop a specific chaos type, you may also find [Development Guide](https://chaos-mesh.org/docs/development_guides/development_overview) useful.

## Submitting a PR

If you have a specific idea of a fix or update, follow these steps below to submit a PR:

- [Step 1: Make the change](#step-1-make-the-change)
- [Step 2: Run unit tests](#step-2-run-unit-tests)
- [Step 3: Start Chaos Mesh locally and do manual tests](#step-3-start-chaos-mesh-locally-and-do-manual-tests)
- [Step 4: Commit and push your changes](#step-4-commit-and-push-your-changes)
- [Step 5: Create a pull request](#step-5-create-a-pull-request)
- [Step 6: Get a code review](#step-6-get-a-code-review)

### Step 1: Make the change

1. Fork the Chaos Mesh repo, and then clone it:

   ```bash
   $ export user={your github. profile name}
   $ git clone git@github.com:${user}/chaos-mesh.git
   ```

2. Set your cloned local to track the upstream repository:

   ```bash
   $ cd chaos-mesh
   $ git remote add upstream https://github.com/chaos-mesh/chaos-mesh
   ```

3. Disable pushing to upstream master:

   ```bash
   $ git remote set-url --push upstream no_push
   $ git remote -v
   ```

   The output should look like:

   ```bash
   origin    git@github.com:$(user)/chaos-mesh.git (fetch)
   origin    git@github.com:$(user)/chaos-mesh.git (push)
   upstream  https://github.com/chaos-mesh/chaos-mesh (fetch)
   upstream  no_push (push)
   ```

4. Get your local master up-to-date and create your working branch:

   ```bash
   $ git fetch upstream
   $ git checkout master
   $ git rebase upstream/master
   $ git checkout -b myfeature
   ```

5. Make the change on the code.

   You can new edit the code on the `myfeature` branch.

   If you want to update the `crd.yaml` according the the CRD structs, run the following commands:

   ```bash
   $ make generate
   $ make manifests/crd.yaml
   ```

6. Check the code change by running the following command:

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

### Step 3: Start Chaos Mesh locally and do manual tests

1. Start a Kubernetes cluster locally. There are two options:

   - Use [kind](https://kind.sigs.k8s.io/docs/user/quick-start/#installation) to start a Kubernetes cluster locally and [kubectl](https://kubernetes.io/docs/reference/kubectl/overview/) to access the cluster. If you install these manually, run `kind`: `kind create cluster`to start the cluster.

   - Install the above dependencies in `~/local/bin` using [`install.sh`](https://github.com/chaos-mesh/chaos-mesh/blob/master/install.sh):

     ```bash
     $ ./install.sh --local kind --dependency-only
     ```

2. Make sure the installation in step 1 is successful:

   ```bash
   $ source ~/.bash_profile
   $ kind --version
   ...
   $ kubectl version
   ...
   ```

3. Install Chaos Mesh:

   Following command will rebuild project code and reinstall chaos mesh.

   ```bash
   $ ./hack/local-up-chaos-mesh.sh
   ```

Now you can test your code update on the deployed cluster.

### Step 4: Commit and push your changes

Congratulations! Now you have finished all tests and are ready to commit your code.

1. Run the following commands to keep your branch in sync:

   ```bash
   $ git fetch upstream
   $ git rebase upstream/master
   ```

2. Commit your changes:

   ```bash
   $ git add -A
   $ git commit --signoff
   ```

3. Push your changes to the remote branch:

   ```bash
   $ git push -f origin myfeature
   ```

### Step 5: Create a pull request

1. Visit your fork at <https://github.com/chaos-mesh/chaos-mesh> (replace the first chaos-mesh with your username).
2. Click the Compare & pull request button next to your `myfeature` branch.
3. Edit the description of the pull request to match your changes.

### Step 6: Get a code review

Once your pull request has been opened, it will be assigned to at least two reviewers. The reviewers will do a thorough code review of correctness, bugs, opportunities for improvement, documentation and comments, and style.

Commit changes made in response to review comments to the same branch on your fork.
