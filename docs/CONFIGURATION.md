# Mockzure Configuration Guide

This guide explains how to configure Mockzure using a `config.yaml` or `config.json` file.

## Overview

Mockzure uses a YAML or JSON configuration file to define resources and service accounts. The configuration file is user-provided; there are no defaults.

**Security Note:** `config.yaml`/`config.json` are excluded from version control via `.gitignore` to prevent accidentally committing secrets. Never commit actual credentials to the repository.

## Configuration File Location

- **Local/Manual Installation:** `--config ./config.yaml` (or `.json`)
- **Docker/Docker Compose:** Mount to `/app/config.yaml` and set `MOCKZURE_CONFIG=/app/config.yaml`
- **RPM Installation:** Place at `/etc/mockzure/config.yaml` and run with `--config /etc/mockzure/config.yaml`

## Configuration Schema

The configuration supports four top-level arrays: `resourceGroups`, `vms`, `users`, and `serviceAccounts`.

```yaml
resourceGroups:
  - id: string
    name: string
    location: string
    tags: { string: string }

vms:
  - id: string
    name: string
    resourceGroup: string
    location: string
    vmSize: string
    osType: string
    provisioningState: string
    powerState: string
    status: string
    tags: { string: string }

users:
  - id: string
    displayName: string
    userPrincipalName: string
    mail: string
    jobTitle: string
    department: string
    officeLocation: string
    userType: string
    accountEnabled: bool
    roles: [string]
    azureRoles:
      - id: string
        name: string
        description: string
        actions: [string]
        scope: string
    permissions:
      - resource: string
        actions: [string]
        resourceGroup: string
    resourceGroups: [string]
    subscriptions: [string]

serviceAccounts:
  - id: string
    applicationId: string
    secret: string
    displayName: string
    description: string
    accountEnabled: bool
    createdDateTime: string (RFC3339)
    servicePrincipal: bool
    permissions:
      - resourceGroup: string | "*"
        permissions: [read, write, start, stop, restart, delete]
    graphPermissions: [string]
```

### Service Accounts and Graph Permissions
Service accounts include `applicationId` and `secret` for authentication and may optionally include `graphPermissions` which control access to `/mock/azure/users`.

## Microsoft Graph Permissions

Graph permissions control what Microsoft Graph API operations the service account can perform. Common permissions include:

| Permission | Description | Use Case |
|------------|-------------|----------|
| `User.Read.All` | Read all users' full profiles | List users, get user details |
| `Directory.Read.All` | Read directory data | Read all directory information including users, groups |
| `User.ReadWrite.All` | Read and write all users' full profiles | Create, update, delete users |
| `Group.Read.All` | Read all groups | List groups, get group membership |

### Graph API Authentication

Service accounts with Graph permissions can access the `/mock/azure/users` endpoint:

```bash
# Get access token using client credentials
curl -X POST http://localhost:8090/oauth2/v2.0/token \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "grant_type=client_credentials" \
  -d "client_id=your-app-id" \
  -d "client_secret=your-secret" \
  -d "scope=https://graph.microsoft.com/.default"

# Use the token to access Graph API
curl http://localhost:8090/mock/azure/users \
  -H "Authorization: Bearer mock_access_token_your-app-id"
```

## Example Configurations

### Minimal Configuration (YAML)

```yaml
serviceAccounts:
  - applicationId: my-app-id
    secret: my-secret-key
```

### Full Configuration (YAML)

```yaml
resourceGroups:
  - id: "/subscriptions/000.../resourceGroups/rg-dev"
    name: rg-dev
    location: East US
vms:
  - id: "/subscriptions/000.../resourceGroups/rg-dev/providers/Microsoft.Compute/virtualMachines/vm-web-01"
    name: vm-web-01
    resourceGroup: rg-dev
    location: East US
    vmSize: Standard_B2s
    osType: linux
    provisioningState: Succeeded
    powerState: VM running
    status: running
users:
  - id: user-1
    displayName: John Doe
    userPrincipalName: john.doe@company.com
    roles: [Developer]
serviceAccounts:
  - applicationId: sandman-app-id-12345
    secret: sandman-secret-key-development-only
    displayName: Sandman Service Account
    graphPermissions: [User.Read.All]
```

### Multiple Service Accounts

You can define multiple service accounts with different permissions:

```json
{
  "serviceAccounts": [
    {
      "applicationId": "readonly-app",
      "secret": "readonly-secret",
      "displayName": "Read-Only Service Account",
      "description": "Limited access for monitoring",
      "graphPermissions": []
    },
    {
      "applicationId": "admin-app",
      "secret": "admin-secret",
      "displayName": "Admin Service Account",
      "description": "Full access for administration",
      "graphPermissions": ["User.Read.All", "Directory.Read.All"]
    },
    {
      "applicationId": "integration-app",
      "secret": "integration-secret",
      "displayName": "Integration Service Account",
      "description": "For third-party integrations",
      "graphPermissions": ["User.Read.All"]
    }
  ]
}
```

