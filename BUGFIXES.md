# Bug Fixes and Discovered Issues

This document tracks bugs discovered during initial testing and their resolutions.

## OAuth2 Client Secret Retrieval

**Issue**: OAuth2 client creation returned `null` for `client_secret` instead of the actual secret.

**Root Cause**: Kanidm's OAuth2 creation API (`POST /v1/oauth2/_basic`) returns `null` in the response body. The secret must be retrieved separately using `GET /v1/oauth2/{name}/_basic_secret`.

**Impact**:
- OAuth2 clients created successfully in Kanidm
- Terraform state had empty/null client_secret
- 1Password items created without secrets
- Users unable to use OAuth2 clients

**Resolution**:
- Added `GetOAuth2BasicSecret()` function to retrieve secrets via GET request
- Modified `CreateOAuth2BasicClient()` to call `GetOAuth2BasicSecret()` immediately after creation
- Updated `Read()` function to retrieve secrets for imported clients (when `client_secret` is null in state)

**Test Case**: Create OAuth2 client, verify `client_secret` output is populated with valid secret string.

---

## OAuth2 Client Type Detection

**Issue**: Provider incorrectly identified OAuth2 basic (confidential) clients as public clients after creation.

**Root Cause**: The `oauth2_rs_basic_secret` attribute value is "hidden" in API responses for security. Checking `GetString("oauth2_rs_basic_secret") == ""` returned true even for basic clients because the value was empty, not the attribute itself.

**Impact**:
- Import of OAuth2 basic clients failed with "Invalid Client Type" error
- Read operations after creation returned incorrect client type

**Resolution**:
- Changed detection from checking attribute value to checking attribute key presence:
  ```go
  _, hasBasicSecret := entry.Attrs["oauth2_rs_basic_secret"]
  isPublic := !hasBasicSecret
  ```

**Test Case**: Create OAuth2 basic client, import it, verify no "Invalid Client Type" error.

---

## OAuth2 Secret Retrieval Function Naming

**Issue**: Function named `RegenerateOAuth2BasicSecret()` was misleading - it actually retrieved the current secret without regenerating.

**Root Cause**: Initial implementation used GET request but named function "Regenerate" which implies POST/modification.

**Impact**: Code clarity and maintainability - developers might avoid calling it thinking it would invalidate existing secrets.

**Resolution**:
- Split into two functions:
  - `GetOAuth2BasicSecret()` - GET request, retrieves current secret (non-destructive)
  - `RegenerateOAuth2BasicSecret()` - POST request, generates new secret (destructive)
- Updated `CreateOAuth2BasicClient()` to use `GetOAuth2BasicSecret()`

**Test Case**: Call `GetOAuth2BasicSecret()` multiple times, verify secret remains consistent.

---

## Imported OAuth2 Clients Missing Secrets

**Issue**: OAuth2 clients imported via `terraform import` had null `client_secret` in state.

**Root Cause**: The `Read()` function didn't retrieve client secrets - it only read name, displayname, origin, and redirect URIs.

**Impact**:
- Imported clients couldn't be used (no secret available)
- 1Password items for imported clients created with null secrets
- Users had to manually retrieve secrets via Kanidm CLI

**Resolution**:
- Updated `Read()` function to retrieve secret if not already in state:
  ```go
  if state.ClientSecret.IsNull() || state.ClientSecret.ValueString() == "" {
      secret, err := r.client.GetOAuth2BasicSecret(ctx, state.Name.ValueString())
      if err == nil {
          state.ClientSecret = types.StringValue(secret)
      }
  }
  ```

**Test Case**:
1. Create OAuth2 client via Kanidm CLI
2. Import into Terraform: `terraform import kanidm_oauth2_basic.test test-client`
3. Verify `client_secret` is populated in state
4. Verify `terraform output -raw test_client_secret` returns valid secret

---

## Group Members Ordering Issues

**Issue**: Terraform detected spurious changes in group members due to ordering differences.

**Root Cause**:
- Group `members` was defined as `ListAttribute` (ordered collection)
- Kanidm API returns members in arbitrary order
- Terraform compared ordered lists and detected drift even when membership was unchanged

**Impact**:
- `terraform plan` always showed changes even when no actual changes occurred
- Example: `["alice", "bob"]` vs `["bob", "alice"]` detected as different
- Unnecessary updates triggered on every apply

