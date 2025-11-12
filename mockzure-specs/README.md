# Mockzure Specs (fetched)

This folder contains the minimal specs required by Mockzure MVP.

## Structure
- identity/
  - oidc-configuration.json (live Microsoft discovery mirror)
  - oidc-jwks.json (live Microsoft JWKS mirror)
  - oidc-openapi.yaml (minimal OIDC OAS for Prism)
- graph/
  - graph-users.yaml (pruned OpenAPI: only /users + components)
- arm/
  - arm-compute.json (Microsoft.Compute stable 2024-07-01)
  - arm-resources.json (Microsoft.Resources stable 2021-04-01)
  - arm-authorization.json (Microsoft.Authorization stable 2022-04-01)

## Suggested Prism run (see ../docker-compose.yaml)
- ARM on :4010
- Graph on :4011
- OIDC on :4012
