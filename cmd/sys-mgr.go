package main

import (
	"fmt"
	"os"
	"path"
	"strings"

	sysmgr "github.com/infra-whizz/sys-mgr"
	wzlib_logger "github.com/infra-whizz/wzlib/logger"
	"github.com/sirupsen/logrus"
	"github.com/thoas/go-funk"
	"github.com/urfave/cli/v2"
)

var sm *sysmgr.SysrootManager
var VERSION string = "1.0"

func init() {
	sm = sysmgr.NewSysrootManager(path.Base(os.Args[0]))

	// setup logger
	if funk.Contains(os.Args, "--verbose") || funk.Contains(os.Args, "--debug") {
		wzlib_logger.GetCurrentLogger().SetLevel(logrus.TraceLevel)
	} else {
		wzlib_logger.GetCurrentLogger().SetLevel(logrus.ErrorLevel)
	}

	if err := sm.RunArchGate(); err != nil {
		wzlib_logger.GetCurrentLogger().Errorf("Error: %s", err.Error())
		os.Exit(1)
	}
}

func main() {
	app := &cli.App{
		Version: VERSION,
		Name:    sm.AppName(),
		Usage:   fmt.Sprintf("System root manager (via %s)", sm.PkgManager().Name()),
	}

	app.Commands = []*cli.Command{
		{
			Name:   "sysroot",
			Usage:  "Manage sysroot",
			Action: sm.RunSystemManager,
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
					Usage:   fmt.Sprintf("Set architecture for the system root. Choices: %s.", strings.Join(sm.Architectures(), ", ")),
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
		err = sm.RunPackageManager()
	}
	if err != nil {
		wzlib_logger.GetCurrentLogger().Errorf("General error: %s", err.Error())
	}
}
