package sysmgr_pm

import (
	"fmt"
	"os"
	"strings"

	wzlib_logger "github.com/infra-whizz/wzlib/logger"
	wzlib_subprocess "github.com/infra-whizz/wzlib/subprocess"
)

// BasePackageManager mixin
type BasePackageManager struct {
	env map[string]string
	wzlib_logger.WzLogger
}

func (bpm *BasePackageManager) callPackageManager(name string, args ...string) error {
	cmd := wzlib_subprocess.ExecCommand(name, args...)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin

	if bpm.env != nil {
		cmd.Env = os.Environ()
		for ek, ev := range bpm.env {
			if strings.Contains(ev, " ") {
				ev = fmt.Sprintf("\"%s\"", ev)
			}
			cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", strings.TrimSpace(ek), ev))
		}
		bpm.GetLogger().Debugf("Environment set for '%s' with '%v' as follows: '%v'", name, args, cmd.Env)
	}

	return cmd.Run()
}
