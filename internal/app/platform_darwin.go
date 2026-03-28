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

@interface MaskerApplication : NSApplication
@end

@implementation MaskerApplication

- (void)sendEvent:(NSEvent *)event {
	if ([self isActive] && [event type] == NSEventTypeSystemDefined && [event subtype] == maskerMediaKeySubtype) {
		NSInteger data1 = [event data1];
		int keyType = (int)((data1 & 0xFFFF0000) >> 16);
		int keyState = (int)((data1 & 0x0000FF00) >> 8);
		if (keyState == maskerMediaKeyDown &&
			(keyType == maskerKeyTypeNextTrack || keyType == maskerKeyTypePreviousTrack)) {
			maskerHandleTrackCommand(keyType);
			return;
		}
	}

	[super sendEvent:event];
}

@end

static void prepareMaskerApp(void) {
	[MaskerApplication sharedApplication];
	[NSApp setActivationPolicy:NSApplicationActivationPolicyAccessory];
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
}

func clearTrackCommandHandlers() {
	nextTrackHandler = nil
	previousTrackHandler = nil
}
