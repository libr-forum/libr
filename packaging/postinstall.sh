#!/bin/bash
# Post-install script for Libr

# Update desktop database
if command -v update-desktop-database >/dev/null 2>&1; then
    update-desktop-database /usr/share/applications
fi

# Update icon cache
if command -v gtk-update-icon-cache >/dev/null 2>&1; then
    gtk-update-icon-cache -f -t /usr/share/icons/hicolor
fi

# Update MIME database if needed
if command -v update-mime-database >/dev/null 2>&1; then
    update-mime-database /usr/share/mime
fi

exit 0
