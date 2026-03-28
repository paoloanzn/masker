package app

/*
#cgo darwin CFLAGS: -x objective-c -fobjc-arc
#cgo darwin LDFLAGS: -framework Cocoa -framework MediaPlayer

#import <Cocoa/Cocoa.h>
#import <MediaPlayer/MediaPlayer.h>
#include <stdlib.h>

enum {
	maskerKeyTypeNextTrack = 17,
	maskerKeyTypePreviousTrack = 18,
};

extern void maskerHandleTrackCommand(int keyType);

static id maskerNextTrackToken = nil;
static id maskerPreviousTrackToken = nil;

static void prepareMaskerApp(void) {
	[NSApplication sharedApplication];
	[NSApp setActivationPolicy:NSApplicationActivationPolicyAccessory];
}

static void installMaskerTrackCommandHandlers(void) {
	MPRemoteCommandCenter *center = [MPRemoteCommandCenter sharedCommandCenter];
	if (maskerNextTrackToken == nil) {
		center.nextTrackCommand.enabled = YES;
		maskerNextTrackToken = [center.nextTrackCommand addTargetWithHandler:^MPRemoteCommandHandlerStatus(MPRemoteCommandEvent *event) {
			maskerHandleTrackCommand(maskerKeyTypeNextTrack);
			return MPRemoteCommandHandlerStatusSuccess;
		}];
	}
	if (maskerPreviousTrackToken == nil) {
		center.previousTrackCommand.enabled = YES;
		maskerPreviousTrackToken = [center.previousTrackCommand addTargetWithHandler:^MPRemoteCommandHandlerStatus(MPRemoteCommandEvent *event) {
			maskerHandleTrackCommand(maskerKeyTypePreviousTrack);
			return MPRemoteCommandHandlerStatusSuccess;
		}];
	}
}

static void clearMaskerTrackCommandHandlers(void) {
	MPRemoteCommandCenter *center = [MPRemoteCommandCenter sharedCommandCenter];
	if (maskerNextTrackToken != nil) {
		[center.nextTrackCommand removeTarget:maskerNextTrackToken];
		maskerNextTrackToken = nil;
	}
	if (maskerPreviousTrackToken != nil) {
		[center.previousTrackCommand removeTarget:maskerPreviousTrackToken];
		maskerPreviousTrackToken = nil;
	}
	center.nextTrackCommand.enabled = NO;
	center.previousTrackCommand.enabled = NO;
	[MPNowPlayingInfoCenter defaultCenter].nowPlayingInfo = nil;
	if (@available(macOS 10.13.2, *)) {
		[MPNowPlayingInfoCenter defaultCenter].playbackState = MPNowPlayingPlaybackStateStopped;
	}
}

static void updateMaskerNowPlayingInfo(const char *modeName) {
	NSString *mode = modeName != NULL ? [NSString stringWithUTF8String:modeName] : @"";
	NSMutableDictionary *info = [NSMutableDictionary dictionary];
	info[MPMediaItemPropertyTitle] = @"Masker";
	if ([mode length] > 0) {
		info[MPMediaItemPropertyArtist] = mode;
	}
	info[MPNowPlayingInfoPropertyPlaybackRate] = @1.0;
	info[MPNowPlayingInfoPropertyElapsedPlaybackTime] = @0.0;
	[MPNowPlayingInfoCenter defaultCenter].nowPlayingInfo = info;
	if (@available(macOS 10.13.2, *)) {
		[MPNowPlayingInfoCenter defaultCenter].playbackState = MPNowPlayingPlaybackStatePlaying;
	}
}
*/
import "C"

import "unsafe"

func preparePlatform() {
	C.prepareMaskerApp()
}

var nextTrackHandler func()
var previousTrackHandler func()

func installTrackCommandHandlers(next, previous func()) {
	nextTrackHandler = next
	previousTrackHandler = previous
	C.installMaskerTrackCommandHandlers()
}

func clearTrackCommandHandlers() {
	C.clearMaskerTrackCommandHandlers()
	nextTrackHandler = nil
	previousTrackHandler = nil
}

func updateTrackCommandMode(mode string) {
	cMode := C.CString(mode)
	defer C.free(unsafe.Pointer(cMode))
	C.updateMaskerNowPlayingInfo(cMode)
}
