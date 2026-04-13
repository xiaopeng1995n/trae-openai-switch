package hosts

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os"
	"runtime"
	"strings"
)

const (
	HostsMarkerStart = "# === Trae Switch Start ==="
	HostsMarkerEnd   = "# === Trae Switch End ==="
	TargetDomain     = "api.openai.com"
	TargetIP         = "127.0.0.1"
)

var (
	ErrNotWindows = errors.New("this feature is only supported on Windows")
)

type HostsManager struct {
	hostsPath string
	original  []byte
}

func NewHostsManager() *HostsManager {
	hm := &HostsManager{}
	if runtime.GOOS == "windows" {
		hm.hostsPath = `C:\Windows\System32\drivers\etc\hosts`
	} else {
		hm.hostsPath = "/etc/hosts"
	}
	return hm
}

func (hm *HostsManager) GetHostsPath() string {
	return hm.hostsPath
}

func (hm *HostsManager) ReadHosts() ([]byte, error) {
	return os.ReadFile(hm.hostsPath)
}

func (hm *HostsManager) WriteHosts(data []byte) error {
	return os.WriteFile(hm.hostsPath, data, 0644)
}

func (hm *HostsManager) Backup() error {
	data, err := hm.ReadHosts()
	if err != nil {
		return err
	}
	hm.original = data
	return nil
}

func (hm *HostsManager) IsSet() (bool, error) {
	data, err := hm.ReadHosts()
	if err != nil {
		return false, err
	}
	return bytes.Contains(data, []byte(HostsMarkerStart)), nil
}

func (hm *HostsManager) Set() error {
	isSet, err := hm.IsSet()
	if err != nil {
		return err
	}
	if isSet {
		return nil
	}

	if err := hm.Backup(); err != nil {
		return err
	}

	data, err := hm.ReadHosts()
	if err != nil {
		return err
	}

	entry := fmt.Sprintf("\n%s\n%s %s\n%s\n", HostsMarkerStart, TargetIP, TargetDomain, HostsMarkerEnd)

	newData := append(data, []byte(entry)...)

	return hm.WriteHosts(newData)
}

func (hm *HostsManager) Restore() error {
	data, err := hm.ReadHosts()
	if err != nil {
		return err
	}

	startIdx := bytes.Index(data, []byte(HostsMarkerStart))
	if startIdx == -1 {
		return nil
	}

	endIdx := bytes.Index(data, []byte(HostsMarkerEnd))
	if endIdx == -1 {
		return errors.New("malformed hosts file: found start marker but no end marker")
	}

	endIdx += len(HostsMarkerEnd)

	var newData []byte
	newData = append(newData, data[:startIdx]...)

	if endIdx < len(data) {
		if data[endIdx] == '\n' {
			endIdx++
		}
		newData = append(newData, data[endIdx:]...)
	}

	return hm.WriteHosts(newData)
}

func (hm *HostsManager) GetEntries() ([]string, error) {
	data, err := hm.ReadHosts()
	if err != nil {
		return nil, err
	}

	var entries []string
	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		entries = append(entries, line)
	}
	return entries, scanner.Err()
}
