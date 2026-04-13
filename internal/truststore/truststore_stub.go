//go:build !windows
// +build !windows

package truststore

import (
	"errors"
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
	return false, errors.New("not implemented on this platform")
}

func (tm *TrustStoreManager) Install() error {
	return errors.New("not implemented on this platform")
}

func (tm *TrustStoreManager) Uninstall() error {
	return errors.New("not implemented on this platform")
}

func IsRunningAsAdmin() bool {
	return false
}
