package sysmgr

import (
	"fmt"
	"os"
	"path"
	"sort"

	sysmgr_arch "github.com/infra-whizz/sys-mgr/arch"
	sysmgr_pm "github.com/infra-whizz/sys-mgr/pm"
	sysmgr_sr "github.com/infra-whizz/sys-mgr/sr"
	wzlib_logger "github.com/infra-whizz/wzlib/logger"
	wzlib_subprocess "github.com/infra-whizz/wzlib/subprocess"
	"github.com/isbm/go-nanoconf"
	"github.com/thoas/go-funk"
	"github.com/urfave/cli/v2"
)

// SysrootManager object
type SysrootManager struct {
	appname       string
	pkgman        sysmgr_pm.PackageManager
	architectures []string
	mgr           *sysmgr_sr.SysrootManager
	binfmt        *sysmgr_arch.BinFormat

	wzlib_logger.WzLogger
}

// NewSysrootManager constructor
func NewSysrootManager(appname string) *SysrootManager {
	srm := new(SysrootManager)
	srm.pkgman = GetCurrentPackageManager()
	srm.binfmt = sysmgr_arch.NewBinFormat()
	srm.appname = appname

	srm.architectures = []string{}
	for _, arch := range srm.binfmt.Architectures {
		srm.architectures = append(srm.architectures, arch.Name)
	}

	sort.Strings(srm.architectures)

	confpath := nanoconf.NewNanoconfFinder("sysroots").DefaultSetup(nil)
	srm.mgr = sysmgr_sr.NewSysrootManager(nanoconf.NewConfig(confpath.SetDefaultConfig(confpath.FindFirst()).FindDefault())).
		SetSupportedArchitectures(srm.architectures)

	if srm.appname != fmt.Sprintf("%s-sysroot", srm.pkgman.Name()) && srm.appname != "sysroot-manager" {
		wzlib_logger.GetCurrentLogger().Errorf("This app should be called '%s-sysroot'.", srm.pkgman.Name())
		os.Exit(1)
	}

	return srm
}

// AppName returns a name of the binary, as it should have multiple ones
func (srm SysrootManager) AppName() string {
	return srm.appname
}

// PkgManager underneath the system root manager
func (srm SysrootManager) PkgManager() sysmgr_pm.PackageManager {
	return srm.pkgman
}

// Architectures returns a list of supported archs
func (srm SysrootManager) Architectures() []string {
	return srm.architectures
}

// ExitOnNonRootUID will terminate program immediately if caller is not UID root.
func (srm *SysrootManager) ExitOnNonRootUID() {
	if !funk.Contains(os.Args, "-h") && !funk.Contains(os.Args, "--help") {
		if err := CheckUser(0, 0); err != nil {
			wzlib_logger.GetCurrentLogger().Error("Root privileges are required to run this command.")
			os.Exit(1)
		}
	}
}

// RunArchGate runs every time to check if it should intercept any external calls
func (srm SysrootManager) RunArchGate() error {
	// intercept itself as a
	if srm.appname == "sysroot-manager" {
		if len(os.Args) == 1 || funk.Contains(os.Args, "-h") || funk.Contains(os.Args, "--help") {
			fmt.Printf("This is a helper utility and should not be directly used.\nYou are looking for '%s-sysroot' instead.\n", srm.pkgman.Name())
			os.Exit(0)
		}
		dr, _ := srm.mgr.GetDefaultSysroot()
		var args []string
		if _, err := os.Stat("/etc/sysroot.conf"); os.IsNotExist(err) {
			if dr == nil {
				return fmt.Errorf("Sysroot was not found though")
			}
			// Call natively
			args = append([]string{
				path.Join(dr.Path, "/lib/ld-linux-armhf.so.3"), "--library-path",
				fmt.Sprintf("%s:%s", path.Join(dr.Path, "/usr/lib"), path.Join(dr.Path, "/lib")),
			}, os.Args[1:]...)
		} else {
			// Call chrooted
			args = os.Args[1:]
		}

		cmd := wzlib_subprocess.ExecCommand("/usr/bin/qemu-arm", args...)
		cmd.Stderr = os.Stderr
		cmd.Stdout = os.Stdout
		cmd.Stdin = os.Stdin
		if err := cmd.Run(); err != nil {
			fmt.Println("Error:", err.Error())
			os.Exit(1)
		}

		os.Exit(0)
	}

	// no-op
	return nil
}

