package sysmgr

import (
	"fmt"
	"os"
	"os/user"
	"strconv"

	"github.com/elastic/go-sysinfo"
	sysmgr_pm "github.com/infra-whizz/sys-mgr/pm"
	"github.com/thoas/go-funk"
)

func CheckUser(uid int, gid int) error {
	var err error
	var u *user.User
	if u, err = user.Current(); err != nil {
		return err
	}

	userId, _ := strconv.ParseInt(u.Uid, 10, 64)
	groupId, _ := strconv.ParseInt(u.Gid, 10, 64)

	if int(userId) != uid {
		return fmt.Errorf("User ID does not match")
	}

	if int(groupId) != gid {
		return fmt.Errorf("Group ID does not match")
	}

	return nil
}

func GetCurrentPackageManager() sysmgr_pm.PackageManager {
	host, err := sysinfo.Host()
	if err != nil {
		panic(err)
	}

	platform := host.Info().OS.Platform
	var pkgman sysmgr_pm.PackageManager
	switch platform {
	case "ubuntu":
		os.Stderr.WriteString("This is Ubuntu platform. Currently 'apt' is not supported.\n")
		os.Exit(1)
	case "opensuse-leap":
		pkgman = sysmgr_pm.NewZypperPackageManager()
	default:
		os.Stderr.WriteString("The '%s' platform is not supported.\n")
		os.Exit(1)
	}

	return pkgman
}

// Any of the occurrences
func Any(in interface{}, args ...interface{}) bool {
	for _, arg := range args {
		if funk.Contains(in, arg) {
			return true
		}
	}
	return false
}
