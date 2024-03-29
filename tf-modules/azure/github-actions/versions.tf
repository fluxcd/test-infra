terraform {
  required_providers {
    azurerm = {
      source  = "hashicorp/azurerm"
      version = ">= 3.62.1"
    }

    azuread = {
      source  = "hashicorp/azuread"
      version = ">= 2.39.0"
    }

    github = {
      source  = "integrations/github"
      version = ">= 5.28.1"
    }
  }
}
