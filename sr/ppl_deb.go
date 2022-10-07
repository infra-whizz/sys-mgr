package sysmgr_sr

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strings"
)

type DebianSysrootProvisioner struct {
	BaseSysrootProvisioner
}

func NewDebianSysrootProvisioner(name, arch, root string) *DebianSysrootProvisioner {
	dsp := new(DebianSysrootProvisioner)
	dsp.qemuPattern = "qemu-%s-static"

	dsp.SetArch(arch)
	dsp.SetName(name)
	dsp.SetSysPath(root)

	return dsp
}

func (dsp *DebianSysrootProvisioner) beforePopulate() error {
	var err error
	if dsp.qemuPath, err = exec.LookPath(fmt.Sprintf(dsp.qemuPattern, dsp.arch)); err != nil {
		return err
	}

	dsp.sysrootPath = path.Join(dsp.sysPath, fmt.Sprintf("%s.%s", dsp.name, dsp.arch))
	dsp.confPath = path.Join(dsp.sysrootPath, ChildSysrootConfig)

	return dsp.CheckExisting(true)
}

func (dsp *DebianSysrootProvisioner) onPopulate() error {
	return nil
}

func (dsp *DebianSysrootProvisioner) afterPopulate() error {
	for _, d := range []string{"/etc", "/proc", "/dev", "/sys", "/run", "/tmp"} {
		if err := os.MkdirAll(path.Join(dsp.sysrootPath, d), 0755); err != nil {
			return err
		}
	}

	// Create sysroot configuration
	if err := ioutil.WriteFile(dsp.confPath, []byte(fmt.Sprintf("name: %s\narch: %s\ndefault: false\n", dsp.name, dsp.arch)), 0644); err != nil {
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
			tgtpath := path.Join(dsp.sysrootPath, "etc", f.Name())
			dsp.GetLogger().Debugf("Copying %s to %s", srcpath, tgtpath)

			data, err := ioutil.ReadFile(srcpath)
			if err != nil {
				return err
			}
			if err := ioutil.WriteFile(tgtpath, data, 0644); err != nil {
				return err
			}
		}
	}

	return nil
}
