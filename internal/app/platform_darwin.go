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
	maskerKeyTypePlay = 19,
	maskerKeyTypePause = 20,
	maskerKeyTypeTogglePlay = 21,
};

extern void maskerHandleTrackCommand(int keyType);

static id maskerNextTrackToken = nil;
static id maskerPreviousTrackToken = nil;
static id maskerPlayToken = nil;
static id maskerPauseToken = nil;
static id maskerTogglePlayToken = nil;

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
	if (maskerPlayToken == nil) {
		center.playCommand.enabled = YES;
		maskerPlayToken = [center.playCommand addTargetWithHandler:^MPRemoteCommandHandlerStatus(MPRemoteCommandEvent *event) {
			maskerHandleTrackCommand(maskerKeyTypePlay);
			return MPRemoteCommandHandlerStatusSuccess;
		}];
	}
	if (maskerPauseToken == nil) {
		center.pauseCommand.enabled = YES;
		maskerPauseToken = [center.pauseCommand addTargetWithHandler:^MPRemoteCommandHandlerStatus(MPRemoteCommandEvent *event) {
			maskerHandleTrackCommand(maskerKeyTypePause);
			return MPRemoteCommandHandlerStatusSuccess;
		}];
	}
	if (maskerTogglePlayToken == nil) {
		center.togglePlayPauseCommand.enabled = YES;
		maskerTogglePlayToken = [center.togglePlayPauseCommand addTargetWithHandler:^MPRemoteCommandHandlerStatus(MPRemoteCommandEvent *event) {
			maskerHandleTrackCommand(maskerKeyTypeTogglePlay);
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
	if (maskerPlayToken != nil) {
		[center.playCommand removeTarget:maskerPlayToken];
		maskerPlayToken = nil;
	}
	if (maskerPauseToken != nil) {
		[center.pauseCommand removeTarget:maskerPauseToken];
		maskerPauseToken = nil;
	}
	if (maskerTogglePlayToken != nil) {
		[center.togglePlayPauseCommand removeTarget:maskerTogglePlayToken];
		maskerTogglePlayToken = nil;
	}
	center.nextTrackCommand.enabled = NO;
	center.previousTrackCommand.enabled = NO;
	center.playCommand.enabled = NO;
	center.pauseCommand.enabled = NO;
	center.togglePlayPauseCommand.enabled = NO;
	[MPNowPlayingInfoCenter defaultCenter].nowPlayingInfo = nil;
	if (@available(macOS 10.13.2, *)) {
		[MPNowPlayingInfoCenter defaultCenter].playbackState = MPNowPlayingPlaybackStateStopped;
	}
}

static void updateMaskerNowPlayingInfo(const char *modeName, _Bool paused) {
	NSString *mode = modeName != NULL ? [NSString stringWithUTF8String:modeName] : @"";
	NSMutableDictionary *info = [NSMutableDictionary dictionary];
	info[MPMediaItemPropertyTitle] = @"Masker";
	if ([mode length] > 0) {
		info[MPMediaItemPropertyArtist] = mode;
	}
	info[MPNowPlayingInfoPropertyPlaybackRate] = paused ? @0.0 : @1.0;
	info[MPNowPlayingInfoPropertyElapsedPlaybackTime] = @0.0;
	[MPNowPlayingInfoCenter defaultCenter].nowPlayingInfo = info;
	if (@available(macOS 10.13.2, *)) {
		[MPNowPlayingInfoCenter defaultCenter].playbackState = paused ? MPNowPlayingPlaybackStatePaused : MPNowPlayingPlaybackStatePlaying;
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
var playHandler func()
var pauseHandler func()
var togglePlayPauseHandler func()

func installTrackCommandHandlers(next, previous, play, pause, togglePlayPause func()) {
	nextTrackHandler = next
	previousTrackHandler = previous
	playHandler = play
	pauseHandler = pause
	togglePlayPauseHandler = togglePlayPause
	C.installMaskerTrackCommandHandlers()
}

func clearTrackCommandHandlers() {
	C.clearMaskerTrackCommandHandlers()
	nextTrackHandler = nil
	previousTrackHandler = nil
	playHandler = nil
	pauseHandler = nil
	togglePlayPauseHandler = nil
}

func updateTrackCommandState(mode string, paused bool) {
	cMode := C.CString(mode)
	defer C.free(unsafe.Pointer(cMode))
	C.updateMaskerNowPlayingInfo(cMode, C._Bool(paused))
}
