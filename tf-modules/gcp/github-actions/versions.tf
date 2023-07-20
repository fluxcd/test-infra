terraform {
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "4.73.1"
    }

    github = {
      source  = "integrations/github"
      version = ">= 5.28.1"
    }
  }
}
