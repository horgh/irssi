# irssi 1.5.0-beta1 — release-safety review summary

Reviewed all **79** NEWS entries (one subagent per change, read-only, diffing the
exact commit that landed in master). Full per-change reviews are in `reviewed/<label>.md`;
the checklist is `todo.md`.

**Verdict tally:** 0 BLOCKER · **3 NEEDS_ATTENTION** · **7 MINOR** · 69 SAFE.

Nothing is an outright "do-not-ship" blocker, but three items deserve a decision
before release, and the resolver rewrite (#1580) carries the one genuinely serious
runtime bug found in the whole set.

---

## NEEDS_ATTENTION (decide before release)

### 1. #1580 — GIO resolver migration  → `reviewed/pr1580.md`  ⚠️ highest priority
The fork()-based resolver was replaced with GLib's GResolver. Two defects survive at
master HEAD even after its follow-up fixes (#1605, #1607, and 88afc78a/51322d6a):

- **Use-after-free of `SERVER_REC` on a cancelled DNS lookup.** The async resolve
  stores the raw `server` pointer as callback data without taking a ref. On
  `/disconnect`, a connect timeout, or `server_connect_failed()` while a lookup is
  still in flight, the code unrefs and frees the server, but GResolver still dispatches
  the async callback afterward (`G_IO_ERROR_CANCELLED`), which dereferences the freed
  server. Reachable on slow/flaky DNS — i.e. realistic, not theoretical. The old
  fork-based design avoided this by removing the IO source synchronously. **This is the
  most serious finding in the review.** Fix: hold a ref across the async op (release in
  the continuation) or add a cancellation guard in `net_gethostbyname_callback`.
- **FreeBSD/Capsicum build break — confirmed by me.** `net_gethostbyname` is now
  `static` in `network.c` with a new 2-arg signature, but `capsicum.c:282`
  (`symbiont_gethostbyname`) still calls the old `net_gethostbyname(addr, &ip4, &ip6)`.
  The symbol is no longer declared in `network.h` and is static elsewhere, so
  `capsicum.c` will not compile when `have_capsicum` is set. Default builds are fine;
  the FreeBSD/Capsicum configuration is broken.

Behavior changes worth a NEWS line: `resolve_prefer_ipv6` is **removed** for server
connections (not renamed — the new `irssiproxy_prefer_ipv6` only covers the proxy bind
path); address selection changed from per-attempt randomization to resolver-ordered
sequential failover with a cached host list across reconnects (affects DNS round-robin
/ failover). Also bundles an ABI bump (56→57), GLib 2.74.7 + new gio/gobject deps, and
a `glib_log_domains` default change — none mentioned in NEWS.

### 2. an#42 — IRC protocol module autoloading  → `reviewed/an42.md`
A much larger change than NEWS implies: IRC (and dcc/flood/notifylist) are no longer
statically linked into the binary — they load **only** via the `autoload_modules`
setting, whose default changed from `"perl otr"` to `"irc dcc flood notifylist perl otr"`.

- **Upgrade hazard:** any user who customized `autoload_modules` (common) keeps their
  saved value, so on first 1.5.0 run `irc` won't load — a client with no IRC support
  and no clear error. No migration shim. Needs a release note / migration handling.
- Latent NULL function-pointer regression: the dummy "unknown" protocol now leaves
  `create/destroy_server_connect` NULL (were safe stubs). Guarded on default paths;
  reachable only via unload-module-then-reconnect. Minor, but a regression vs. before.
- ABI bumped 44→45 (correct). NEWS understates scope (ABI bump + default change).

### 3. #1546 — CI pinning + make-dist.sh change  → `reviewed/pr1546.md`
Mostly CI-only, **but** it also rewrote the `tar --delete` step in `utils/make-dist.sh`
(the release-tarball builder). The new `grep -F` approach can enumerate both the
`.egg-info/` directory member and its children; GNU `tar --delete` cascade-removes the
directory, then errors on the now-missing children and exits 2, which under `set -e`
aborts `make-dist.sh` → no tarball. The reviewer couldn't install setuptools in-sandbox
to confirm the production sdist layout. **Action:** run `./utils/make-dist.sh` (or check
a recent `dist` CI run) and confirm it produces `irssi-*.tar.gz` and exits 0. If the
dist job has been green, this is latent fragility; if not exercised since this change,
the tarball step is broken. NEWS omits the make-dist.sh change entirely.

---

## MINOR (not release blockers)

Recurring theme — **small, bounded, server- or user-triggerable memory leaks on
fe-common/irc display paths** (no crash/UB; each leaks a few bytes per event):
- **an#81** (`reviewed/an81.md`) — `nickmode` not freed in the public-notice branch.
- **#1457** (`reviewed/pr1457.md`) — `params` not freed on the non-quiet 344 early
  return; plus a cosmetic swapped date/relative-time in the `ebanlist_long` format.
- **#1525** (`reviewed/pr1525.md`) — `my_asctime` result not freed in the `/BAN` list
  loop. (The PR's actual purpose — fixing a real `/BAN` listing crash — is correct.)

Other MINOR:
- **#1519** (`reviewed/pr1519.md`) — OpenSSL 3 path: `char cname[50]` used without
  checking `EVP_PKEY_get_group_name()`'s return; on failure it reads uninitialized
  stack into the TLS info string. Latent (real ECDH groups always succeed); a one-line
  init/return-check fixes it.
- **#1500** (`reviewed/pr1500.md`) — SCRAM SASL: no crash/auth-bypass found and it fails
  closed; but the username isn't SCRAM-escaped (usernames with `,`/`=` silently break
  auth), and a `&&`-vs-`||` prefix check is ineffective (harmless — still fails closed).
  Installed `irc-servers.h` now transitively includes `<openssl/evp.h>`.
- **#1518** (`reviewed/pr1518.md`) — module-load error-reporting regression: ABI-mismatch
  / dlopen failures now also run the `_core` prefix fallback, producing misleading or
  duplicated error messages (common when upgrading with stale third-party modules).
  Cosmetic; valid modules load fine.
- **an#69** (`reviewed/an69.md`) — always-on 24-bit color: runtime behavior preserved
  (still gated by `colors_ansi_24bit`, default off). CI-only side effect: `abicheck.yml`
  still passes the removed `-Denable-true-color=yes` and will fail until updated.

---

## Cross-cutting notes

- **CI workflows reference removed meson options.** `an#69` removed `enable-true-color`;
  `abicheck.yml` still passes `-Denable-true-color=yes` (with `meson<0.59` pinned, an
  unknown `-Doption` is a hard error). Doesn't affect the shipped product, but the
  abicheck job will fail until fixed. Worth a sweep of the workflows for stale options.
- **The resolver feature spans 4 commits** (#1580 + 51322d6a/e9281b2f + #1605 + #1607).
  All required follow-ups are in master, but the UAF and Capsicum break are *not* fixed
  by them.
- **NEWS accuracy:** an#42, #1546, and #1580 materially understate their changes
  (ABI bumps, default-setting changes, dependency bumps, the make-dist.sh change).
  Consider expanding those NEWS lines.

## The 69 SAFE changes
The remaining 69 (build/portability fixes, CI updates, doc/help/typo fixes, and the
bulk of the bug fixes incl. the crash fixes #1405/#1450/#1491 and the netsplit/color
fixes) were reviewed and found to carry no release-safety concern. See their individual
`reviewed/<label>.md` files.
