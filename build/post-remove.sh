#!/bin/sh

remove() {
    printf "\033[32m removing openvoxview\033[0m\n"
    systemctl stop openvoxview.service
    systemctl disable openvoxview.service
    rm -rf /etc/systemd/system/openvoxview.service
    systemctl daemon-reload
}

purge() {
    printf "\033[32m Purgins config files\033[0m\n"
    rm -rf /etc/voxpupuli/openvoxview.yml
}

upgrade() {
    echo ""
}

echo "$@"

action="$1"

case "$action" in
  "0" | "remove")
    remove
    ;;
  "1" | "upgrade")
    upgrade
    ;;
  "purge")
    purge
    ;;
  *)
    printf "\033[32m Alpine\033[0m"
    remove
    ;;
esac
