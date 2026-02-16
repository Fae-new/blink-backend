package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

// installAgent configures the agent to start automatically on login
func installAgent() error {
	executablePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	// Resolve symlinks just in case
	executablePath, err = filepath.EvalSymlinks(executablePath)
	if err != nil {
		return fmt.Errorf("failed to resolve symlinks: %w", err)
	}

	switch runtime.GOOS {
	case "darwin": // macOS
		return installMacOS(executablePath)
	case "linux": // Linux
		return installLinux(executablePath)
	case "windows": // Windows
		return installWindows(executablePath)
	default:
		return fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}
}

// uninstallAgent removes the auto-start configuration
func uninstallAgent() error {
	switch runtime.GOOS {
	case "darwin":
		return uninstallMacOS()
	case "linux":
		return uninstallLinux()
	case "windows":
		return uninstallWindows()
	default:
		return fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}
}

// --- macOS Implementation ---

func installMacOS(exePath string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	launchAgentsDir := filepath.Join(home, "Library", "LaunchAgents")
	if err := os.MkdirAll(launchAgentsDir, 0755); err != nil {
		return fmt.Errorf("failed to create LaunchAgents directory: %w", err)
	}

	plistPath := filepath.Join(launchAgentsDir, "com.blink.agent.plist")
	
	// Create plist content
	plistContent := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>com.blink.agent</string>
    <key>ProgramArguments</key>
    <array>
        <string>%s</string>
    </array>
    <key>RunAtLoad</key>
    <true/>
    <key>KeepAlive</key>
    <true/>
    <key>StandardOutPath</key>
    <string>/tmp/blink-agent.log</string>
    <key>StandardErrorPath</key>
    <string>/tmp/blink-agent.err</string>
</dict>
</plist>`, exePath)

	if err := os.WriteFile(plistPath, []byte(plistContent), 0644); err != nil {
		return fmt.Errorf("failed to write plist file: %w", err)
	}

	// Load the agent immediately
	cmd := exec.Command("launchctl", "load", plistPath)
	if output, err := cmd.CombinedOutput(); err != nil {
		// If it's already loaded, try unloading first
		_ = exec.Command("launchctl", "unload", plistPath).Run()
		if err := exec.Command("launchctl", "load", plistPath).Run(); err != nil {
			return fmt.Errorf("failed to load launch agent: %v (%s)", err, string(output))
		}
	}

	fmt.Println("✅ macOS LaunchAgent installed at:", plistPath)
	return nil
}

func uninstallMacOS() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	plistPath := filepath.Join(home, "Library", "LaunchAgents", "com.blink.agent.plist")
	
	// Unload first
	exec.Command("launchctl", "unload", plistPath).Run()

	if err := os.Remove(plistPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove plist file: %w", err)
	}

	fmt.Println("✅ Blink Agent removed from auto-start")
	return nil
}

// --- Linux Implementation ---

func installLinux(exePath string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	autostartDir := filepath.Join(home, ".config", "autostart")
	if err := os.MkdirAll(autostartDir, 0755); err != nil {
		return fmt.Errorf("failed to create autostart directory: %w", err)
	}

	desktopPath := filepath.Join(autostartDir, "blink-agent.desktop")
	
	desktopContent := fmt.Sprintf(`[Desktop Entry]
Type=Application
Name=Blink Agent
Exec=%s
Hidden=false
NoDisplay=false
X-GNOME-Autostart-enabled=true
Comment=Local agent for Blink API testing
`, exePath)

	if err := os.WriteFile(desktopPath, []byte(desktopContent), 0644); err != nil {
		return fmt.Errorf("failed to write desktop entry: %w", err)
	}

	fmt.Println("✅ Linux autostart entry created at:", desktopPath)
	
	// Try to start it immediately in background
	cmd := exec.Command(exePath)
	if err := cmd.Start(); err != nil {
		fmt.Println("⚠️ Note: Could not start agent immediately, it will start on next login.")
	} else {
		fmt.Println("🚀 Agent started in background")
	}

	return nil
}

func uninstallLinux() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	desktopPath := filepath.Join(home, ".config", "autostart", "blink-agent.desktop")
	if err := os.Remove(desktopPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove desktop entry: %w", err)
	}

	fmt.Println("✅ Blink Agent removed from auto-start")
	return nil
}

// --- Windows Implementation ---

func installWindows(exePath string) error {
	// PowerShell command to add registry key
	cmdStr := fmt.Sprintf(`Set-ItemProperty -Path 'HKCU:\Software\Microsoft\Windows\CurrentVersion\Run' -Name 'BlinkAgent' -Value '"%s"'`, exePath)
	
	cmd := exec.Command("powershell", "-Command", cmdStr)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to add registry key: %v (%s)", err, string(output))
	}

	fmt.Println("✅ Windows Registry key added for auto-start")
	
	// Try to start immediately
	startCmd := exec.Command("powershell", "-Command", "Start-Process", fmt.Sprintf(`"%s"`, exePath))
	startCmd.Start()

	return nil
}

func uninstallWindows() error {
	cmdStr := `Remove-ItemProperty -Path 'HKCU:\Software\Microsoft\Windows\CurrentVersion\Run' -Name 'BlinkAgent' -ErrorAction SilentlyContinue`
	
	cmd := exec.Command("powershell", "-Command", cmdStr)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to remove registry key: %v (%s)", err, string(output))
	}

	fmt.Println("✅ Blink Agent removed from auto-start")
	return nil
}
