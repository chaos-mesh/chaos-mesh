# Release Guide

This document introduces how to publish a new release of Chaos Mesh.

## How to publish a new release

Here are several steps to publish a new release, and several of them are optional for bugfix/patch release.

### Step 1 Draft Release Notes

This step is required for all the releases(major, minor, bugfix/patch).

As we maintain a `CHANGELOG.md` on each active branch, we could use it to draft release notes. Changelog is used for developers and Release Notes is used for users, so there are some little different between them.

- Draft a Release Notes on google docs, here is a Release Notes template: [Google Docs](https://docs.google.com/document/d/1v0P5NQyepEyT4CH8usouyJup_fvOYtsYAz8nbJfn3Jk/edit?usp=sharing)
- Collect changelogs from `CHANGELOG.md` on the target branch into Release Notes. There are several patterns for mapping them:
  - Section `Add` in `CHANGELOG.md` is mapped to `New Features` in Release Notes.
  - Section `Changed` with actual "enhanced changes" in `CHANGELOG.md` is mapped to `Enhancements` in Release Notes.
  - Section `Fixed` in `CHANGELOG.md` is mapped to `Bug Fixes` in Release Notes.
  - For the rest of `CHANGELOG.md`, put them into `Others` in Release Notes.
- After finishing the draft, send it to [Chaos Mesh Committers and Maintainers](https://github.com/chaos-mesh/chaos-mesh/blob/master/MAINTAINERS.md), ask them to review it.
- Once 2 (or more) of committer/maintainers left "Approve" on the draft Release Note, it's ready for making a release on GitHub.

### Step 2 Checkout New Branch

This step is only required for major or minor version release. You should skip this step if you are going to release a bugfix/patch version.

Checkout a new branch with the name of `release-x.y` from `master` where `x.y` is the semantic version number of the release.

### Step 3 Update Versions in Helm Charts and install.sh

This step is required for all the releases(major, minor, bugfix/patch).

You should make a new PR for updating the version in helm charts and `install.sh`:

- `version` and `appVersion` in `helm/chaos-mesh/Chart.yaml`. Please notice that do NOT use prefix `v` in here.
- Docker image tags in helm charts:
  - After `2.1.0`, there is only one place need to change: `images.tag` in `helm/chaos-mesh/values.yaml`.
  - Before `2.1.0`, you should update the image tag for **each image**
- `version` in help messages of `install.sh`, like `Version of chaos-mesh, default value: <replace-with-version>`.
- Execute `make check` for updating the versions in generated files.

Then you could make a PR with above changes into `release-x.y` branch, here is an example: https://github.com/chaos-mesh/chaos-mesh/pull/2631

### Step 4 Update CHANGELOG.md on target branch

This step is required for all the releases(major, minor, bugfix/patch).

You should update the `CHANGELOG.md` on the target branch:

- rename `[Unreleased]` to `[x.y.z] - YYYY-MM-DD`, and slim empty changes types with only `-Nothing`.
- create new `[Unreleased]` at the top of `CHANGELOG.md`, with `- Nothing` placeholder in each type of changes.

For more detail, see RFC [keep-a-changelog](https://github.com/chaos-mesh/rfcs/blob/main/text/2022-01-17-keep-a-changelog.md#changelogmd-in-release--branches)

### Step 5 Create Release on GitHub

This step is required for all the releases(major, minor, bugfix/patch).

Draft a new Release on GitHub Release: https://github.com/chaos-mesh/chaos-mesh/releases/new with the Release Notes from Step 1, and choose the `release-x.y` branch for release/tag `vx.y.z`.

Please note that here the prefix `v` on the tag is required.

### Step 6 Build and Publish Docker Images

This step is required for all the releases(major, minor, bugfix/patch).

After `2.1.0`, a GitHub Action would automatically run after GitHub Release is published: https://github.com/chaos-mesh/chaos-mesh/actions?query=event%3Arelease

Before `2.1.0`, you should manually trigger a Jenkins Pipeline with several parameters: `tag` and `branch`. https://ci.pingcap.net/view/chaos-mesh/job/release_chaos_mesh/

### Step 7 Upload crd.yaml and install.sh to CDN

This step is required for all the releases(major, minor, bugfix/patch).

A GitHub Action would automatically run after GitHub Release is published: https://github.com/chaos-mesh/chaos-mesh/actions/workflows/upload_release_files.yml

### Step 8 Build Helm Charts

This step is required for all the releases(major, minor, bugfix/patch).

- Pull the latest code from `release-x.y` branch
- `git tag chart-x.y.z`
- `git push upstream chart-x.y.z`(`upstream` is the remote repo `github.com/chaos-mesh/chaos-mesh`)
- A GitHub Action would automatically run: https://github.com/chaos-mesh/chaos-mesh/actions/workflows/release_helm_chart.yml. And new helm artifact will be published to https://github.com/chaos-mesh/charts/tree/gh-pages.

### Step 9 Update TiChi Bot Configuration

This step is only required for major or minor version release. You should skip this step if you are going to release a bugfix/patch version.

If new branches are created, you should update the TiChi Bot configuration:

- Make a PR for setting up a new branch: https://github.com/ti-community-infra/configs/blob/992ff03161a42c0b517e4b4239adbf5f94e96a50/prow/config/config.yaml#L1105. Configuration for new branch could be found in existing settings for other branches.

### Step 10 Update CHANGELOG.md on master branch

This step is required for all the releases(major, minor, bugfix/patch).

Create a new PR for updating the `CHANGELOG.md` on master branch, you could copy the content from "Step 4".

## What should I do if any step failed

If some CI job failed, please profile the reason, and

- if the build environment is not stable, please give it another retry.
- if code changes are required to fix the issue, please ask any of the committers or maintainers for help.
- for other situations, please ask any of committers or maintainers for help.
