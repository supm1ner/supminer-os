package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ── Styles ────────────────────────────────────────────────────────────────────

var (
	white     = lipgloss.Color("#FFFFFF")
	lightGray = lipgloss.Color("#CCCCCC")
	gray      = lipgloss.Color("#888888")
	darkGray  = lipgloss.Color("#444444")
	black     = lipgloss.Color("#000000")

	titleStyle = lipgloss.NewStyle().
			Foreground(white).
			Bold(true).
			Padding(0, 2)

	boxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(white).
			Padding(1, 3)

	selectedStyle = lipgloss.NewStyle().
			Foreground(black).
			Background(white).
			Bold(true).
			Padding(0, 1)

	normalStyle = lipgloss.NewStyle().
			Foreground(lightGray).
			Padding(0, 1)

	dimStyle = lipgloss.NewStyle().
			Foreground(gray)

	errorStyle = lipgloss.NewStyle().
			Foreground(lightGray).
			Bold(true)

	successStyle = lipgloss.NewStyle().
			Foreground(white).
			Bold(true)

	headerStyle = lipgloss.NewStyle().
			Foreground(white).
			Bold(true).
			BorderBottom(true).
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(darkGray).
			Width(60)

	progressBarFill  = lipgloss.NewStyle().Background(white).Foreground(black)
	progressBarEmpty = lipgloss.NewStyle().Background(darkGray).Foreground(gray)
)

// ── ASCII Logo ────────────────────────────────────────────────────────────────

const logo = `
  ██████████████████████████████████████████████
  ██                                          ██
  ██   ███████╗██╗   ██╗██████╗ ███╗   ███╗  ██
  ██   ██╔════╝██║   ██║██╔══██╗████╗ ████║  ██
  ██   ███████╗██║   ██║██████╔╝██╔████╔██║  ██
  ██   ╚════██║██║   ██║██╔═══╝ ██║╚██╔╝██║  ██
  ██   ███████║╚██████╔╝██║     ██║ ╚═╝ ██║  ██
  ██   ╚══════╝ ╚═════╝ ╚═╝     ╚═╝     ╚═╝  ██
  ██                  ░▒▓█ OS █▓▒░            ██
  ██████████████████████████████████████████████
  ▓▓▓▓▒▒▒▒░░░░  I N S T A L L E R  ░░░░▒▒▒▒▓▓▓▓
`

// ── Step definitions ──────────────────────────────────────────────────────────

type step int

const (
	stepWelcome step = iota
	stepDisk
	stepPartitionMode
	stepTimezone
	stepLocale
	stepHostname
	stepUser
	stepPassword
	stepConfirm
	stepInstalling
	stepDone
)

var stepNames = []string{
	"Welcome", "Disk", "Partition", "Timezone",
	"Locale", "Hostname", "User", "Password", "Confirm", "Install", "Done",
}

// ── Model ─────────────────────────────────────────────────────────────────────

type model struct {
	step         step
	cursor       int
	disks        []string
	selectedDisk string
	partMode     string // auto | manual
	timezone     string
	locale       string
	hostname     string
	username     string
	password     string
	passConfirm  string
	inputMode    bool
	inputBuffer  string
	inputField   string // which field is being edited
	errMsg       string
	progress     int
	progressMax  int
	progressMsg  string
	installLog   []string
	width        int
	height       int
}

// ── Init ──────────────────────────────────────────────────────────────────────

func initialModel() model {
	disks := getDisks()
	return model{
		step:        stepWelcome,
		disks:       disks,
		timezone:    "Europe/Moscow",
		locale:      "en_US.UTF-8",
		hostname:    "supminer",
		username:    "user",
		progressMax: 20,
	}
}

func getDisks() []string {
	out, err := exec.Command("lsblk", "-dpno", "NAME,SIZE,MODEL").Output()
	if err != nil {
		return []string{"/dev/sda", "/dev/nvme0n1"}
	}
	var disks []string
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if line != "" {
			disks = append(disks, line)
		}
	}
	if len(disks) == 0 {
		return []string{"No disks found"}
	}
	return disks
}

func (m model) Init() tea.Cmd {
	return nil
}

// ── Messages ──────────────────────────────────────────────────────────────────

type progressMsg struct {
	step int
	msg  string
}
type doneMsg struct{}
type errMsg struct{ err string }

