# de46fee8 — Add a GitHub Actions workflow to trigger the website pages build

- Commit: de46fee8
- Verdict: SAFE

## What it changes
Adds a new GitHub Actions workflow file `.github/workflows/trigger-pages.yml`.
On every push to `master` of the `irssi/irssi` repo, it dispatches the
`pages.yml` workflow in the `irssi.github.io` repo (ref `main`) using
`actions/github-script@v6` authenticated with the `PAT_TOKEN` secret.

## Files touched
- `.github/workflows/trigger-pages.yml` (new)

## Release-safety findings
None — safe to release. This is CI infrastructure only and touches no
compiled/shipped code, build system, or runtime behavior. It cannot affect
the released artifact, ABI, or end users.

The `if: github.repository == 'irssi/irssi'` guard prevents forks from
running it. The script is static (no untrusted input interpolated into the
`script:` body), so there is no command/expression injection vector. The
PAT_TOKEN is referenced via the `secrets` context, not echoed. Worst case of
misconfiguration (missing/invalid token) is a failed CI job, not a broken
release.

## Interactions with related changes
None.

## NEWS accuracy
Accurate. The NEWS line describes adding a workflow to trigger the website
pages build, which matches the diff.
