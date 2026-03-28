//go:build !darwin

package app

func preparePlatform() {}

func installTrackCommandHandlers(_, _ func()) {}

func clearTrackCommandHandlers() {}

func updateTrackCommandMode(string) {}
