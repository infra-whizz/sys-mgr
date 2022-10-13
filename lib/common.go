package sysmgr_lib

import (
	"fmt"
	"os/user"
	"strconv"

	"github.com/elastic/go-sysinfo"
	"github.com/elastic/go-sysinfo/types"
	"github.com/thoas/go-funk"
)

// Any of the occurrences
func Any(in interface{}, args ...interface{}) bool {
	for _, arg := range args {
		if funk.Contains(in, arg) {
			return true
		}
	}
	return false
}

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

var _currentHostInfo types.Host

// GetCurrentPlatform returns a current platform class
func GetCurrentPlatform() string {
	var err error
	if _currentHostInfo == nil {
		_currentHostInfo, err = sysinfo.Host()
		if err != nil {
			panic(err)
		}
	}

	return _currentHostInfo.Info().OS.Platform
}
