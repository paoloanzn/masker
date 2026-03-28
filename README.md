# Masker

Masker is a small Go menu bar utility that plays continuous masking audio for open-plan work, travel, and other distracting environments.

## Features

- Lives in the macOS menu bar with a minimal tray UI
- Starts in a slow-beat focus mode with fixed-tempo instrumental audio
- Adjust focus density without changing the macro-tempo
- Switch between focus, brown, pink, and speech-shaped masking modes
- Adjust volume without reopening the app
- Detaches from the terminal by default when started from the command line

## Requirements

- Go 1.25+
- macOS for full audio playback support

The repository builds on other operating systems, but the audio engine is currently implemented only for macOS.

## Install

Build and install the binary with:

```sh
make install
```

The `install` target detects a local bin directory for the current operating system and environment. It prefers an existing writable path such as `XDG_BIN_HOME`, `~/.local/bin`, `~/bin`, `/opt/homebrew/bin`, or `/usr/local/bin`, then falls back to `~/.local/bin`.

You can override the destination explicitly:

```sh
make install INSTALL_DIR="$HOME/bin"
```

## Run

For a normal development launch:

```sh
make run
```

If you run the compiled binary directly:

```sh
masker
```

Masker will relaunch itself in the background by default so the shell prompt returns immediately. For debugging, keep it attached to the terminal with:

```sh
masker --foreground
```

The default `Focus` mode uses a steady 4/4 slow beat with low-density, medium-density, and high-density variations:

- `Low`: sparse pad plus soft kick pulse
- `Medium`: adds a soft harmonic bed
- `High`: adds subtle rhythmic texture while keeping the same tempo

## Development

Useful targets:

```sh
make build
make test
make fmt
make clean
```
