#!/usr/bin/env bash
# shellcheck disable=SC2034

iso_name="supminer-os"
iso_label="SUPMINER_$(date --date="@${SOURCE_DATE_EPOCH:-$(date +%s)}" +%Y%m)"
iso_publisher="SupMiner OS <https://github.com/supminer-os>"
iso_application="SupMiner OS Live/Rescue DVD"
iso_version="$(date --date="@${SOURCE_DATE_EPOCH:-$(date +%s)}" +%Y.%m.%d)"
install_dir="arch"
buildmodes=('iso')
bootmodes=('bios.syslinux' 'uefi.systemd-boot')
arch="x86_64"
pacman_conf="pacman.conf"
airootfs_image_type="squashfs"
airootfs_image_tool_options=('-comp' 'xz' '-Xbcj' 'x86' '-b' '1M' '-Xdict-size' '1M')
bootstrap_tarball_compression=('zstd' '-c' '-T0' '--auto-threads=logical' '--long' '-19')
file_permissions=(
  ["/etc/shadow"]="0:0:400"
  ["/usr/local/bin/install"]="0:0:755"
  ["/usr/local/bin/supminer-install"]="0:0:755"
  ["/usr/local/bin/supminer-welcome"]="0:0:755"
  ["/root/customize_airootfs.sh"]="0:0:755"
)
