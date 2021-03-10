package sysmgr

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strconv"

	"github.com/isbm/go-nanoconf"
)

type SysRoot struct {
	Name    string
	Arch    string
	Path    string
	Default bool

	confPath string
	sysPath  string
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
	if _, err := os.Stat(sr.sysPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("No system root found at %s", sr.sysPath)
	}

	sr.confPath = path.Join(sr.sysPath, "/etc/sysroot.conf")
	if _, err := os.Stat(sr.confPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("Invalid or unknown system root. Configuration missing at %s", sr.confPath)
	}
	conf := nanoconf.NewConfig(sr.confPath)
	sr.Name = conf.Root().String("name", "")
	sr.Arch = conf.Root().String("arch", "")

	isDefault := (*conf.Root().Raw())["default"]
	if isDefault != nil {
		sr.Default = isDefault.(bool)
	}

	if sr.Name == "" || sr.Arch == "" {
		return nil, fmt.Errorf("Invalid or unknown system root at %s", sr.Path)
	}

	sr.Path = path.Join(sr.sysPath, fmt.Sprintf("%s.%s", sr.Name, sr.Arch))

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

// Create a system root
func (sr *SysRoot) Create() error {
	sr.Path = path.Join(sr.sysPath, fmt.Sprintf("%s.%s", sr.Name, sr.Arch))
	sr.confPath = path.Join(sr.Path, "/etc/sysroot.conf")

	if err := sr.checkExistingSysroot(true); err != nil {
		return err
	}

	if err := os.MkdirAll(path.Join(sr.Path, "/etc"), 0755); err != nil {
		return err
	}

	return ioutil.WriteFile(sr.confPath, []byte(fmt.Sprintf("name: %s\narch: %s\ndefault: false\n", sr.Name, sr.Arch)), 0644)
}

// Delete a system root
func (sr *SysRoot) Delete() error {
	if err := sr.checkExistingSysroot(false); err != nil {
		return err
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
