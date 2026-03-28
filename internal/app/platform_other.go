//go:build !darwin

package app

func preparePlatform() {}

func installTrackCommandHandlers(_, _, _, _, _ func()) {}

func clearTrackCommandHandlers() {}

func updateTrackCommandState(string, bool) {}
