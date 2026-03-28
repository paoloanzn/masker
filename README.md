# Masker

Masker is a small Go menu bar utility that plays continuous masking audio for open-plan work, travel, and other distracting environments.

## Features

- Lives in the macOS menu bar with a minimal tray UI
- Starts in `Focus`, the recommended slow-beat mode for general productivity
- Provides a separate `ADHD / attention problems` mode with evidence-linked steady-state `White` and `Pink` presets
- Adjust Focus presets without changing the macro-tempo
- Switch between Focus and masking-oriented Brown and Speech-shaped modes
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

The default `Focus` mode is the main recommended preset for general productivity. It uses a steady 4/4 slow beat with a soft beat-synced pulse overlay that gently contours the sustained focus bed across all focus presets:

- `Low`: sparse pad plus a soft carrier pulse that reinforces the beat without obvious tremolo
- `Medium`: the main recommended preset; adds a soft harmonic bed under the same structured pulse scaffold
- `High`: adds very subtle background motion while keeping the same tempo and low salience
- `High cognitive load`: keeps the same BPM range while further reducing melodic novelty and harmonic motion

The preset controls apply only to `Focus`. `Brown` and `Speech-shaped` remain masking-style options rather than slow-beat productivity presets.

Masker also exposes a separate `ADHD / attention problems` mode. That mode intentionally limits itself to steady-state `White` and `Pink` presets:

- `White`: broadband white noise with a flat spectral density
- `Pink`: steady pink noise with an approximate 1/f spectral tilt

## Development

Useful targets:

```sh
make build
make test
make fmt
make clean
```
