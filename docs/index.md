---
page_title: "Kanidm Provider"
subcategory: ""
description: |-
  Terraform provider for managing Kanidm identity and access resources
---

# Kanidm Provider

The Kanidm provider enables Infrastructure-as-Code management of [Kanidm](https://kanidm.com) identity and access resources.

## Features

- **Person Accounts** - Manage user accounts with password or passkey authentication
- **Service Accounts** - Automated systems with API token generation
- **Groups** - Organize users and service accounts with membership management
- **OAuth2 Clients** - Configure OAuth2/OIDC integration with scope mapping

## Example Usage

```terraform
terraform {
  required_providers {
    kanidm = {
      source = "ssoriche/kanidm"
      version = "~> 0.1"
    }
  }
}

provider "kanidm" {
  url   = "https://idm.example.com"
  token = var.kanidm_token
}

# Create a person account with passkey authentication
resource "kanidm_person" "alice" {
  id                              = "alice"
  displayname                     = "Alice Smith"
  mail                            = ["alice@example.com"]
  generate_credential_reset_token = true
}

# Create a service account
resource "kanidm_service_account" "terraform" {
  id          = "terraform-automation"
  displayname = "Terraform Automation Account"
}

# Create a group with members
resource "kanidm_group" "developers" {
  id          = "developers"
  description = "Development team members"

  members = [
    kanidm_person.alice.id,
    kanidm_service_account.terraform.id,
  ]
}

# Create an OAuth2 client
resource "kanidm_oauth2_basic" "grafana" {
  name        = "grafana"
  displayname = "Grafana"
  origin      = "https://grafana.example.com"

  redirect_uris = [
    "https://grafana.example.com/login/generic_oauth"
  ]

  scope_map {
    group  = "developers"
    scopes = ["openid", "profile", "email"]
  }
}
```

## Authentication

The provider requires a service account API token for authentication with Kanidm.

### Creating a Service Account Token

```bash
# Create a service account
kanidm service-account create terraform "Terraform Automation" \
  --name idm_admin \
  -H https://idm.example.com

# Generate an API token
kanidm service-account api-token generate terraform \
  --name idm_admin \
  -H https://idm.example.com
```

Store the API token securely (e.g., in 1Password, HashiCorp Vault, or environment variables).

## Schema

### Required

- `url` (String) - Kanidm server URL (e.g., `https://idm.example.com`)
- `token` (String, Sensitive) - Service account API token for authentication

### Optional

None

## Environment Variables

- `KANIDM_URL` - Alternative to provider `url` argument
- `KANIDM_TOKEN` - Alternative to provider `token` argument

## Using with 1Password Provider

```terraform
data "onepassword_item" "kanidm_admin_token" {
  vault = "Infrastructure"
  title = "Kanidm Admin Token"
}

provider "kanidm" {
  url   = "https://idm.example.com"
  token = data.onepassword_item.kanidm_admin_token.credential
}
```
