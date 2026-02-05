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

// Get current input source
static CFStringRef getCurrentInputSource() {
    TISInputSourceRef source = TISCopyCurrentKeyboardInputSource();
    if (!source) return NULL;

    CFStringRef sourceID = (CFStringRef)TISGetInputSourceProperty(source, kTISPropertyInputSourceID);
    CFStringRef result = CFStringCreateCopy(kCFAllocatorDefault, sourceID);
    CFRelease(source);
    return result;
}

// Select input source by ID
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

// List all available input sources
static void listInputSources() {
    CFArrayRef sources = TISCreateInputSourceList(NULL, false);
    if (!sources) return;

    CFIndex count = CFArrayGetCount(sources);

    for (CFIndex i = 0; i < count; i++) {
        TISInputSourceRef source = (TISInputSourceRef)CFArrayGetValueAtIndex(sources, i);
        CFStringRef sourceID = (CFStringRef)TISGetInputSourceProperty(source, kTISPropertyInputSourceID);
        CFBooleanRef selectable = (CFBooleanRef)TISGetInputSourceProperty(source, kTISPropertyInputSourceIsSelectCapable);

        if (selectable && CFBooleanGetValue(selectable) && sourceID) {
            char buffer[256];
            CFStringGetCString(sourceID, buffer, sizeof(buffer), kCFStringEncodingUTF8);
            printf("%s\n", buffer);
        }
    }

    CFRelease(sources);
}

// Wrapper for Go
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
	"flag"
	"fmt"
	"os"
	"unsafe"
)

const (
	fnKeyFlag = 0x800000 // NSEventModifierFlagFunction
)

var (
	buildDate string
	commit    string
	version   string
)

var fnPressed = false

// Default layouts - can be changed via flags
var (
	layout1 = "com.apple.keylayout.ABC"
	layout2 = "com.apple.keylayout.Russian"
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

func listLayouts() {
	C.listInputSources()
}

func switchInputSource() {
	current := getCurrentLayout()

	var target string
	if current == layout1 {
		target = layout2
	} else {
		target = layout1
	}

	if err := setLayout(target); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
	}
}

// TODO: Поставить на вывод buildDate и commit
func printUsage() {
	fmt.Printf(`fn-switcher v%s - Fast Fn key input source switcher for macOS

Usage:
  fn-switcher [flags]           Start the switcher daemon
  fn-switcher -list             List available input sources
  fn-switcher -get              Show current input source
  fn-switcher -set <source_id>  Set input source

Flags:
  -l1 <source_id>   First layout (default: com.apple.keylayout.ABC)
  -l2 <source_id>   Second layout (default: com.apple.keylayout.Russian)
  -list             List all available input sources
  -get              Get current input source
  -set <source_id>  Set input source
  -version          Show version
  -help             Show this help

Examples:
  fn-switcher                          # Start with default layouts (ABC <-> Russian)
  fn-switcher -l1 com.apple.keylayout.US -l2 com.apple.keylayout.German
  fn-switcher -list
  fn-switcher -get
  fn-switcher -set com.apple.keylayout.Russian

Note: Requires Accessibility permissions in System Settings.
`, version)
}

func main() {
	l1 := flag.String("l1", layout1, "First layout")
	l2 := flag.String("l2", layout2, "Second layout")
	list := flag.Bool("list", false, "List available input sources")
	get := flag.Bool("get", false, "Get current input source")
	set := flag.String("set", "", "Set input source")
	showVersion := flag.Bool("version", false, "Show version")
	help := flag.Bool("help", false, "Show help")

	flag.Parse()

	if *help {
		printUsage()
		return
	}

	if *showVersion {
		fmt.Printf("fn-switcher v%s\n", version)
		return
	}

	if *list {
		listLayouts()
		return
	}

	if *get {
		fmt.Println(getCurrentLayout())
		return
	}

	if *set != "" {
		if err := setLayout(*set); err != nil {
			fmt.Fprintln(os.Stderr, "Error:", err)
			os.Exit(1)
		}
		return
	}

	// Update layouts from flags
	layout1 = *l1
	layout2 = *l2

	fmt.Printf("fn-switcher v%s started\n", version)
	fmt.Printf("Layouts: %s <-> %s\n", layout1, layout2)
	fmt.Printf("Current: %s\n", getCurrentLayout())
	fmt.Println("Press Fn to switch. Ctrl+C to exit.")

	C.startEventTap()
}
