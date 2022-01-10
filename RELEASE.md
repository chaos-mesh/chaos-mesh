# Release Guide

This document introduces how to publish a new release of Chaos Mesh.

## How to publish a new release

Here are several steps to publish a new release, and several of them are optional for bugfix/patch relese.

### Step 1 Draft Release Notes

This step is required for all the release(major, minor, bugfix/patch).

- Draft a Release Notes on google docs, here is a Release Notes template: [Google Docs](https://docs.google.com/document/d/1v0P5NQyepEyT4CH8usouyJup_fvOYtsYAz8nbJfn3Jk/edit?usp=sharing)
- Collect changelogs from each commits on the target branch into Release Notes.
  - For Example, if you are going to release a new patch version `vx.y.z`, you should collect the commits on branch `release-x.y` which committed after tag `vx.y.z-1`.
  - If you are going to release a new major or minor version, you should collect the commits on branch `master`.
- After finishing the draft, send it to Chaos Mesh Committers and Maintainers, ask them to review it.
- Once 2 (or more) of committer/maintainers left "Approve" on the draft Release Note, it's ready for making a release on GitHub.

### Step 2 Checkout New Branch

This step is only required for major or minor version release. You should skip this step if you are going to release a bugfix/patch version.

Checkout a new branch with the name of `release-x.y` from `master` where `x.y` is the semantic version number of the release.

### Step 3 Update Versions in Helm Charts and install.sh

This step is required for all the release(major, minor, bugfix/patch).

You should make a new PR for updating the version in helm charts and `install.sh`:

- `version` and `appVersion` in `helm/chaos-mesh/Chart.yaml`. Please notice that do NOT use prefix `v` in here.
- Docker image tags in helm charts:
  - After `2.1.0`, there is only one place need to change: `images.tag` in `helm/chaos-mesh/values.yaml`.
  - Before `2.1.0`, you should update the image tag for **each images**
- `version` in help messages of `install.sh`, about line 46.
- Execute `make check` for update the versions in generated files.

Then you could make a PR with above changes into `release-x.y` branch, here is an example: https://github.com/chaos-mesh/chaos-mesh/pull/2631

### Step 4 Create Release on GitHub

This step is required for all the release(major, minor, bugfix/patch).

Draft a new Release on GitHub Release: https://github.com/chaos-mesh/chaos-mesh/releases/new with the Release Notes from Step 1, and choose the `release-x.y` branch for release/tag `vx.y.z`.

Please note that here requires prefix `v` on the tag.

### Step 5 Build and Publish Docker Images

This step is required for all the release(major, minor, bugfix/patch).

After 2.1.0, a GitHub Action would automatically run after GitHub Release is published: https://github.com/chaos-mesh/chaos-mesh/actions?query=event%3Arelease

Before 2.1.0, you should manually trigger a Jenkins Pipeline with several parameters: `tag` and `branch`. https://ci.pingcap.net/view/chaos-mesh/job/release_chaos_mesh/

### Step 6 Upload crd.yaml and install.sh to CDN

This step is required for all the release(major, minor, bugfix/patch).

A GitHub Action would automatically run after GitHub Release is published: https://github.com/chaos-mesh/chaos-mesh/actions/workflows/upload_release_files.yml

### Step 7 Build Helm Charts

This step is required for all the release(major, minor, bugfix/patch).

- Pull the latest code from `release-x.y` branch
- `git tag chart-vx.y.z`
- `git push upstream chart-vx.y.z`(`upstream` is the a remote points to `github.com/chaos-mesh/chaos-mesh`)
- A GitHub Action would automatically run: https://github.com/chaos-mesh/chaos-mesh/actions/workflows/release_helm_chart.yml. And new helm artifact will be published to https://github.com/chaos-mesh/charts/tree/gh-pages.

### Step 8 Update TiChi Bot Configuration

This step is only required for major or minor version release. You should skip this step if you are going to release a bugfix/patch version.

If new branches are created, you should update the TiChi Bot configuration:

- Make a PR for setting up new branch: https://github.com/ti-community-infra/configs/blob/992ff03161a42c0b517e4b4239adbf5f94e96a50/prow/config/config.yaml#L1105. Configuration for new branch could be found in existing settings for other branches.

## What should I do if any step failed

If some CI job failed, please profile the reason, and

- if the build environment is not stable, please take a retry.
- if it requires code changes to fix the issue, please ask any of committers or maintainers for help.
- for other situations, please ask any of committers or maintainers for help.
