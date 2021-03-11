package sysmgr_pm

import (
	"fmt"
	"os"
	"strings"

	wzlib_subprocess "github.com/infra-whizz/wzlib/subprocess"
)

// BasePackageManager mixin
type BasePackageManager struct {
	env map[string]string
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
	}

	return cmd.Run()
}
