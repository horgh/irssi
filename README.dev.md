# Irssi Developer Guide

This document provides an architecture overview for developers new to the irssi codebase.

## Architecture Overview

Irssi is a **signal-driven, layered, modular IRC client**. Each layer only talks to adjacent layers, and communication happens through **signals** (an event system), not direct function calls.

```
Frontend (fe-text, fe-none, fe-fuzz)
         |
FE-Common (windows, themes, text formatting)
         |
IRC Module (protocol implementation)
         |
Core (signals, settings, servers, channels)
         |
lib-config (configuration file I/O)
```

## The Signal System

This is the most important concept. Instead of direct function calls between modules, modules communicate through signals:

```c
// Emit an event
signal_emit("message public", 5, server, msg, nick, address, target);

// Listen for events
signal_add("message public", (SIGNAL_FUNC) my_handler);
```

This loose coupling means you can add features (Perl scripts, proxy mode, alternative UIs) without modifying core code.

**Signal priorities** control execution order:
- `SIGNAL_PRIORITY_HIGH` (-100) - runs first
- `SIGNAL_PRIORITY_DEFAULT` (0) - normal order
- `SIGNAL_PRIORITY_LOW` (100) - runs last

See `docs/signals.txt` for the complete signal list.

## Key Data Structures

- **SERVER_REC** - an IRC server connection (tag, nick, channels, lag)
- **CHANNEL_REC** - a joined channel (topic, nicks, modes)
- **NICK_REC** - a user in a channel (nick, host, op/voice flags)
- **WINDOW_REC** - a terminal window displaying items
- **IRC_SERVER_REC** - extends SERVER_REC with IRC-specific data

These are defined in `src/common.h`.

## Directory Structure

| Directory | Purpose |
|-----------|---------|
| `src/core/` | Infrastructure: signals, settings, commands, servers, channels |
| `src/irc/core/` | IRC protocol: messages, CTCP, modes, capabilities |
| `src/irc/dcc/` | File transfers, DCC chat |
| `src/irc/flood/` | Flood detection and automatic ignoring |
| `src/irc/notifylist/` | Monitor away status of friends |
| `src/irc/proxy/` | IRC proxy server mode |
| `src/fe-common/core/` | Common UI: windows, themes, printtext |
| `src/fe-common/irc/` | IRC-specific UI: channel windows, nick completion |
| `src/fe-text/` | Terminal UI: the main frontend |
| `src/fe-none/` | Headless mode (bot, server mode) |
| `src/fe-fuzz/` | Fuzzing harness for security testing |
| `src/lib-config/` | Config file parser (`.conf` format) |
| `src/otr/` | OTR encryption support |
| `src/perl/` | Perl scripting integration |
| `tests/` | Unit tests organized by module |

## Example Flow: User Joins a Channel

1. **fe-text**: User types `/join #irssi`
2. **core**: Command dispatched via signal
3. **irc/core**: Sends `JOIN #irssi` over socket
4. **irc/core**: Receives response, creates `CHANNEL_REC`
5. **signal**: `"channel joined"` emitted
6. **fe-common**: Creates window item
7. **fe-text**: Renders in terminal

## Key Files

| File | Purpose |
|------|---------|
| `src/common.h` | Core typedefs (SERVER_REC, CHANNEL_REC, etc.) and ABI version |
| `src/core/signals.h` | Signal API (emit, add, remove, priorities) |
| `docs/signals.txt` | Complete list of all signals |
| `docs/design.txt` | Architecture overview |
| `src/core/core.c` | Core module initialization |
| `src/fe-text/irssi.c` | Main entry point for text UI |
| `src/irc/core/irc.c` | IRC protocol parser and event dispatcher |

## Contributing Pattern

```c
#define MODULE_NAME "mymodule"

static void sig_channel_joined(CHANNEL_REC *channel) {
    // Handle the event
}

void mymodule_init(void) {
    signal_add("channel joined", (SIGNAL_FUNC) sig_channel_joined);
}

void mymodule_deinit(void) {
    signal_remove("channel joined", (SIGNAL_FUNC) sig_channel_joined);
}
```

## Working with Servers and Channels

```c
// Check if it's an IRC server
if (IS_IRC_SERVER(server)) {
    IRC_SERVER_REC *irc = (IRC_SERVER_REC *) server;
    // IRC-specific operations
}

// Find a channel
CHANNEL_REC *channel = channel_find(server, "#irssi");

// Iterate nicks in a channel
for (GSList *list = nicklist_getnicks(channel); list != NULL; list = list->next) {
    NICK_REC *nick = list->data;
}
```

## Module Data Attachment

Each server/channel/nick can store module-specific data:

```c
MODULE_DATA_SET(server, my_module_data);
my_data = MODULE_DATA(server);
MODULE_DATA_UNSET(server);
```

## Mental Model

Think of irssi as an **event bus** where modules publish and subscribe to domain events rather than calling each other directly. This is what makes it so extensible.
