// clipboard_windows.go
//go:build windows
// +build windows

package tui

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"
)

func copyToClipboard(text string) error {
    if runtime.GOOS != "windows" {
        return nil
    }
    
    // Попробуем несколько способов
    cmd := exec.Command("cmd", "/c", "echo", text, "|", "clip")
    if err := cmd.Run(); err == nil {
        return nil
    }
    
    cmd = exec.Command("powershell", "-Command", "Set-Clipboard", "-Value", text)
    if err := cmd.Run(); err == nil {
        return nil
    }
    
    escapedText := strings.ReplaceAll(text, `"`, `\"`)
    cmd = exec.Command("powershell", "-Command", fmt.Sprintf(`Add-Type -AssemblyName System.Windows.Forms; [System.Windows.Forms.Clipboard]::SetText("%s")`, escapedText))
    if err := cmd.Run(); err == nil {
        return nil
    }
    
    return fmt.Errorf("все методы копирования не сработали")
}
