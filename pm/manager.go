package sysmgr_pm

import (
	"os"

	sysmgr_sr "github.com/infra-whizz/sys-mgr/sr"
)

// PackageManager class interface
type PackageManager interface {
	// Call underlying package manager.
	Call(args ...string) error

	// Return a name of the package manager. E.g. on Ubuntu or Debian it will return "apt", on Fedora "dnf" etc.
	Name() string

	// SetSysroot to the package manager and lock on it
	SetSysroot(sysroot *sysmgr_sr.SysRoot) PackageManager

	// Setup package manager, once sysroot is given. This will write required configurations
	Setup() error
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
