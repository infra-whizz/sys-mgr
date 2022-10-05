package sysmgr_pm

import sysmgr_sr "github.com/infra-whizz/sys-mgr/sr"

type AptPackageManager struct {
	sysroot *sysmgr_sr.SysRoot
	archFix map[string]string

	BasePackageManager
}

// Apt/dpkg package manager
func NewAptPackageManager() *AptPackageManager {
	pm := new(AptPackageManager)
	pm.archFix = map[string]string{}
	pm.env = map[string]string{}

	return pm
}

// Call apt/dpkg
func (pm *AptPackageManager) Call(args ...string) error {
	return nil
}

// Name of the package manager
func (pm *AptPackageManager) Name() string {
	return "apt"
}

// SetSysroot to work with
func (pm *AptPackageManager) SetSysroot(sysroot *sysmgr_sr.SysRoot) PackageManager {
	pm.sysroot = sysroot
	pm.sysroot.GetLogger().Debug("Apt environment:", pm.env)
	return pm
}

// Setup package manager
// This is used to pre-setup a package manager for a multiarch
func (pm *AptPackageManager) Setup() error {
	return nil
}
