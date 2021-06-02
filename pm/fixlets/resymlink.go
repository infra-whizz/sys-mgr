package sysmgr_fixlets

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	sysmgr_sr "github.com/infra-whizz/sys-mgr/sr"
	"github.com/karrick/godirwalk"
)

// ReSymlink type
type ReSymlink struct {
	skipTopDirs []string
	sysroot     *sysmgr_sr.SysRoot
	here        string
}

// NewReSymlink constructor
func NewReSymlink(sysroot *sysmgr_sr.SysRoot) *ReSymlink {
	rsl := new(ReSymlink)
	rsl.sysroot = sysroot
	rsl.skipTopDirs = []string{"proc", "sys", "dev", "run", "boot", "tmp", "mnt"}

	var err error
	rsl.here, err = os.Getwd()
	if err != nil {
		panic("Unable to get current directory: " + err.Error())
	}

	return rsl
}

// a2r converts an existing pathname link on absolute target to a relative counterpart
func (rsl *ReSymlink) a2r(pathname string, target string) string {
	inner := path.Dir(pathname[len(rsl.sysroot.Path):])
	levels := strings.Split(inner, "/")
	if levels[0] == "" {
		levels = levels[1:]
	}
	rjump := "./"
	for i := 0; i < len(levels); i++ {
		rjump += "../"
	}

	return path.Clean(path.Join(rjump, target))
}

// Callback on each dirwalk event
func (rsl *ReSymlink) callback(pathname string, dirEntry *godirwalk.Dirent) error {
	for _, skipTopDir := range rsl.skipTopDirs {
		pref := path.Join(rsl.sysroot.Path, skipTopDir)
		if strings.HasPrefix(pathname, pref) {
			return filepath.SkipDir
		}
	}

	if dirEntry.IsSymlink() {
		brokenPtr, _ := os.Readlink(pathname)
		ptrDir := path.Dir(pathname)

		if strings.HasPrefix(brokenPtr, "/") {
			if err := os.Chdir(ptrDir); err != nil {
				return fmt.Errorf("Cannot change directory to %s: %s", ptrDir, err.Error())
			}

			if err := os.Remove(pathname); err != nil {
				return fmt.Errorf("Broken link (%s) removal error: %s", pathname, err.Error())
			}

			if err := os.Symlink(rsl.a2r(pathname, brokenPtr), path.Base(pathname)); err != nil {
				return fmt.Errorf("Symlink error: %s", err.Error())
			}
		}
	}

	return nil
}

// Relink absolute symlinks to relative
func (rsl *ReSymlink) Relink() error {
	opts := &godirwalk.Options{
		Callback: rsl.callback,
	}
	return godirwalk.Walk(rsl.sysroot.Path, opts)
}
