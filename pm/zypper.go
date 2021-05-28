package sysmgr_pm

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"

	sysmgr_sr "github.com/infra-whizz/sys-mgr/sr"
)

// ZypperPackageManager object
type ZypperPackageManager struct {
	sysroot *sysmgr_sr.SysRoot
	archFix map[string]string

	BasePackageManager
}

// NewZypperPackageManager creates a zypper caller object
func NewZypperPackageManager() *ZypperPackageManager {
	pm := new(ZypperPackageManager)
	pm.archFix = map[string]string{"arm": "armv7hl"}
	pm.env = make(map[string]string)
	return pm
}

// Call zypper
func (pm *ZypperPackageManager) Call(args ...string) error {
	if pm.sysroot == nil {
		return fmt.Errorf("No default sysroot has been found. Please specify one.")
	}

	args = append([]string{"--root", pm.sysroot.Path}, args...)
	return pm.callPackageManager(pm.Name(), args...)
}

// Name of the package manager
func (pm *ZypperPackageManager) Name() string {
	return "zypper"
}

// SetSysroot to work with
func (pm *ZypperPackageManager) SetSysroot(sysroot *sysmgr_sr.SysRoot) PackageManager {
	pm.sysroot = sysroot
	pm.env["ZYPP_CONF"] = path.Join(pm.sysroot.Path, "/etc/zypp/zypp.conf")
	pm.sysroot.GetLogger().Debug("Zypper environment: ", pm.env)

	return pm
}

// Setup package manager
func (pm *ZypperPackageManager) Setup() error {
	zyppConf := path.Join(pm.sysroot.Path, "/etc/zypp")
	if _, err := os.Stat(zyppConf); os.IsNotExist(err) {
		if err := os.MkdirAll(zyppConf, 0755); err != nil {
			return err
		}
	}
	zyppConf = path.Join(zyppConf, "zypp.conf")

	// Fix arch names
	arch, ex := pm.archFix[pm.sysroot.Arch]
	if !ex {
		arch = pm.sysroot.Arch
		pm.sysroot.GetLogger().Infof("Setting default architecture to Zypper: %s", arch)
	}

	var buff strings.Builder
	buff.WriteString("[main]\n")
	buff.WriteString(fmt.Sprintf("arch = %s\n", arch))
	buff.WriteString("multiversion = provides:multiversion(kernel)\n")
	buff.WriteString("multiversion.kernels = latest,latest-1,running\n")

	return ioutil.WriteFile(zyppConf, []byte(buff.String()), 0644)
}
