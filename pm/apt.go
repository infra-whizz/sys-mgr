package sysmgr_pm

import (
	sysmgr_lib "github.com/infra-whizz/sys-mgr/lib"
	sysmgr_sr "github.com/infra-whizz/sys-mgr/sr"
)

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
	return sysmgr_lib.StdoutExec("chroot", append([]string{pm.sysroot.Path, "apt"}, args...)...)
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
	// Nothing here for apt to do
	return nil
}

func (pm *AptPackageManager) GetHelpFlags() map[string]string {
	return map[string]string{
		"list":         "List packages based on package names",
		"search":       "Search in package descriptions",
		"show":         "Show package details",
		"install":      "Install packages",
		"reinstall":    "Reinstall packages",
		"remove":       "Remove packages",
		"autoremove":   "Remove automatically all unused packages",
		"update":       "Update list of available packages",
		"upgrade":      "Upgrade the system by installing/upgrading packages",
		"full-upgrade": "Upgrade the system by removing/installing/upgrading packages",
		"edit-sources": "Edit the source information file",
		"satisfy":      "Satisfy dependency strings",
	}
}
