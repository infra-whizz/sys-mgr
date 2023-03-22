#
# spec file for package sysroot-mgr
#
# Copyright (c) 2022 Bo Maryniuk
#
# All modifications and additions to the file contributed by third parties
# remain the property of their copyright owners, unless otherwise agreed
# upon. The license for this file, and modifications and additions to the
# file, is the same license as for the pristine package itself (unless the
# license for the pristine package is not an Open Source License, in which
# case the license is the MIT License). An "Open Source License" is a
# license that conforms to the Open Source Definition (Version 1.9)
# published by the Open Source Initiative.

Name:           sysroot-manager
Version:        2.0
Release:        0
Summary:        Manage auxiliary system roots
License:        MIT
Group:          Embedded/Tools
Url:            https://gitlab.com/infra-whizz/sys-mgr
Source:         %{name}-%{version}.tar.gz


Requires:       qemu-user-static
Requires:       debootstrap

BuildRequires:  debbuild
BuildRequires:  golang

%description
Manage auxiliary system roots, adding unlimited amount of them for cross-compilation,
regular development or playground purposes, where containers or VM is an overhead.

%prep
%setup -q

%build
# Output for log purpuses
go env

# Build the binary
CGO_ENABLED=0 go build -a -mod=vendor -tags netgo -ldflags '-w -extldflags "-static"' -o %{name} ./cmd/*go

%install
install -D -m 0755 %{name} %{buildroot}%{_bindir}/%{name}
mkdir -p %{buildroot}%{_sysconfdir}
mkdir %{buildroot}/usr/sysroots
install -m 0644 ./etc/sysroots.conf %{buildroot}%{_sysconfdir}/sysroots.conf
ln -s %{name} %{buildroot}%{_bindir}/apt-sysroot

%files
%defattr(-,root,root)
%{_bindir}/%{name}
%{_bindir}/apt-sysroot
%dir /usr/sysroots
%config /etc/sysroots.conf

%changelog
