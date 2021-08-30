package blackdarkaws

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type AwsOIDCProviderClientIDEntry struct {
	ClientID types.String `tfsdk:"client_id"`
	OIDCArn  types.String `tfsdk:"oidc_arn"`
}
