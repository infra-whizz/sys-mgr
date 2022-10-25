package sysmgr_pm

import (
	"fmt"
	"path"

	sysmgr_lib "github.com/infra-whizz/sys-mgr/lib"
	sysmgr_sr "github.com/infra-whizz/sys-mgr/sr"
	"github.com/urfave/cli/v2"
)

type AptPackageManager struct {
	sysroot *sysmgr_sr.SysRoot
	archFix map[string]string

	dpkgConverse map[string]string
	dpkgCommands []string
	chrooted     []string

	BasePackageManager
}

// Apt/dpkg package manager
func NewAptPackageManager() *AptPackageManager {
	pm := new(AptPackageManager)
	pm.archFix = map[string]string{}
	pm.env = map[string]string{}

	pm.dpkgCommands = []string{"list-installed", "installed", "files", "content"}
	pm.dpkgConverse = map[string]string{"list-installed": "-l", "installed": "-l", "files": "-L", "content": "-L"}
	pm.chrooted = []string{"install", "reinstall", "remove", "autoremove", "update", "upgrade", "full-upgrade", "satisfy", "purge"}

	return pm
}

// Call apt/dpkg
func (pm *AptPackageManager) Call(args ...string) error {
	if sysmgr_lib.Any([]string{"chroot", "c"}, args[0]) {
		cmd := []string{"chroot", pm.sysroot.Path}
		if err := sysmgr_lib.CheckUser(0, 0); err != nil {
			cmd = append([]string{"sudo"}, cmd...)
		}
		return sysmgr_lib.StdoutExec(cmd[0], append(cmd[1:], args[1:]...)...)
	} else if sysmgr_lib.Any(pm.chrooted, args[0]) {
		cmd := append([]string{"chroot", pm.sysroot.Path, "apt"}, args...)
		if err := sysmgr_lib.CheckUser(0, 0); err != nil {
			cmd = append([]string{"sudo"}, cmd...)
		}
		return sysmgr_lib.StdoutExec(cmd[0], cmd[1:]...)
	} else if sysmgr_lib.Any(pm.dpkgCommands, args[0]) {
		return sysmgr_lib.StdoutExec(path.Join(pm.sysroot.Path, "usr", "bin", "dpkg"),
			append([]string{"--root", pm.sysroot.Path, pm.dpkgConverse[args[0]]}, args[1:]...)...)
	} else {
		return sysmgr_lib.StdoutExec(path.Join(pm.sysroot.Path, "usr", "bin", "apt"),
			append([]string{"-o", fmt.Sprintf("RootDir=%s", pm.sysroot.Path)}, args...)...)
	}
}

// Name of the package manager
func (pm *AptPackageManager) Name() string {
	return "apt"
}

// SetSysroot to work with
func (pm *AptPackageManager) SetSysroot(sysroot *sysmgr_sr.SysRoot) PackageManager {
	pm.sysroot = sysroot
	pm.sysroot.GetLogger().Debug("Apt environment:", pm.env)
	return pm
}

// Setup package manager
// This is used to pre-setup a package manager for a multiarch
func (pm *AptPackageManager) Setup() error {
	// Nothing here for apt to do
	return nil
}

func (pm *AptPackageManager) GetHelpFlags() *PmCommands {
	return &PmCommands{
		&PmCommand{
			Chroot: true,
			Command: cli.Command{
				Name:    "list",
				Usage:   "List packages based on package names",
				Aliases: []string{"ls"},
			},
		},
		&PmCommand{
			Chroot: true,
			Command: cli.Command{
				Name:    "search",
				Usage:   "Search in package descriptions",
				Aliases: []string{"se"},
			},
		},
		&PmCommand{
			Chroot: true,
			Command: cli.Command{
				Name:    "show",
				Usage:   "Show package details",
				Aliases: []string{"info"},
			},
		},
		&PmCommand{
			Chroot: true,
			Command: cli.Command{
				Name:    "install",
				Usage:   "Install packages",
				Aliases: []string{"in"},
			},
		},
		&PmCommand{
			Chroot: true,
			Command: cli.Command{
				Name:    "purge",
				Usage:   "Completely uninstall selected packages",
				Aliases: []string{"pg"},
			},
		},
		&PmCommand{
			Chroot: true,
			Command: cli.Command{
				Name:    "reinstall",
				Usage:   "Reinstall packages",
				Aliases: []string{"rein", "rin"},
			},
		},
		&PmCommand{
			Chroot: true,
			Command: cli.Command{
				Name:    "remove",
				Usage:   "Remove packages",
				Aliases: []string{"rm"},
			},
		},
		&PmCommand{
			Chroot: true,
			Command: cli.Command{
				Name:    "autoremove",
				Usage:   "Remove automatically all unused packages",
				Aliases: []string{"arm"},
			},
		},
		&PmCommand{
			Chroot: true,
			Command: cli.Command{
				Name:    "update",
				Usage:   "Update list of available packages",
				Aliases: []string{"up"},
			},
		},
		&PmCommand{
			Chroot: true,
			Command: cli.Command{
				Name:    "upgrade",
				Usage:   "Upgrade the system by installing/upgrading packages",
				Aliases: []string{"dup"},
			},
		},
		&PmCommand{
			Chroot: true,
			Command: cli.Command{
				Name:    "full-upgrade",
				Usage:   "Upgrade the system by removing/installing/upgrading packages",
				Aliases: []string{"fdup"},
			},
		},
		&PmCommand{
			Chroot: true,
			Command: cli.Command{
				Name:    "edit-sources",
				Usage:   "Edit the source information file",
				Aliases: []string{"esr"},
			},
		},
		&PmCommand{
			Chroot: true,
			Command: cli.Command{
				Name:    "list-installed",
				Usage:   "List installed packages",
				Aliases: []string{"ils", "ils", "installed"},
			},
		},
		&PmCommand{
			Chroot: true,
			Command: cli.Command{
				Name:    "files",
				Usage:   "List contents of a specific package",
				Aliases: []string{"content"},
			},
		},
		&PmCommand{
			Chroot: true,
			Command: cli.Command{
				Name:    "chroot",
				Usage:   "Change root to the selected sysroot",
				Aliases: []string{"c", "run"},
			},
		},
		&PmCommand{
			Chroot: true,
			Command: cli.Command{
				Name:    "satisfy",
				Usage:   "Satisfy dependency strings",
				Aliases: []string{"sat"},
			},
		},
	}
}
