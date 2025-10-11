# Mockzure RPM Package

This directory contains the RPM package specifications and build scripts for Mockzure.

## Building the RPM

To build the Mockzure RPM package:

```bash
cd Mockzure/deploy/rpm
./build-rpm.sh
```

This will:
1. Generate a timestamp-based version (e.g., `20241011.143055`)
2. Create a source tarball with all necessary files
3. Build the RPM package using `rpmbuild`
4. Output the RPM to this directory
5. Create a symlink `mockzure-latest.rpm` pointing to the most recent build

## Requirements

- `rpmbuild` tool installed
- Go 1.20 or later
- Standard RPM build environment
- Works on both macOS and Linux (portable tar usage)

## Installation

On the target system (Azure Linux with tdnf):

```bash
sudo tdnf install -y mockzure-20241011.143055-1.x86_64.rpm
```

Or using the latest symlink:

```bash
sudo tdnf install -y mockzure-latest.rpm
```

## Package Contents

The RPM installs:
- Binary: `/usr/bin/mockzure`
- Configuration: `/etc/mockzure/config.json`
- Systemd service: `/etc/systemd/system/mockzure.service`
- Data directory: `/var/lib/mockzure/`

## Post-Installation

After installing the RPM:

1. Review the configuration (optional):
   ```bash
   sudo nano /etc/mockzure/config.json
   ```

2. Enable and start the service:
   ```bash
   sudo systemctl enable mockzure
   sudo systemctl start mockzure
   ```

3. Check status:
   ```bash
   sudo systemctl status mockzure
   sudo journalctl -u mockzure -f
   ```

4. Access the web portal:
   - Local: http://localhost:8090
   - Behind reverse proxy: Configure as needed

## Versioning

The package uses timestamp-based versioning in the format `YYYYMMDD.HHMMSS`. This ensures each build has a unique version number and facilitates automatic upgrades.

## Notes

- Mockzure runs as a systemd service using the `sandman` user
- The service listens on port 8090 by default
- Configuration can be customized in `/etc/mockzure/config.json`

