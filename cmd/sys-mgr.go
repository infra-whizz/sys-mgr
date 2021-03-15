package main

import (
	"fmt"
	"os"
	"path"
	"sort"
	"strings"

	sysmgr "github.com/infra-whizz/sys-mgr"
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

func init() {
	appname = path.Base(os.Args[0])
	pkgman = sysmgr.GetCurrentPackageManager()
	wzlib_logger.GetCurrentLogger().SetLevel(logrus.InfoLevel)
	architectures = []string{
		"x86_64", "aarch64", "ppc64", "ppc64le", "s390x", "riscv64", "mips64", "sparc64",
	}
	sort.Strings(architectures)

	if appname != fmt.Sprintf("%s-sysroot", pkgman.Name()) && appname != "sysroot-manager" {
		wzlib_logger.GetCurrentLogger().Errorf("This app should be called '%s-sysroot'.", pkgman.Name())
		os.Exit(1)
	}

	if !funk.Contains(os.Args, "-h") && !funk.Contains(os.Args, "--help") {
		if err := sysmgr.CheckUser(0, 0); err != nil {
			wzlib_logger.GetCurrentLogger().Error("Root privileges are required to run this command.")
			os.Exit(1)
		}
	}

	confpath := nanoconf.NewNanoconfFinder("sysroots").DefaultSetup(nil)
	mgr = sysmgr_sr.NewSysrootManager(nanoconf.NewConfig(confpath.SetDefaultConfig(confpath.FindFirst()).FindDefault())).SetSupportedArchitectures(architectures)

	if err := runArchGate(appname, mgr); err != nil {
		wzlib_logger.GetCurrentLogger().Errorf("Error: %s", err.Error())
		os.Exit(1)
	}
}

// If the command is called as "sysroot-manager", it will run as qemu-<
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
		roots, err := mgr.GetSysRoots()
		if err != nil {
			return err
		}

		name, arch := getNameArch(ctx)
		wzlib_logger.GetCurrentLogger().Infof("Creating system root: %s (%s)", name, arch)
		sysroot, err := mgr.CreateSysRoot(name, arch)
		if err != nil {
			return err
		}
		if err := sysroot.SetDefault(len(roots) == 0); err != nil {
			return err
		}
		return pkgman.SetSysroot(sysroot).Setup()
	} else if ctx.Bool("delete") {
		name, arch := getNameArch(ctx)
		wzlib_logger.GetCurrentLogger().Infof("Deleting system root: %s (%s)", name, arch)
		return mgr.DeleteSysRoot(name, arch)
	} else if ctx.Bool("set") {
		name, arch := getNameArch(ctx)
		wzlib_logger.GetCurrentLogger().Infof("Setting selected system root '%s' (%s) as default", name, arch)
		return mgr.SetDefaultSysRoot(name, arch)
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
					Name:    "set",
					Aliases: []string{"s"},
					Usage:   "Set default system root by name",
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
