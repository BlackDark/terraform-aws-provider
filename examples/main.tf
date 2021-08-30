terraform {
  required_providers {
    blackdarkaws = {
      version = "~> 0.3.1"
      source  = "hashicorp.com/edu/blackdark-aws"
    }
  }
  required_version = "~> 1.0.3"
}

provider "blackdarkaws" {
  role_arn     = "arn"
  session_name = "Test"
}

resource "blackdarkaws_oidc_provider_client_id" "edu" {
  client_id = "test"
  oidc_arn  = "arn"
}
