module "tags" {
  source = "git::https://github.com/fluxcd/test-infra.git//tf-modules/utils/tags"

  tags = var.tags
}

data "google_project" "this" {}


resource "google_sourcerepo_repository" "this" {
  name = var.name
}