// ── Update ────────────────────────────────────────────────────────────────────

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case progressMsg:
		m.progress = msg.step
		m.progressMsg = msg.msg
		m.installLog = append(m.installLog, msg.msg)
		if msg.step < m.progressMax {
			return m, runInstallStep(m, msg.step+1)
		}
		return m, func() tea.Msg { return doneMsg{} }

	case doneMsg:
		m.step = stepDone

	case errMsg:
		m.errMsg = msg.err
		m.step = stepConfirm

	case tea.KeyMsg:
		return m.handleKey(msg)
	}
	return m, nil
}

func (m model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.inputMode {
		return m.handleInput(msg)
	}

	switch msg.String() {
	case "ctrl+c", "q":
		if m.step != stepInstalling {
			return m, tea.Quit
		}
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.step == stepDisk && m.cursor < len(m.disks)-1 {
			m.cursor++
		}
	case "enter", " ":
		return m.handleEnter()
	case "tab":
		return m.handleTab()
	}
	return m, nil
}

func (m model) handleInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		switch m.inputField {
		case "timezone":
			m.timezone = m.inputBuffer
		case "locale":
			m.locale = m.inputBuffer
		case "hostname":
			m.hostname = m.inputBuffer
		case "username":
			m.username = m.inputBuffer
		case "password":
			m.password = m.inputBuffer
		case "passconfirm":
			m.passConfirm = m.inputBuffer
		}
		m.inputMode = false
		m.inputBuffer = ""
		m.errMsg = ""
	case "esc":
		m.inputMode = false
		m.inputBuffer = ""
	case "backspace":
		if len(m.inputBuffer) > 0 {
			m.inputBuffer = m.inputBuffer[:len(m.inputBuffer)-1]
		}
	default:
		if len(msg.String()) == 1 {
			m.inputBuffer += msg.String()
		}
	}
	return m, nil
}

func (m model) handleTab() (model, tea.Cmd) {
	// cycle through fields on multi-field screens
	return m, nil
}

func (m model) handleEnter() (model, tea.Cmd) {
	m.errMsg = ""
	switch m.step {
	case stepWelcome:
		m.step = stepDisk
		m.cursor = 0

	case stepDisk:
		if len(m.disks) > 0 && m.disks[0] != "No disks found" {
			parts := strings.Fields(m.disks[m.cursor])
			if len(parts) > 0 {
				m.selectedDisk = parts[0]
			}
		}
		m.step = stepPartitionMode
		m.cursor = 0

	case stepPartitionMode:
		if m.cursor == 0 {
			m.partMode = "auto"
		} else {
			m.partMode = "manual"
		}
		m.step = stepTimezone
		m.inputMode = true
		m.inputField = "timezone"
		m.inputBuffer = m.timezone

	case stepTimezone:
		m.step = stepLocale
		m.inputMode = true
		m.inputField = "locale"
		m.inputBuffer = m.locale

	case stepLocale:
		m.step = stepHostname
		m.inputMode = true
		m.inputField = "hostname"
		m.inputBuffer = m.hostname

	case stepHostname:
		m.step = stepUser
		m.inputMode = true
		m.inputField = "username"
		m.inputBuffer = m.username

	case stepUser:
		m.step = stepPassword
		m.inputMode = true
		m.inputField = "password"
		m.inputBuffer = ""

	case stepPassword:
		if m.password == "" {
			m.errMsg = "Password cannot be empty"
			return m, nil
		}
		m.step = stepConfirm

	case stepConfirm:
		if m.cursor == 0 {
			m.step = stepInstalling
			m.progress = 0
			m.installLog = nil
			return m, runInstallStep(m, 0)
		}
		return m, tea.Quit
	}
	return m, nil
}

// ── Install steps ─────────────────────────────────────────────────────────────

