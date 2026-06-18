# 9f71a391 — Add an OBS workflow trigger

- Commit: 9f71a391
- Verdict: SAFE

## What it changes
Adds a new file `.obs/workflows.yml` defining an openSUSE Build Service (OBS) SCM/CI workflow. On a `pull_request` event it branches the `irssi-git-an` package into the `home:ailin_nemui:CI` project and enables the publish flag. This is external build-service automation config only.

## Files touched
- `.obs/workflows.yml` (new, 13 lines)

## Release-safety findings
None — safe to release. The file is consumed only by the external OBS service and is not referenced by the build system, source, or release tarball. It cannot break compilation, runtime, or packaging. The referenced projects are the author's personal OBS namespaces; no secrets are embedded and there is no injection surface.

## Interactions with related changes
None.

## NEWS accuracy
Accurate. The change adds an OBS workflow trigger as described.
