# Contributing to Chaos Mesh

Thank you for helping improve Chaos Mesh. This guide covers the basic contribution workflow; use the directory-specific README files for implementation details.

All contributors must follow the [Code of Conduct](CODE_OF_CONDUCT.md).

## Before you start

For a bug fix or small improvement, you can open a pull request directly. For a new feature, new chaos type, public API change, or other substantial work, open or comment on a GitHub issue first so the design and ownership can be agreed before implementation.

The [Chaos Mesh developer guide](https://chaos-mesh.org/docs/developer-guide-overview/) provides additional background for developing chaos types and setting up a development environment.

## Set up your fork

Fork the repository, clone your fork, add the upstream repository, and create a branch from the latest `master`:

```bash
git clone git@github.com:YOUR_GITHUB_USERNAME/chaos-mesh.git
cd chaos-mesh
git remote add upstream https://github.com/chaos-mesh/chaos-mesh.git
git fetch upstream
git switch --create BRANCH_NAME upstream/master
```

Keep each pull request focused on one problem and preserve unrelated changes in your working tree.

## Read the relevant guide

Start with the documentation closest to the code you plan to change:

| Area | Guide |
| --- | --- |
| Project overview | [README.md](README.md) |
| Controller architecture, reconciliation, and chaos implementations | [controllers/README.md](controllers/README.md) |
| Executable entry points and repository generators | [cmd/README.md](cmd/README.md) |
| Helm chart configuration and development | [helm/chaos-mesh/README.md](helm/chaos-mesh/README.md) |
| Dashboard frontend | [ui/README.md](ui/README.md) |
| Integration tests | [test/integration_test/README.md](test/integration_test/README.md) |

README files in subdirectories contain more focused design and lifecycle details. Follow existing package boundaries and patterns before introducing a new abstraction.

## Make and verify the change

Add or update tests with the implementation. During development, run the narrowest relevant tests, builds, linters, or Helm rendering commands described by the area's README instead of repeatedly running every repository-wide check.

For focused Go work, run tests from the module that owns the package:

```bash
go test ./path/to/package
```

Do not edit generated files directly. Change their source and run the corresponding generator. For example, `make generate` refreshes generated API code, clients, CRDs, the combined CRD manifest, and related artifacts after an API or CRD change.

Use the broad checks when the change requires them and for final validation when practical:

```bash
make test
make check
```

These targets can regenerate or reformat tracked files and may build the containerized development environment. Inspect `git status` and `git diff` afterward, and keep only changes related to your contribution.

Do not run end-to-end or fault-injection tests unless you have a suitable disposable Kubernetes cluster and the change requires them.

## Submit a pull request

Review the final diff, stage only the intended files, and sign off every commit for DCO compliance:

```bash
git status --short
git diff --check
git add CHANGED_FILES
git commit --signoff
git push --set-upstream origin BRANCH_NAME
```

Follow [Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/) for commit messages and the pull request title. Complete the [pull request template](.github/pull_request_template.md), link the relevant issue, describe how the change was verified, and update `CHANGELOG.md` when required by the template.

During review, push follow-up commits to the same branch. Keep the branch up to date with `upstream/master` when necessary, and avoid mixing unrelated cleanup into the pull request.

## Getting help

Use GitHub issues for bugs and feature discussions. For broader development questions, see the community channels listed in the [project README](README.md#community).
