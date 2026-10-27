# GHCR image cleanup

`cleanup_ghcr.py` removes old untagged container image versions from GHCR packages linked to `chaos-mesh/chaos-mesh`.
This prevents older build outputs from accumulating while retaining tagged images and their dependencies.

The script retains:

- Every tagged version.
- Every version created or updated within the last 30 days.
- Every manifest referenced directly or indirectly by those retained versions.

A tagged multi-platform image can depend on untagged manifests, so being untagged alone does not make a version eligible for deletion.

The [Clean GHCR images workflow](../workflows/cleanup_ghcr.yml) runs on the first day of every other month at 10:00 China Standard Time, starting in January.
You can also start it manually from GitHub Actions.
Both scheduled and manual runs delete eligible versions.

The workflow shares a concurrency group with the image publishing workflows.
The script deletes candidates selected by the initial scan and accepts a successful HTTP 204 response as deletion success.
It waits 2 seconds after each successful deletion.
It stops without retrying if a request fails or returns an invalid response.
Progress appears in the run logs, and a new run reads the current package inventory.
The logs show initial tagged and untagged counts, and counts after cleanup unless the package has no old untagged versions and is skipped.
The after counts use the initial inventory minus successful deletions, without an additional API scan.

Local execution requires Python 3, GitHub CLI, Docker with Buildx, and authenticated access to GitHub and GHCR.
The credentials need permission to read and delete package versions.
Run `python3 -B .github/scripts/cleanup_ghcr.py` from the repository root to perform the same cleanup locally.
