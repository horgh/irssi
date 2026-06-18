* Document the Perl version() and parse_special() functions and the
  $abiversion special variable. This was incorrectly listed in the 1.4.1
  NEWS but was not actually present until now (an#38, 9f9d49a2). By bw1.
  * Only docs changes.
* Load the IRC protocol and its dcc, flood and notifylist submodules as
  dynamically loaded modules instead of linking them into the binary, and
  read their config once the chat protocol is created. The default
  autoload_modules is now "irc dcc flood notifylist perl otr"; setups that
  have customised autoload_modules must add "irc dcc flood notifylist" to
  keep IRC support (an#42, d9b181e4). By Ailin Nemui.
  * Seems like a substantial change.
* Remove the autotools build system in favour of meson (an#64, 58efc637). By
  Ailin Nemui.
* Remove the enable-true-color compile-time switch and always enable 24-bit
  true color (an#69, 3a5f93bb). By Ailin Nemui.
* Remove deprecated compile-time build options such as PERL_STATIC_LIBS,
  HAVE_SOCKS and USE_GREGEX (an#73, 4115e474). By Ailin Nemui.
* Add some missing files to make-dist (an#74, 4bf12908). By Ailin Nemui.
* Simplify the termux GitHub Actions workflow (an#75, 4986a4f9). By Ailin
  Nemui.
* Abort the meson build (and bump the required meson version) when run in an
  autotools-configured tree (an#80, ed4ec313). By Ailin Nemui.
* Fix bound IRC queries not being restored on reconnect (an#79, ae5a9283). By
  Ailin Nemui.
* Add a new PUBNOTICES message level to separate public channel notices from
  private notices (an#81, e300fee9). By Ailin Nemui.
* Add an OBS workflow trigger (9f71a391). By Ailin Nemui.
* Fix /names sending a raw server event by binding it first (an#87, 7285da27).
  By Ailin Nemui.
* Revert the Neirssi branding change back to Irssi in the README and startup
  banner (f2c5e607). By Ailin Nemui.
* Add a GitHub Actions workflow to trigger the website pages build (de46fee8).
  By Ailin Nemui.
* Speed up the Termux package CI by no longer building the perl package
  separately (#1394). By Ailin Nemui.
* Fix a crash when loading a server set up without a chatnet (#1405). By Ailin
  Nemui.
* Exit gracefully on SIGTERM by emitting "gui exit" and quitting (#1403). By
  Andreas Lundin.
* Add clearer error messages to the config parser and internal groundwork for
  list-of-lists (not yet enabled), and fix a time zone assertion by parsing
  ISO 8601 times as UTC (#1430). By Ailin Nemui.
* Build the Termux package from a git source URL (#1440). By Ailin Nemui.
* Update GitHub Actions workflows for deprecated set-output and artifact
  actions (#1441). By Ailin Nemui.
* Prefer IPv6 by default when resolving. The resolve_prefer_ipv6 setting this
  changed was later removed by the GIO resolver migration, which now leaves
  server address order to the resolver (#1445). By jesopo.
* Fix a crash on connect during startup when no chat protocol is available
  (#1450). By Ailin Nemui.
* Update the GitHub Actions workflows to a newer Ubuntu runner image (#1463).
  By Ailin Nemui.
* Properly format listmodes and their timestamps, add human-readable relative
  times, and support ircd-hybrid quiet lists (#1457). By David Schultz.
* Add a command_history_editable setting to enable bash-like editing of
  command history entries (#1486). By Jari Matilainen.
* Fix a crash when adding a server (#1491). By Ailin Nemui.
* Add SET window_default_ to the "See also" section of the window help
  (#1502). By Ailin Nemui.
* Improve the bind help text to better explain upper- and lowercase key usage
  (#1505). By Jari Matilainen.
* Document pkg-config as a build dependency required by meson to find glib
  (#1509). By Ailin Nemui.
* Restore the locale if Perl breaks it (#1510). By Ailin Nemui.
* Fix typos in the scrollback help text (#1511). By Gunter Labes.
* Add a missing include to fix the fuzz build under clang-18 (#1515). By
  maflcko.
* Fix missing shell quotes in the irssi-version.sh helper script (#1512). By
  Ailin Nemui.
* Add support for the SCRAM-SHA-1, SCRAM-SHA-256 and SCRAM-SHA-512 SASL
  mechanisms (#1500). By patrick-irc.
* Update the minimum required Perl version in the README (#1520). By Ailin
  Nemui.
* Update GitHub Actions to newer Node.js versions (#1521). By Ailin Nemui.
* Fix a glib deprecation in module loading, which may allow loading Apple
  dylibs with new enough glib (#1518). By Ailin Nemui.
* Replace a deprecated OpenSSL 3 function (#1519). By Ailin Nemui.
* Ensure all text files end with a newline (#1522). By Doug Freed.
* Fix irssi switching to a Unix socket when the network name contains a slash
  (#1523). By Andrej Kacian.
* Fix the /BAN command, which was broken by an earlier change (#1525). By
  Ailin Nemui.
* Document the capsicum settings section in the capsicum docs (#1527). By
  Arrigo Marchiori.
* Bump actions/upload-artifact from the deprecated v1 to v4 in CI (#1545). By
  Pontus Lundkvist.
* Fix the GitHub workflows by pinning setuptools and meson and adding the glib
  dependency, and change how make-dist.sh strips the Python sdist metadata
  from the release tarball (#1546). By Ailin Nemui.
* Fix uninitialised memory in the CTCP ping reply (#1552). By Ailin Nemui.
* Fix SASL negotiation failing when multiple CAP ACK lines are received
  (#1556). By Steering7253.
* Add ipaddr and chosen_family fields to the Perl connection record (#1543).
  By Steering7253.
* Fix building on Solaris by namespacing TERM_REC functions that conflict with
  curses.h, and add a Solaris CI build (#1557). By Ailin Nemui.
* Include <sys/select.h> to get select(3), fixing the build on strict POSIX
  libcs (#1540). By Jonas Termansen.
* Replace obsolescent inet_addr/inet_aton with inet_pton (#1541). By Jonas
  Termansen.
* Use EAI_NONAME instead of HOST_NOT_FOUND for getaddrinfo(3) errors; this
  code path was later replaced by the GIO resolver migration (#1542). By
  Jonas Termansen.
* Remove const qualifiers in fe-text terminfo for better Darwin support
  (#1558). By Emil Engler.
* Update GitHub Actions to Ubuntu 22.04 (#1561). By Ailin Nemui.
* Make bitfields unsigned to silence compiler warnings (#1559). By Ailin
  Nemui.
* Adapt the PEP 440 dev version scheme in make-dist.sh (#1564). By Ailin
  Nemui.
* Add a Void Linux Docker build to GitHub Actions (#1563). By Ailin Nemui.
* Fix the Void Linux GitHub Actions workflow (#1565). By Ailin Nemui.
* Fix a build issue by adding a missing wchar.h include in the Perl core
  (#1560). By Ailin Nemui.
* Fix two memory leaks when creating a main window with not enough space
  (#1566). By Christian Carey.
* Fix a `_GNU_SOURCE` redefined warning caused by Perl ccflags (#1572). By
  Ailin Nemui.
* Fix the Solaris GitHub Actions test by detecting the third argument type of
  puts (#1576). By Ailin Nemui.
* Document the existing -priority option in the HILIGHT syntax help (#1575).
  By Jari Matilainen.
* Fix the xbps Void Linux GitHub Actions workflow (#1579). By Ailin Nemui.
* Use g_string_free_and_steal to avoid an unused-result warning on GLib 2.76
  (#1577). By Ailin Nemui.
* Fix a typo in the LUSERS help text (#1590). By Juha Remes.
* Increase the default scrollback_lines from 500 to 5000 (#1588). By
  arza-zara.
* Fix nickname truncation in netsplit messages to avoid a trailing comma-space
  (#1591). By soulseller.
* Fix hide_text_style and hide_colors to mitigate color bleed on the reset
  control code (#1594). By soulseller.
* Update bundled scripts autoop, mail, and scriptassist (#1597). By Ailin
  Nemui.
* Update package lists in the abicheck GitHub Actions workflow (#1600). By
  Ailin Nemui.
* Update syncdocs.sh to sync the New-users and qna docs, dropping faq and
  startup-HOWTO (#1599). By Ailin Nemui.
* Make compilation work on Cygwin by using shared_library instead of
  shared_module (#1578). By Ailin Nemui.
* Use the GIO resolver (GResolver) for name resolution instead of forking a
  resolver child process, adding gio and gobject as dependencies. Remove the
  resolve_prefer_ipv6 setting for server connections (IPv4/IPv6 order now
  follows the resolver) and add an irssiproxy_prefer_ipv6 setting for the
  irssiproxy bind. Server connections now try resolved addresses in the
  resolver's order with sequential failover instead of a random address
  (#1580). By Ailin Nemui.
* Document additional compile dependencies (utf8proc, libgcrypt, complete
  Perl) in INSTALL (#1568). By nikolas.
* Move the Perl DCC glue into a separate Irssi::Irc::Dcc module so Irssi::Irc
  can be used with dcc unloaded (#1604). By Ailin Nemui.
* Fix the clang-format-xs tooling boot code and comment spacing (#1606). By
  Ailin Nemui.
* Add compatibility code for older GResolver versions (#1605). By Ailin Nemui.
* Fix several issues with the GIO resolver (#1607). By Ailin Nemui.
* Add a muon meson-format GitHub Actions workflow and reformat the meson build
  files (#1611). By William Storey.
