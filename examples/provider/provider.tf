terraform {
  required_providers {
    kanidm = {
      source = "ssoriche/kanidm"
    }
  }
}

# Configure the Kanidm Provider with explicit credentials
provider "kanidm" {
  url   = "https://idm.s8i.ca"
  token = var.kanidm_token
}

# Alternative: Configure using environment variables
# export KANIDM_URL="https://idm.s8i.ca"
# export KANIDM_TOKEN="your-api-token"
provider "kanidm" {}

# Recommended: Use 1Password provider to retrieve credentials
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
  url   = "https://idm.s8i.ca"
  token = data.onepassword_item.kanidm_admin_token.credential
}
