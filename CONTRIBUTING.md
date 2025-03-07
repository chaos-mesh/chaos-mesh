# Contributing to Chaos Mesh

Thanks for your interest in improving the project! This document provides a step-by-step guide for general contributions to Chaos Mesh.

## Communications

Before starting work on something major, please reach out to us via GitHub, Slack, email, etc. We will make sure no one else is already working on it and ask you to open a GitHub issue. Also, we will provide necessary guidance should you need it.

Specifically, if you want to develop a specific chaos type, you may also find [Development Guide](https://chaos-mesh.org/docs/developer-guide-overview) useful.

## Submitting a PR

If you have a specific idea of a fix or update, follow these steps below to submit a PR:

### Table of Contents

- [Step 1: Make changes](#step-1-make-changes)
- [Step 2: Run unit tests](#step-2-run-unit-tests)
- [Step 3: Start Chaos Mesh locally and do manual tests](#step-3-start-chaos-mesh-locally-and-do-manual-tests)
- [Step 4: Commit and push your changes](#step-4-commit-and-push-your-changes)
- [Step 5: Create a pull request](#step-5-create-a-pull-request)
- [Step 6: Get a code review](#step-6-get-a-code-review)

### Step 1: Make changes

1. Fork the Chaos Mesh repo, and then clone it:

   ```bash
   git clone git@github.com:your-github-username/chaos-mesh.git
   ```

2. Set your cloned local to track the upstream repository:

   ```bash
   cd chaos-mesh
   git remote add upstream https://github.com/chaos-mesh/chaos-mesh
   ```

3. Disable pushing to upstream master:

   ```bash
   git remote set-url --push upstream no_push
   git remote -v
   ```

   The output should look like:

   ```bash
   origin    git@github.com:your-github-username/chaos-mesh.git (fetch)
   origin    git@github.com:your-github-username/chaos-mesh.git (push)
   upstream  https://github.com/chaos-mesh/chaos-mesh (fetch)
   upstream  no_push (push)
   ```

4. Get your local master up-to-date and create your working branch:

   ```bash
   git fetch upstream
   git checkout master
   git rebase upstream/master
   git checkout -b new-feature
   ```

5. Make the change on the code.

   You can now edit the code on the `new-feature` branch.

   If you want to update the `crd.yaml` according to the CRD structs, run the following commands:

   ```bash
   make generate
   make manifests/crd.yaml
   ```

6. Check the code change by running the following command:

   ```bash
   make check
   ```

This will show errors if your changes do not pass the check (e.g. fmt, lint). Please fix them before submitting the PR.

### Step 2: Run unit tests

Before running your code in a real Kubernetes cluster, make sure it passes all unit tests:

```bash
make test
```

### Step 3: Start Chaos Mesh locally and do manual tests

Referring to the [Configure the Development Environment](https://chaos-mesh.org/docs/configure-development-environment/).

Now you can test your changes on the deployed cluster.

### Step 4: Commit and push your changes

Congratulations! Now you have finished all tests and are ready to commit your code.

1. Run the following commands to keep your branch in sync:

   ```bash
   git fetch upstream
   git rebase upstream/master
   ```

2. Commit your changes:

   ```bash
   git add -A
   git commit --signoff
   ```

3. Push your changes to the remote branch:

   ```bash
   git push origin new-feature
   ```

### Step 5: Create a pull request

Please follow the pull request template when creating a pull request.

### Step 6: Get a code review

Once your pull request has been opened, it will be assigned to at least two reviewers. The reviewers will do a thorough code review of correctness, bugs, opportunities for improvement, documentation and comments, and style.

Commit changes made in response to review comments to the same branch on your fork.
