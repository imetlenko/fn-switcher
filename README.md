# fn-switcher

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
- Go 1.22+ (for building)
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
sudo cp fn-switcher /usr/local/bin/
```

For autostart, see the [Autostart](#autostart) section below.

## Setup

### 1. Grant Accessibility permissions

**System Settings → Privacy & Security → Accessibility**

Add `/usr/local/bin/fn-switcher` (click "+", then Cmd+Shift+G to enter path).

### 2. Disable system Fn popup

**System Settings → Keyboard → "Press 🌐 key to"** → set to **"Do Nothing"**

### 3. Run

If installed via Homebrew:

```bash
brew services start fn-switcher
```

If installed from source or manually, see the [Autostart](#autostart) section to run as a background service, or test in the foreground:

```bash
fn-switcher
```

## Usage

```bash
# Start switcher (auto-detects layouts, MRU mode)
fn-switcher

# Use custom layouts
fn-switcher -layouts "ABC,Russian"

# Cycle mode with custom layouts
fn-switcher -cycle -layouts "ABC,Russian,German"

# List available input sources
fn-switcher -list

# Get current input source
fn-switcher -get

# Set input source
fn-switcher -set com.apple.keylayout.Russian

# Show help
fn-switcher -help
```

## Configuration

fn-switcher uses layered configuration (highest priority first):

1. **CLI flags** — `-layouts "ABC,Russian" -cycle`
2. **Environment variables** — `FN_SWITCHER_LAYOUTS`, `FN_SWITCHER_CYCLE`
3. **Config file** — `~/.config/fn-switcher/config.json`
4. **Defaults** — auto-detect all layouts, MRU mode

### Config file (recommended for brew services)

A default config file is created automatically on first run at `~/.config/fn-switcher/config.json` with auto-detected layouts and MRU mode. You can edit it to customize:

```json
{
  "layouts": ["ABC", "Russian"],
  "cycle": true
}
```

The config file is ideal when running as a brew service, since the service plist passes no flags and gets overwritten on upgrades.

### Environment variables

| Variable | Description | Example |
|---|---|---|
| `FN_SWITCHER_LAYOUTS` | Comma-separated layout names | `ABC,Russian` |
| `FN_SWITCHER_CYCLE` | Enable cycle mode | `true` or `1` |

```bash
FN_SWITCHER_LAYOUTS="ABC,Russian" FN_SWITCHER_CYCLE=true fn-switcher
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

### Homebrew

```bash
brew services stop fn-switcher
brew uninstall fn-switcher
```

### From source

```bash
make uninstall
```

### Manual

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

By default, fn-switcher auto-detects all system layouts. Use custom layouts to:
- **Limit the list** — only switch between specific layouts, ignoring the rest
- **Set the order** — control the cycle sequence for long press switching

Configure your layouts (use short names without the `com.apple.keylayout.` prefix):

```bash
# CLI flag
fn-switcher -layouts "US,German"

# Or config file (recommended for service use)
echo '{"layouts": ["US", "German"]}' > ~/.config/fn-switcher/config.json
```

> **Tip:** When running as a brew service, use the config file instead of editing the plist — the plist gets overwritten on brew upgrades.

## Switching modes

### MRU (default)

Toggle between the two most recently used layouts:

- **Short press** (< 500ms) — instantly toggles between current and previous layout
- **Long press** (≥ 500ms) — cycles to the next layout in the list (skipping the previous one), triggers at the 500ms mark without waiting for release

This means short press always jumps between your two main layouts, and long press lets you reach additional layouts when you have 3+.

### Cycle (`-cycle`)

Each Fn press cycles through all layouts in order. Triggers instantly on key down.

## How it works

1. Uses `CGEventTap` to intercept Fn key modifier flag changes
2. On Fn press, starts a 500ms timer (MRU mode) or switches immediately (Cycle mode)
3. If Fn released before 500ms — MRU toggle with previous layout
4. If timer fires while Fn still held — cycle to next layout
5. Calls Carbon `TISSelectInputSource` API directly — no shell commands, no external dependencies

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

## License

MIT

## Credits

Inspired by frustration with macOS language switcher popup.
