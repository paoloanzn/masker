package app

/*
#cgo darwin CFLAGS: -x objective-c -fobjc-arc
#cgo darwin LDFLAGS: -framework Cocoa

#import <Cocoa/Cocoa.h>

enum {
	maskerMediaKeySubtype = 8,
	maskerMediaKeyDown = 0xA,
	maskerKeyTypeNextTrack = 17,
	maskerKeyTypePreviousTrack = 18,
};

extern void maskerHandleTrackCommand(int keyType);

static id maskerMediaKeyMonitor = nil;

static void prepareMaskerApp(void) {
	[NSApplication sharedApplication];
	[NSApp setActivationPolicy:NSApplicationActivationPolicyAccessory];
}

static void installMaskerMediaKeyMonitor(void) {
	if (maskerMediaKeyMonitor != nil) {
		return;
	}

	maskerMediaKeyMonitor = [NSEvent addLocalMonitorForEventsMatchingMask:NSEventMaskSystemDefined handler:^NSEvent * _Nullable(NSEvent *event) {
		if (![NSApp isActive] || [event subtype] != maskerMediaKeySubtype) {
			return event;
		}

		NSInteger data1 = [event data1];
		int keyType = (int)((data1 & 0xFFFF0000) >> 16);
		int keyState = (int)((data1 & 0x0000FF00) >> 8);
		if (keyState != maskerMediaKeyDown) {
			return event;
		}

		if (keyType == maskerKeyTypeNextTrack || keyType == maskerKeyTypePreviousTrack) {
			maskerHandleTrackCommand(keyType);
			return nil;
		}

		return event;
	}];
}

static void clearMaskerMediaKeyMonitor(void) {
	if (maskerMediaKeyMonitor == nil) {
		return;
	}

	[NSEvent removeMonitor:maskerMediaKeyMonitor];
	maskerMediaKeyMonitor = nil;
}
*/
import "C"

func preparePlatform() {
	C.prepareMaskerApp()
}

var nextTrackHandler func()
var previousTrackHandler func()

func installTrackCommandHandlers(next, previous func()) {
	nextTrackHandler = next
	previousTrackHandler = previous
	C.installMaskerMediaKeyMonitor()
}

func clearTrackCommandHandlers() {
	nextTrackHandler = nil
	previousTrackHandler = nil
	C.clearMaskerMediaKeyMonitor()
}
