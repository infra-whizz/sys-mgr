package sysmgr_sr

import (
	"fmt"
	"io/ioutil"
	"os/exec"
	"sort"
	"strings"

	sysmgr_lib "github.com/infra-whizz/sys-mgr/lib"
	wzlib_traits "github.com/infra-whizz/wzlib/traits"
	wzlib_traits_attributes "github.com/infra-whizz/wzlib/traits/attributes"
	wzlib_utils "github.com/infra-whizz/wzlib/utils"
)

type repodata struct {
	components []string
	url        string
	codename   string
}

type DebianSysrootProvisioner struct {
	BaseSysrootProvisioner
}

func NewDebianSysrootProvisioner(name, arch, root string) *DebianSysrootProvisioner {
	dsp := new(DebianSysrootProvisioner)
	dsp.qemuPattern = "qemu-%s-static"

	dsp.SetArch(arch)
	dsp.SetName(name)
	dsp.SetSysPath(root)

	dsp.qemuPath, _ = exec.LookPath(fmt.Sprintf(dsp.qemuPattern, dsp.arch))
	dsp.ref = dsp

	dsp.sysinfo = wzlib_traits.NewWzTraitsContainer()
	wzlib_traits_attributes.NewSysInfo().Load(dsp.sysinfo)

	return dsp
}

func (dsp *DebianSysrootProvisioner) getQemuPath() string {
	return dsp.qemuPath
}

func (dsp *DebianSysrootProvisioner) getSysPath() string {
	return dsp.sysPath
}

func (dsp *DebianSysrootProvisioner) beforePopulate() error {
	if dsp.getQemuPath() == "" {
		return fmt.Errorf("No static QEMU found: %s", fmt.Sprintf(dsp.qemuPattern, dsp.arch))
	}

	return nil
}

func (dsp *DebianSysrootProvisioner) getRepoData() (*repodata, error) {
	sourcesList := "/etc/apt/sources.list"
	if !wzlib_utils.FileExists(sourcesList) {
		return nil, fmt.Errorf("File %s is not accessible", sourcesList)
	}

	data, err := ioutil.ReadFile(sourcesList)
	if err != nil {
		return nil, err
	}

	r := &repodata{
		codename: dsp.sysinfo.Get("os.codename").(string),
	}

	components := map[string]interface{}{}

	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "deb ") {
			continue
		}

		tkn := strings.Fields(line)
		if len(tkn) < 3 || tkn[2] != r.codename {
			continue
		}

		if r.url == "" {
			if strings.Contains(line, "main") {
				r.url = tkn[1]
			}
		}

		for _, cmpt := range tkn[3:] {
			components[cmpt] = nil
		}
	}

	// Set to array
	for k := range components {
		r.components = append(r.components, k)
	}
	sort.Strings(r.components)

	return r, nil
}

// Populate sysroot according to the current package manager specifics
func (dsp *DebianSysrootProvisioner) onPopulate() error {
	repo, err := dsp.getRepoData()
	if err != nil {
		return err
	}

	if err = sysmgr_lib.LoggedExec("debootstrap", "--arch", "i386", "--no-check-gpg", "--variant=minbase",
		fmt.Sprintf("--components=%s", strings.Join(repo.components, ",")), repo.codename, dsp.sysrootPath, repo.url); err != nil {
		return err
	}

	return sysmgr_lib.LoggedExec("chroot", dsp.sysrootPath, "apt", "--fix-broken", "install")
}

func (dsp *DebianSysrootProvisioner) afterPopulate() error {
	/*
		for _, d := range []string{"/etc", "/proc", "/dev", "/sys", "/run", "/tmp"} {
			if err := os.MkdirAll(path.Join(dsp.sysrootPath, d), 0755); err != nil {
				return err
			}
		}
	*/

	// Create sysroot configuration
	if err := ioutil.WriteFile(dsp.confPath, []byte(fmt.Sprintf("name: %s\narch: %s\ndefault: false\n", dsp.name, dsp.arch)), 0644); err != nil {
		return err
	}
	return nil
}
