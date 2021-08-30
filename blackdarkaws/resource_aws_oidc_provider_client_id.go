package blackdarkaws

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"

	// "github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

type resourceAwsOidcProviderClientIdEntryType struct{}

// Order Resource schema
func (r resourceAwsOidcProviderClientIdEntryType) GetSchema(_ context.Context) (schema.Schema, []*tfprotov6.Diagnostic) {
	return schema.Schema{
		Attributes: map[string]schema.Attribute{
			"client_id": {
				Type:     types.StringType,
				Required: true,
			},
			"oidc_arn": {
				Type:     types.StringType,
				Required: true,
			},
		},
	}, nil
}

// New resource instance
func (r resourceAwsOidcProviderClientIdEntryType) NewResource(_ context.Context, p tfsdk.Provider) (tfsdk.Resource, []*tfprotov6.Diagnostic) {
	return resourceAwsOidcProviderClientIdEntry{
		p: *(p.(*provider)),
	}, nil
}

type resourceAwsOidcProviderClientIdEntry struct {
	p provider
}

// Create a new resource
func (r resourceAwsOidcProviderClientIdEntry) Create(ctx context.Context, req tfsdk.CreateResourceRequest, resp *tfsdk.CreateResourceResponse) {
	if !r.p.configured {
		resp.Diagnostics = append(resp.Diagnostics, &tfprotov6.Diagnostic{
			Severity: tfprotov6.DiagnosticSeverityError,
			Summary:  "Provider not configured",
			Detail:   "The provider hasn't been configured before apply, likely because it depends on an unknown value from another resource. This leads to weird stuff happening, so we'd prefer if you didn't do that. Thanks!",
		})
		return
	}

	// Retrieve values from plan
	var plan AwsOIDCProviderClientIDEntry
	err := req.Plan.Get(ctx, &plan)
	if err != nil {
		resp.Diagnostics = append(resp.Diagnostics, &tfprotov6.Diagnostic{
			Severity: tfprotov6.DiagnosticSeverityError,
			Summary:  "Error reading plan",
			Detail:   "An unexpected error was encountered while reading the plan: " + err.Error(),
		})
		return
	}

	iamClient := iam.NewFromConfig(*r.p.client)

	iamResult, err := iamClient.AddClientIDToOpenIDConnectProvider(context.TODO(),
		&iam.AddClientIDToOpenIDConnectProviderInput{ClientID: &plan.ClientID.Value, OpenIDConnectProviderArn: &plan.OIDCArn.Value})

	_ = iamResult
	if err != nil {
		resp.Diagnostics = append(resp.Diagnostics, &tfprotov6.Diagnostic{
			Severity: tfprotov6.DiagnosticSeverityError,
			Summary:  "Failed adding ClientID",
			Detail:   "An unexpected error was encountered while adding ClientID: " + err.Error(),
		})
		return
	}

	var result = AwsOIDCProviderClientIDEntry{
		ClientID: plan.ClientID,
		OIDCArn:  plan.OIDCArn,
	}

	err = resp.State.Set(ctx, result)

	if err != nil {
		resp.Diagnostics = append(resp.Diagnostics, &tfprotov6.Diagnostic{
			Severity: tfprotov6.DiagnosticSeverityError,
			Summary:  "Error saving state",
			Detail:   "Could not set state, unexpected error: " + err.Error(),
		})
		return
	}
}

