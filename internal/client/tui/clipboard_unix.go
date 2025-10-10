//go:build !windows
// +build !windows

package tui

import (
	"os/exec"
	"runtime"
)

func copyToClipboard(text string) error {
    var cmd *exec.Cmd
    
    switch runtime.GOOS {
    case "darwin":
        cmd = exec.Command("pbcopy")
    case "linux":
        // Проверяем доступность xclip или xsel
        if _, err := exec.LookPath("xclip"); err == nil {
            cmd = exec.Command("xclip", "-selection", "clipboard")
        } else if _, err := exec.LookPath("xsel"); err == nil {
            cmd = exec.Command("xsel", "--clipboard", "--input")
        } else {
            return nil // Не поддерживается
        }
    default:
        return nil // Не поддерживается
    }
    
    stdin, err := cmd.StdinPipe()
    if err != nil {
        return err
    }
    
    if err := cmd.Start(); err != nil {
        return err
    }
    
    if _, err := stdin.Write([]byte(text)); err != nil {
        return err
    }
    
    stdin.Close()
    return cmd.Wait()
}
