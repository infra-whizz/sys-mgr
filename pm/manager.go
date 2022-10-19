package sysmgr_pm

import (
	"fmt"
	"os"

	sysmgr_lib "github.com/infra-whizz/sys-mgr/lib"
	sysmgr_sr "github.com/infra-whizz/sys-mgr/sr"
	"github.com/urfave/cli/v2"
)

type PmCommand struct {
	Chroot  bool
	ZeroUID bool

	cli.Command
}

type PmCommands []*PmCommand

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

	// Extract help flags to override package manager
	GetHelpFlags() *PmCommands
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

func GetCurrentPackageManager() PackageManager {
	platform := sysmgr_lib.GetCurrentPlatform()
	var pkgman PackageManager
	switch platform {
	case "ubuntu", "debian":
		pkgman = NewAptPackageManager()
	case "opensuse-leap":
		pkgman = NewZypperPackageManager()
	default:
		os.Stderr.WriteString(fmt.Sprintf("The '%s' platform is not supported. :-(\n", platform))
		os.Exit(1)
	}

	return pkgman
}
