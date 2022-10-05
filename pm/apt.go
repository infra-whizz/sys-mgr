package sysmgr_pm

import sysmgr_sr "github.com/infra-whizz/sys-mgr/sr"

type AptPackageManager struct {
	sysroot *sysmgr_sr.SysRoot
	archFix map[string]string

	BasePackageManager
}

func NewAptPackageManager() *AptPackageManager {
	pm := new(AptPackageManager)
	pm.archFix = map[string]string{}
	pm.env = map[string]string{}

	return pm
}
