#
# spec file for package sysroot-mgr
#
# Copyright (c) 2021 Elektrobit Automotive, Erlangen, Germany.
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
Version:        1.1
Release:        0
Summary:        System root manager
License:        MIT
Group:          Embedded/Tools
Url:            https://gitlab.com/infra-whizz/sys-mgr
Source:         %{name}-%{version}.tar.gz
Source1:        vendor.tar.gz

BuildRequires:  git-core
BuildRequires:  golang-packaging
BuildRequires:  golang(API) >= 1.13
Requires:       qemu-linux-user

%description
System root manager allows installing and fully manage a separate, alternative system roots for cross-compilation purposes.

%prep
%setup -q
%setup -q -T -D -a 1

%build
CGO_ENABLED=0 go build -a -mod=vendor -tags netgo -ldflags '-w -extldflags "-static"' -o %{name} ./cmd/*go

%install
install -D -m 0755 %{name} %{buildroot}%{_bindir}/%{name}
mkdir -p %{buildroot}%{_sysconfdir}
mkdir %{buildroot}/usr/sysroots
install -m 0644 ./etc/sysroots.conf %{buildroot}%{_sysconfdir}/sysroots.conf
ln -s %{name} %{buildroot}%{_bindir}/zypper-sysroot

%files
%defattr(-,root,root)
%{_bindir}/%{name}
%{_bindir}/zypper-sysroot
%dir %{_sysconfdir}
%dir /usr/sysroots
%config /etc/sysroots.conf

%changelog
