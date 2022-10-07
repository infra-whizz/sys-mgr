package sysmgr_sr

type ZypperSysrootProvisioner struct {
	BaseSysrootProvisioner
}

func NewZypperSysrootProvisioner(name, arch, root string) *ZypperSysrootProvisioner {
	zsp := new(ZypperSysrootProvisioner)

	zsp.SetArch(arch)
	zsp.SetName(name)
	zsp.SetSysPath(root)

	return zsp
}

func (zsp *ZypperSysrootProvisioner) beforePopulate() error {
	return nil
}

func (zsp *ZypperSysrootProvisioner) onPopulate() error {
	return nil
}

func (zsp *ZypperSysrootProvisioner) afterPopulate() error {
	return nil
}
