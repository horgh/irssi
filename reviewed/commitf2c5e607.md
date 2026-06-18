# f2c5e607 — Revert the Neirssi branding change back to Irssi in the README and startup banner

- Commit: f2c5e607
- Verdict: SAFE

## What it changes
Reverts the "Neirssi" branding (introduced in 2cdcf861) back to "Irssi":

- NEWS: top changelog header text restored from `v1.5-head-an 2022-xx-xx` to `v1.5-head 202x-xx-xx`.
- README.md: title, build-status badge URL, prose, clone/release/download/docs/modules links changed from `ailin-nemui` forks/pages back to `irssi/irssi` and `irssi.org`.
- src/fe-text/module-formats.c: the `irssi_banner` startup format string changed from the single-line `"Neırssi v$J"` back to the 5-line ASCII-art block ending in `"Irssi v$J - https://irssi.org"`.
- utils/make-dist.sh: the generated `setup.cfg` `url` and `maintainer` metadata fields changed from `ailin-nemui.github.io` / `Ailin Nemui` back to `https://irssi.org/` / `The Irssi team`.

## Files touched
- NEWS
- README.md
- src/fe-text/module-formats.c
- utils/make-dist.sh

## Release-safety findings
None — safe to release.

The only compiled change is the `irssi_banner` format string. I verified:
- The restored banner is byte-for-byte identical to the pre-branding original (the parent of 2cdcf861).
- The `$J` expando used in the string is a registered version expando (`src/core/expandos.c:649`, `expando_create("J", expando_version, ...)`), so it renders correctly.
- The FORMAT_REC keeps parameter count `0` (only the `$J` expando, no positional `$N` args), consistent with the original and with neighboring entries; `%:` is the standard newline separator. No change to the number of format entries, so `TXT_COUNT` / the `module-formats.h` enum stay aligned (no ABI/format-table mismatch).

The make-dist.sh edits are static literal strings inside a quoted heredoc context; they are not derived from user input, so there is no shell/template injection concern. README and NEWS edits are documentation/metadata text only.

## Interactions with related changes
None. This is a self-contained revert of the 2cdcf861 branding commit. NEWS, README, the banner string, and the dist metadata are independent; no functional code path depends on the changed text.

## NEWS accuracy
Accurate. The change reverts the README and startup banner branding back to Irssi as described. (The NEWS line is itself one of the reverted items; the description's scope matches the diff. The README also incidentally diverges from the exact pre-branding text in unrelated spots, e.g. Meson/Ninja version mentions, but those predate this commit and are not introduced by it.)
