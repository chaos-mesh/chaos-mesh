# GHCR image cleanup

`cleanup_ghcr.py` verifies access to the fixed GHCR packages in its `PACKAGE_NAMES` tuple.
It queries each package directly, checks its name and association with `chaos-mesh/chaos-mesh`, and reads all active version pages.
The default command prints tagged and untagged counts for each package without inspecting manifests or deleting versions.

The [Verify GHCR package queries workflow](../workflows/cleanup_ghcr.yml) runs on the first day of every other month at 10:00 China Standard Time, starting in January.
It also runs for pull requests that change the script or workflow and supports manual runs from GitHub Actions.
All runs verify queries with the default `GITHUB_TOKEN` and `packages: read` permission.
The current workflow cannot delete versions and shares a concurrency group with the image publishing workflows.

Local query verification requires Python 3, GitHub CLI, and authenticated access to read the packages.
The command is `python3 -B .github/scripts/cleanup_ghcr.py` from the repository root.
The script stops without retrying if a request fails, a response is invalid, or a package does not match the expected repository.
It reports completion only after all configured packages pass verification.

## Explicit cleanup

The `--apply` option enables deletion.
Cleanup requires Docker with Buildx and authenticated access to GHCR.
In GitHub Actions, cleanup requires `packages: write` permission and package admin access.
Local cleanup with a classic personal access token requires `read:packages` and `delete:packages` scopes and package admin access.
The current workflow does not enable this option or grant these permissions.

With `--apply`, the script retains:

- Every tagged version.
- Every version created or updated within the last 30 days.
- Every manifest referenced directly or indirectly by those retained versions.

A tagged multi-platform image can depend on untagged manifests, so being untagged alone does not make a version eligible for deletion.

The script deletes candidates selected by the initial scan and accepts a successful HTTP 204 response as deletion success.
It waits 2 seconds after each successful deletion.
It stops without retrying if a request fails or returns an invalid response.
Progress appears in the run logs, and a new run reads the current package inventory.
The logs show initial tagged and untagged counts, and counts after cleanup unless the package has no old untagged versions and is skipped.
The after counts use the initial inventory minus successful deletions, without an additional API scan.
