package sysmgr_arch

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	sysmgr_pm "github.com/infra-whizz/sys-mgr/pm"
)

type SystemdService struct {
	fname  string
	pkgman sysmgr_pm.PackageManager
}

func NewSystemdService() *SystemdService {
	s := new(SystemdService)
	s.fname = "/etc/systemd/system/sysroot-manager.service"
	return s
}

// SetPackageManager of the system
func (s *SystemdService) SetPackageManager(pkgman sysmgr_pm.PackageManager) *SystemdService {
	s.pkgman = pkgman
	return s
}

// Create service and enable default sysroot
func (s SystemdService) Create(arch string) error {
	var buff strings.Builder

	for _, line := range []string{
		"[Unit]", fmt.Sprintf("Description=%s arch activation via %s", arch, s.pkgman.Name()), "",
		"[Service]", "Type=oneshot", fmt.Sprintf("ExecStart=/usr/bin/%s-sysroot sysroot --init", s.pkgman.Name()), "",
		"[Install]", "WantedBy=default.target",
	} {
		buff.WriteString(fmt.Sprintf("%s\n", line))
	}

	return ioutil.WriteFile(s.fname, []byte(buff.String()), 0644)
}

// Remove service
func (s SystemdService) Remove() error {
	info, _ := os.Stat(s.fname)
	if info != nil {
		return os.Remove(s.fname)
	}

	return nil
}
