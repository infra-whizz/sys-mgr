package main

import (
	"fmt"
	"os"
	"path"
	"sort"
	"strings"

	sysmgr "github.com/infra-whizz/sys-mgr"
	wzlib_logger "github.com/infra-whizz/wzlib/logger"
	"github.com/isbm/go-nanoconf"
	"github.com/urfave/cli/v2"
)

var appname string
var pkgman string
var architectures []string

func init() {
	appname = path.Base(os.Args[0])
	pkgman = sysmgr.GetCurrentPackageManager()
	architectures = []string{
		"x86_64", "aarch64", "ppc64", "ppc64le", "s390x", "riscv64", "mips64", "sparc64",
	}
	sort.Strings(architectures)

	if appname != fmt.Sprintf("%s-sysroot", pkgman) {
		os.Stderr.WriteString(fmt.Sprintf("This app should be called '%s-sysroot'.\n", pkgman))
		os.Exit(1)
	}
}

func runPackageManager(ctx *cli.Context) error {
	fmt.Print("Package manager \n")
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
				fmt.Printf("%d. %s (%s)\n", idx+1, sr.Name, sr.Arch)
			}
		}
	} else if ctx.Bool("create") {
		wzlib_logger.GetCurrentLogger().Info("Creating system root...")
		name, arch := getNameArch(ctx)
		return mgr.CreateSysRoot(name, arch)
	} else if ctx.Bool("delete") {
		wzlib_logger.GetCurrentLogger().Info("Creating system root...")
		name, arch := getNameArch(ctx)
		return mgr.DeleteSysRoot(name, arch)
	} else {
		cli.ShowSubcommandHelpAndExit(ctx, 1)
	}

	return nil
}

func main() {
	app := &cli.App{
		Version: "0.1 Alpha",
		Name:    appname,
		Usage:   fmt.Sprintf("System root manager (via %s)", pkgman),
		Action:  runPackageManager,
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

	if err := app.Run(os.Args); err != nil {
		wzlib_logger.GetCurrentLogger().Errorf("General error: %s", err.Error())
	}
}