var installSteps = []struct {
	msg string
	fn  func(m model) error
}{
	{"Wiping disk signatures...", stepWipeDisk},
	{"Creating partition table...", stepPartitionDisk},
	{"Formatting EFI partition...", stepFormatEFI},
	{"Formatting root partition...", stepFormatRoot},
	{"Mounting filesystems...", stepMount},
	{"Updating mirrors...", stepMirrors},
	{"Installing base system...", stepPacstrap},
	{"Generating fstab...", stepFstab},
	{"Setting timezone...", stepTimezoneSet},
	{"Generating locales...", stepLocaleGen},
	{"Setting hostname...", stepHostnameSet},
	{"Configuring hosts...", stepHosts},
	{"Installing bootloader...", stepGrub},
	{"Creating user...", stepCreateUser},
	{"Enabling services...", stepServices},
	{"Cloning wallpapers...", stepWallpapers},
	{"Applying GNOME theme...", stepGnomeTheme},
	{"Installing GNOME extensions...", stepGnomeExtensions},
	{"Copying installer...", stepCopyInstaller},
	{"Finalizing...", stepFinalize},
}

func runInstallStep(m model, idx int) tea.Cmd {
	return func() tea.Msg {
		if idx >= len(installSteps) {
			return doneMsg{}
		}
		s := installSteps[idx]
		if err := s.fn(m); err != nil {
			return errMsg{err: fmt.Sprintf("Step %d (%s): %v", idx, s.msg, err)}
		}
		return progressMsg{step: idx, msg: s.msg}
	}
}

