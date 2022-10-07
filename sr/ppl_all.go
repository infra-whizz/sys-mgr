package sysmgr_sr

import (
	"fmt"
	"os"
	"path"
	"syscall"

	wzlib_logger "github.com/infra-whizz/wzlib/logger"
)

type BaseSysrootProvisioner struct {
	qemuPattern string
	qemuPath    string
	name        string
	arch        string
	sysrootPath string
	sysPath     string // Path of the root
	confPath    string

	// Self reference to reuse subclass implementations.
	// This needs to be initialised in the child object.
	ref SysrootProvisioner
	wzlib_logger.WzLogger
}

func (bsp *BaseSysrootProvisioner) SetSysrootPath(p string) {
	bsp.sysrootPath = p
}

func (bsp *BaseSysrootProvisioner) Activate() error {
	for _, src := range []string{"/proc", "/sys", "/dev", "/run"} {
		if err := syscall.Mount(src, path.Join(bsp.sysrootPath, src), "", syscall.MS_BIND, ""); err != nil {
			return err
		}
	}
	return nil
}

// Checks an existing system root
func (bsp *BaseSysrootProvisioner) CheckExisting(checkExists bool) error {
	if checkExists {
		if _, err := os.Stat(bsp.sysrootPath); !os.IsNotExist(err) {
			return fmt.Errorf("System root at %s already exists", bsp.sysrootPath)
		}
	}

	if bsp.name == "" {
		return fmt.Errorf("Name was not set for new sysroot")
	} else if bsp.arch == "" {
		return fmt.Errorf("Architecture was not set for the new sysroot")
	}
	return nil
}

// SetSysPath of the system root
func (bsp *BaseSysrootProvisioner) SetSysPath(p string) {
	bsp.sysPath = p
}

// SetName of the system root
func (bsp *BaseSysrootProvisioner) SetName(name string) {
	bsp.name = name
}

// SetArch of the system root
func (bsp *BaseSysrootProvisioner) SetArch(arch string) {
	bsp.arch = arch
}

func (dsp *BaseSysrootProvisioner) Populate() error {
	if dsp.ref != nil {
		panic("Sysroot populator is not properly initialised: no implementation reference found")
	}

	if err := dsp.ref.beforePopulate(); err != nil {
		return err
	}

	if err := dsp.ref.onPopulate(); err != nil {
		return err
	}

	return dsp.ref.afterPopulate()
}
