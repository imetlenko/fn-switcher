package main

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework CoreGraphics -framework CoreFoundation -framework Carbon

#include <CoreGraphics/CoreGraphics.h>
#include <Carbon/Carbon.h>

extern void goKeyCallback(int keyCode, int flags);

static CGEventRef eventCallback(CGEventTapProxy proxy, CGEventType type, CGEventRef event, void *refcon) {
    if (type == kCGEventFlagsChanged) {
        CGEventFlags flags = CGEventGetFlags(event);
        int keyCode = (int)CGEventGetIntegerValueField(event, kCGKeyboardEventKeycode);
        goKeyCallback(keyCode, (int)flags);
    }
    return event;
}

static void startEventTap() {
    CGEventMask mask = CGEventMaskBit(kCGEventFlagsChanged);
    CFMachPortRef tap = CGEventTapCreate(
        kCGSessionEventTap,
        kCGHeadInsertEventTap,
        kCGEventTapOptionDefault,
        mask,
        eventCallback,
        NULL
    );
    
    if (!tap) {
        printf("Failed to create event tap. Check Accessibility permissions.\n");
        return;
    }
    
    CFRunLoopSourceRef source = CFMachPortCreateRunLoopSource(kCFAllocatorDefault, tap, 0);
    CFRunLoopAddSource(CFRunLoopGetCurrent(), source, kCFRunLoopCommonModes);
    CGEventTapEnable(tap, true);
    CFRunLoopRun();
}

// Получить текущую раскладку
static CFStringRef getCurrentInputSource() {
    TISInputSourceRef source = TISCopyCurrentKeyboardInputSource();
    if (!source) return NULL;
    
    CFStringRef sourceID = (CFStringRef)TISGetInputSourceProperty(source, kTISPropertyInputSourceID);
    CFStringRef result = CFStringCreateCopy(kCFAllocatorDefault, sourceID);
    CFRelease(source);
    return result;
}

// Переключить на раскладку по ID
static int selectInputSource(const char *sourceID) {
    CFStringRef targetID = CFStringCreateWithCString(kCFAllocatorDefault, sourceID, kCFStringEncodingUTF8);
    
    CFArrayRef sources = TISCreateInputSourceList(NULL, false);
    if (!sources) {
        CFRelease(targetID);
        return -1;
    }
    
    CFIndex count = CFArrayGetCount(sources);
    int found = 0;
    
    for (CFIndex i = 0; i < count; i++) {
        TISInputSourceRef source = (TISInputSourceRef)CFArrayGetValueAtIndex(sources, i);
        CFStringRef sourceIDProp = (CFStringRef)TISGetInputSourceProperty(source, kTISPropertyInputSourceID);
        
        if (sourceIDProp && CFStringCompare(sourceIDProp, targetID, 0) == kCFCompareEqualTo) {
            TISSelectInputSource(source);
            found = 1;
            break;
        }
    }
    
    CFRelease(sources);
    CFRelease(targetID);
    return found ? 0 : -1;
}

// Обёртка для Go
static const char* getCurrentInputSourceString() {
    CFStringRef source = getCurrentInputSource();
    if (!source) return "";
    
    static char buffer[256];
    CFStringGetCString(source, buffer, sizeof(buffer), kCFStringEncodingUTF8);
    CFRelease(source);
    return buffer;
}
*/
import "C"

import (
	"fmt"
	"unsafe"
)

const fnKeyFlag = 0x800000

var fnPressed = false

const (
	layoutABC     = "com.apple.keylayout.ABC"
	layoutRussian = "com.apple.keylayout.Russian"
)

//export goKeyCallback
func goKeyCallback(keyCode C.int, flags C.int) {
	fnNow := (int(flags) & fnKeyFlag) != 0
	
	if fnNow && !fnPressed {
		switchInputSource()
	}
	fnPressed = fnNow
}

func getCurrentLayout() string {
	cstr := C.getCurrentInputSourceString()
	return C.GoString(cstr)
}

func setLayout(layoutID string) error {
	cstr := C.CString(layoutID)
	defer C.free(unsafe.Pointer(cstr))
	
	result := C.selectInputSource(cstr)
	if result != 0 {
		return fmt.Errorf("failed to select input source: %s", layoutID)
	}
	return nil
}

func switchInputSource() {
	current := getCurrentLayout()
	
	var target string
	if current == layoutABC {
		target = layoutRussian
	} else {
		target = layoutABC
	}
	
	if err := setLayout(target); err != nil {
		fmt.Println("Error:", err)
	}
}

func main() {
	fmt.Println("Fn Switcher started (pure Go + Carbon)")
	fmt.Println("Current layout:", getCurrentLayout())
	C.startEventTap()
}
