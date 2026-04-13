//go:build windows
// +build windows

package truststore

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"syscall"
)

var (
	ErrNotAdmin = errors.New("administrator privileges required")
)

type TrustStoreManager struct {
	certPath string
}

func NewTrustStoreManager(certPath string) *TrustStoreManager {
	return &TrustStoreManager{
		certPath: certPath,
	}
}

func (tm *TrustStoreManager) IsInstalled() (bool, error) {
	cmd := exec.Command("certutil", "-store", "-user", "Root")
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	output, err := cmd.CombinedOutput()
	if err != nil {
		return false, fmt.Errorf("failed to check certificate store: %w", err)
	}

	return strings.Contains(string(output), "Trae Switch Root CA"), nil
}

func (tm *TrustStoreManager) Install() error {
	installed, err := tm.IsInstalled()
	if err != nil {
		return err
	}
	if installed {
		return nil
	}

	cmd := exec.Command("certutil", "-addstore", "-user", "Root", tm.certPath)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to install certificate: %w, output: %s", err, string(output))
	}

	return nil
}

func (tm *TrustStoreManager) Uninstall() error {
	installed, err := tm.IsInstalled()
	if err != nil {
		return err
	}
	if !installed {
		return nil
	}

	cmd := exec.Command("certutil", "-delstore", "-user", "Root", "Trae Switch Root CA")
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to uninstall certificate: %w, output: %s", err, string(output))
	}

	return nil
}

func IsRunningAsAdmin() bool {
	cmd := exec.Command("net", "session")
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	err := cmd.Run()
	return err == nil
}
