# SupMiner OS

Minimal Arch-based distro with GNOME and a beautiful TUI installer.

## Build

Requires: `archiso`, `go` (on an Arch Linux host)

```bash
sudo bash build.sh
```

ISO will be in `out/`.

## Installer

Boot the ISO, login as root, then:

```bash
install
```

## What gets installed

- Arch Linux base + GNOME
- SupMiner black/white theme
- Wallpapers cloned from GitHub
- fastfetch + neofetch with custom SupMiner logo
- NetworkManager, pipewire, all essentials
