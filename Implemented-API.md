

# Mockzure – Implemented API (Summary)

This document lists all services and API areas Mockzure must implement for its MVP.

---

## 1. Microsoft Identity Platform (OIDC)
**Purpose:** Authentication, token issuance, and user sign-in (SSO).

**Endpoints:**
- /.well-known/openid-configuration
- /oauth2/v2.0/authorize
- /oauth2/v2.0/token
- /discovery/v2.0/keys

**Objects:**
- application
- servicePrincipal
- user
- appRoleAssignment

**Actions:**
- Issue ID/access tokens (OIDC)
- Validate scopes and audiences
- Enforce user assignment if enabled

---

## 2. Microsoft Graph API
**Purpose:** Directory access for listing users.

**Base:** https://graph.mockzure.local/v1.0

**Resources:**
- /users

**Objects:**
- user

**Actions:**
- GET /users
- Support basic filtering and pagination
- Enforce Graph scopes (User.Read.All)

---

## 3. Azure Resource Manager (ARM)
**Purpose:** Virtual Machine management and status simulation.

**Base:** https://management.mockzure.local

**Resources:**
- /subscriptions/{subId}/resourceGroups/{rg}/providers/Microsoft.Compute/virtualMachines
- /operations

**Objects:**
- virtualMachine
- operation (for LRO)
- roleAssignment (RBAC)

**Actions:**
- GET: list/get VM and instance view
- POST: start/deallocate/restart VM
- GET: check operation status
- Enforce role-based authorization

---

## 4. RBAC & Directory Data
**Purpose:** Authorization and scope-based access control.

**Objects:**
- roleAssignment
- roleDefinition
- tenant
- user/group mapping

**Actions:**
- Resolve principal roles at subscription/RG/resource scope
- Authorize based on token roles

---

## Out of Scope (Future)
- Teams Adaptive Cards / Bot Framework API
- Other Graph or Azure resource types