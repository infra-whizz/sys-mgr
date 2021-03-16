package sysmgr_arch

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"

	wzlib_logger "github.com/infra-whizz/wzlib/logger"
)

type Arch struct {
	Magic string
	Mask  string
	Name  string
}

type BinFormat struct {
	Arch_ARM    *Arch
	Arch_ARM64  *Arch
	Arch_x86_64 *Arch
	Arch_MIPS   *Arch
	Arch_MIPS32 *Arch
	Arch_MIPS64 *Arch

	Architectures []*Arch
	bfmtMisc      string

	wzlib_logger.WzLogger
}

func NewBinFormat() *BinFormat {
	bf := new(BinFormat)

	bf.Arch_ARM = &Arch{
		Magic: "\x7fELF\x01\x01\x01\x00\x00\x00\x00\x00\x00\x00\x00\x00\x02\x00\x28\x00",
		Mask:  "\xff\xff\xff\xff\xff\xff\xff\x00\xff\xff\xff\xff\xff\xff\xff\xff\xfe\xff\xff\xff",
		Name:  "arm",
	}

	bf.Arch_ARM64 = &Arch{
		Magic: "\x7fELF\x02\x01\x01\x00\x00\x00\x00\x00\x00\x00\x00\x00\x02\x00\xb7\x00",
		Mask:  "\xff\xff\xff\xff\xff\xff\xff\x00\xff\xff\xff\xff\xff\xff\xff\xff\xfe\xff\xff\xff",
		Name:  "aarch64",
	}

	bf.Arch_x86_64 = &Arch{
		Magic: "\x7fELF\x02\x01\x01\x00\x00\x00\x00\x00\x00\x00\x00\x00\x02\x00\x3e\x00",
		Mask:  "\xff\xff\xff\xff\xff\xfe\xfe\x00\xff\xff\xff\xff\xff\xff\xff\xff\xfe\xff\xff\xff",
		Name:  "x86_64",
	}

	bf.Arch_MIPS = &Arch{
		Magic: "\x7fELF\x01\x02\x01\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x02\x00\x08",
		Mask:  "\xff\xff\xff\xff\xff\xff\xff\x00\xff\xff\xff\xff\xff\xff\xff\xff\xff\xfe\xff\xff",
		Name:  "mips",
	}

	bf.Arch_MIPS32 = &Arch{
		Magic: "\x7fELF\x01\x02\x01\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x02\x00\x08",
		Mask:  "\xff\xff\xff\xff\xff\xff\xff\x00\xff\xff\xff\xff\xff\xff\xff\xff\xff\xfe\xff\xff",
		Name:  "mips32",
	}

	bf.Arch_MIPS64 = &Arch{
		Magic: "\x7fELF\x02\x02\x01\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x02\x00\x08",
		Mask:  "\xff\xff\xff\xff\xff\xff\xff\x00\xff\xff\xff\xff\xff\xff\xff\xff\xff\xfe\xff\xff",
		Name:  "mips64",
	}

	// Supported architectures
	bf.Architectures = []*Arch{
		bf.Arch_x86_64,
		bf.Arch_ARM, bf.Arch_ARM64,
		bf.Arch_MIPS, bf.Arch_MIPS32, bf.Arch_MIPS64,
	}

	bf.bfmtMisc = "/proc/sys/fs/binfmt_misc"

	return bf
}

// GetArch if the architecture naming matches
func (bf BinFormat) GetArch(arch string) (*Arch, error) {
	for _, a := range bf.Architectures {
		if a.Name == arch {
			return a, nil
		}
	}
	return nil, fmt.Errorf("Unknown architecture: %s", arch)
}

// Get formatted registrar string for the binfmt
func (bf BinFormat) format(arch string) (string, string, error) {
	a, err := bf.GetArch(arch)
	if err != nil {
		return "", "", err
	}

	target := fmt.Sprintf("sysroot_%s", a.Name)
	return target, fmt.Sprintf(":%s:M::%s:%s:/usr/bin/sysroot-manager:", target, a.Magic, a.Mask), nil
}

// Unregister specific architecture. If architecture registration does not exist yet, just pass-through.
func (bf BinFormat) Unregister(arch string) error {
	target, _, err := bf.format(arch)
	if err != nil {
		return err
	}

	regNodePath := path.Join(bf.bfmtMisc, target)
	node, _ := os.Stat(regNodePath)
	if node != nil {
		if err := ioutil.WriteFile(regNodePath, []byte("-1"), 0200); err != nil {
			return err
		}
	}

	return nil
}

// Register architecture. This previously un-registers possible target.
func (bf BinFormat) Register(arch string) error {
	if err := bf.Unregister(arch); err != nil {
		return err
	}

	_, binfmt, err := bf.format(arch)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(path.Join(bf.bfmtMisc, "register"), []byte(binfmt), 0200)
}
