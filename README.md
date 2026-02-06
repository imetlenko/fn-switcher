# fn-switcher

Fast Fn key input source switcher for macOS. No Python, no Karabiner, no bloat.

Instantly switches keyboard layout when you press the Fn (Globe) key — without the annoying macOS popup.

## Why?

macOS shows an irritating language switcher popup every time you press Fn/Globe. It adds delay and breaks your flow. Apple provides no way to disable it.

This tool:
- Intercepts Fn key presses before macOS handles them
- Switches input source directly via Carbon API
- No popup, no delay, instant switching
- Single binary, ~2MB, pure Go + cgo

## Requirements

- macOS 10.13+ (tested on Sonoma)
- Go 1.21+ (for building)
- Xcode Command Line Tools (`xcode-select --install`)

## Installation

### Homebrew (Recommended)

```bash
brew install imetlenko/apps/fn-switcher
brew services start fn-switcher
```

### From source

```bash
git clone https://github.com/imetlenko/fn-switcher.git
cd fn-switcher
make build
sudo make install
```

### Manual

```bash
go build -o fn-switcher
env GOPATH=/usr/local/go sudo cp fn-switcher /usr/local/bin/
```

## Setup

### 1. Grant Accessibility permissions

**System Settings → Privacy & Security → Accessibility**

Add `/usr/local/bin/fn-switcher` (click "+", then Cmd+Shift+G to enter path).

### 2. Disable system Fn popup

**System Settings → Keyboard → "Press 🌐 key to"** → set to **"Do Nothing"**

### 3. Run

```bash
fn-switcher
```

## Usage

```bash
# Start switcher with default layouts (ABC <-> Russian)
fn-switcher

# Use custom layouts
fn-switcher -l1 com.apple.keylayout.US -l2 com.apple.keylayout.German

# List available input sources
fn-switcher -list

# Get current input source
fn-switcher -get

# Set input source
fn-switcher -set com.apple.keylayout.Russian

# Show help
fn-switcher -help
```

## Autostart

### Option 1: LaunchAgent (recommended)

```bash
make install-agent
```

Or manually create `~/Library/LaunchAgents/com.user.fnswitcher.plist`:

```xml
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>com.user.fnswitcher</string>
    <key>ProgramArguments</key>
    <array>
        <string>/usr/local/bin/fn-switcher</string>
    </array>
    <key>RunAtLoad</key>
    <true/>
    <key>KeepAlive</key>
    <true/>
</dict>
</plist>
```

Then load it:

```bash
launchctl load ~/Library/LaunchAgents/com.user.fnswitcher.plist
```

### Option 2: Login Items

System Settings → General → Login Items → add fn-switcher

## Uninstall

```bash
make uninstall
```

Or manually:

```bash
launchctl unload ~/Library/LaunchAgents/com.user.fnswitcher.plist
rm ~/Library/LaunchAgents/com.user.fnswitcher.plist
sudo rm /usr/local/bin/fn-switcher
```

## Custom layouts

Find your layout IDs:

```bash
fn-switcher -list
```

Common layouts:
- `com.apple.keylayout.ABC` — ABC (default English)
- `com.apple.keylayout.US` — U.S.
- `com.apple.keylayout.Russian` — Russian
- `com.apple.keylayout.German` — German
- `com.apple.keylayout.French` — French

Then run with your layouts:

```bash
fn-switcher -l1 com.apple.keylayout.US -l2 com.apple.keylayout.German
```

For autostart with custom layouts, edit the plist:

```xml
<key>ProgramArguments</key>
<array>
    <string>/usr/local/bin/fn-switcher</string>
    <string>-l1</string>
    <string>com.apple.keylayout.US</string>
    <string>-l2</string>
    <string>com.apple.keylayout.German</string>
</array>
```

## How it works

1. Uses `CGEventTap` to intercept Fn key modifier flag changes
2. Calls `TISCopyCurrentKeyboardInputSource` to get current layout
3. Calls `TISSelectInputSource` to switch layout
4. No shell commands, no external dependencies

## Troubleshooting

### "Failed to create event tap"

Add fn-switcher to Accessibility in System Settings.

### Fn key still shows popup

Set "Press 🌐 key to" → "Do Nothing" in Keyboard settings.

### First character in wrong layout

This is a known macOS quirk. The switch happens fast but sometimes the first keypress races ahead. Usually not noticeable in practice.

### Not working after macOS update

Re-add fn-switcher to Accessibility permissions — macOS sometimes resets them.

## Security

This tool requires Accessibility permissions to intercept key events. The code is open source — audit it yourself:

- No network calls
- No data collection
- No external dependencies
- Single-purpose: intercept Fn, switch layout

## Alternatives

| Tool | Language | Size | Intercepts Fn | No popup |
|------|----------|------|---------------|----------|
| fn-switcher | Go | ~2MB | ✅ | ✅ |
| Karabiner-Elements | C++ | ~50MB | ✅ | ✅ |
| issw + pynput | Python | ~30MB | ✅ | ✅ |
| Built-in Caps Lock | - | 0 | ❌ | ✅ |

## License

MIT

## Credits

Inspired by frustration with macOS language switcher popup.
