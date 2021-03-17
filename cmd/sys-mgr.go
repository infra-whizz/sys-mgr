package main

import (
	"fmt"
	"os"
	"path"
	"sort"
	"strings"

	sysmgr "github.com/infra-whizz/sys-mgr"
	sysmgr_arch "github.com/infra-whizz/sys-mgr/arch"
	sysmgr_pm "github.com/infra-whizz/sys-mgr/pm"
	sysmgr_sr "github.com/infra-whizz/sys-mgr/sr"
	wzlib_logger "github.com/infra-whizz/wzlib/logger"
	wzlib_subprocess "github.com/infra-whizz/wzlib/subprocess"
	"github.com/isbm/go-nanoconf"
	"github.com/sirupsen/logrus"
	"github.com/thoas/go-funk"
	"github.com/urfave/cli/v2"
)

var appname string
var pkgman sysmgr_pm.PackageManager
var architectures []string
var mgr *sysmgr_sr.SysrootManager
var binfmt *sysmgr_arch.BinFormat

func init() {
	appname = path.Base(os.Args[0])
	pkgman = sysmgr.GetCurrentPackageManager()
	binfmt = sysmgr_arch.NewBinFormat()

	// setup logger
	if funk.Contains(os.Args, "--verbose") || funk.Contains(os.Args, "-v") {
		wzlib_logger.GetCurrentLogger().SetLevel(logrus.TraceLevel)
	} else {
		wzlib_logger.GetCurrentLogger().SetLevel(logrus.ErrorLevel)
	}

	architectures = []string{}
	for _, arch := range binfmt.Architectures {
		architectures = append(architectures, arch.Name)
	}

	sort.Strings(architectures)

	if appname != fmt.Sprintf("%s-sysroot", pkgman.Name()) && appname != "sysroot-manager" {
		wzlib_logger.GetCurrentLogger().Errorf("This app should be called '%s-sysroot'.", pkgman.Name())
		os.Exit(1)
	}

	confpath := nanoconf.NewNanoconfFinder("sysroots").DefaultSetup(nil)
	mgr = sysmgr_sr.NewSysrootManager(nanoconf.NewConfig(confpath.SetDefaultConfig(confpath.FindFirst()).FindDefault())).SetSupportedArchitectures(architectures)

	if err := runArchGate(appname, mgr); err != nil {
		wzlib_logger.GetCurrentLogger().Errorf("Error: %s", err.Error())
		os.Exit(1)
	}
}

// Exit if the current user is not root
func exitOnNonRootUID() {
	if !funk.Contains(os.Args, "-h") && !funk.Contains(os.Args, "--help") {
		if err := sysmgr.CheckUser(0, 0); err != nil {
			wzlib_logger.GetCurrentLogger().Error("Root privileges are required to run this command.")
			os.Exit(1)
		}
	}
}

