#!/usr/bin/env bash
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
OUT_DIR="$SCRIPT_DIR/out"
WORK_DIR="$SCRIPT_DIR/work"

echo "▓▓▒▒░░ Building SupMiner OS ░░▒▒▓▓"
echo ""

# Refresh pacman keyring on host to avoid marginal trust errors
echo "[0/3] Refreshing pacman keyring..."
pacman-key --init
pacman-key --populate archlinux
pacman-key --refresh-keys 2>/dev/null || true

# Build the Go installer first
echo "[1/3] Building installer..."
cd "$SCRIPT_DIR/installer"
go mod tidy
GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o "$SCRIPT_DIR/airootfs/usr/local/bin/install" .
chmod +x "$SCRIPT_DIR/airootfs/usr/local/bin/install"
cd "$SCRIPT_DIR"

echo "[2/3] Preparing airootfs..."
# Create symlink for supminer-welcome in profile.d
chmod +x "$SCRIPT_DIR/airootfs/usr/local/bin/supminer-welcome"
chmod +x "$SCRIPT_DIR/airootfs/etc/profile.d/welcome.sh"

echo "[3/3] Running mkarchiso..."
mkdir -p "$OUT_DIR" "$WORK_DIR"
mkarchiso -v -w "$WORK_DIR" -o "$OUT_DIR" "$SCRIPT_DIR"

echo ""
echo "░░▒▒▓▓ Build complete! ISO: $OUT_DIR ▓▓▒▒░░"
