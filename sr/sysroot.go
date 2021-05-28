package sysmgr_sr

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"syscall"

	wzlib_logger "github.com/infra-whizz/wzlib/logger"
	"github.com/isbm/go-nanoconf"
	"github.com/isbm/go-shutil"
	"golang.org/x/sys/unix"
)

type SysRoot struct {
	Name    string
	Arch    string
	Path    string
	Default bool

	confPath string
	sysPath  string
	qemuPath string

	wzlib_logger.WzLogger
}

func NewSysRoot(syspath string) *SysRoot {
	sr := new(SysRoot)
	sr.sysPath = syspath

	return sr
}

// SetName alias
func (sr *SysRoot) SetName(name string) *SysRoot {
	sr.Name = name
	return sr
}

// SetArch alias
func (sr *SysRoot) SetArch(arch string) *SysRoot {
	sr.Arch = arch
	return sr
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
	sr.Name = conf.Root().String("name", "")
	sr.Arch = conf.Root().String("arch", "")

	isDefault := (*conf.Root().Raw())["default"]
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

// replicate self
func (sr *SysRoot) replicate() error {
	selfPath, err := os.Executable()
	if err != nil {
		return err
	}

	for _, bin := range []string{selfPath, sr.qemuPath} {
		sr.GetLogger().Debugf("Preparing %s", bin)

		// Setup target dir
		target := path.Join(sr.Path, path.Dir(bin))
		if _, err = os.Stat(target); os.IsNotExist(err) {
			sr.GetLogger().Debugf("Creating directory %s", target)
			if err = os.MkdirAll(target, 0755); err != nil {
				return err
			}
		}

		// Copy required utility
		target = path.Join(target, path.Base(bin))
		sr.GetLogger().Debugf("Copying %s to %s", bin, target)
		if err = shutil.CopyFile(bin, target, false); err != nil {
			return err
		}
		sr.GetLogger().Debugf("Setting %s as 0755", target)
		if err = syscall.Chmod(target, 0755); err != nil {
			return err
		}
	}

	return nil
}

// Create a system root
func (sr *SysRoot) Create() error {
	var err error
	if sr.qemuPath, err = exec.LookPath(fmt.Sprintf("qemu-%s", sr.Arch)); err != nil {
		return err
	}

	sr.Path = path.Join(sr.sysPath, fmt.Sprintf("%s.%s", sr.Name, sr.Arch))
	sr.confPath = path.Join(sr.Path, ChildSysrootConfig)

	if err = sr.checkExistingSysroot(true); err != nil {
		return err
	}

	for _, d := range []string{"/etc", "/proc", "/dev", "/sys", "/run", "/tmp"} {
		if err = os.MkdirAll(path.Join(sr.Path, d), 0755); err != nil {
			return err
		}
	}

	// Create sysroot configuration
	if err = ioutil.WriteFile(sr.confPath, []byte(fmt.Sprintf("name: %s\narch: %s\ndefault: false\n", sr.Name, sr.Arch)), 0644); err != nil {
		return err
	}

	// Add basic skeleton for package manager should work
	files, err := ioutil.ReadDir("/etc")
	if err != nil {
		return err
	}

	// Copy *-relese files
	for _, f := range files {
		if strings.HasSuffix(f.Name(), "-release") {
			srcpath := path.Join("/etc", f.Name())
			tgtpath := path.Join(sr.Path, "etc", f.Name())
			sr.GetLogger().Debugf("Copying %s to %s", srcpath, tgtpath)

			data, err := ioutil.ReadFile(srcpath)
			if err != nil {
				return err
			}
			if err := ioutil.WriteFile(tgtpath, data, 0644); err != nil {
				return err
			}
		}
	}

	return sr.replicate()
}

// UmountBinds removes proc, dev, sys and run
func (sr *SysRoot) UmountBinds() error {
	if _, err := sr.Init(); err != nil {
		return err
	}

	// pre-umount, if anything
	for _, d := range []string{"/proc", "/dev", "/sys", "/run"} {
		d = path.Join(sr.Path, d)
		if err := syscall.Unmount(d, syscall.MNT_DETACH|syscall.MNT_FORCE|unix.UMOUNT_NOFOLLOW); err != nil {
			sr.GetLogger().Warnf("Unable to unmount %s", d)
		}
		files, err := ioutil.ReadDir(d)
		if err != nil {
			return err
		}
		if len(files) > 0 {
			return fmt.Errorf("Failed to unmount %s. Please umount it manually.", d)
		}
	}

	return nil
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

	return ioutil.WriteFile(sr.confPath, []byte(fmt.Sprintf("name: %s\narch: %s\ndefault: %s\n",
		sr.Name, sr.Arch, strconv.FormatBool(isDefault))), 0644)
}

// Activate default sysroot (mount runtime directories)
func (sr *SysRoot) Activate() error {
	for _, src := range []string{"/proc", "/sys", "/dev", "/run"} {
		if err := syscall.Mount(src, path.Join(sr.Path, src), "", syscall.MS_BIND, ""); err != nil {
			return err
		}
	}
	return nil
}
