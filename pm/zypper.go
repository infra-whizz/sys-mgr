package sysmgr_pm

import (
	wzlib_subprocess "github.com/infra-whizz/wzlib/subprocess"
)

// ZypperPackageManager object
type ZypperPackageManager struct {
}

// NewZypperPackageManager creates a zypper caller object
func NewZypperPackageManager() *ZypperPackageManager {
	pm := new(ZypperPackageManager)
	return pm
}

// Call zypper
func (pm *ZypperPackageManager) Call(args ...string) (string, string, error) {
	stdout, stderr := wzlib_subprocess.StreamedExec(NewStdProcessStream(), pm.Name(), args...)
	return stdout, stderr, nil
}

// Name of the package manager
func (pm *ZypperPackageManager) Name() string {
	return "zypper"
}
