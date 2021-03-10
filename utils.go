package sysmgr

import (
	"os"

	"github.com/elastic/go-sysinfo"
)

func GetCurrentPackageManager() string {
	host, err := sysinfo.Host()
	if err != nil {
		panic(err)
	}

	platform := host.Info().OS.Platform
	pkgman := ""
	switch platform {
	case "ubuntu":
		os.Stderr.WriteString("This is Ubuntu platform. Currently 'apt' is not supported.\n")
		os.Exit(1)
	case "opensuse-leap":
		pkgman = "zypper"
	default:
		os.Stderr.WriteString("The '%s' platform is not supported.\n")
		os.Exit(1)
	}

	return pkgman
}