func runArchGate(k string, m *sysmgr_sr.SysrootManager) error {
	// intercept itself as a
	if k == "sysroot-manager" {
		if len(os.Args) == 1 || funk.Contains(os.Args, "-h") || funk.Contains(os.Args, "--help") {
			fmt.Printf("This is a helper utility and should not be directly used.\nYou are looking for '%s-sysroot' instead.\n", pkgman.Name())
			os.Exit(0)
		}
		dr, _ := m.GetDefaultSysroot()
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
func runPackageManager() error {
	sysroot, err := mgr.GetDefaultSysroot()
	if err != nil {
		return err
	}
	return pkgman.SetSysroot(sysroot).Call(os.Args[1:]...)
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
func runSystemManager(ctx *cli.Context) error {
	if ctx.Bool("list") {
		roots, err := mgr.GetSysRoots()
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
		exitOnNonRootUID()
		roots, err := mgr.GetSysRoots()
		if err != nil {
			return err
		}

		isDefault := len(roots) == 0
		name, arch := getNameArch(ctx)
		wzlib_logger.GetCurrentLogger().Infof("Creating system root: %s (%s)", name, arch)
		sysroot, err := mgr.CreateSysRoot(name, arch)
		if err != nil {
			return err
		}
		if err := sysroot.SetDefault(isDefault); err != nil {
			return err
		}
		if err := pkgman.SetSysroot(sysroot).Setup(); err != nil {
			return err
		}
		if isDefault {
			return sysroot.Activate()
		}
	} else if ctx.Bool("delete") {
		exitOnNonRootUID()
		name, arch := getNameArch(ctx)
		wzlib_logger.GetCurrentLogger().Infof("Deleting system root: %s (%s)", name, arch)
		return mgr.DeleteSysRoot(name, arch)
	} else if ctx.Bool("set") {
		exitOnNonRootUID()
		name, arch := getNameArch(ctx)

		// Detach current default
		psr, err := mgr.GetDefaultSysroot()
		if err != nil {
			return err
		}
		if psr != nil {
			if err := psr.UmountBinds(); err != nil {
				return err
			}
		}

		wzlib_logger.GetCurrentLogger().Infof("Setting selected system root '%s' (%s) as default", name, arch)
		if err := mgr.SetDefaultSysRoot(name, arch); err != nil {
			return err
		}
		if err := binfmt.Register(arch); err != nil {
			return err
		}

		sr, err := mgr.GetDefaultSysroot()
		if err != nil {
			return err
		}

		// Setup systemd
		sds := sysmgr_arch.NewSystemdService().SetPackageManager(pkgman)
		if err := sds.Remove(); err != nil {
			return err
		}
		if err := sds.Create(sr.Arch); err != nil {
			return err
		}

		// Activate
		return sr.Activate()
	} else if ctx.Bool("path") {
		sr, err := mgr.GetDefaultSysroot()
		if err != nil {
			return err
		}
		fmt.Println(sr.Path)
	} else if ctx.Bool("init") {
		exitOnNonRootUID()
		sr, err := mgr.GetDefaultSysroot()

		if err != nil {
			return err
		}

		if err := binfmt.Register(sr.Arch); err != nil {
			return err
		}

		if err := sysmgr_arch.NewSystemdService().SetPackageManager(pkgman).Create(sr.Arch); err != nil {
			return err
		}
		return sr.Activate()
	} else {
		cli.ShowSubcommandHelpAndExit(ctx, 1)
	}

	return nil
}

func main() {
	app := &cli.App{
		Version: "0.1 Alpha",
		Name:    appname,
		Usage:   fmt.Sprintf("System root manager (via %s)", pkgman.Name()),
	}

	app.Commands = []*cli.Command{
		{
			Name:   "sysroot",
			Usage:  "Manage sysroot",
			Action: runSystemManager,
			Flags: []cli.Flag{
				&cli.BoolFlag{
					Name:    "list",
					Aliases: []string{"l"},
					Usage:   "List available system roots",
				},
				&cli.BoolFlag{
					Name:    "create",
					Aliases: []string{"c"},
					Usage:   "Create a new system root",
				},
				&cli.BoolFlag{
					Name:    "delete",
					Aliases: []string{"d"},
					Usage:   "Delete a system root by name",
				},
				&cli.BoolFlag{
					Name:    "init",
					Aliases: []string{"i"},
					Usage:   "Init default system root",
				},
				&cli.BoolFlag{
					Name:    "set",
					Aliases: []string{"s"},
					Usage:   "Set default system root by name",
				},
				&cli.BoolFlag{
					Name:    "path",
					Aliases: []string{"p"},
					Usage:   "Display path of an active system root",
				},
				&cli.StringFlag{
					Name:    "name",
					Aliases: []string{"n"},
					Usage:   "Set name of the system root",
				},
				&cli.StringFlag{
					Name:    "arch",
					Aliases: []string{"a"},
					Usage:   fmt.Sprintf("Set architecture for the system root. Choices: %s.", strings.Join(architectures, ", ")),
				},
				&cli.BoolFlag{
					Name:    "verbose",
					Aliases: []string{"v"},
					Usage:   "Show debugging log",
				},
			},
		},
	}

	var err error
	if len(os.Args) == 1 || sysmgr.Any(os.Args, "sysroot", "-h", "--help") {
		err = app.Run(os.Args)
	} else {
		err = runPackageManager()
	}
	if err != nil {
		wzlib_logger.GetCurrentLogger().Errorf("General error: %s", err.Error())
	}
}
