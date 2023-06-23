package sysmgr_sr

import (
	"fmt"
	"os"
	"path"
	"syscall"

	wzlib_logger "github.com/infra-whizz/wzlib/logger"
	wzlib_traits "github.com/infra-whizz/wzlib/traits"
	"github.com/isbm/go-shutil"
)

type BaseSysrootProvisioner struct {
	qemuPattern string
	qemuPath    string
	name        string
	arch        string
	sysrootPath string
	sysPath     string // Path of the root
	confPath    string

	sysinfo *wzlib_traits.WzTraitsContainer

	// Self reference to reuse subclass implementations.
	// This needs to be initialised in the child object.
	ref SysrootProvisioner
	wzlib_logger.WzLogger
}

func (bsp *BaseSysrootProvisioner) SetSysrootPath(p string) {
	bsp.sysrootPath = p
}

// Checks an existing system root
func (bsp *BaseSysrootProvisioner) CheckExisting(checkExists bool) error {
	if checkExists {
		if _, err := os.Stat(bsp.sysrootPath); !os.IsNotExist(err) {
			return fmt.Errorf("system root at %s already exists", bsp.sysrootPath)
		}
	}

	if bsp.name == "" {
		return fmt.Errorf("name was not set for new sysroot")
	} else if bsp.arch == "" {
		return fmt.Errorf("architecture was not set for the new sysroot")
	}
	return nil
}

// SetSysPath of all system roots together, default is "/usr/syroots"
func (bsp *BaseSysrootProvisioner) SetSysPath(p string) {
	// Due to nature of this class, certain configuration methods should be called separately,
	// and later after the constructor, but this particular one requires to be called last in order.
	if bsp.name == "" {
		panic("SetName() should be called first")
	}

	if bsp.arch == "" {
		panic("SetArch() should be called first")
	}

	bsp.sysPath = p
	bsp.sysrootPath = path.Join(bsp.sysPath, fmt.Sprintf("%s.%s", bsp.name, bsp.arch))
	bsp.confPath = path.Join(bsp.sysrootPath, ChildSysrootConfig)
}

// SetName of the system root
func (bsp *BaseSysrootProvisioner) SetName(name string) {
	bsp.name = name
}

// SetArch of the system root
func (bsp *BaseSysrootProvisioner) SetArch(arch string) {
	bsp.arch = arch
}

func (bsp *BaseSysrootProvisioner) beforePopulate() error {
	return bsp.CheckExisting(true)
}

func (bsp *BaseSysrootProvisioner) afterPopulate() error {
	return nil
}

func (dsp *BaseSysrootProvisioner) Populate() error {
	if dsp.ref == nil {
		panic("Sysroot populator is not properly initialised: no implementation reference found")
	}

	if err := dsp.beforePopulate(); err != nil {
		return err
	}

	if err := dsp.ref.beforePopulate(); err != nil {
		return err
	}

	if err := dsp.ref.onPopulate(); err != nil {
		return err
	}

	if err := dsp.afterPopulate(); err != nil {
		return err
	}

	if err := dsp.ref.afterPopulate(); err != nil {
		return err
	}

	return dsp.replicate()
}

// Replicate self, i.e. copy utils, other important stuff etc.
func (dsp *BaseSysrootProvisioner) replicate() error {
	selfPath, err := os.Executable()
	if err != nil {
		return err
	}

	for _, bin := range []string{selfPath, dsp.ref.getQemuPath()} {
		dsp.GetLogger().Debugf("Preparing %s", bin)

		// Setup target dir
		target := path.Join(dsp.sysrootPath, "usr", "bin")
		if _, err = os.Stat(target); os.IsNotExist(err) {
			dsp.GetLogger().Debugf("Creating directory %s", target)
			if err = os.MkdirAll(target, 0755); err != nil {
				return err
			}
		}

		// Copy required utility
		target = path.Join(target, path.Base(bin))
		dsp.GetLogger().Debugf("Copying %s to %s", bin, target)
		if err = shutil.CopyFile(bin, target, false); err != nil {
			return err
		}
		dsp.GetLogger().Debugf("Setting %s as 0755", target)
		if err = syscall.Chmod(target, 0755); err != nil {
			return err
		}
	}

	return nil
}

func (dsp *BaseSysrootProvisioner) GetConfigPath() string {
	return dsp.confPath
}
