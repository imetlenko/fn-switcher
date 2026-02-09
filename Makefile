BINARY_NAME=fn-switcher
INSTALL_PATH=/usr/local/bin
PLIST_NAME=com.user.fnswitcher.plist
PLIST_PATH=~/Library/LaunchAgents/$(PLIST_NAME)

.PHONY: build install uninstall install-agent uninstall-agent clean

build:
	go build -o $(BINARY_NAME)

install: build
	sudo cp $(BINARY_NAME) $(INSTALL_PATH)/
	@echo "Installed to $(INSTALL_PATH)/$(BINARY_NAME)"
	@echo ""
	@echo "Next steps:"
	@echo "1. Add $(INSTALL_PATH)/$(BINARY_NAME) to System Settings → Privacy & Security → Accessibility"
	@echo "2. Set 'Press 🌐 key to' → 'Do Nothing' in System Settings → Keyboard"
	@echo "3. Run: fn-switcher"
	@echo ""
	@echo "For autostart run: make install-agent"

uninstall: uninstall-agent
	sudo rm -f $(INSTALL_PATH)/$(BINARY_NAME)
	@echo "Uninstalled"

install-agent:
	@mkdir -p ~/Library/LaunchAgents
	@echo '<?xml version="1.0" encoding="UTF-8"?>' > $(PLIST_PATH)
	@echo '<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">' >> $(PLIST_PATH)
	@echo '<plist version="1.0">' >> $(PLIST_PATH)
	@echo '<dict>' >> $(PLIST_PATH)
	@echo '    <key>Label</key>' >> $(PLIST_PATH)
	@echo '    <string>com.user.fnswitcher</string>' >> $(PLIST_PATH)
	@echo '    <key>ProgramArguments</key>' >> $(PLIST_PATH)
	@echo '    <array>' >> $(PLIST_PATH)
	@echo '        <string>$(INSTALL_PATH)/$(BINARY_NAME)</string>' >> $(PLIST_PATH)
	@echo '    </array>' >> $(PLIST_PATH)
	@echo '    <key>RunAtLoad</key>' >> $(PLIST_PATH)
	@echo '    <true/>' >> $(PLIST_PATH)
	@echo '    <key>KeepAlive</key>' >> $(PLIST_PATH)
	@echo '    <true/>' >> $(PLIST_PATH)
	@echo '</dict>' >> $(PLIST_PATH)
	@echo '</plist>' >> $(PLIST_PATH)
	launchctl load $(PLIST_PATH)
	@echo "LaunchAgent installed and loaded"

uninstall-agent:
	launchctl unload $(PLIST_PATH) 2>/dev/null
	rm -f $(PLIST_PATH)
	@echo "LaunchAgent removed"

clean:
	rm -f $(BINARY_NAME)

status:
	@launchctl list | grep fnswitcher || echo "Not running"

restart: uninstall-agent install-agent
