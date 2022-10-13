package main

import (
	"fmt"
	"os"
	"path"
	"sort"
	"strings"

	sysmgr "github.com/infra-whizz/sys-mgr"
	wzlib_logger "github.com/infra-whizz/wzlib/logger"
	"github.com/sirupsen/logrus"
	"github.com/thoas/go-funk"
	"github.com/urfave/cli/v2"
)

var sm *sysmgr.SysrootManager
var VERSION string = "1.1"

func init() {
	sm = sysmgr.NewSysrootManager(path.Base(os.Args[0]))

	// setup logger
	if funk.Contains(os.Args, "--verbose") || funk.Contains(os.Args, "--debug") {
		wzlib_logger.GetCurrentLogger().SetLevel(logrus.TraceLevel)
	} else {
		wzlib_logger.GetCurrentLogger().SetLevel(logrus.ErrorLevel)
	}

	if err := sm.RunArchGate(); err != nil {
		wzlib_logger.GetCurrentLogger().Errorf("Gate arch error: %s", err.Error())
		os.Exit(1)
	}
}

func buildAppHelpCommands(app *cli.App) string {
	flags := sm.PkgManager().GetHelpFlags()
	keys := []string{}
	for k := range flags {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	out := ""
	for _, k := range keys {
		out += fmt.Sprintf("  %s\t%s\n", k, flags[k])
	}
	for _, c := range app.Commands {
		out += fmt.Sprintf("  %s\t%s", c.Name, c.Usage)
	}

	return out
}

func main() {
	app := &cli.App{
		Version: VERSION,
		Name:    sm.AppName(),
		Usage:   fmt.Sprintf("System root manager (via %s)", sm.PkgManager().Name()),
	}

	app.Commands = []*cli.Command{
		{
			CustomHelpTemplate: `NAME:
{{$v := offset .HelpName 6}}{{wrap .HelpName 3}}{{if .Usage}} - {{wrap .Usage $v}}{{end}}

USAGE:
	{{if .UsageText}}{{wrap .UsageText 3}}{{else}}{{.HelpName}}{{if .VisibleFlags}} [command options]{{end}} {{if .ArgsUsage}}{{.ArgsUsage}}{{else}}[arguments...]{{end}}{{end}}{{if .Category}}

CATEGORY:
	{{.Category}}{{end}}{{if .Description}}

DESCRIPTION:
	{{wrap .Description 3}}{{end}}{{if .VisibleFlagCategories}}

OPTIONS:
	{{range .VisibleFlagCategories}}
	{{if .Name}}{{.Name}}
	{{end}}{{range .Flags}}{{.}}
	{{end}}{{end}}{{else}}{{if .VisibleFlags}}

OPTIONS:
	{{range .VisibleFlags}}{{.}}
	{{end}}{{end}}{{end}}
`,

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
					Name:  "verbose",
					Usage: "Show debugging log",
				},
			},
		},
	}

	app.CustomAppHelpTemplate = `NAME:
	{{template "helpNameTemplate" .}}

USAGE:
	{{if .UsageText}}{{wrap .UsageText 3}}{{else}}{{.HelpName}} {{if .VisibleFlags}}[global options]{{end}}{{if .Commands}} command [command options]{{end}} {{if .ArgsUsage}}{{.ArgsUsage}}{{else}}[arguments...]{{end}}{{end}}{{if .Version}}{{if not .HideVersion}}

VERSION:
	{{.Version}}{{end}}{{end}}{{if .Description}}

DESCRIPTION:
   {{template "descriptionTemplate" .}}{{end}}
{{- if len .Authors}}

AUTHOR{{template "authorsTemplate" .}}{{end}}{{if .VisibleCommands}}

COMMANDS:
` +
		buildAppHelpCommands(app) +

		`{{end}}{{if .VisibleFlagCategories}}

GLOBAL OPTIONS:{{template "visibleFlagCategoryTemplate" .}}{{else if .VisibleFlags}}

GLOBAL OPTIONS:{{template "visibleFlagTemplate" .}}{{end}}{{if .Copyright}}

COPYRIGHT:{{template "copyrightTemplate" .}}{{end}}

   `

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
