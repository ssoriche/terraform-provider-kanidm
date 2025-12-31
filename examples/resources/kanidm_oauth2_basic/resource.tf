# Example: OAuth2 client for Grafana
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
  }
}

# Store the client secret securely
output "grafana_client_secret" {
  description = "OAuth2 client secret for Grafana"
  value       = kanidm_oauth2_basic.grafana.client_secret
  sensitive   = true
}

# Example: OAuth2 client for Authentik
resource "kanidm_oauth2_basic" "authentik" {
  name        = "authentik"
  displayname = "Authentik SSO"
  origin      = "https://auth.example.com"

  redirect_uris = [
    "https://auth.example.com/source/oauth/callback/kanidm/"
  ]

  scope_map {
    group  = "all-users"
    scopes = ["openid", "profile", "email"]
  }
}

# Example: OAuth2 client for GitLab
resource "kanidm_oauth2_basic" "gitlab" {
  name        = "gitlab"
  displayname = "GitLab"
  origin      = "https://gitlab.example.com"

  redirect_uris = [
    "https://gitlab.example.com/users/auth/openid_connect/callback"
  ]

  scope_map {
    group  = "developers"
    scopes = ["openid", "profile", "email", "groups"]
  }

  scope_map {
    group  = "project-managers"
    scopes = ["openid", "profile", "email"]
  }
}

# Example: Simple OAuth2 client without scope maps
resource "kanidm_oauth2_basic" "simple_app" {
  name        = "simple-app"
  displayname = "Simple Application"
  origin      = "https://app.example.com"

  redirect_uris = [
    "https://app.example.com/callback"
  ]
}

# Example: Imported existing OAuth2 client
# Import command: terraform import kanidm_oauth2_basic.existing client_name
# Note: Client secret will not be available after import
resource "kanidm_oauth2_basic" "existing" {
  name        = "existing-client"
  displayname = "Existing OAuth2 Client"
  origin      = "https://existing.example.com"
}
