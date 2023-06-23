package sysmgr_sr

import (
	"fmt"
	"io/ioutil"
	"path"
	"syscall"

	"golang.org/x/sys/unix"
)

type ZypperSysrootProvisioner struct {
	BaseSysrootProvisioner
}

func NewZypperSysrootProvisioner(name, arch, root string) *ZypperSysrootProvisioner {
	zsp := new(ZypperSysrootProvisioner)

	zsp.SetArch(arch)
	zsp.SetName(name)
	zsp.SetSysPath(root)
	zsp.ref = zsp

	return zsp
}

func (zsp *ZypperSysrootProvisioner) beforePopulate() error {
	return nil
}

func (zsp *ZypperSysrootProvisioner) onPopulate() error {
	return nil
}

func (zsp *ZypperSysrootProvisioner) afterPopulate() error {
	return nil
}

func (dsp *ZypperSysrootProvisioner) getQemuPath() string {
	return ""
}

func (dsp *ZypperSysrootProvisioner) getSysPath() string {
	return dsp.sysPath
}

func (dsp *ZypperSysrootProvisioner) GetArch() string {
	return ""
}

func (dsp *ZypperSysrootProvisioner) Activate() error {
	dsp.GetLogger().Info("Activating system root")
	for _, src := range []string{"/proc", "/sys", "/dev", "/run"} {
		dst := path.Join(dsp.sysrootPath, src)
		dsp.GetLogger().Debugf("Mounting %s to %s", src, dst)
		if err := syscall.Mount(src, dst, "", syscall.MS_BIND, ""); err != nil {
			return err
		}
	}
	return nil
}

func (dsp *ZypperSysrootProvisioner) UnmountBinds() error {
	// pre-umount, if anything
	for _, d := range []string{"/proc", "/dev", "/sys", "/run"} {
		d = path.Join(dsp.sysrootPath, d)
		if err := syscall.Unmount(d, syscall.MNT_DETACH|syscall.MNT_FORCE|unix.UMOUNT_NOFOLLOW); err != nil {
			dsp.GetLogger().Warnf("Unable to unmount %s", d)
		}
		files, err := ioutil.ReadDir(d)
		if err != nil {
			return err
		}
		if len(files) > 0 {
			return fmt.Errorf("failed to unmount %s. Please umount it manually", d)
		}
	}

	return nil
}
