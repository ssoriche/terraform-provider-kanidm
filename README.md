# Terraform Provider for Kanidm

The official Terraform provider for [Kanidm](https://kanidm.com), enabling Infrastructure-as-Code management of identity and access resources.

[![License: MPL 2.0](https://img.shields.io/badge/License-MPL%202.0-brightgreen.svg)](https://opensource.org/licenses/MPL-2.0)
[![Go Report Card](https://goreportcard.com/badge/github.com/ssoriche/terraform-provider-kanidm)](https://goreportcard.com/report/github.com/ssoriche/terraform-provider-kanidm)

## Features

- **Person Accounts** - Manage user accounts with password or passkey authentication
- **Service Accounts** - Automated systems with API token generation
- **Groups** - Organize users and service accounts with membership management
- **OAuth2 Clients** - Configure OAuth2/OIDC integration with scope mapping

## Requirements

- [Terraform](https://www.terraform.io/downloads.html) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.24 (for development)
- [Kanidm](https://kanidm.com) >= 1.8.5

## Installation

### Terraform Registry (Recommended)

```hcl
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
  token = var.kanidm_token  # Service account API token
}
```

### Local Development

```bash
git clone https://github.com/ssoriche/terraform-provider-kanidm
cd terraform-provider-kanidm
go build -o terraform-provider-kanidm
```

## Usage Examples

### Person Account with Passkey Authentication (Recommended)

```hcl
resource "kanidm_person" "alice" {
  id                              = "alice"
  displayname                     = "Alice Smith"
  mail                            = ["alice@example.com"]
  generate_credential_reset_token = true
  credential_reset_token_ttl      = 7200  # 2 hours
}

# Output the credential reset token for user setup
output "alice_reset_token" {
  value     = kanidm_person.alice.credential_reset_token
  sensitive = true
}
```

### Service Account with API Token

```hcl
resource "kanidm_service_account" "terraform" {
  id          = "terraform-automation"
  displayname = "Terraform Automation Account"
}

output "terraform_api_token" {
  value     = kanidm_service_account.terraform.api_token
  sensitive = true
}
```

### Group with Members

```hcl
resource "kanidm_group" "developers" {
  id          = "developers"
  description = "Development team members"

  members = [
    kanidm_person.alice.id,
    kanidm_person.bob.id,
    kanidm_service_account.ci.id,
  ]
}
```

### OAuth2 Client for Application Integration

```hcl
resource "kanidm_oauth2_basic" "grafana" {
  name        = "grafana"
  displayname = "Grafana"
  origin      = "https://grafana.example.com"

  redirect_uris = [
    "https://grafana.example.com/login/generic_oauth"
  ]

  scope_map {
    group  = "admins"
    scopes = ["openid", "profile", "email", "groups"]
  }

  scope_map {
    group  = "developers"
    scopes = ["openid", "profile", "email"]
  ]
}

output "grafana_client_secret" {
  value     = kanidm_oauth2_basic.grafana.client_secret
  sensitive = true
}
```

## Provider Configuration

### Arguments

- `url` - (Required) Kanidm server URL (e.g., `https://idm.example.com`)
- `token` - (Required) Service account API token for authentication

### Environment Variables

You can also configure the provider using environment variables:

```bash
export KANIDM_URL="https://idm.example.com"
export KANIDM_TOKEN="your-api-token"
```

### Using with 1Password Provider

```hcl
terraform {
  required_providers {
    kanidm = {
      source = "ssoriche/kanidm"
    }
    onepassword = {
      source  = "1Password/onepassword"
      version = "~> 2.1"
    }
  }
}

data "onepassword_item" "kanidm_admin_token" {
  vault = "Infrastructure"
  title = "Kanidm Admin Token"
}

provider "kanidm" {
  url   = "https://idm.example.com"
  token = data.onepassword_item.kanidm_admin_token.credential
}
```

## Resources

- `kanidm_person` - Person accounts with credential management
- `kanidm_service_account` - Service accounts with API tokens
- `kanidm_group` - Groups with membership management
- `kanidm_oauth2_basic` - OAuth2 basic (confidential) clients

## Development

### Building

```bash
go build -o terraform-provider-kanidm
```

### Testing

```bash
# Unit tests
go test -v ./...

# Acceptance tests (requires running Kanidm instance)
TF_ACC=1 go test -v -timeout 30m ./internal/provider/
```

### Code Quality

```bash
# Format code
go fmt ./...

# Lint
golangci-lint run ./...
```

## Contributing

Contributions are welcome! Please follow these guidelines:

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes using conventional commits
4. Ensure all tests pass and code is formatted
5. Push to your branch (`git push origin feature/amazing-feature`)
6. Open a Pull Request

### Commit Message Format

This project follows [Conventional Commits](https://www.conventionalcommits.org/):

```
feat(resource): add new OAuth2 public client resource
fix(client): handle 404 errors in person resource
docs(readme): update usage examples
chore(deps): update terraform-plugin-framework to v1.12.0
```

## Architecture

This provider is built using:

- [Terraform Plugin Framework](https://github.com/hashicorp/terraform-plugin-framework) - Modern provider SDK
- [Kanidm REST API](https://kanidm.github.io/kanidm/stable/) - Identity management backend
- Go 1.24 - Implementation language

### Project Structure

```
terraform-provider-kanidm/
├── internal/
│   ├── client/          # Kanidm API client
│   │   ├── client.go    # HTTP client and auth
│   │   ├── person.go    # Person account operations
│   │   ├── service_account.go
│   │   ├── group.go
│   │   └── oauth2.go
│   └── provider/        # Terraform resources
│       ├── provider.go  # Provider configuration
│       ├── person_resource.go
│       ├── service_account_resource.go
│       ├── group_resource.go
│       └── oauth2_basic_resource.go
├── examples/            # Usage examples
│   ├── provider/
│   └── resources/
├── devbox.json          # Development environment
└── main.go              # Provider entry point
```

## Roadmap

- [x] Person resource with password and passkey authentication
- [x] Service account resource with API tokens
- [x] Group resource with membership management
- [x] OAuth2 basic (confidential) client resource
- [ ] OAuth2 public client resource
- [ ] Data sources for reading existing resources
- [ ] Unit and acceptance test suite
- [ ] Terraform Registry publication
- [ ] Community contribution to Kanidm project

## License

This project is licensed under the Mozilla Public License 2.0 - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- [Kanidm Project](https://kanidm.com) - Identity management platform
- [HashiCorp Terraform](https://www.terraform.io/) - Infrastructure as Code

## Support

- GitHub Issues: [https://github.com/ssoriche/terraform-provider-kanidm/issues](https://github.com/ssoriche/terraform-provider-kanidm/issues)
- Kanidm Documentation: [https://kanidm.github.io/kanidm/stable/](https://kanidm.github.io/kanidm/stable/)

---

**Note:** This provider is currently in active development. APIs may change between releases until v1.0.0.
