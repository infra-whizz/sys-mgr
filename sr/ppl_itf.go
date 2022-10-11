package sysmgr_sr

// SysrootProvisioner is an interface per a distribution that allows to populate
// sysroot according to a specific package manager. For example, on openSUSE
// while using zypper, not much needs to be done other than reconfigure it,
// especially for other arch. On Debian series apt cannot do this, therefore
// "debootstrap" is used.
type SysrootProvisioner interface {
	Populate() error
	Activate() error
	SetSysrootPath(pt string)
	SetArch(a string)
	SetName(n string)
	SetSysPath(p string) // Path of the root
	GetConfigPath() string

	// Internal hooks, should be private and used only in implementation.
	beforePopulate() error // Called before population
	onPopulate() error     // Population implenetation
	afterPopulate() error  // Called after populate is finished
	getQemuPath() string   // Get static QEMU path
	getSysPath() string    // Get sysroot path
}
