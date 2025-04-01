# Configure the Azure provider
terraform {
  required_providers {
    azurerm = {
      source  = "hashicorp/azurerm"
      version = "~> 2.65"
    }
    azuread = {
      source  = "hashicorp/azuread"
      version = "~> 2.15.0"
    }
  }
}

provider "azurerm" {
  features {}
}

provider "azuread" {
}

# Create a secure Azure AD application
resource "azuread_application" "secure_app" {
  display_name = "SecureApp"

  web {
    homepage_url  = "https://example.com"
    redirect_uris = ["https://example.com/auth"]

    implicit_grant {
      access_token_issuance_enabled = true
      id_token_issuance_enabled     = true
    }
  }

  required_resource_access {
    resource_app_id = "00000003-0000-0000-c000-000000000000" # Microsoft Graph

    resource_access {
      id   = "e1fe6dd8-ba31-4d61-89e7-88639da4683d" # User.Read
      type = "Scope"
    }
  }

  sign_in_audience = "AzureADandPersonalMicrosoftAccount"

  prevent_duplicate_names = false
}

# Assign a strong password policy
resource "azuread_directory_password_policy" "secure_policy" {
  min_password_length        = 14
  require_lowercase          = true
  require_uppercase          = true
  require_numbers            = true
  require_symbols            = true
  password_history_enforced  = 24
  max_password_age_in_days   = 60
  min_password_age_in_hours  = 24
}