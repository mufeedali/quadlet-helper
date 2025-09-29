#!/usr/bin/fish

# Colors for output
set RED '\033[0;31m'
set GREEN '\033[0;32m'
set YELLOW '\033[1;33m'
set BLUE '\033[0;34m'
set NC '\033[0m' # No Color

echo -e "$BLUE""Installing Cloudflare IP Updater...$NC"

# Check Python dependencies
echo -e "$YELLOW""Checking Python dependencies...$NC"
if not python3 -c "import requests, yaml" 2>/dev/null
    echo -e "$RED""Missing required Python packages: requests, pyyaml$NC"
    echo -e "$RED""Please install with: pip install requests pyyaml$NC"
    exit 1
end
echo -e "$GREEN""✓ Python dependencies found$NC"

# Copy systemd files
echo -e "$YELLOW""Installing systemd service and timer...$NC"

# Get the absolute path of the current directory
set INSTALL_DIR (pwd)
echo -e "$BLUE""Installing from: $INSTALL_DIR$NC"

# Create a temporary service file with the correct path
sed "s|__INSTALL_DIR__|$INSTALL_DIR|g" cloudflare-ip-updater.service > /tmp/cloudflare-ip-updater.service.tmp
cp /tmp/cloudflare-ip-updater.service.tmp ~/.config/systemd/user/cloudflare-ip-updater.service
rm /tmp/cloudflare-ip-updater.service.tmp

cp cloudflare-ip-updater.timer ~/.config/systemd/user/

# Reload systemd and enable timer
systemctl --user daemon-reload
systemctl --user enable cloudflare-ip-updater.timer
systemctl --user start cloudflare-ip-updater.timer

echo -e "$GREEN""✓ Installation complete!$NC"
echo -e "$BLUE""Timer status:$NC"
systemctl --user --no-pager status cloudflare-ip-updater.timer

echo -e "\n$YELLOW""Usage:$NC"
echo -e "  Manual run:    $BLUE""./update-cloudflare-ips.py$NC"
echo -e "  Check timer:   $BLUE""systemctl --user list-timers cloudflare-ip-updater.timer$NC"
echo -e "  View logs:     $BLUE""journalctl --user -u cloudflare-ip-updater.service$NC"
echo -e "  Test now:      $BLUE""systemctl --user start cloudflare-ip-updater.service$NC"
