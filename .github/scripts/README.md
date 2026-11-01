# GHCR image cleanup

`cleanup_ghcr.py` queries the packages in `PACKAGE_NAMES` and checks their association with `chaos-mesh/chaos-mesh`.
Cleanup deletes old untagged versions while retaining:

- Every tagged version.
- Every version created or updated within the last 30 days.
- Every manifest referenced directly or indirectly by those retained versions.

Run `python3 -B .github/scripts/cleanup_ghcr.py` to verify queries without deletion.
Add `--apply` to delete eligible versions.

The [workflow](../workflows/cleanup_ghcr.yml) uses the default `GITHUB_TOKEN`.
PR checks only verify queries with `packages: read`.
Scheduled and manual runs perform cleanup with `packages: write`.
The schedule is 10:00 China Standard Time on the first day of every other month, starting in January.

Local use requires Python 3, GitHub CLI, and package read access.
Cleanup also requires Docker with Buildx, GHCR authentication, and package admin access.
Classic personal access tokens need `read:packages` and `delete:packages` scopes for cleanup.
