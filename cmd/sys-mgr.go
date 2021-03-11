package main

import (
	"fmt"
	"os"
	"path"
	"sort"
	"strings"

	sysmgr "github.com/infra-whizz/sys-mgr"
	sysmgr_pm "github.com/infra-whizz/sys-mgr/pm"
	wzlib_logger "github.com/infra-whizz/wzlib/logger"
	"github.com/isbm/go-nanoconf"
	"github.com/sirupsen/logrus"
	"github.com/thoas/go-funk"
	"github.com/urfave/cli/v2"
)

var appname string
var pkgman sysmgr_pm.PackageManager
var architectures []string

func init() {
	appname = path.Base(os.Args[0])
	pkgman = sysmgr.GetCurrentPackageManager()
	wzlib_logger.GetCurrentLogger().SetLevel(logrus.ErrorLevel)
	architectures = []string{
		"x86_64", "aarch64", "ppc64", "ppc64le", "s390x", "riscv64", "mips64", "sparc64",
	}
	sort.Strings(architectures)

	if appname != fmt.Sprintf("%s-sysroot", pkgman.Name()) {
		os.Stderr.WriteString(fmt.Sprintf("This app should be called '%s-sysroot'.\n", pkgman.Name()))
		os.Exit(1)
	}

	if !funk.Contains(os.Args, "-h") && !funk.Contains(os.Args, "--help") {
		if err := sysmgr.CheckUser(0, 0); err != nil {
			os.Stderr.WriteString("Root privileges are required to run this command.\n")
			os.Exit(1)
		}
	}
}

func runPackageManager() error {
	pkgman.Call(os.Args[1:]...)
	return nil
}

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
	confpath := nanoconf.NewNanoconfFinder("sysroots").DefaultSetup(nil)
	mgr := sysmgr.NewSysrootManager(nanoconf.NewConfig(confpath.SetDefaultConfig(confpath.FindFirst()).FindDefault())).SetSupportedArchitectures(architectures)
	if ctx.Bool("list") {
		roots, err := mgr.GetSysRoots()
		if err != nil {
			wzlib_logger.GetCurrentLogger().Errorf("Error while getting system roots: %s", err.Error())
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
		name, arch := getNameArch(ctx)
		wzlib_logger.GetCurrentLogger().Infof("Creating system root: %s (%s)", name, arch)
		return mgr.CreateSysRoot(name, arch)
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

	if len(os.Args) == 1 || sysmgr.Any(os.Args, "sysroot", "-h", "--help") {
		if err := app.Run(os.Args); err != nil {
			wzlib_logger.GetCurrentLogger().Errorf("General error: %s", err.Error())
		}
	} else {
		runPackageManager()
	}
}
