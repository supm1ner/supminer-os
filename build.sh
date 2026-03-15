#!/usr/bin/env bash
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
OUT_DIR="$SCRIPT_DIR/out"
WORK_DIR="$SCRIPT_DIR/work"

echo "▓▓▒▒░░ Building SupMiner OS ░░▒▒▓▓"
echo ""

# Fix keyring trust issues (marginal trust on Daniel Bermond's key)
echo "[0/3] Fixing pacman keyring..."
pacman-key --init
pacman-key --populate archlinux
# Locally sign the problematic key to elevate from marginal to full trust
pacman-key --lsign-key E85B8683EB48BC95 2>/dev/null || true

# Build the Go installer first
echo "[1/3] Building installer..."
cd "$SCRIPT_DIR/installer"
go mod tidy
GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o "$SCRIPT_DIR/airootfs/usr/local/bin/supminer-install" .
chmod +x "$SCRIPT_DIR/airootfs/usr/local/bin/supminer-install"
cd "$SCRIPT_DIR"

echo "[2/3] Preparing airootfs..."
# Create symlink for supminer-welcome in profile.d
chmod +x "$SCRIPT_DIR/airootfs/usr/local/bin/supminer-welcome"
chmod +x "$SCRIPT_DIR/airootfs/etc/profile.d/welcome.sh"

echo "[3/3] Running mkarchiso..."
mkdir -p "$OUT_DIR"
# Unmount any leftover mounts from previous failed build
umount -R "$WORK_DIR/x86_64/airootfs/proc" 2>/dev/null || true
umount -R "$WORK_DIR/x86_64/airootfs/sys"  2>/dev/null || true
umount -R "$WORK_DIR/x86_64/airootfs/dev"  2>/dev/null || true
umount -R "$WORK_DIR/x86_64/airootfs/tmp"  2>/dev/null || true
umount -R "$WORK_DIR/x86_64/airootfs/run"  2>/dev/null || true
rm -rf "$WORK_DIR" 2>/dev/null || true
mkdir -p "$WORK_DIR"
mkarchiso -v -w "$WORK_DIR" -o "$OUT_DIR" "$SCRIPT_DIR"

echo ""
echo "░░▒▒▓▓ Build complete! ISO: $OUT_DIR ▓▓▒▒░░"
