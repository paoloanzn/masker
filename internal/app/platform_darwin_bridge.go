//go:build darwin

package app

/*
#include <stdint.h>
*/
import "C"

const (
	keyTypeNextTrack     = 17
	keyTypePreviousTrack = 18
	keyTypePlay          = 19
	keyTypePause         = 20
	keyTypeTogglePlay    = 21
)

//export maskerHandleTrackCommand
func maskerHandleTrackCommand(keyType C.int) {
	switch int(keyType) {
	case keyTypeNextTrack:
		if nextTrackHandler != nil {
			nextTrackHandler()
		}
	case keyTypePreviousTrack:
		if previousTrackHandler != nil {
			previousTrackHandler()
		}
	case keyTypePlay:
		if playHandler != nil {
			playHandler()
		}
	case keyTypePause:
		if pauseHandler != nil {
			pauseHandler()
		}
	case keyTypeTogglePlay:
		if togglePlayPauseHandler != nil {
			togglePlayPauseHandler()
		}
	}
}
