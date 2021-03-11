package sysmgr_pm

import (
	"os"

	sysmgr_sr "github.com/infra-whizz/sys-mgr/sr"
)

// PackageManager class interface
type PackageManager interface {
	Call(args ...string) (string, string, error)
	Name() string
	SetSysroot(sysroot *sysmgr_sr.SysRoot) PackageManager
}

// StdProcessStream is just a generic pipe to the STDOUT and nothing else at this time
type StdProcessStream struct {
	filePipe *os.File
}

// NewStdProcessStream creates a ProcessStream instance.
func NewStdProcessStream() *StdProcessStream {
	zs := new(StdProcessStream)
	zs.filePipe = os.Stdout
	return zs
}

// Write data to the underlying pipe file
func (zs *StdProcessStream) Write(data []byte) (n int, err error) {
	return zs.filePipe.Write(data)
}

// Close stream
func (zs *StdProcessStream) Close() error {
	return zs.filePipe.Close()
}
