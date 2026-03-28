package main

import (
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/term"

	"masker/internal/app"
)

const backgroundEnv = "MASKER_BACKGROUND_PROCESS"

func run() error {
	args, foreground := stripForegroundFlag(os.Args[1:])
	if shouldBackground(foreground) {
		return launchDetached(args)
	}
	return app.Run()
}

func stripForegroundFlag(args []string) ([]string, bool) {
	filtered := make([]string, 0, len(args))
	foreground := false

	for _, arg := range args {
		if arg == "--foreground" {
			foreground = true
			continue
		}
		filtered = append(filtered, arg)
	}

	return filtered, foreground
}

func shouldBackground(foreground bool) bool {
	if foreground || os.Getenv(backgroundEnv) == "1" {
		return false
	}
	if launchedFromAppBundle() {
		return false
	}
	return attachedToTerminal()
}

func launchedFromAppBundle() bool {
	executable, err := os.Executable()
	if err != nil {
		return false
	}

	return strings.Contains(filepath.ToSlash(executable), ".app/Contents/MacOS/")
}

func attachedToTerminal() bool {
	return term.IsTerminal(int(os.Stdin.Fd())) ||
		term.IsTerminal(int(os.Stdout.Fd())) ||
		term.IsTerminal(int(os.Stderr.Fd()))
}
