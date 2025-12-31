# Terraform Provider for Kanidm

This provider enables Infrastructure-as-Code management of [Kanidm](https://kanidm.com/) identity management resources using Terraform.

## Status

🚧 **Under Development** - This provider is currently in early development.

## Features (Planned)

- **Person Accounts**: Manage user accounts with credentials
- **Service Accounts**: Create service accounts with API tokens
- **Groups**: Manage groups and memberships
- **OAuth2 Clients**: Configure OIDC relying parties

## Requirements

- Terraform >= 1.0
- Go >= 1.24 (for development)
- Kanidm >= 1.1.0

## Development

### Setup

This project uses [devbox](https://www.jetpack.io/devbox/) for development environment management.

```bash
# Enter devbox shell (installs all dependencies)
devbox shell

# Initialize Go dependencies
go mod tidy
```

### Building

```bash
# Build the provider
devbox run build

# Or use make
make build
```

### Testing

```bash
# Run unit tests
devbox run test

# Run acceptance tests (requires KANIDM_URL and KANIDM_TOKEN)
devbox run testacc
```

### Code Generation

This provider uses HashiCorp's official code generation tools to scaffold from Kanidm's OpenAPI schema.

```bash
# Regenerate provider code from OpenAPI schema
devbox run generate
```

## Project Structure

```
terraform-provider-kanidm/
├── internal/
│   ├── provider/          # Terraform provider implementation
│   ├── client/            # Kanidm API client
│   └── spec/              # OpenAPI schema and code generation config
├── examples/              # Usage examples
├── docs/                  # Auto-generated documentation
└── devbox.json           # Development environment configuration
```

## License

Mozilla Public License 2.0 (MPL-2.0)

## Acknowledgments

Built using:
- [HashiCorp Terraform Plugin Framework](https://github.com/hashicorp/terraform-plugin-framework)
- [HashiCorp Terraform Plugin Code Generation](https://github.com/hashicorp/terraform-plugin-codegen-openapi)
- [Kanidm](https://kanidm.com/)
