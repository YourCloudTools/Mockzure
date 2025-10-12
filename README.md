# Mockzure - Mock Azure Service

A lightweight mock Azure service for development and testing. Mockzure provides a simulated Azure environment with resource management, user authentication, and OIDC/OAuth2 support.

## Features

### Resource Organization

Mockzure organizes resources into two main categories:

1. **Resource Group Managed Resources** - VMs and other compute resources
2. **Global Resources (Entra ID)** - Users and App Registrations

### Web Portal

The Mockzure portal provides a modern, tabbed interface similar to the Azure Portal:

#### 🗂️ Resource Groups Tab
- View all resource groups (rg-dev, rg-prod)
- List VMs organized by resource group
- Display VM details: Name, Size, OS, Owner, Status
- Quick actions: Start/Stop VMs directly from the UI

#### 👥 Entra ID Tab
- **Users Section**: View all users with details (Name, UPN, Job Title, Department, Roles, Status)
- **App Registrations Section**: View registered applications with Client IDs, Redirect URIs, and Scopes

#### ⚙️ Settings Tab
- Statistics dashboard showing VM counts
- API endpoint documentation
- Data management tools (Reset/Clear data)

## Quick Start

### Configuration

Before running Mockzure, copy the example configuration file:

```bash
cp config.json.example config.json
```

Edit `config.json` to add your service account credentials. The config file contains:
- **applicationId**: The client ID for the service account
- **secret**: The client secret for authentication
- **graphPermissions**: Optional Microsoft Graph API permissions (e.g., "User.Read.All")

**Note:** `config.json` is excluded from version control for security. Never commit actual secrets to the repository.

### Using the Web Portal

Open your browser and navigate to:
```
http://localhost:8090
```

You'll see three tabs:
- **Resource Groups**: Manage VMs by resource group
- **Entra ID**: View users and app registrations
- **Settings**: Statistics and data management

## Installation

### Docker (Recommended)

The easiest way to run Mockzure is using Docker. Pre-built multi-platform images are available on GitHub Container Registry:

```bash
# Pull the latest version
docker pull ghcr.io/yourcloudtools/mockzure:latest

# Run Mockzure
docker run -d -p 8090:8090 --name mockzure ghcr.io/yourcloudtools/mockzure:latest

# Or run a specific version
docker pull ghcr.io/yourcloudtools/mockzure:v1.0.0
docker run -d -p 8090:8090 ghcr.io/yourcloudtools/mockzure:v1.0.0
```

Access the web portal at: http://localhost:8090

Supported platforms: `linux/amd64`, `linux/arm64`

### Pre-built Binaries

