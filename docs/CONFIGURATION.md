# Mockzure Configuration Guide

This guide explains how to configure Mockzure using the `config.json` file.

## Overview

Mockzure uses a JSON configuration file to define service account credentials and permissions. The configuration file is located at the root of the project and is named `config.json`.

**Security Note:** `config.json` is excluded from version control via `.gitignore` to prevent accidentally committing secrets. Never commit actual credentials to the repository.

## Configuration File Location

- **Local/Manual Installation:** `./config.json` (in the Mockzure directory)
- **Docker/Docker Compose:** Mounted to `/app/config.json` inside the container
- **RPM Installation:** `/etc/mockzure/config.json`

## Configuration Schema

The configuration file has the following structure:

```json
{
  "serviceAccounts": [
    {
      "applicationId": "string",
      "secret": "string",
      "displayName": "string (optional)",
      "description": "string (optional)",
      "graphPermissions": ["string array (optional)"]
    }
  ]
}
```

### Field Descriptions

#### `serviceAccounts` (array, required)
Array of service account (service principal) configurations. Each service account represents an Azure Service Principal that can authenticate to Mockzure.

#### Service Account Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `applicationId` | string | Yes | The Client ID / Application ID of the service principal. This is used for authentication. |
| `secret` | string | Yes | The client secret for the service principal. Used to authenticate requests. |
| `displayName` | string | No | Human-readable name for the service account. Displayed in the portal. |
| `description` | string | No | Description of the service account's purpose. |
| `graphPermissions` | array | No | Array of Microsoft Graph API permissions granted to this service account. |

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

### Minimal Configuration

```json
{
  "serviceAccounts": [
    {
      "applicationId": "my-app-id",
      "secret": "my-secret-key"
    }
  ]
}
```

### Full Configuration

```json
{
  "serviceAccounts": [
    {
      "applicationId": "sandman-app-id-12345",
      "secret": "sandman-secret-key-development-only",
      "displayName": "Sandman Service Account",
      "description": "Service account for Sandman to manage VMs",
      "graphPermissions": ["User.Read.All"]
    },
    {
      "applicationId": "admin-automation-app-id",
      "secret": "admin-secret-key-development-only",
      "displayName": "Admin Automation Service Account",
      "description": "Service account for administrative automation",
      "graphPermissions": ["User.Read.All", "Directory.Read.All"]
    }
  ]
}
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

Restrict access to `config.json`:

```bash
# Local installation
chmod 600 config.json

# RPM installation
sudo chmod 600 /etc/mockzure/config.json
sudo chown mockzure:mockzure /etc/mockzure/config.json
```

### Docker/Docker Compose

When using Docker, mount the config file as read-only:

```bash
docker run -v $(pwd)/config.json:/app/config.json:ro ...
```

In Docker Compose, the `:ro` flag is already included in the provided `compose.yml`.

### Secret Management

- Never commit `config.json` to version control
- Use strong, randomly generated secrets
- Rotate secrets regularly
- Use different secrets for different environments
- Consider using secret management tools (HashiCorp Vault, Azure Key Vault) for production

### Development vs Production

For development:
```json
{
  "serviceAccounts": [
    {
      "applicationId": "dev-app",
      "secret": "dev-secret-not-for-production",
      "displayName": "Development Account"
    }
  ]
}
```

For production-like testing:
```json
{
  "serviceAccounts": [
    {
      "applicationId": "prod-like-app",
      "secret": "strong-randomly-generated-secret-32chars",
      "displayName": "Production-Like Account"
    }
  ]
}
```

## Troubleshooting

### Configuration Not Loaded

**Problem:** Mockzure creates a default `config.json` on startup.

**Solution:** Ensure your `config.json` exists before starting Mockzure.

```bash
# Verify file exists
ls -la config.json

# Check file permissions
chmod 600 config.json
```

### Authentication Fails

**Problem:** Requests return `401 Unauthorized` or `invalid_client`.

**Solution:** Verify credentials match exactly:

```bash
# Check if applicationId exists in config
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
ls -la $(pwd)/config.json

# Test mount
docker run --rm -v $(pwd)/config.json:/app/config.json:ro alpine cat /app/config.json
```

## Validation

To validate your configuration:

1. Start Mockzure and check the logs:
```bash
# Local
./mockzure

# Docker
docker compose logs -f

# RPM
sudo journalctl -u mockzure -f
```

2. Look for the configuration load message:
```
Loaded X service account secrets from config
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

