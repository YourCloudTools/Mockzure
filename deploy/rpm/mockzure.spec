Name:           mockzure
Version:        %{_version}
Release:        1%{?dist}
Summary:        Mock Azure API server for testing and development

License:        MIT
URL:            https://github.com/yourcloudtools/sandman
Source0:        %{name}-%{version}.tar.gz
BuildArch:      x86_64

# Note: golang BuildRequires removed for cross-platform builds
# Ensure Go 1.25+ is installed before building
Requires:       systemd

%description
Mockzure is a lightweight mock Azure API server that simulates
Azure Resource Manager and Entra ID (Azure AD) endpoints for
testing and development purposes.

%prep
%setup -q

%build
# Build the Go binary
# Note: CGO_ENABLED=1 required for Azure Linux (CBL-Mariner) FIPS crypto
export GO111MODULE=on
go mod download
CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -o mockzure main.go

%install
rm -rf %{buildroot}

# Create directories
mkdir -p %{buildroot}%{_bindir}
mkdir -p %{buildroot}%{_sysconfdir}/mockzure
mkdir -p %{buildroot}%{_unitdir}
mkdir -p %{buildroot}%{_sharedstatedir}/mockzure

# Install binary
install -m 0755 mockzure %{buildroot}%{_bindir}/mockzure

# Install example configuration (YAML default)
install -m 0644 config.yaml.example %{buildroot}%{_sysconfdir}/mockzure/config.yaml

# Install systemd service
install -m 0644 deploy/systemd/mockzure.service %{buildroot}%{_unitdir}/mockzure.service

%pre
# Create mockzure user and group (uses sandman user from main package)
getent group sandman >/dev/null || groupadd -r sandman
getent passwd sandman >/dev/null || \
    useradd -r -g sandman -d /opt/sandman -s /sbin/nologin \
    -c "Sandman service account" sandman
exit 0

%post
# Set proper permissions
chown -R sandman:sandman %{_sharedstatedir}/mockzure
chmod 700 %{_sharedstatedir}/mockzure
chown root:sandman %{_sysconfdir}/mockzure/config.yaml
chmod 640 %{_sysconfdir}/mockzure/config.yaml

# Reload systemd
%systemd_post mockzure.service

%preun
%systemd_preun mockzure.service

%postun
%systemd_postun_with_restart mockzure.service

%files
%{_bindir}/mockzure
%config(noreplace) %attr(0640,root,sandman) %{_sysconfdir}/mockzure/config.yaml
%{_unitdir}/mockzure.service
%dir %attr(0700,sandman,sandman) %{_sharedstatedir}/mockzure

%changelog
* %(date "+%a %b %d %Y") Auto Build <build@yourcloudtools.com> - %{version}-1
- Automated build from timestamp version

