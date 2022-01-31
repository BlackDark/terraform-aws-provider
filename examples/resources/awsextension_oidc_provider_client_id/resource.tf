resource "aws-extension_oidc_provider_client_id" "example" {
  client_id = "SomeAudience"
  oidc_arn  = "arn:aws:iam::xxx:oidc-provider/xxx"
}
