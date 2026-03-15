#!/usr/bin/env bash
set -e

# Generate initramfs
mkinitcpio -P

# Enable services
systemctl enable NetworkManager
systemctl enable gdm

# Set root password to empty (auto-login on live)
passwd -d root

# Set locale
echo "en_US.UTF-8 UTF-8" > /etc/locale.gen
locale-gen
echo "LANG=en_US.UTF-8" > /etc/locale.conf
