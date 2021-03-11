# sys-mgr

The "sys-mgr" is a system root manager utility, allowing you to install separate, alternative system roots for cross-compilation purposes.

## Basic Usage

Depends on what system you are, e.g. on openSUSE Leap it will be called `zypper-sys`, on Debian/Ubuntu `apt-sys`, on Fedora/RHEL `dnf-sys` etc:

    # zypper-sysroot sysroot --create --name my_sysroot --arch aarch64
    # zypper-sysroot ar http://download.opensuse.org/..... arm_build
    # zypper-sysroot ref
    # zypper-sysroot in emacs

Last three lines belong to the underlying package manager, i.e. `zypper` in this case. For Ubuntu it would be something like:

    # apt-sysroot update
    # apt-sysroot install emacs

It will create a sysroot labeled `my_sysroot` for ARM architecture and install there Emacs for that architecture with all the dependencies.

## "_______ doesn't work/not implemented!"

Yup. Your pull request is highly welcome.
