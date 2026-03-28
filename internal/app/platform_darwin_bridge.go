//go:build darwin

package app

/*
#include <stdint.h>
*/
import "C"

const (
	keyTypeNextTrack     = 17
	keyTypePreviousTrack = 18
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
	}
}
