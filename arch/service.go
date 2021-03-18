package sysmgr_arch

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"

	sysmgr_pm "github.com/infra-whizz/sys-mgr/pm"
	wzlib_logger "github.com/infra-whizz/wzlib/logger"
)

type SystemdService struct {
	serviceName string
	servicePath string
	levelPath   string
	pkgman      sysmgr_pm.PackageManager

	wzlib_logger.WzLogger
}

func NewSystemdService() *SystemdService {
	s := new(SystemdService)
	s.serviceName = "sysroot-manager.service"
	s.servicePath = "/etc/systemd/system"
	s.levelPath = "multi-user.target.wants"
	return s
}

// SetPackageManager of the system
func (s *SystemdService) SetPackageManager(pkgman sysmgr_pm.PackageManager) *SystemdService {
	s.pkgman = pkgman
	return s
}

// Create service and enable default sysroot
func (s SystemdService) Create(arch string) error {
	var err error
	if err = s.Disable(); err != nil {
		s.GetLogger().Debugf("Unable to disable service: %s", err.Error())
	}

	if err = s.Remove(); err != nil {
		s.GetLogger().Debugf("Unable to remove the service file: %s", err.Error())
	}

	var buff strings.Builder

	for _, line := range []string{
		"[Unit]", fmt.Sprintf("Description=%s arch activation via %s", arch, s.pkgman.Name()), "",
		"[Service]", "Type=oneshot", fmt.Sprintf("ExecStart=/usr/bin/%s-sysroot sysroot --init", s.pkgman.Name()), "",
		"[Install]", "WantedBy=default.target",
	} {
		buff.WriteString(fmt.Sprintf("%s\n", line))
	}

	if err = ioutil.WriteFile(path.Join(s.servicePath, s.serviceName), []byte(buff.String()), 0644); err != nil {
		return err
	}
	s.GetLogger().Debugf("Wrote service file")
	return s.Enable()
}

// Enable service
func (s SystemdService) Enable() error {
	target := path.Join(s.servicePath, s.levelPath, s.serviceName)
	if _, err := os.Stat(target); os.IsNotExist(err) {
		if err := os.Symlink(path.Join(s.servicePath, s.serviceName), target); err != nil {
			return err
		}
		s.GetLogger().Debugf("Service has been activated")
	}
	return nil
}

// Disable service
func (s SystemdService) Disable() error {
	target := path.Join(s.servicePath, s.levelPath, s.serviceName)
	info, err := os.Stat(target)
	if info != nil {
		return os.Remove(target)
	}

	return err
}

// Remove service
func (s SystemdService) Remove() error {
	target := path.Join(s.servicePath, s.serviceName)
	info, _ := os.Stat(target)
	if info != nil {
		return os.Remove(target)
	}

	return nil
}
