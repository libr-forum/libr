#!/bin/bash
# Pre-upgrade script for Libr on Arch Linux

echo "Preparing to upgrade libr..."

# Stop any running libr processes
if pgrep -x "libr" > /dev/null; then
    echo "Stopping running libr processes..."
    pkill -f libr || true
fi

# Backup user configuration if it exists
if [ -d "$HOME/.config/libr" ]; then
    echo "Backing up user configuration..."
    cp -r "$HOME/.config/libr" "$HOME/.config/libr.backup.$(date +%s)" || true
fi

exit 0
