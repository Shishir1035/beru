//go:build linux

package config

func applyPlatformDefaults(cfg *Config) {
	// Clear default global hotkeys to avoid collisions with terminal and
	// application shortcuts (e.g. Ctrl+Shift+C / copy, Ctrl+Shift+V / paste).
	// Users can enable global hotkeys in their config, or bind `neru <mode>`
	// in their desktop environment / window manager.
	cfg.Hotkeys.Bindings = make(map[string][]string)

	// Hint the KDE panel, tray and desktop icons, which belong to plasmashell.
	// Ignored on desktops that don't run it.
	cfg.Hints.IncludeMenubarHints = true
	cfg.Hints.AdditionalMenubarHintsTargets = append(
		cfg.Hints.AdditionalMenubarHintsTargets, "plasmashell")
}
