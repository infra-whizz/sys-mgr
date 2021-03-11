package sysmgr

import (
	"fmt"
	"io/ioutil"
	"path"
	"strings"

	"github.com/isbm/go-nanoconf"
	"github.com/thoas/go-funk"
)

var defaultSysrootPath string = "/usr/sysroots"

type SysrootManager struct {
	sysroots      string
	architectures []string
}

func NewSysrootManager(conf *nanoconf.Config) *SysrootManager {
	srm := new(SysrootManager)
	srm.sysroots = conf.Root().String("sysroots", "")
	if srm.sysroots == "" {
		srm.sysroots = defaultSysrootPath
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
func (srm *SysrootManager) CreateSysRoot(name string, arch string) error {
	if err := srm.checkArch(arch); err != nil {
		return err
	}

	return NewSysRoot(srm.sysroots).SetName(name).SetArch(arch).Create()
}

// DeleteSysroot deletes the entire system root
func (srm *SysrootManager) DeleteSysRoot(name string, arch string) error {
	if err := srm.checkArch(arch); err != nil {
		return err
	}

	return NewSysRoot(srm.sysroots).SetName(name).SetArch(arch).Delete()
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
		return nil, err
	}
	roots := []*SysRoot{}
	for _, fn := range data {
		r, err := NewSysRoot(path.Join(srm.sysroots, fn.Name())).Init()
		if err != nil {
			return nil, err
		}
		roots = append(roots, r)
	}

	return roots, nil
}