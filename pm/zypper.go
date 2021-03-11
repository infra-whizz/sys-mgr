package sysmgr_pm

import (
	"fmt"

	sysmgr_sr "github.com/infra-whizz/sys-mgr/sr"
	wzlib_subprocess "github.com/infra-whizz/wzlib/subprocess"
)

// ZypperPackageManager object
type ZypperPackageManager struct {
	sysroot *sysmgr_sr.SysRoot
}

// NewZypperPackageManager creates a zypper caller object
func NewZypperPackageManager() *ZypperPackageManager {
	pm := new(ZypperPackageManager)
	return pm
}

// Call zypper
func (pm *ZypperPackageManager) Call(args ...string) (string, string, error) {
	if pm.sysroot == nil {
		return "", "", fmt.Errorf("No default sysroot has been found. Please specify one.")
	}

	args = append([]string{"--root", pm.sysroot.Path}, args...)

	stdout, stderr := wzlib_subprocess.StreamedExec(NewStdProcessStream(), pm.Name(), args...)
	return stdout, stderr, nil
}

// Name of the package manager
func (pm *ZypperPackageManager) Name() string {
	return "zypper"
}

// SetSysroot to work with
func (pm *ZypperPackageManager) SetSysroot(sysroot *sysmgr_sr.SysRoot) PackageManager {
	pm.sysroot = sysroot
	return pm
}
