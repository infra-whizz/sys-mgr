package sysmgr_sr

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"

	wzlib_logger "github.com/infra-whizz/wzlib/logger"
	"github.com/isbm/go-nanoconf"
	"github.com/thoas/go-funk"
)

var DefaultSysrootPath string = "/usr/sysroots"
var HostSysrootConfig string = "/etc/sysroots.conf"
var ChildSysrootConfig string = "/etc/sysroot.conf"

type SysrootManager struct {
	sysroots      string
	architectures []string
	wzlib_logger.WzLogger
}

func NewSysrootManager(conf *nanoconf.Config) *SysrootManager {
	srm := new(SysrootManager)
	srm.sysroots = conf.Root().String("sysroots", "")
	if srm.sysroots == "" {
		srm.sysroots = DefaultSysrootPath
	}
	srm.architectures = []string{}
	return srm
}

// SetSupported Architectures
func (srm *SysrootManager) SetSupportedArchitectures(architectures []string) *SysrootManager {
	srm.architectures = architectures
	return srm
}

func (srm SysrootManager) checkArch(arch string) error {
	arch = strings.ToLower(arch)
	if !funk.Contains(srm.architectures, arch) {
		return fmt.Errorf("Unsupported architecture: %s", arch)
	}
	return nil
}

// CreateSysRoot creates a system root placeholder
func (srm *SysrootManager) CreateSysRoot(name string, arch string) (*SysRoot, error) {
	if err := srm.checkArch(arch); err != nil {
		return nil, err
	}

	srm.GetLogger().Debugf("Placing sysroot into %s", srm.sysroots)

	sysroot := NewSysRoot(srm.sysroots).SetName(name).SetArch(arch)
	if err := sysroot.Create(); err != nil {
		return nil, err
	}
	return sysroot, nil
}

// CheckWithinSysroot returns an error, if current working directory is within sysroot.
// Useful to prevent destructive operations, such as sysroot removal, while still working in it.
func (srm SysrootManager) CheckWithinSysroot(sysroot *SysRoot) error {
	here, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("Unable to obtain current working directory: %s", err.Error())
	}

	if strings.HasPrefix(here, sysroot.Path) {
		return fmt.Errorf("This operation is not permitted while still inside the '%s' directory", sysroot.Path)
	}

	return nil
}

// DeleteSysroot deletes the entire system root
func (srm *SysrootManager) DeleteSysRoot(name string, arch string) error {
	if err := srm.checkArch(arch); err != nil {
		return err
	}

	sysroot, err := NewSysRoot(srm.sysroots).SetName(name).SetArch(arch).Init()
	if err != nil {
		return err
	}

	if err := srm.CheckWithinSysroot(sysroot); err != nil {
		return err
	}

	if err := sysroot.UmountBinds(); err != nil {
		return err
	}

	return sysroot.Delete()
}

// SetDefaultSysRoot to be locked on the particular package manager.
// This option only sets configured default sysroot, but it still can be overridden.
func (srm *SysrootManager) SetDefaultSysRoot(name string, arch string) error {
	if err := srm.checkArch(arch); err != nil {
		return err
	}

	roots, err := srm.GetSysRoots()
	if err != nil {
		return err
	}

	var dsr *SysRoot
	for _, sr := range roots {
		if sr.Name == name && sr.Arch == arch {
			dsr = sr
		}
	}
	if dsr != nil {
		for _, sr := range roots {
			if err := sr.SetDefault(sr.Name == dsr.Name && sr.Arch == dsr.Arch); err != nil {
				return err
			}
		}
	} else {
		return fmt.Errorf("Sysroot you want to make default was not found")
	}

	return nil
}

// GetSysRoots returns all available sysroots
func (srm *SysrootManager) GetSysRoots() ([]*SysRoot, error) {
	data, err := ioutil.ReadDir(srm.sysroots)
	if err != nil {
		return nil, fmt.Errorf("Unable to read directory '%s': %s", srm.sysroots, err.Error())
	}

	roots := []*SysRoot{}
	for _, fn := range data {
		na := strings.Split(fn.Name(), ".")
		if len(na) != 2 {
			return nil, fmt.Errorf("Unknown sysroot found at %s", path.Join(srm.sysroots, fn.Name()))
		}

		r, err := NewSysRoot(srm.sysroots).SetName(na[0]).SetArch(na[1]).Init()
		if err != nil {
			return nil, err
		}
		roots = append(roots, r)
	}

	return roots, nil
}

// GetDefaultSysroot. If chrooted, returns current
func (srm *SysrootManager) GetDefaultSysroot() (*SysRoot, error) {
	isChrooted, err := srm.IsChrooted()

	if err != nil {
		return nil, fmt.Errorf("Unable to determine the chroot environment: %s", err.Error())
	}

	if !isChrooted {
		sysroots, err := srm.GetSysRoots()
		if err != nil {
			return nil, err
		}

		for _, sr := range sysroots {
			if sr.Default {
				return sr, nil
			}
		}
	} else {
		return NewSysRoot("/").Init()
	}

	return nil, fmt.Errorf("No default system root has been found. Please setup one.")
}

// fileExists or not. This needs to be moved to utils, but importing them causes cycle.
// Utils needs to be moved into a separate sub-package.
func (srm *SysrootManager) fileExists(filepath string) (bool, error) {
	_, err := os.Stat(filepath)
	if err == nil {
		return true, nil
	} else if os.IsNotExist(err) {
		return false, nil
	}

	// Schrodinger cat, look for the error
	return false, err
}

// IsChrooted returns false, if the current root belongs to host, true if this is a child root
func (srm *SysrootManager) IsChrooted() (bool, error) {
	hasHostSysroot, err := srm.fileExists(HostSysrootConfig)
	if err != nil {
		return false, fmt.Errorf("Unable to determine weather host sysroot config exists or not: %s", err.Error())
	}
	hasChildSysroot, err := srm.fileExists(ChildSysrootConfig)
	if err != nil {
		return false, fmt.Errorf("Unable to determine weather child sysroot config exists or not: %s", err.Error())
	}

	return hasChildSysroot && !hasHostSysroot, nil
}
