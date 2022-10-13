package sysmgr_sr

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strconv"

	"github.com/elastic/go-sysinfo"
	wzlib_logger "github.com/infra-whizz/wzlib/logger"
	"github.com/isbm/go-nanoconf"
)

type SysRoot struct {
	Name    string
	Arch    string
	Path    string
	Default bool

	confPath string
	sysPath  string
	qemuPath string

	_provisioner SysrootProvisioner

	wzlib_logger.WzLogger
}

func NewSysRoot(syspath string) *SysRoot {
	sr := new(SysRoot)
	sr.sysPath = syspath

	return sr
}

// SetName alias
func (sr *SysRoot) SetName(name string) *SysRoot {
	if sr.Name == "" {
		sr.Name = name
	}
	return sr
}

// SetArch alias
func (sr *SysRoot) SetArch(arch string) *SysRoot {
	if sr.Arch == "" {
		sr.Arch = arch
	}
	return sr
}

func (sr *SysRoot) GetProvisioner() (SysrootProvisioner, error) {
	if sr._provisioner == nil {
		// Initialise provisioner
		p := sr.GetCurrentPlatform()
		switch p {
		case "ubuntu", "debian":
			sr._provisioner = NewDebianSysrootProvisioner(sr.Name, sr.Arch, sr.sysPath)
		case "opensuse-leap":
			sr._provisioner = NewZypperSysrootProvisioner(sr.Name, sr.Arch, sr.sysPath)
		default:
			return nil, fmt.Errorf("Unable to initialise provisioner for unsupported platform: %s", p)
		}
	}
	return sr._provisioner, nil
}

// Init system root.
func (sr *SysRoot) Init() (*SysRoot, error) {
	if sr.sysPath != "/" {
		if sr.Name == "" {
			return nil, fmt.Errorf("Name of the sysroot was not specified while looking it up")
		}

		if sr.Arch == "" {
			return nil, fmt.Errorf("Architecture of the sysroot was not specified while looking it up")
		}
	}

	// Already initialised
	if sr.Path != "" {
		return sr, nil
	}

	// Read sysroot from the host root or chrooted
	if sr.sysPath != "/" {
		sr.Path = path.Join(sr.sysPath, fmt.Sprintf("%s.%s", sr.Name, sr.Arch))
		if _, err := os.Stat(sr.sysPath); os.IsNotExist(err) {
			return nil, fmt.Errorf("No system root found at %s", sr.sysPath)
		}
	} else { // chrooted
		sr.Path = sr.sysPath
	}

	sr.confPath = path.Clean(path.Join(sr.Path, ChildSysrootConfig))
	if _, err := os.Stat(sr.confPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("Invalid or unknown child system root. Configuration missing at %s", sr.confPath)
	}
	conf := nanoconf.NewConfig(sr.confPath)

	sr.SetName(conf.Root().String("name", ""))
	sr.SetArch(conf.Root().String("arch", ""))

	isDefault := conf.Root().Raw()["default"]
	if isDefault != nil {
		sr.Default = isDefault.(bool)
	}

	if sr.Name == "" || sr.Arch == "" {
		return nil, fmt.Errorf("Invalid configuration of a system root at %s", sr.Path)
	}

	return sr, nil
}

// Checks an existing system root
func (sr *SysRoot) checkExistingSysroot(checkExists bool) error {
	if checkExists {
		if _, err := os.Stat(sr.Path); !os.IsNotExist(err) {
			return fmt.Errorf("System root at %s already exists", sr.Path)
		}
	}

	if sr.Name == "" {
		return fmt.Errorf("Name was not set for new sysroot")
	} else if sr.Arch == "" {
		return fmt.Errorf("Architecture was not set for the new sysroot")
	}
	return nil
}

// GetCurrentPlatform returns a current platform class
// XXX: Needs to be moved to the utils, but that requires a major refactoring.
func (sr *SysRoot) GetCurrentPlatform() string {
	info, err := sysinfo.Host()
	if err != nil {
		panic(err)
	}

	return info.Info().OS.Platform
}

// Create a system root
func (sr *SysRoot) Create() error {
	provisioner, err := sr.GetProvisioner()
	if err != nil {
		return err
	}

	return provisioner.Populate()
}

// UmountBinds removes proc, dev, sys and run
func (sr *SysRoot) UmountBinds() error {
	if _, err := sr.Init(); err != nil {
		return err
	}

	provisioner, err := sr.GetProvisioner()
	if err != nil {
		return err
	}

	return provisioner.UnmountBinds()
}

// Delete a system root
func (sr *SysRoot) Delete() error {
	if err := sr.checkExistingSysroot(false); err != nil {
		return err
	}

	if _, err := sr.Init(); err != nil {
		return err
	}

	// check if the sysroot still bound to something
	for _, d := range []string{"/proc", "/dev", "/sys", "/run"} {
		d = path.Join(sr.Path, d)
		files, err := ioutil.ReadDir(d)
		if err != nil {
			return err
		}
		if len(files) > 0 {
			return fmt.Errorf("Directory %s seems not properly unmounted. Please check it, unmount manually and try again.", d)
		}
	}

	return os.RemoveAll(sr.Path)
}

// SEtDefault system root
func (sr *SysRoot) SetDefault(isDefault bool) error {
	if err := sr.checkExistingSysroot(false); err != nil {
		return err
	}

	provisioner, err := sr.GetProvisioner()
	if err != nil {
		return err
	}

	return ioutil.WriteFile(provisioner.GetConfigPath(), []byte(fmt.Sprintf("name: %s\narch: %s\ndefault: %s\n",
		sr.Name, sr.Arch, strconv.FormatBool(isDefault))), 0644)
}

// Activate default sysroot (mount runtime directories)
func (sr *SysRoot) Activate() error {
	provisioner, err := sr.GetProvisioner()
	if err != nil {
		return err
	}

	return provisioner.Activate()
}