Download pre-built binaries from the [Releases page](https://github.com/YourCloudTools/Mockzure/releases):

```bash
# Linux (amd64)
wget https://github.com/YourCloudTools/Mockzure/releases/latest/download/mockzure-linux-amd64
chmod +x mockzure-linux-amd64
./mockzure-linux-amd64

# macOS (Apple Silicon)
wget https://github.com/YourCloudTools/Mockzure/releases/latest/download/mockzure-darwin-arm64
chmod +x mockzure-darwin-arm64
./mockzure-darwin-arm64

# Windows
# Download mockzure-windows-amd64.exe from the releases page
```

Available binaries:
- `mockzure-linux-amd64` - Linux x86_64
- `mockzure-linux-arm64` - Linux ARM64
- `mockzure-darwin-amd64` - macOS Intel
- `mockzure-darwin-arm64` - macOS Apple Silicon
- `mockzure-windows-amd64.exe` - Windows x86_64

### RPM Package (RHEL/Azure Linux/Fedora)

Mockzure is available as an RPM package through our GitHub Pages repository.

#### Add Repository

```bash
# Create repository configuration
sudo tee /etc/yum.repos.d/mockzure.repo << 'EOF'
[mockzure]
name=Mockzure RPM Repository
baseurl=https://yourcloudtools.github.io/Mockzure/rpms
enabled=1
gpgcheck=0
EOF
```

#### Install Package

```bash
# Install latest version
sudo dnf install mockzure

# Or install specific version
sudo dnf install mockzure-20241011.143055

# Or download and install directly
curl -LO https://yourcloudtools.github.io/Mockzure/rpms/mockzure-latest.rpm
sudo dnf install -y mockzure-latest.rpm
```

#### Post-Installation

```bash
# Configure the service (optional)
sudo nano /etc/mockzure/config.json

# Enable and start the service
sudo systemctl enable mockzure
sudo systemctl start mockzure

# Check status
sudo systemctl status mockzure

# View logs
sudo journalctl -u mockzure -f
```

The RPM package installs:
- Binary: `/usr/bin/mockzure`
- Configuration: `/etc/mockzure/config.json`
- Systemd service: `/etc/systemd/system/mockzure.service`
- Data directory: `/var/lib/mockzure/`

Browse available packages at: **https://yourcloudtools.github.io/Mockzure/**

### Manual Build

For development or other platforms, build from source:

```bash
cd Mockzure
go build -o mockzure main.go
./mockzure
```

## API Endpoints

### Resource Groups

```bash
# List all resource groups
GET /mock/azure/resource-groups

# Get specific resource group with its VMs
GET /mock/azure/resource-groups/{name}

# Create a new resource group
POST /mock/azure/resource-groups
```

### Virtual Machines

```bash
# List all VMs
GET /mock/azure/vms

# Get specific VM
GET /mock/azure/vms/{name}

# Get VM status
GET /mock/azure/vms/{name}/status

# Create a VM
POST /mock/azure/vms

# Start a VM
POST /mock/azure/vms/{name}/start

# Stop a VM
POST /mock/azure/vms/{name}/stop

# Restart a VM
POST /mock/azure/vms/{name}/restart

# Update VM tags
POST /mock/azure/vms/{name}/tags
PUT /mock/azure/vms/{name}/tags
```

### Users (Entra ID)

```bash
# List all users
GET /mock/azure/users

# Get specific user
GET /mock/azure/users/{id}

# Create a user
POST /mock/azure/users
```

### App Registrations

```bash
# List all app registrations
GET /mock/azure/apps

# Create an app registration
POST /mock/azure/apps
```

### Statistics

```bash
# Get system statistics
GET /mock/azure/stats
```

### Data Management

```bash
# Reset data to defaults
POST /mock/azure/data/reset

# Clear all data
POST /mock/azure/data/clear
```

### OIDC/OAuth2 Endpoints

```bash
# OIDC Discovery
GET /.well-known/openid-configuration

# Authorization endpoint (with user selection)
GET /oauth2/v2.0/authorize

# Token endpoint
POST /oauth2/v2.0/token

# User info endpoint
GET /oidc/userinfo
```

## Default Resources

### Resource Groups

- **rg-dev** - Development environment (East US)
- **rg-prod** - Production environment (West US)

### Virtual Machines

#### rg-dev
- **vm-web-01** - Web server (Running, Standard_B2s, Linux)
- **vm-api-01** - API server (Stopped, Standard_B2s, Linux)

#### rg-prod
- **vm-web-prod-01** - Production web server (Running, Standard_D2s_v3, Linux)

### Users

1. **John Doe** (john.doe@company.com)
   - Senior Developer
   - Roles: Developer, VM Owner
   - Resource Groups: rg-dev

2. **Jane Smith** (jane.smith@company.com)
   - DevOps Engineer
   - Roles: DevOps, VM Owner
   - Resource Groups: rg-dev, rg-prod

3. **Admin User** (admin@company.com)
   - System Administrator
   - Roles: Global Administrator, VM Administrator
   - Resource Groups: rg-dev, rg-prod

## Examples

### Creating a VM

```bash
curl -X POST http://localhost:8090/mock/azure/vms \
  -H "Content-Type: application/json" \
  -d '{
    "id": "/subscriptions/12345678-1234-1234-1234-123456789012/resourceGroups/rg-dev/providers/Microsoft.Compute/virtualMachines/vm-test-01",
    "name": "vm-test-01",
    "resourceGroup": "rg-dev",
    "location": "East US",
    "vmSize": "Standard_B2s",
    "osType": "linux",
    "provisioningState": "Succeeded",
    "powerState": "VM running",
    "status": "running",
    "tags": {
      "Environment": "Development",
      "Owner": "test@company.com"
    },
    "owner": "test@company.com",
    "environment": "Development"
  }'
```

### Starting a VM

```bash
curl -X POST http://localhost:8090/mock/azure/vms/vm-web-01/start
```

### Getting Resource Group VMs

```bash
curl http://localhost:8090/mock/azure/resource-groups/rg-dev
```

### Creating a User

```bash
curl -X POST http://localhost:8090/mock/azure/users \
  -H "Content-Type: application/json" \
  -d '{
    "id": "12345678-1234-1234-1234-123456789004",
    "displayName": "Test User",
    "userPrincipalName": "test.user@company.com",
    "mail": "test.user@company.com",
    "jobTitle": "Tester",
    "department": "QA",
    "accountEnabled": true,
    "roles": ["Tester"]
  }'
```

## Development Mode

To run Mockzure in development mode with auto-reload:

```bash
./bin/dev.sh
```

## Integration with Sandman

Mockzure is designed to work seamlessly with Sandman for local development and testing:

1. Configure Sandman to use Mockzure as the Azure endpoint
2. Use the OAuth2/OIDC endpoints for authentication testing
3. Test VM management operations without affecting real Azure resources
4. Simulate different user roles and permissions

## Architecture

Mockzure uses an in-memory data store with the following structure:

- **Resource Groups**: Top-level containers for resources
- **VMs**: Compute resources associated with resource groups
- **Users**: Entra ID users with roles and permissions
- **Clients**: Registered OAuth2/OIDC applications
- **Auth Codes**: Temporary authorization codes for OAuth2 flow

## UI Features

### Modern Design
- Tailwind CSS for responsive, modern styling
- Purple/blue gradient theme
- Tabbed navigation similar to Azure Portal
- Interactive tables with hover effects

### Resource Management
- Grouped VMs by resource group
- Status indicators (running/stopped)
- Quick actions (start/stop VMs)
- Real-time updates via page refresh

### User Management
- Display all user details
- Role-based access visualization
- Active/disabled status badges

## Releases and Development

### Creating a Release

Releases are automated via GitHub Actions. To create a new release:

```bash
# Create and push a version tag
git tag v1.0.0
git push origin v1.0.0
```

This automatically:
- Builds binaries for all platforms (Linux, macOS, Windows)
- Creates RPM packages
- Builds and publishes multi-platform Docker images to GitHub Container Registry
- Creates a GitHub release with all artifacts

For detailed information, see [CI/CD Guide](docs/CICD_GUIDE.md).

### Development

To contribute to Mockzure:

```bash
# Clone the repository
git clone https://github.com/YourCloudTools/Mockzure.git
cd Mockzure

# Build locally
go build -o mockzure main.go

# Run
./mockzure

# Or use the development script
./bin/dev.sh
```

### Building Docker Images Locally

```bash
# Build for local platform
docker build -t mockzure:dev --target production .

# Build for multiple platforms (requires buildx)
docker buildx build --platform linux/amd64,linux/arm64 -t mockzure:dev --target production .
```

## Notes

- Mockzure runs on port **8090** by default
- All data is stored in memory and will be lost on restart
- Use the "Reset to Defaults" button to restore initial data
- The service provides unsigned JWTs for testing purposes only
- Not suitable for production use - development and testing only

## License


Part of the Sandman project by YourCloudTools.com

