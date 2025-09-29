#!/usr/bin/fish

# Colors for output
set RED '\033[0;31m'
set GREEN '\033[0;32m'
set YELLOW '\033[1;33m'
set BLUE '\033[0;34m'
set NC '\033[0m' # No Color

echo -e "$BLUE""Uninstalling Cloudflare IP Updater...$NC"

# Stop and disable the timer first
echo -e "$YELLOW""Stopping and disabling systemd timer...$NC"
systemctl --user stop cloudflare-ip-updater.timer 2>/dev/null
systemctl --user stop cloudflare-ip-updater.service 2>/dev/null
systemctl --user disable cloudflare-ip-updater.timer 2>/dev/null

# Remove systemd files
echo -e "$YELLOW""Removing systemd service and timer files...$NC"
set service_file "$HOME/.config/systemd/user/cloudflare-ip-updater.service"
set timer_file "$HOME/.config/systemd/user/cloudflare-ip-updater.timer"

if test -f $service_file
    rm $service_file
    echo -e "$GREEN""✓ Removed $service_file$NC"
else
    echo -e "$YELLOW""- Service file not found: $service_file$NC"
end

if test -f $timer_file
    rm $timer_file
    echo -e "$GREEN""✓ Removed $timer_file$NC"
else
    echo -e "$YELLOW""- Timer file not found: $timer_file$NC"
end

# Reload systemd to reflect changes
echo -e "$YELLOW""Reloading systemd user daemon...$NC"
systemctl --user daemon-reload

echo -e "$GREEN""✓ Uninstallation complete!$NC"

echo -e "\n$YELLOW""To view remaining logs:$NC"
echo -e "  $BLUE""journalctl --user -u cloudflare-ip-updater.service$NC"

echo -e "\n$YELLOW""To clean up logs (optional):$NC"
echo -e "  $BLUE""journalctl --user --vacuum-time=1d$NC"