**Resolution**:
- Changed `members` from `ListAttribute` to `SetAttribute` (unordered collection)
- Updated all CRUD methods to use `types.SetValueFrom()` instead of `types.ListValueFrom()`
- Removed all sorting logic (Sets handle ordering internally)

**Test Case**:
1. Create group with members `["alice", "bob"]`
2. Run `terraform plan` twice
3. Verify no changes detected on second plan

---

## Group Empty Members Null vs Empty List

**Issue**: Groups with no members returned `null` instead of empty list, causing Terraform drift.

**Root Cause**: In `internal/client/group.go`, `GetStringSlice()` returned `nil` when attribute was missing instead of empty slice `[]string{}`.

**Impact**:
- Groups without members showed as having null members
- Terraform detected drift between `[]` in config and `null` in state

**Resolution**:
- Fixed `GetStringSlice()` to return empty slice when attribute missing:
  ```go
  members := entry.GetStringSlice("member")
  if members == nil {
      members = []string{}
  }
  ```
- Updated all CRUD methods in `group_resource.go` to consistently use `SetValueFrom()`

**Test Case**: Create empty group, verify `members = []` in state, not `null`.

---

## Service Account DisplayName Not Supported

**Issue**: Service accounts created with `displayname` field, but Kanidm rejected it as unsupported.

**Root Cause**: Kanidm's service account API doesn't support `displayname` attribute - only person accounts have this field.

**Impact**:
- Service account creation failed when `displayname` was provided
- Terraform showed persistent drift for imported service accounts

**Resolution**:
- Removed `displayname` from service account schema and model
- Removed `displayname` parameter from `CreateServiceAccount()`
- Updated all Terraform service account resources to remove displayname attributes

**Test Case**: Create service account without displayname, verify no errors.

---

## OAuth2 Name Collision with Service Accounts

**Issue**: Creating OAuth2 client with same name as existing service account failed with `attributeuniqueness` error.

**Root Cause**: Kanidm uses shared namespace for all account names (persons, service accounts, OAuth2 clients).

**Impact**:
- OAuth2 client `argocd` couldn't be created because service account `argocd` already existed
- Error: `{"attributeuniqueness":["name","spn"]}`

**Resolution**:
- Renamed OAuth2 client from `argocd` to `argocd-oidc` to avoid collision
- Documentation should warn about shared namespace

**Best Practice**: Use suffixes like `-oidc`, `-oauth2` for OAuth2 clients to avoid collisions with service accounts.

**Test Case**: Create service account `test`, then OAuth2 client `test-oidc`, verify both succeed.

---

## Group Members FQN Format

**Issue**: Group members showed drift because config used short names but Kanidm returned fully-qualified names (FQN).

**Root Cause**: Kanidm normalizes all member references to FQN format: `{name}@{domain}`.

**Impact**:
- Config: `members = ["argocd"]`
- State: `members = ["argocd@idm.s8i.ca"]`
- Terraform detected constant drift

**Resolution**:
- Updated all group member references in Terraform configs to use FQN format:
  ```hcl
  members = [
    "${kanidm_service_account.argocd.id}@idm.s8i.ca",
  ]
  ```

**Test Case**: Create group with FQN members, verify no drift on subsequent plans.

---

## Testing Checklist

Based on discovered issues, the following test scenarios should be validated:

### OAuth2 Clients
- [ ] Create OAuth2 basic client and verify `client_secret` is populated
- [ ] Import existing OAuth2 client and verify `client_secret` is retrieved
- [ ] Call `GetOAuth2BasicSecret()` multiple times and verify consistency
- [ ] Verify OAuth2 client type detection (basic vs public)
- [ ] Test OAuth2 client with same name as service account (should fail gracefully)
- [ ] Create OAuth2 client, store secret in 1Password, verify secret is correct

### Groups
- [ ] Create group with members and verify Set behavior (no ordering drift)
- [ ] Create empty group and verify `members = []` not `null`
- [ ] Add/remove members and verify updates
- [ ] Use FQN format for members and verify no drift
- [ ] Import existing group and verify membership

### Service Accounts
- [ ] Create service account without displayname
- [ ] Verify API token generation works
- [ ] Import service account (note: tokens only available on creation)

### General
- [ ] Test all resource imports
- [ ] Verify sensitive values are marked correctly
- [ ] Test provider with 1Password integration
- [ ] Verify state file doesn't contain unencrypted secrets
