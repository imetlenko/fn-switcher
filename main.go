package main

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework CoreGraphics -framework CoreFoundation -framework Carbon

#include <CoreGraphics/CoreGraphics.h>
#include <Carbon/Carbon.h>
#include <stdlib.h>
#include <string.h>

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

// Get selectable keyboard layouts (com.apple.keylayout.* only), newline-separated
static const char* getSelectableKeyboardLayouts() {
    CFArrayRef sources = TISCreateInputSourceList(NULL, false);
    if (!sources) return "";

    static char buffer[4096];
    buffer[0] = '\0';
    int offset = 0;

    CFStringRef prefix = CFSTR("com.apple.keylayout.");
    CFIndex count = CFArrayGetCount(sources);

    for (CFIndex i = 0; i < count; i++) {
        TISInputSourceRef source = (TISInputSourceRef)CFArrayGetValueAtIndex(sources, i);
        CFBooleanRef selectable = (CFBooleanRef)TISGetInputSourceProperty(source, kTISPropertyInputSourceIsSelectCapable);

        if (!selectable || !CFBooleanGetValue(selectable)) continue;

        CFStringRef sourceID = (CFStringRef)TISGetInputSourceProperty(source, kTISPropertyInputSourceID);
        if (!sourceID) continue;

        if (!CFStringHasPrefix(sourceID, prefix)) continue;

        char tmp[256];
        CFStringGetCString(sourceID, tmp, sizeof(tmp), kCFStringEncodingUTF8);

        int len = strlen(tmp);
        if (offset + len + 2 > (int)sizeof(buffer)) break;

        if (offset > 0) {
            buffer[offset++] = '\n';
        }
        memcpy(buffer + offset, tmp, len);
        offset += len;
        buffer[offset] = '\0';
    }

    CFRelease(sources);
    return buffer;
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
	"strings"
	"unsafe"
)

const (
	fnKeyFlag      = 0x800000 // NSEventModifierFlagFunction
	keylayoutPrefix = "com.apple.keylayout."
)

var (
	buildDate string
	commit    string
	version   string
)

var (
	fnPressed      = false
	layouts        []string // ordered list of layouts
	previousLayout string   // for MRU mode
	cycleMode      bool     // false = MRU (default), true = cycle
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

func getKeyboardLayouts() []string {
	cstr := C.getSelectableKeyboardLayouts()
	raw := C.GoString(cstr)
	if raw == "" {
		return nil
	}
	return strings.Split(raw, "\n")
}

func listLayouts() {
	for _, l := range getKeyboardLayouts() {
		fmt.Println(l)
	}
}

func findIndex(list []string, val string) int {
	for i, v := range list {
		if v == val {
			return i
		}
	}
	return -1
}

func switchInputSource() {
	current := getCurrentLayout()

	var target string
	if cycleMode {
		idx := findIndex(layouts, current)
		if idx == -1 {
			target = layouts[0]
		} else {
			target = layouts[(idx+1)%len(layouts)]
		}
	} else {
		// MRU mode
		if previousLayout == "" || previousLayout == current {
			// fallback: go to next in list
			idx := findIndex(layouts, current)
			if idx == -1 {
				target = layouts[0]
			} else {
				target = layouts[(idx+1)%len(layouts)]
			}
		} else {
			target = previousLayout
		}
		previousLayout = current
	}

	if err := setLayout(target); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
	}
}

func printUsage() {
	fmt.Printf(`fn-switcher v%s - Fast Fn key input source switcher for macOS

Usage:
  fn-switcher [flags]           Start the switcher daemon
  fn-switcher -list             List available keyboard layouts
  fn-switcher -get              Show current input source
  fn-switcher -set <source_id>  Set input source

Flags:
  -layouts <list>   Comma-separated list of layouts to switch between
                    (default: auto-detect all com.apple.keylayout.* sources)
                    Use short names without the com.apple.keylayout. prefix.
  -cycle            Use cycle mode instead of MRU (most recently used)
  -list             List all available keyboard layouts
  -get              Get current input source
  -set <source_id>  Set input source
  -version          Show version
  -help             Show this help

Modes:
  MRU (default)     Toggle between current and previous layout.
  Cycle (-cycle)    Cycle through layouts in order.

Examples:
  fn-switcher                                # Auto-detect layouts, MRU mode
  fn-switcher -cycle                         # Auto-detect layouts, cycle mode
  fn-switcher -layouts "ABC,Russian"         # Specific layouts, MRU mode
  fn-switcher -cycle -layouts "ABC,Russian,Kaz"  # Specific layouts, cycle mode
  fn-switcher -list
  fn-switcher -get
  fn-switcher -set com.apple.keylayout.Russian

Note: Requires Accessibility permissions in System Settings.
`, version)
}

func main() {
	layoutsFlag := flag.String("layouts", "", "Comma-separated list of layouts (short names without com.apple.keylayout. prefix)")
	cycle := flag.Bool("cycle", false, "Use cycle mode instead of MRU")
	list := flag.Bool("list", false, "List available keyboard layouts")
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
		if commit != "" {
			fmt.Printf("commit: %s\n", commit)
		}
		if buildDate != "" {
			fmt.Printf("built:  %s\n", buildDate)
		}
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

	cycleMode = *cycle

	// Build layouts list
	if *layoutsFlag != "" {
		parts := strings.Split(*layoutsFlag, ",")
		for _, p := range parts {
			p = strings.TrimSpace(p)
			if p == "" {
				continue
			}
			if !strings.HasPrefix(p, keylayoutPrefix) {
				p = keylayoutPrefix + p
			}
			layouts = append(layouts, p)
		}
	} else {
		layouts = getKeyboardLayouts()
	}

	if len(layouts) < 2 {
		available := getKeyboardLayouts()
		fmt.Fprintln(os.Stderr, "Error: need at least 2 keyboard layouts to switch between.")
		fmt.Fprintln(os.Stderr, "Available layouts:")
		for _, l := range available {
			fmt.Fprintf(os.Stderr, "  %s\n", l)
		}
		os.Exit(1)
	}

	// Initialize previousLayout for MRU mode
	previousLayout = layouts[1]

	mode := "MRU"
	if cycleMode {
		mode = "Cycle"
	}

	fmt.Printf("fn-switcher v%s", version)
	if commit != "" {
		fmt.Printf(" (%s)", commit)
	}
	if buildDate != "" {
		fmt.Printf(" built %s", buildDate)
	}
	fmt.Println(" started")
	fmt.Printf("Mode: %s\n", mode)
	fmt.Printf("Layouts: %s\n", strings.Join(layouts, " -> "))
	fmt.Printf("Current: %s\n", getCurrentLayout())
	fmt.Println("Press Fn to switch. Ctrl+C to exit.")

	C.startEventTap()
}
