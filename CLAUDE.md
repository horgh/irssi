# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build Commands

```bash
# Configure and build
meson Build
ninja -C Build

# Install (requires appropriate permissions)
ninja -C Build install

# Run all tests
ninja -C Build test

# Run a specific test
Build/tests/irc/core/test-irc --tap
```

## Build Options

Common meson options (pass to `meson Build`):
- `-Dwith-proxy=yes` - Build irssi-proxy
- `-Dwith-perl=[yes|no]` - Enable/disable Perl scripting
- `-Dwith-bot=yes` - Build irssi-bot
- `-Dwithout-textui=yes` - Build without text frontend
- `--prefix=/path` - Installation prefix (can install without root)

## Code Formatting

**C code:** Uses clang-format-14 with Linux kernel style (tabs, 8-char width, 100-char line limit)
```bash
CLANG_FORMAT=clang-format-14 git-clang-format-14 FETCH_HEAD HEAD
```

**Meson files:** Uses muon fmt
```bash
muon fmt -c .muon_fmt.ini -i meson.build
```

## Architecture

Irssi uses a layered, signal-based architecture:

```
CORE (src/core/)
  └─ Signals, settings, commands, networking, servers, modules, logging
       │
IRC module (src/irc/)
  ├─ core/     - IRC protocol, channels, nicks, servers, CTCP
  ├─ dcc/      - DCC chat/file transfer
  ├─ flood/    - Flood detection
  ├─ notifylist/
  └─ proxy/    - IRC proxy mode
       │
Common UI (src/fe-common/)
  ├─ core/     - Windows, printtext(), themes, completion
  └─ irc/      - IRC-specific UI
       │
Frontend
  ├─ fe-text/  - Terminal UI (main client)
  ├─ fe-none/  - Headless/bot mode
  └─ fe-fuzz/  - Fuzzing harness
```

**Signal system:** Modules communicate via `signal_emit()` and `signal_add()` - see `docs/signals.txt` for the complete list.

**Configuration library:** `src/lib-config/` handles reading/writing `.conf` files.

## Key Files

- `src/common.h` - Core typedefs (SERVER_REC, CHANNEL_REC, etc.) and ABI version
- `src/core/signals.h` - Signal/event system API
- `docs/design.txt` - Architecture overview
- `docs/signals.txt` - All signal definitions
- `docs/perl.txt` - Perl scripting guide

## Dependencies

Required: GLib 2.32+, OpenSSL, terminfo/ncurses, pkg-config
Optional: Perl 5.8+ (scripting), libotr 4.1+ (encryption), utf8proc (character width)
