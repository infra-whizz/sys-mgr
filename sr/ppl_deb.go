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
	wzlib_logger "github.com/infra-whizz/wzlib/logger"
	wzlib_traits "github.com/infra-whizz/wzlib/traits"
	wzlib_traits_attributes "github.com/infra-whizz/wzlib/traits/attributes"
	wzlib_utils "github.com/infra-whizz/wzlib/utils"
	"github.com/shirou/gopsutil/host"
)

type repodata struct {
	options    map[string]string
	components []string
	url        string
	codename   string
}

type DebianSysrootProvisioner struct {
	BaseSysrootProvisioner
	rd *repodata
	wzlib_logger.WzLogger
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

func (bsp *DebianSysrootProvisioner) Activate() error {
	// Nothing to do here
	return nil
}

func (dsp *DebianSysrootProvisioner) UnmountBinds() error {
	// Nothing to do here
	return nil
}

func (dsp *DebianSysrootProvisioner) getQemuPath() string {
	return dsp.qemuPath
}

func (dsp *DebianSysrootProvisioner) getSysPath() string {
	return dsp.sysPath
}

func (dsp *DebianSysrootProvisioner) beforePopulate() error {
	if dsp.getQemuPath() == "" {
		return fmt.Errorf("no static QEMU found: %s", fmt.Sprintf(dsp.qemuPattern, dsp.arch))
	}

	return nil
}

/*
Get a proper URL for the repo.
This is done the following way:

 1. URLs are searched only in /etc/apt/sources.list (subdir "sources.list.d" is ignored)

 2. Only genuine distribution URL is extracted (i.e. which contains "main" component)

 3. If the current arch is not the same as asked arch, then the system has to be
    setup for multiarch: https://wiki.debian.org/Multiarch/HOWTO
    Specifically the following tag is required to match the architecture:

    deb [arch=....] .....

    Example:

    deb [arch=arm64] http://ports.ubuntu.com/ubuntu-ports jammy main universe multiverse restricted
*/
func (dsp *DebianSysrootProvisioner) getRepoData() (*repodata, error) {
	currArch, err := host.KernelArch()
	if err != nil {
		return nil, err
	}
	currArch = dsp.GetArch(currArch)

	sourcesList := "/etc/apt/sources.list"
	if !wzlib_utils.FileExists(sourcesList) {
		return nil, fmt.Errorf("file %s is not accessible", sourcesList)
	}

	data, err := ioutil.ReadFile(sourcesList)
	if err != nil {
		return nil, err
	}

	components := map[string]interface{}{}
	var r *repodata

	for _, line := range strings.Split(string(data), "\n") {
		r = &repodata{
			codename: dsp.sysinfo.Get("os.codename").(string),
			options:  map[string]string{},
		}
		line = strings.TrimSpace(line)
		// Not a binary repo
		if !strings.HasPrefix(line, "deb ") {
			continue
		}

		// Incomplete format
		tkn := strings.Fields(line)
		if len(tkn) < 3 {
			continue
		}

		// Options?
		offset := 0
		if tkn[1][0] == '[' && tkn[1][len(tkn[1])-1] == ']' {
			for _, kvd := range strings.Fields(tkn[1][1 : len(tkn[1])-1]) {
				kv := strings.Split(kvd, "=")
				if len(kv) == 2 {
					r.options[kv[0]] = kv[1]
				}
			}
			offset++
		}

		optArch := r.options["arch"]
		if currArch == dsp.GetArch("") {
			optArch = ""
		}

		if tkn[2+offset] != r.codename && tkn[2] == "sid" {
			r.codename = "sid"
		}
		if tkn[2+offset] != r.codename {
			continue
		}

		if (r.url == "" && strings.Contains(line, "main")) &&
			((currArch == dsp.GetArch("")) || (currArch != dsp.GetArch("") && optArch == dsp.GetArch(""))) {
			r.url = tkn[1+offset]
		}

		for _, cmpt := range tkn[3+offset:] {
			components[cmpt] = nil
		}

		if r.url != "" {
			break
		}
	}

	if r == nil || r.url == "" {
		return nil, fmt.Errorf("no repo URL found that matches target architecture (%s)", dsp.GetArch(""))
	}

	// Turn sets to arrays
	for k := range components {
		r.components = append(r.components, k)
	}
	sort.Strings(r.components)

	return r, nil
}

func (dsp *DebianSysrootProvisioner) GetArch(arch string) string {
	archfix := map[string]string{
		"x86_64":  "amd64",
		"i586":    "i386",
		"aarch64": "arm64", // Only for Debian repos. Otherwise aarch64. Yes, it is a mess!
	}

	var ex bool
	if arch == "" {
		arch = dsp.arch
	}

	arch, ex = archfix[arch]
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

	if err = sysmgr_lib.LoggedExec("debootstrap", "--arch", dsp.GetArch(""), "--no-check-gpg", "--variant=minbase",
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

	// Add to sources.list if Ubuntu, but don't if Debian
	if dsp.sysinfo.Get("os.platform") == "ubuntu" {
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
	}

	// Upgrade everything
	if err := sysmgr_lib.LoggedExec("chroot", dsp.sysrootPath, "apt-get", "update"); err != nil {
		return err
	}

	if err := sysmgr_lib.LoggedExec("chroot", dsp.sysrootPath, "apt-get", "upgrade", "--yes"); err != nil {
		return err
	}

	return nil
}
