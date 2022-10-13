package sysmgr_sr

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
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
	rd *repodata
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

		if r.url == "" && strings.Contains(line, "main") {
			r.url = tkn[1]
		}

		for _, cmpt := range tkn[3:] {
			components[cmpt] = nil
		}
	}

	// Turn sets to arrays
	for k := range components {
		r.components = append(r.components, k)
	}
	sort.Strings(r.components)

	return r, nil
}

func (dsp *DebianSysrootProvisioner) GetArch() string {
	archfix := map[string]string{
		"x86_64": "amd64",
		"i586":   "i386",
	}
	arch, ex := archfix[dsp.arch]
	if !ex {
		arch = dsp.arch
	}

	return arch
}

// Populate sysroot according to the current package manager specifics
func (dsp *DebianSysrootProvisioner) onPopulate() error {
	var err error
	dsp.rd, err = dsp.getRepoData()
	if err != nil {
		return err
	}

	dsp.GetLogger().Debugf("Populating sysroot into %s", dsp.sysrootPath)

	if err = sysmgr_lib.LoggedExec("debootstrap", "--arch", dsp.GetArch(), "--no-check-gpg", "--variant=minbase",
		fmt.Sprintf("--components=%s", strings.Join(dsp.rd.components, ",")), dsp.rd.codename, dsp.sysrootPath, dsp.rd.url); err != nil {
		return err
	}

	return sysmgr_lib.LoggedExec("chroot", dsp.sysrootPath, "apt", "--fix-broken", "install") // Normally not needed, but mostly who knows? :)
}

func (dsp *DebianSysrootProvisioner) afterPopulate() error {
	for _, d := range []string{"/etc", "/proc", "/dev", "/sys", "/run", "/tmp"} {
		tp := path.Join(dsp.sysrootPath, d)
		if wzlib_utils.FileExists(tp) {
			continue
		}

		if err := os.MkdirAll(tp, 0755); err != nil {
			return err
		}
	}

	// Create sysroot configuration
	if err := ioutil.WriteFile(dsp.confPath, []byte(fmt.Sprintf("name: %s\narch: %s\ndefault: false\n", dsp.name, dsp.arch)), 0644); err != nil {
		return err
	}

	// Add to sources.list if anything
	f, err := os.OpenFile(path.Join(dsp.sysrootPath, "etc", "apt", "sources.list"), os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	for _, section := range []string{"updates", "backports", "security"} {
		if _, err := f.WriteString(fmt.Sprintf("deb %s %s-%s %s\n", dsp.rd.url, dsp.rd.codename, section, strings.Join(dsp.rd.components, " "))); err != nil {
			return err
		}
	}
	f.Close()

	// Upgrade everything
	if err := sysmgr_lib.LoggedExec("chroot", dsp.sysrootPath, "apt-get", "update"); err != nil {
		return err
	}

	if err := sysmgr_lib.LoggedExec("chroot", dsp.sysrootPath, "apt-get", "upgrade", "--yes"); err != nil {
		return err
	}

	return nil
}
