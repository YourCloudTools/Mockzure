# Mockzure - Mock Azure Service

A lightweight mock Azure service for development and testing. Mockzure provides a simulated Azure environment with resource management, user authentication, and OIDC/OAuth2 support.

📚 **[View Documentation](https://yourcloudtools.github.io/Mockzure/)** | 🐳 **[Docker Image](https://github.com/YourCloudTools/Mockzure/pkgs/container/mockzure)** | 📦 **[RPM Repository](https://yourcloudtools.github.io/Mockzure/rpms/)**

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

### Prerequisites

- Docker and Docker Compose installed (recommended), OR
- Go 1.25+ for building from source

### Configuration

Before running Mockzure, create your configuration file:

```bash
# Copy the example configuration
cp config.json.example config.json

# Edit the configuration with your service account credentials
nano config.json  # or use your preferred editor
```

The `config.json` file defines service account credentials for authentication. See [Configuration Guide](docs/CONFIGURATION.md) for detailed documentation.

**Note:** `config.json` is excluded from version control for security. Never commit actual secrets to the repository.

### Run with Docker Compose (Recommended)

The easiest way to run Mockzure:

```bash
# 1. Ensure config.json exists (see Configuration above)
cp config.json.example config.json

# 2. Start Mockzure
docker compose up -d

# 3. View logs
docker compose logs -f

# 4. Stop Mockzure
docker compose down
```

Access the web portal at: http://localhost:8090

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

### Docker Compose (Recommended)

Docker Compose is the recommended way to run Mockzure. It handles configuration mounting and container lifecycle automatically.

**Prerequisites:**
- Docker Engine 20.10+ and Docker Compose V2
- `config.json` file in the project directory

**Steps:**

```bash
# 1. Clone the repository (if not already done)
git clone https://github.com/YourCloudTools/Mockzure.git
cd Mockzure

# 2. Create configuration
cp config.json.example config.json
# Edit config.json with your credentials

# 3. Start Mockzure
docker compose up -d

# 4. Check status
docker compose ps

# 5. View logs
docker compose logs -f mockzure

# 6. Access the portal
# Open http://localhost:8090 in your browser
```

**Managing the Service:**

```bash
# Stop Mockzure
docker compose stop

# Start Mockzure
docker compose start

# Restart Mockzure (e.g., after config changes)
docker compose restart

# Stop and remove container
docker compose down

# View real-time logs
docker compose logs -f

# Rebuild and restart (if using local build)
docker compose up -d --build
```

**Using Local Build:**

To build from the local Dockerfile instead of pulling from GitHub Container Registry, edit `compose.yml`:

```yaml
services:
  mockzure:
    # Comment out the image line
    # image: ghcr.io/yourcloudtools/mockzure:latest
    # Uncomment the build section
    build:
      context: .
      target: production
```

Then run:
```bash
docker compose up -d --build
```

### Docker Run

If you prefer using `docker run` directly instead of Docker Compose:

**With Config File (Recommended):**

```bash
# Pull the image
docker pull ghcr.io/yourcloudtools/mockzure:latest

# Run with config.json mounted from current directory
docker run -d \
  --name mockzure \
  -p 8090:8090 \
  -v $(pwd)/config.json:/app/config.json:ro \
  --restart unless-stopped \
  ghcr.io/yourcloudtools/mockzure:latest

# View logs
docker logs -f mockzure

# Stop container
docker stop mockzure

# Remove container
docker rm mockzure
```

**With Custom Config Path:**

```bash
# If config.json is in a different location
docker run -d \
  --name mockzure \
  -p 8090:8090 \
  -v /path/to/your/config.json:/app/config.json:ro \
  --restart unless-stopped \
  ghcr.io/yourcloudtools/mockzure:latest
```

**Without Config File (Uses Defaults):**

```bash
# Run without mounting config (will create default config)
docker run -d \
  --name mockzure \
  -p 8090:8090 \
  --restart unless-stopped \
  ghcr.io/yourcloudtools/mockzure:latest
```

**Run Specific Version:**

```bash
# Run a specific tagged version
docker run -d \
  --name mockzure \
  -p 8090:8090 \
  -v $(pwd)/config.json:/app/config.json:ro \
  ghcr.io/yourcloudtools/mockzure:v1.0.0
```

**Notes:**
- The `:ro` flag mounts the config file as read-only for security
- `--restart unless-stopped` automatically restarts the container after system reboot
- Access the portal at: http://localhost:8090
- Supported platforms: `linux/amd64`, `linux/arm64`

### Docker Image Information

Pre-built multi-platform images are available on GitHub Container Registry:

- **Latest:** `ghcr.io/yourcloudtools/mockzure:latest`
- **Specific Version:** `ghcr.io/yourcloudtools/mockzure:v1.0.0`
- **Platforms:** `linux/amd64`, `linux/arm64`

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

## Configuration

For detailed information about configuring Mockzure, see the [Configuration Guide](docs/CONFIGURATION.md).

For Azure API compatibility information, see the [Azure API Compatibility Report](docs/AZURE_API_COMPATIBILITY.md).

Quick configuration reference:
- **Config File:** `config.json` (required for service account authentication)
- **Schema:** Service accounts with applicationId, secret, and optional Graph permissions
- **Location:** Project root (local), `/app/config.json` (Docker), `/etc/mockzure/config.json` (RPM)
- **Security:** File is excluded from git, use read-only mounts in Docker

Example minimal configuration:
```json
{
  "serviceAccounts": [
    {
      "applicationId": "your-app-id",
      "secret": "your-secret-key",
      "graphPermissions": ["User.Read.All"]
    }
  ]
}
```

## Development Mode

To run Mockzure in development mode with auto-reload:

```bash
./bin/dev.sh
```

### Development with Docker Compose

```bash
# Use the development target with live code mounting
docker compose -f compose.dev.yml up

# Or edit compose.yml to use the development stage
# Change target: production to target: development
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

# Run your local build
docker run -d -p 8090:8090 -v $(pwd)/config.json:/app/config.json:ro mockzure:dev

# Build for multiple platforms (requires buildx)
docker buildx build --platform linux/amd64,linux/arm64 -t mockzure:dev --target production .

# Build development image with hot reload
docker build -t mockzure:dev --target development .
```

## Notes

- Mockzure runs on port **8090** by default
- All data is stored in memory and will be lost on restart
- Use the "Reset to Defaults" button to restore initial data
- The service provides unsigned JWTs for testing purposes only
- Not suitable for production use - development and testing only

## License


Part of the Sandman project by YourCloudTools.com

