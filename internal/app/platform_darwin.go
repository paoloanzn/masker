package app

/*
#cgo darwin CFLAGS: -x objective-c -fobjc-arc
#cgo darwin LDFLAGS: -framework Cocoa

#import <Cocoa/Cocoa.h>

static void prepareMaskerApp(void) {
	[NSApplication sharedApplication];
	[NSApp setActivationPolicy:NSApplicationActivationPolicyAccessory];
}
*/
import "C"

func preparePlatform() {
	C.prepareMaskerApp()
}