// Read resource information
func (r resourceAwsOidcProviderClientIdEntry) Read(ctx context.Context, req tfsdk.ReadResourceRequest, resp *tfsdk.ReadResourceResponse) {
	// Get current state
	var state AwsOIDCProviderClientIDEntry
	err := req.State.Get(ctx, &state)
	if err != nil {
		resp.Diagnostics = append(resp.Diagnostics, &tfprotov6.Diagnostic{
			Severity: tfprotov6.DiagnosticSeverityError,
			Summary:  "Error reading state",
			Detail:   "An unexpected error was encountered while reading the state: " + err.Error(),
		})
		return
	}

	// Get order from API and then update what is in state from what the API returns
	orderID := state.ClientID.Value
	oidcArn := state.OIDCArn

	iamClient := iam.NewFromConfig(*r.p.client)

	oidcProvider, err := iamClient.GetOpenIDConnectProvider(context.TODO(), &iam.GetOpenIDConnectProviderInput{OpenIDConnectProviderArn: &oidcArn.Value})

	if err != nil {
		resp.Diagnostics = append(resp.Diagnostics, &tfprotov6.Diagnostic{
			Severity: tfprotov6.DiagnosticSeverityError,
			Summary:  "Error requesting OpenIDProvider",
			Detail:   "An unexpected error was encountered while requesting oidc provider for client ids: " + err.Error(),
		})
		return
	}

	var newOrder AwsOIDCProviderClientIDEntry

	if contains(oidcProvider.ClientIDList, orderID) {
		newOrder = state
	}

	err = resp.State.Set(ctx, &newOrder)

	if err != nil {
		resp.Diagnostics = append(resp.Diagnostics, &tfprotov6.Diagnostic{
			Severity: tfprotov6.DiagnosticSeverityError,
			Summary:  "Error setting state",
			Detail:   "Unexpected error encountered trying to set new state: " + err.Error(),
		})
		return
	}
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

// Update resource
func (r resourceAwsOidcProviderClientIdEntry) Update(ctx context.Context, req tfsdk.UpdateResourceRequest, resp *tfsdk.UpdateResourceResponse) {
	if !r.p.configured {
		resp.Diagnostics = append(resp.Diagnostics, &tfprotov6.Diagnostic{
			Severity: tfprotov6.DiagnosticSeverityError,
			Summary:  "Provider not configured",
			Detail:   "The provider hasn't been configured before apply, likely because it depends on an unknown value from another resource. This leads to weird stuff happening, so we'd prefer if you didn't do that. Thanks!",
		})
		return
	}

	// Retrieve values from plan
	var plan AwsOIDCProviderClientIDEntry
	err := req.Plan.Get(ctx, &plan)
	if err != nil {
		resp.Diagnostics = append(resp.Diagnostics, &tfprotov6.Diagnostic{
			Severity: tfprotov6.DiagnosticSeverityError,
			Summary:  "Error reading plan",
			Detail:   "An unexpected error was encountered while reading the plan: " + err.Error(),
		})
		return
	}

	iamClient := iam.NewFromConfig(*r.p.client)

	iamResult, err := iamClient.AddClientIDToOpenIDConnectProvider(context.TODO(),
		&iam.AddClientIDToOpenIDConnectProviderInput{ClientID: &plan.ClientID.Value, OpenIDConnectProviderArn: &plan.OIDCArn.Value})

	_ = iamResult
	if err != nil {
		resp.Diagnostics = append(resp.Diagnostics, &tfprotov6.Diagnostic{
			Severity: tfprotov6.DiagnosticSeverityError,
			Summary:  "Failed adding ClientID",
			Detail:   "An unexpected error was encountered while adding ClientID: " + err.Error(),
		})
		return
	}

	var result = AwsOIDCProviderClientIDEntry{
		ClientID: plan.ClientID,
		OIDCArn:  plan.OIDCArn,
	}

	err = resp.State.Set(ctx, result)

	if err != nil {
		resp.Diagnostics = append(resp.Diagnostics, &tfprotov6.Diagnostic{
			Severity: tfprotov6.DiagnosticSeverityError,
			Summary:  "Error saving state",
			Detail:   "Could not set state, unexpected error: " + err.Error(),
		})
		return
	}
}

// Delete resource
func (r resourceAwsOidcProviderClientIdEntry) Delete(ctx context.Context, req tfsdk.DeleteResourceRequest, resp *tfsdk.DeleteResourceResponse) {
	var state AwsOIDCProviderClientIDEntry
	err := req.State.Get(ctx, &state)
	if err != nil {
		resp.Diagnostics = append(resp.Diagnostics, &tfprotov6.Diagnostic{
			Severity: tfprotov6.DiagnosticSeverityError,
			Summary:  "Error reading state",
			Detail:   "An unexpected error was encountered while reading the state: " + err.Error(),
		})
		return
	}

	// Get order ID from state
	orderID := state.ClientID.Value
	oidcArn := state.OIDCArn.Value

	iamClient := iam.NewFromConfig(*r.p.client)

	iamResult, err := iamClient.RemoveClientIDFromOpenIDConnectProvider(context.TODO(),
		&iam.RemoveClientIDFromOpenIDConnectProviderInput{ClientID: &orderID, OpenIDConnectProviderArn: &oidcArn})

	_ = iamResult
	if err != nil {
		resp.Diagnostics = append(resp.Diagnostics, &tfprotov6.Diagnostic{
			Severity: tfprotov6.DiagnosticSeverityError,
			Summary:  "Failed deleting client id from oidc provider",
			Detail:   "An unexpected error was encountered while deleting client id: " + err.Error(),
		})
		return
	}

	// Remove resource from state
	resp.State.RemoveResource(ctx)
}
