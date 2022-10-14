package sysmgr_lib

import (
	"io/ioutil"
	"strings"
)

// IsMounted checks if a directory is still mounted or not
func IsMounted(pth string) bool {
	data, err := ioutil.ReadFile("/proc/mounts")
	if err != nil {
		// This should not happen, so we panic
		panic(err)
	}

	for _, l := range strings.Split(string(data), "\n") {
		l = strings.TrimSpace(l)
		if strings.Contains(l, pth) {
			return true
		}
	}

	return false
}