## Authentication Methods

Mockzure supports two authentication methods for service accounts:

### 1. Bearer Token (OAuth 2.0)

```bash
# Get token first
TOKEN=$(curl -X POST http://localhost:8090/oauth2/v2.0/token \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "grant_type=client_credentials" \
  -d "client_id=your-app-id" \
  -d "client_secret=your-secret" \
  | jq -r '.access_token')

# Use token in requests
curl http://localhost:8090/mock/azure/vms \
  -H "Authorization: Bearer $TOKEN"
```

### 2. Basic Authentication

```bash
# Direct authentication with credentials
curl http://localhost:8090/mock/azure/vms \
  -u "your-app-id:your-secret"
```

## Resource Permissions

In addition to Graph permissions, service accounts have permissions on Azure resources (VMs, resource groups). These are configured in the code via the `Permissions` field of service accounts and control what VM operations can be performed.

Default service accounts in Mockzure have the following resource permissions:

- **Sandman Service Account**: Can read, start, stop, and restart VMs in `rg-dev`, read-only in `rg-prod`
- **Admin Automation Service Account**: Full permissions on all resource groups (wildcard `*`)

## Security Best Practices

### File Permissions

Restrict access to `config.yaml`:

```bash
# Local installation
chmod 600 config.yaml

# RPM installation
sudo chmod 600 /etc/mockzure/config.yaml
sudo chown mockzure:mockzure /etc/mockzure/config.yaml
```

### Docker/Docker Compose

When using Docker, mount the config file as read-only and pass its path:

```bash
docker run -v $(pwd)/config.yaml:/app/config.yaml:ro -e MOCKZURE_CONFIG=/app/config.yaml ...
```

In Docker Compose, the `:ro` flag is already included in the provided `compose.yml`.

### Secret Management

- Never commit `config.yaml`/`config.json` to version control
- Use strong, randomly generated secrets
- Rotate secrets regularly
- Use different secrets for different environments
- Consider using secret management tools (HashiCorp Vault, Azure Key Vault) for production

### Development vs Production

For development:
```yaml
serviceAccounts:
  - applicationId: dev-app
    secret: dev-secret-not-for-production
    displayName: Development Account
```

For production-like testing:
```yaml
serviceAccounts:
  - applicationId: prod-like-app
    secret: strong-randomly-generated-secret-32chars
    displayName: Production-Like Account
```

## Troubleshooting

### Configuration Not Loaded

**Problem:** Mockzure errors that no config was provided.

**Solution:** Provide a config path via `--config` or set `MOCKZURE_CONFIG` and ensure the file exists.

```bash
# Verify file exists
ls -la config.yaml

# Check file permissions
chmod 600 config.json
```

### Authentication Fails

**Problem:** Requests return `401 Unauthorized` or `invalid_client`.

**Solution:** Verify credentials match exactly:

```bash
# Check if applicationId exists in config (JSON example)
cat config.json | jq '.serviceAccounts[].applicationId'

# Test authentication
curl -X POST http://localhost:8090/oauth2/v2.0/token \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "grant_type=client_credentials" \
  -d "client_id=your-app-id" \
  -d "client_secret=your-secret"
```

### Graph API Access Denied

**Problem:** Requests return `403 Forbidden` when accessing `/mock/azure/users`.

**Solution:** Add required Graph permissions to the service account:

```json
{
  "graphPermissions": ["User.Read.All"]
}
```

### Docker Volume Mount Issues

**Problem:** Config file not found in Docker container.

**Solution:** Ensure the mount path is correct and file exists on host:

```bash
# Verify config exists on host
ls -la $(pwd)/config.yaml

# Test mount
docker run --rm -v $(pwd)/config.yaml:/app/config.yaml:ro alpine cat /app/config.yaml
```

## Validation

To validate your configuration:

1. Start Mockzure and check the logs:
```bash
# Local
./mockzure --config ./config.yaml

# Docker
docker compose logs -f

# RPM
sudo journalctl -u mockzure -f
```

2. Look for a configuration load message like:
```
Config loaded: 1 RGs, 1 VMs, 1 users, 1 service accounts
```

3. Test authentication:
```bash
curl -X POST http://localhost:8090/oauth2/v2.0/token \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "grant_type=client_credentials" \
  -d "client_id=your-app-id" \
  -d "client_secret=your-secret"
```

Expected response:
```json
{
  "access_token": "mock_access_token_your-app-id",
  "token_type": "Bearer",
  "expires_in": 3600,
  "scope": "..."
}
```

## Further Reading

- [README.md](../README.md) - Main documentation
- [Azure Service Principals Documentation](https://learn.microsoft.com/en-us/azure/active-directory/develop/app-objects-and-service-principals)
- [Microsoft Graph Permissions Reference](https://learn.microsoft.com/en-us/graph/permissions-reference)

