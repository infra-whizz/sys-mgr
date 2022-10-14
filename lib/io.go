package sysmgr_lib

import (
	"os"
	"os/exec"
	"strings"

	wzlib_logger "github.com/infra-whizz/wzlib/logger"
)

type StdoutLogger struct {
	wzlib_logger.WzLogger
}

func (sl *StdoutLogger) Write(p []byte) (n int, err error) {
	sl.GetLogger().Info(strings.TrimSpace(string(p)))
	return len(p), nil
}

func LoggedExec(cmd string, args ...string) error {
	wzlib_logger.GetCurrentLogger().Debugf("Calling: %s %v", cmd, args)
	out := exec.Command(cmd, args...)
	out.Stdin = os.Stdin
	out.Stdout = &StdoutLogger{}
	out.Stderr = os.Stderr
	return out.Run()
}

func StdoutExec(cmd string, args ...string) error {
	out := exec.Command(cmd, args...)
	out.Stdin = os.Stdin
	out.Stdout = os.Stdout
	out.Stderr = os.Stderr
	return out.Run()
}
