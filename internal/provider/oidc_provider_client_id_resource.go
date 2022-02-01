package provider

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

type oidcProviderClientIdResourceType struct{}

func (t oidcProviderClientIdResourceType) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Resource for synching the audience/ClientID in the AWS OIDC providers. The official AWS provider does currently not support this feature therefore this resource.",

		Attributes: map[string]tfsdk.Attribute{
			"client_id": {
				MarkdownDescription: "The ClientID/Audience to be added to the OIDC provider.",
				Optional:            false,
				Required:            true,
				Type:                types.StringType,
				PlanModifiers: tfsdk.AttributePlanModifiers{
					tfsdk.RequiresReplace(),
				},
			},
			"oidc_arn": {
				MarkdownDescription: "The target OIDC provider ARN where the ClientID should be added.",
				Optional:            false,
				Required:            true,
				Type:                types.StringType,
			},
		},
	}, nil
}

func (t oidcProviderClientIdResourceType) NewResource(ctx context.Context, in tfsdk.Provider) (tfsdk.Resource, diag.Diagnostics) {
	provider, diags := convertProviderType(in)

	return oidcProviderClientIdResource{
		provider: provider,
	}, diags
}

type oidcProviderClientIdResourceData struct {
	ClientId types.String `tfsdk:"client_id"`
	OidcArn  types.String `tfsdk:"oidc_arn"`
}

type oidcProviderClientIdResource struct {
	provider provider
}

func (r oidcProviderClientIdResource) Create(ctx context.Context, req tfsdk.CreateResourceRequest, resp *tfsdk.CreateResourceResponse) {
	var data oidcProviderClientIdResourceData

	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	_, err := addClientId(r, &data.ClientId.Value, &data.OidcArn.Value)

	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to add audience to provider, got error: %s", err))
		return
	}

	// write logs using the tflog package
	// see https://pkg.go.dev/github.com/hashicorp/terraform-plugin-log/tflog
	// for more information
	tflog.Trace(ctx, "created a resource")

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r oidcProviderClientIdResource) Read(ctx context.Context, req tfsdk.ReadResourceRequest, resp *tfsdk.ReadResourceResponse) {
	var data oidcProviderClientIdResourceData

	diags := req.State.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	plan := data

	iamClient := r.provider.iamClient

	oidcProvider, err := iamClient.GetOpenIDConnectProvider(context.TODO(), &iam.GetOpenIDConnectProviderInput{OpenIDConnectProviderArn: &data.OidcArn.Value})

	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Failed retrieving current audiences, got error: %s", err))
		return
	}

	if !contains(oidcProvider.ClientIDList, data.ClientId.Value) {
		resp.Diagnostics.AddWarning(fmt.Sprintf("ClientID: %s not found on provider.", data.ClientId.Value), "ClientID was already removed manually or from another script. Will skip during delete.")
		plan.ClientId.Null = true
	}

	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r oidcProviderClientIdResource) Update(ctx context.Context, req tfsdk.UpdateResourceRequest, resp *tfsdk.UpdateResourceResponse) {
	var data oidcProviderClientIdResourceData

	diags := req.Plan.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	_, err := addClientId(r, &data.ClientId.Value, &data.OidcArn.Value)

	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to add audience to provider, got error: %s", err))
		return
	}

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r oidcProviderClientIdResource) Delete(ctx context.Context, req tfsdk.DeleteResourceRequest, resp *tfsdk.DeleteResourceResponse) {
	var data oidcProviderClientIdResourceData

	diags := req.State.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	if !data.ClientId.Null {
		_, err := removeClientId(r, &data.ClientId.Value, &data.OidcArn.Value)

		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete audience from provider, got error: %s", err))
			return
		}
	} else {
		tflog.Info(ctx, "Skip deletion because already removed from audience.")
	}

	resp.State.RemoveResource(ctx)
}

func (r oidcProviderClientIdResource) ImportState(ctx context.Context, req tfsdk.ImportResourceStateRequest, resp *tfsdk.ImportResourceStateResponse) {
	tfsdk.ResourceImportStatePassthroughID(ctx, tftypes.NewAttributePath().WithAttributeName("id"), req, resp)
}

// CUSTOM FUNCTIONS

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func addClientId(r oidcProviderClientIdResource, clientId *string, oidcArn *string) (*iam.AddClientIDToOpenIDConnectProviderOutput, error) {
	r.provider.lock.Lock()
	defer r.provider.lock.Unlock()

	iamClient := r.provider.iamClient

	return iamClient.AddClientIDToOpenIDConnectProvider(context.TODO(),
		&iam.AddClientIDToOpenIDConnectProviderInput{ClientID: clientId, OpenIDConnectProviderArn: oidcArn})
}

func removeClientId(r oidcProviderClientIdResource, clientId *string, oidcArn *string) (*iam.RemoveClientIDFromOpenIDConnectProviderOutput, error) {
	r.provider.lock.Lock()
	defer r.provider.lock.Unlock()

	iamClient := r.provider.iamClient

	return iamClient.RemoveClientIDFromOpenIDConnectProvider(context.TODO(),
		&iam.RemoveClientIDFromOpenIDConnectProviderInput{ClientID: clientId, OpenIDConnectProviderArn: oidcArn})
}
