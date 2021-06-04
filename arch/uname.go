package sysmgr_arch

import (
	"fmt"
	"strings"
	"syscall"
)

// Uname object
type Uname struct {
	Nodename   string
	Release    string
	Sysname    string
	Version    string
	Machine    string
	Domainname string
}

// NewUname instance
func NewUname() *Uname {
	return new(Uname)
}

// a2s converts C string to a Go string. Compatible with 32 and 64 bit.
func (un *Uname) a2s(data interface{}) string {
	var buf [65]byte
	switch data := data.(type) {
	case [65]uint8:
		for i, b := range data {
			buf[i] = byte(b)
		}
	case [65]int8:
		for i, b := range data {
			buf[i] = byte(b)
		}
	}
	val := string(buf[:])
	if i := strings.Index(val, "\x00"); i != -1 {
		val = val[:i]
	}
	return val
}

// Init Uname
func (un *Uname) Init() error {
	var uname syscall.Utsname
	if err := syscall.Uname(&uname); err != nil {
		return fmt.Errorf("Error init uname: %v", err)
	}

	un.Nodename = un.a2s(uname.Nodename)
	un.Release = un.a2s(uname.Release)
	un.Sysname = un.a2s(uname.Sysname)
	un.Version = un.a2s(uname.Version)
	un.Machine = un.a2s(uname.Machine)
	un.Domainname = un.a2s(uname.Domainname)

	return nil
}
