# Azure API Compatibility Report

**Generated:** 2025-10-20 14:38:41 UTC
**Mockzure Version:** v1.0.0

## Executive Summary

Mockzure provides comprehensive compatibility with Azure APIs for development and testing scenarios. Overall compatibility score: **84.2%** (16/19 endpoints fully supported).

## Compatibility Matrix

| API Category | Support Level | Coverage | Notes |
|--------------|--------------|----------|-------|
| Microsoft Identity Platform (OIDC) | ✅ Full | 6/6 | Full OIDC/OAuth2 support with user selection UI. Uses unsigned JWTs for testing. |
| Microsoft Graph API | ⚠️ Partial | 3/3 | Requires proper Graph permissions (User.Read.All). Supports service account authentication. |
| Azure Resource Manager (ARM) | ✅ Full | 6/6 | Full ARM compatibility with proper RBAC enforcement. Supports both ARM and simplified endpoints. |
| RBAC & Authorization | ✅ Full | 4/4 | Comprehensive RBAC with service account permissions, Graph API scopes, and resource-level access control. |

## Detailed Breakdown

### ✅ Microsoft Identity Platform (OIDC)

**Support Level:** Fully Supported  
**Coverage:** 6/6  
**Notes:** Full OIDC/OAuth2 support with user selection UI. Uses unsigned JWTs for testing.

| Endpoint | Method | Description | Status |
|----------|--------|-------------|--------|
| `/.well-known/openid-configuration` | GET | OIDC Discovery | ✅ |
| `/common/v2.0/.well-known/openid-configuration` | GET | Tenant-specific Discovery | ✅ |
| `/oauth2/v2.0/authorize` | GET | Authorization Code Flow | ✅ |
| `/oauth2/v2.0/token` | POST | Token Issuance (Client Credentials) | ✅ |
| `/oauth2/v2.0/token` | POST | Token Issuance (Authorization Code) | ✅ |
| `/oidc/userinfo` | GET | User Information | ✅ |

### ⚠️ Microsoft Graph API

**Support Level:** Partially Supported  
**Coverage:** 3/3  
**Notes:** Requires proper Graph permissions (User.Read.All). Supports service account authentication.

| Endpoint | Method | Description | Status |
|----------|--------|-------------|--------|
| `/mock/azure/users` | GET | List Users | ⚠️ |
| `/mock/azure/users/{id}` | GET | Get User | ⚠️ |
| `/mock/azure/users` | POST | Create User | ⚠️ |

### ✅ Azure Resource Manager (ARM)

**Support Level:** Fully Supported  
**Coverage:** 6/6  
**Notes:** Full ARM compatibility with proper RBAC enforcement. Supports both ARM and simplified endpoints.

| Endpoint | Method | Description | Status |
|----------|--------|-------------|--------|
| `/subscriptions/{subId}/resourceGroups/{rg}/providers/Microsoft.Compute/virtualMachines` | GET | List VMs (ARM format) | ✅ |
| `/subscriptions/{subId}/resourceGroups/{rg}/providers/Microsoft.Compute/virtualMachines/{vmName}` | GET | Get VM (ARM format) | ✅ |
| `/subscriptions/{subId}/resourceGroups/{rg}/providers/Microsoft.Compute/virtualMachines/{vmName}?$expand=instanceView` | GET | Get VM with Instance View | ✅ |
| `/mock/azure/vms/{vmName}/start` | POST | Start VM | ✅ |
| `/mock/azure/vms/{vmName}/stop` | POST | Stop VM | ✅ |
| `/mock/azure/vms/{vmName}/restart` | POST | Restart VM | ✅ |

### ✅ RBAC & Authorization

**Support Level:** Fully Supported  
**Coverage:** 4/4  
**Notes:** Comprehensive RBAC with service account permissions, Graph API scopes, and resource-level access control.

| Endpoint | Method | Description | Status |
|----------|--------|-------------|--------|
| `Service Account Authentication` | BASIC | Basic Auth for Service Principals | ✅ |
| `Graph API Permissions` | CHECK | User.Read.All permission enforcement | ✅ |
| `Resource Group Permissions` | CHECK | Read/Write/Start/Stop permissions | ✅ |
| `Role-based Access Control` | CHECK | Scope-based authorization | ✅ |

## Known Limitations

The following limitations apply to Mockzure compared to real Azure APIs:

- Uses unsigned JWTs for testing (not suitable for production)
- Simplified user roles compared to real Azure RBAC
- Limited Graph API scope (only User.Read.All implemented)
- No long-running operations (LRO) support
- No Azure CLI integration
- Mock data only - no persistence to real Azure

## Testing Instructions

To run the compatibility tests:

```bash
# Run all compatibility tests
go test -v -run "TestMicrosoft|TestAzure|TestRBAC" ./...

# Generate this report
go run generate_compatibility_report.go
```

## Report Generation

This report is automatically generated from compatibility tests. To update this report, run `go run generate_compatibility_report.go` from the project root.