// Run underlying package manager
func (srm SysrootManager) RunPackageManager() error {
	sysroot, err := srm.mgr.GetDefaultSysroot()
	if err != nil {
		return err
	}
	return srm.pkgman.SetSysroot(sysroot).Call(os.Args[1:]...)
}

// Get the name of the architecture
func getNameArch(ctx *cli.Context) (string, string) {
	name := ctx.String("name")
	if name == "" {
		wzlib_logger.GetCurrentLogger().Errorf("The name of the sysroot is missing.")
		os.Exit(1)
	}
	arch := ctx.String("arch")
	if arch == "" {
		wzlib_logger.GetCurrentLogger().Errorf("Architecture of the sysroot is missing.")
		os.Exit(1)
	}
	return name, arch
}

// Run system manager
func (srm SysrootManager) RunSystemManager(ctx *cli.Context) error {
	if ctx.Bool("list") {
		roots, err := srm.mgr.GetSysRoots()
		if err != nil {
			return err
		}
		if len(roots) > 0 {
			fmt.Printf("Found %d system roots:\n", len(roots))
			for idx, sr := range roots {
				d := " "
				if sr.Default {
					d = "*"
				}
				fmt.Printf("%s  %d. %s (%s)\n", d, idx+1, sr.Name, sr.Arch)

			}
		}
	} else if ctx.Bool("create") {
		srm.ExitOnNonRootUID()
		roots, err := srm.mgr.GetSysRoots()
		if err != nil {
			return err
		}

		isDefault := len(roots) == 0
		name, arch := getNameArch(ctx)
		wzlib_logger.GetCurrentLogger().Infof("Creating system root: %s (%s)", name, arch)
		sysroot, err := srm.mgr.CreateSysRoot(name, arch)
		if err != nil {
			return err
		}
		if err := sysroot.SetDefault(isDefault); err != nil {
			return err
		}
		if err := srm.pkgman.SetSysroot(sysroot).Setup(); err != nil {
			return err
		}
		if isDefault {
			return sysroot.Activate()
		}
	} else if ctx.Bool("delete") {
		srm.ExitOnNonRootUID()
		name, arch := getNameArch(ctx)
		wzlib_logger.GetCurrentLogger().Infof("Deleting system root: %s (%s)", name, arch)
		return srm.mgr.DeleteSysRoot(name, arch)
	} else if ctx.Bool("set") {
		srm.ExitOnNonRootUID()
		name, arch := getNameArch(ctx)

		// Detach current default
		psr, err := srm.mgr.GetDefaultSysroot()
		if err != nil {
			return err
		}
		if psr != nil {
			if err := psr.UmountBinds(); err != nil {
				return err
			}
		}

		wzlib_logger.GetCurrentLogger().Infof("Setting selected system root '%s' (%s) as default", name, arch)
		if err := srm.mgr.SetDefaultSysRoot(name, arch); err != nil {
			return err
		}
		if err := srm.binfmt.Register(arch); err != nil {
			return err
		}

		sr, err := srm.mgr.GetDefaultSysroot()
		if err != nil {
			return err
		}

		// Setup systemd
		sds := sysmgr_arch.NewSystemdService().SetPackageManager(srm.pkgman)
		if err := sds.Remove(); err != nil {
			return err
		}
		if err := sds.Create(sr.Arch); err != nil {
			return err
		}

		// Activate
		return sr.Activate()
	} else if ctx.Bool("path") {
		sr, err := srm.mgr.GetDefaultSysroot()
		if err != nil {
			return err
		}
		fmt.Println(sr.Path)
	} else if ctx.Bool("init") {
		srm.ExitOnNonRootUID()
		sr, err := srm.mgr.GetDefaultSysroot()

		if err != nil {
			return err
		}

		if err := srm.binfmt.Register(sr.Arch); err != nil {
			return err
		}

		if err := sysmgr_arch.NewSystemdService().SetPackageManager(srm.pkgman).Create(sr.Arch); err != nil {
			return err
		}
		return sr.Activate()
	} else {
		cli.ShowSubcommandHelpAndExit(ctx, 1)
	}

	return nil
}