func chroot(m model, args ...string) error {
	cmd := exec.Command("arch-chroot", append([]string{"/mnt"}, args...)...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func run(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func writeFile(path, content string) error {
	return os.WriteFile(path, []byte(content), 0644)
}

func stepWipeDisk(m model) error {
	return run("wipefs", "-af", m.selectedDisk)
}

func stepPartitionDisk(m model) error {
	script := fmt.Sprintf(`g
n
1

+512M
t
1
n
2


w
`)
	cmd := exec.Command("fdisk", m.selectedDisk)
	cmd.Stdin = strings.NewReader(script)
	return cmd.Run()
}

func efiPart(disk string) string {
	if strings.Contains(disk, "nvme") {
		return disk + "p1"
	}
	return disk + "1"
}

func rootPart(disk string) string {
	if strings.Contains(disk, "nvme") {
		return disk + "p2"
	}
	return disk + "2"
}

func stepFormatEFI(m model) error {
	return run("mkfs.fat", "-F32", efiPart(m.selectedDisk))
}

func stepFormatRoot(m model) error {
	return run("mkfs.ext4", "-F", rootPart(m.selectedDisk))
}

func stepMount(m model) error {
	if err := run("mount", rootPart(m.selectedDisk), "/mnt"); err != nil {
		return err
	}
	if err := run("mkdir", "-p", "/mnt/boot/efi"); err != nil {
		return err
	}
	return run("mount", efiPart(m.selectedDisk), "/mnt/boot/efi")
}

func stepMirrors(m model) error {
	return run("reflector", "--latest", "10", "--sort", "rate", "--save", "/etc/pacman.d/mirrorlist")
}

func stepPacstrap(m model) error {
	return run("pacstrap", "-K", "/mnt",
		"base", "base-devel", "linux", "linux-firmware",
		"grub", "efibootmgr", "networkmanager", "gnome", "gnome-tweaks",
		"gnome-shell-extensions", "gdm", "git", "curl", "fastfetch",
		"sudo", "nano", "vim", "bash-completion", "noto-fonts", "noto-fonts-emoji",
		"ttf-jetbrains-mono-nerd", "pipewire", "pipewire-pulse", "wireplumber",
		"go",
	)
}

func stepFstab(m model) error {
	out, err := exec.Command("genfstab", "-U", "/mnt").Output()
	if err != nil {
		return err
	}
	return os.WriteFile("/mnt/etc/fstab", out, 0644)
}

func stepTimezoneSet(m model) error {
	if err := chroot(m, "ln", "-sf",
		"/usr/share/zoneinfo/"+m.timezone, "/etc/localtime"); err != nil {
		return err
	}
	return chroot(m, "hwclock", "--systohc")
}

func stepLocaleGen(m model) error {
	content := m.locale + " UTF-8\n"
	if err := writeFile("/mnt/etc/locale.gen", content); err != nil {
		return err
	}
	if err := writeFile("/mnt/etc/locale.conf", "LANG="+m.locale+"\n"); err != nil {
		return err
	}
	return chroot(m, "locale-gen")
}

func stepHostnameSet(m model) error {
	return writeFile("/mnt/etc/hostname", m.hostname+"\n")
}

func stepHosts(m model) error {
	content := fmt.Sprintf("127.0.0.1\tlocalhost\n::1\t\tlocalhost\n127.0.1.1\t%s.localdomain\t%s\n",
		m.hostname, m.hostname)
	return writeFile("/mnt/etc/hosts", content)
}

func stepGrub(m model) error {
	if err := chroot(m, "grub-install", "--target=x86_64-efi",
		"--efi-directory=/boot/efi", "--bootloader-id=SUPMINER"); err != nil {
		return err
	}
	// Custom GRUB theme colors
	grubConf := `GRUB_DEFAULT=0
GRUB_TIMEOUT=3
GRUB_DISTRIBUTOR="SupMiner OS"
GRUB_CMDLINE_LINUX_DEFAULT="quiet splash"
GRUB_CMDLINE_LINUX=""
GRUB_TERMINAL_OUTPUT="console"
GRUB_COLOR_NORMAL="white/black"
GRUB_COLOR_HIGHLIGHT="black/white"
`
	if err := writeFile("/mnt/etc/default/grub", grubConf); err != nil {
		return err
	}
	return chroot(m, "grub-mkconfig", "-o", "/boot/grub/grub.cfg")
}

func stepCreateUser(m model) error {
	if err := chroot(m, "useradd", "-m", "-G", "wheel,audio,video,storage",
		"-s", "/bin/bash", m.username); err != nil {
		return err
	}
	// set password via chpasswd
	cmd := exec.Command("arch-chroot", "/mnt", "chpasswd")
	cmd.Stdin = strings.NewReader(fmt.Sprintf("%s:%s\nroot:%s\n", m.username, m.password, m.password))
	return cmd.Run()
}

func stepServices(m model) error {
	services := []string{"NetworkManager", "gdm"}
	for _, svc := range services {
		if err := chroot(m, "systemctl", "enable", svc); err != nil {
			return err
		}
	}
	// sudoers
	sudoers := "%wheel ALL=(ALL:ALL) ALL\n"
	return writeFile("/mnt/etc/sudoers.d/wheel", sudoers)
}

func stepWallpapers(m model) error {
	return chroot(m, "git", "clone", "--depth=1",
		"https://github.com/supminer-os/wallpapers",
		"/usr/share/backgrounds/supminer")
}

func stepGnomeTheme(m model) error {
	// dconf settings via script
	script := `#!/bin/bash
export DBUS_SESSION_BUS_ADDRESS="unix:path=/run/user/$(id -u)/bus"
gsettings set org.gnome.desktop.interface color-scheme 'prefer-dark'
gsettings set org.gnome.desktop.interface gtk-theme 'Adwaita-dark'
gsettings set org.gnome.desktop.interface icon-theme 'Adwaita'
gsettings set org.gnome.desktop.background picture-uri-dark 'file:///usr/share/backgrounds/supminer/default.jpg'
gsettings set org.gnome.shell.extensions.user-theme name 'Adwaita-dark'
`
	path := "/mnt/tmp/gnome-theme.sh"
	if err := writeFile(path, script); err != nil {
		return err
	}
	_ = os.Chmod(path, 0755)
	return nil
}

func stepGnomeExtensions(m model) error {
	// Install popular extensions via gnome-extensions-cli or direct download
	// We'll set up a post-install script that runs on first login
	script := `#!/bin/bash
# SupMiner OS - First login GNOME setup
EXT_IDS=(
  "user-theme@gnome-shell-extensions.gcampax.github.com"
  "dash-to-dock@micxgx.gmail.com"
  "blur-my-shell@aunetx"
)
for id in "${EXT_IDS[@]}"; do
  gnome-extensions enable "$id" 2>/dev/null || true
done
# Remove this script after first run
rm -f ~/.config/autostart/supminer-firstrun.desktop
`
	if err := os.MkdirAll("/mnt/etc/skel/.config/autostart", 0755); err != nil {
		return err
	}
	if err := writeFile("/mnt/tmp/firstrun.sh", script); err != nil {
		return err
	}
	desktop := `[Desktop Entry]
Type=Application
Name=SupMiner First Run
Exec=/tmp/firstrun.sh
Hidden=false
NoDisplay=false
X-GNOME-Autostart-enabled=true
`
	return writeFile("/mnt/etc/skel/.config/autostart/supminer-firstrun.desktop", desktop)
}

func stepCopyInstaller(m model) error {
	// Copy welcome script and installer binary
	data, err := os.ReadFile("/usr/local/bin/supminer-welcome")
	if err == nil {
		_ = writeFile("/mnt/usr/local/bin/supminer-welcome", string(data))
		_ = os.Chmod("/mnt/usr/local/bin/supminer-welcome", 0755)
	}
	// Copy fastfetch & neofetch configs
	_ = run("cp", "-r", "/root/.config/fastfetch", "/mnt/etc/skel/.config/")
	_ = run("cp", "-r", "/root/.config/neofetch", "/mnt/etc/skel/.config/")
	return nil
}

func stepFinalize(m model) error {
	// motd
	motd := "\n  Welcome to SupMiner OS. Type 'fastfetch' for system info.\n\n"
	return writeFile("/mnt/etc/motd", motd)
}

// ── View ──────────────────────────────────────────────────────────────────────

func (m model) View() string {
	var b strings.Builder

	// Logo header
	b.WriteString(lipgloss.NewStyle().Foreground(white).Render(logo))
	b.WriteString("\n")

	switch m.step {
	case stepWelcome:
		b.WriteString(m.viewWelcome())
	case stepDisk:
		b.WriteString(m.viewDisk())
	case stepPartitionMode:
		b.WriteString(m.viewPartitionMode())
	case stepTimezone:
		b.WriteString(m.viewInput("Timezone", "Enter timezone (e.g. Europe/Moscow):", m.timezone))
	case stepLocale:
		b.WriteString(m.viewInput("Locale", "Enter locale (e.g. en_US.UTF-8):", m.locale))
	case stepHostname:
		b.WriteString(m.viewInput("Hostname", "Enter hostname:", m.hostname))
	case stepUser:
		b.WriteString(m.viewInput("Username", "Enter username:", m.username))
	case stepPassword:
		b.WriteString(m.viewPassword())
	case stepConfirm:
		b.WriteString(m.viewConfirm())
	case stepInstalling:
		b.WriteString(m.viewInstalling())
	case stepDone:
		b.WriteString(m.viewDone())
	}

	b.WriteString("\n")
	if m.step != stepInstalling && m.step != stepDone {
		b.WriteString(dimStyle.Render("  [↑↓] navigate  [Enter] confirm  [q] quit"))
	}
	return b.String()
}

func (m model) viewWelcome() string {
	content := lipgloss.JoinVertical(lipgloss.Left,
		headerStyle.Render("Welcome to SupMiner OS Installer"),
		"",
		normalStyle.Render("This installer will guide you through setting up"),
		normalStyle.Render("SupMiner OS on your machine."),
		"",
		normalStyle.Render("What will be installed:"),
		dimStyle.Render("  • Arch Linux base system"),
		dimStyle.Render("  • GNOME desktop environment"),
		dimStyle.Render("  • SupMiner theme & wallpapers"),
		dimStyle.Render("  • Essential applications"),
		"",
		selectedStyle.Render("  Press Enter to begin  "),
	)
	return boxStyle.Render(content)
}

func (m model) viewDisk() string {
	var items []string
	items = append(items, headerStyle.Render("Select Installation Disk"))
	items = append(items, "")
	items = append(items, dimStyle.Render("  WARNING: Selected disk will be ERASED"))
	items = append(items, "")
	for i, d := range m.disks {
		if i == m.cursor {
			items = append(items, selectedStyle.Render("  ▶ "+d))
		} else {
			items = append(items, normalStyle.Render("    "+d))
		}
	}
	return boxStyle.Render(lipgloss.JoinVertical(lipgloss.Left, items...))
}

func (m model) viewPartitionMode() string {
	modes := []string{"Auto (recommended) — wipe & partition automatically", "Manual — use fdisk/cfdisk"}
	var items []string
	items = append(items, headerStyle.Render("Partition Mode"))
	items = append(items, "")
	for i, mode := range modes {
		if i == m.cursor {
			items = append(items, selectedStyle.Render("  ▶ "+mode))
		} else {
			items = append(items, normalStyle.Render("    "+mode))
		}
	}
	return boxStyle.Render(lipgloss.JoinVertical(lipgloss.Left, items...))
}

func (m model) viewInput(title, prompt, current string) string {
	display := current
	if m.inputMode {
		display = m.inputBuffer + "█"
	}
	content := lipgloss.JoinVertical(lipgloss.Left,
		headerStyle.Render(title),
		"",
		normalStyle.Render(prompt),
		"",
		boxStyle.Copy().BorderForeground(lightGray).Render("  "+display),
		"",
		dimStyle.Render("  Press Enter to confirm"),
	)
	if m.errMsg != "" {
		content += "\n" + errorStyle.Render("  ✗ "+m.errMsg)
	}
	return boxStyle.Render(content)
}

func (m model) viewPassword() string {
	display := strings.Repeat("•", len(m.password))
	if m.inputMode {
		display = strings.Repeat("•", len(m.inputBuffer)) + "█"
	}
	content := lipgloss.JoinVertical(lipgloss.Left,
		headerStyle.Render("Set Password"),
		"",
		normalStyle.Render("Enter password for root and your user:"),
		"",
		boxStyle.Copy().BorderForeground(lightGray).Render("  "+display),
		"",
		dimStyle.Render("  Press Enter to confirm"),
	)
	if m.errMsg != "" {
		content += "\n" + errorStyle.Render("  ✗ "+m.errMsg)
	}
	return boxStyle.Render(content)
}

func (m model) viewConfirm() string {
	opts := []string{"  ▶ Install now", "  ✗ Cancel"}
	var items []string
	items = append(items, headerStyle.Render("Confirm Installation"))
	items = append(items, "")
	items = append(items, dimStyle.Render(fmt.Sprintf("  Disk:     %s", m.selectedDisk)))
	items = append(items, dimStyle.Render(fmt.Sprintf("  Mode:     %s", m.partMode)))
	items = append(items, dimStyle.Render(fmt.Sprintf("  Timezone: %s", m.timezone)))
	items = append(items, dimStyle.Render(fmt.Sprintf("  Locale:   %s", m.locale)))
	items = append(items, dimStyle.Render(fmt.Sprintf("  Hostname: %s", m.hostname)))
	items = append(items, dimStyle.Render(fmt.Sprintf("  User:     %s", m.username)))
	items = append(items, "")
	if m.errMsg != "" {
		items = append(items, errorStyle.Render("  ✗ "+m.errMsg))
		items = append(items, "")
	}
	for i, opt := range opts {
		if i == m.cursor {
			items = append(items, selectedStyle.Render(opt))
		} else {
			items = append(items, normalStyle.Render(opt))
		}
	}
	return boxStyle.Render(lipgloss.JoinVertical(lipgloss.Left, items...))
}

func (m model) viewInstalling() string {
	total := len(installSteps)
	pct := 0
	if total > 0 {
		pct = (m.progress * 100) / total
	}
	barWidth := 50
	filled := (barWidth * pct) / 100
	bar := progressBarFill.Render(strings.Repeat("█", filled)) +
		progressBarEmpty.Render(strings.Repeat("░", barWidth-filled))

	var logLines []string
	start := len(m.installLog) - 6
	if start < 0 {
		start = 0
	}
	for _, l := range m.installLog[start:] {
		logLines = append(logLines, dimStyle.Render("  "+l))
	}

	content := lipgloss.JoinVertical(lipgloss.Left,
		headerStyle.Render("Installing SupMiner OS"),
		"",
		normalStyle.Render(fmt.Sprintf("  %d%%  %s", pct, m.progressMsg)),
		"",
		"  "+bar,
		"",
	)
	content += lipgloss.JoinVertical(lipgloss.Left, logLines...)
	return boxStyle.Render(content)
}

func (m model) viewDone() string {
	content := lipgloss.JoinVertical(lipgloss.Left,
		headerStyle.Render("Installation Complete"),
		"",
		successStyle.Render("  ✓ SupMiner OS has been installed successfully!"),
		"",
		normalStyle.Render("  Remove the installation media and reboot:"),
		"",
		selectedStyle.Render("  reboot  "),
		"",
		dimStyle.Render("  Press q to exit the installer"),
	)
	return boxStyle.Render(content)
}

// ── Main ──────────────────────────────────────────────────────────────────────

func main() {
	if os.Geteuid() != 0 {
		fmt.Fprintln(os.Stderr, "\033[1;37mSupMiner Installer requires root privileges.\033[0m")
		fmt.Fprintln(os.Stderr, "Run: sudo install")
		os.Exit(1)
	}

	p := tea.NewProgram(
		initialModel(),
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
