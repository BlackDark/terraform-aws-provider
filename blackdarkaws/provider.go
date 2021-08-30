package blackdarkaws

import (
	"context"
	"os"

	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/hashicorp/terraform-plugin-framework/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
)

var stderr = os.Stderr

func New() tfsdk.Provider {
	return &provider{}
}

type provider struct {
	configured bool
	client     *aws.Config
}

// GetSchema
func (p *provider) GetSchema(_ context.Context) (schema.Schema, []*tfprotov6.Diagnostic) {
	return schema.Schema{
		Attributes: map[string]schema.Attribute{
			"role_arn": {
				Type:     types.StringType,
				Optional: true,
				Computed: true,
			},
			"session_name": {
				Type:     types.StringType,
				Optional: true,
				Computed: true,
			},
		},
	}, nil
}

// Provider schema struct
type providerData struct {
	RoleArn     types.String `tfsdk:"role_arn"`
	SessionName types.String `tfsdk:"session_name"`
}

// STSAssumeRoleAPI defines the interface for the AssumeRole function.
// We use this interface to test the function using a mocked service.
type STSAssumeRoleAPI interface {
	AssumeRole(ctx context.Context,
		params *sts.AssumeRoleInput,
		optFns ...func(*sts.Options)) (*sts.AssumeRoleOutput, error)
}

// TakeRole gets temporary security credentials to access resources.
// Inputs:
//     c is the context of the method call, which includes the AWS Region.
//     api is the interface that defines the method call.
//     input defines the input arguments to the service call.
// Output:
//     If successful, an AssumeRoleOutput object containing the result of the service call and nil.
//     Otherwise, nil and an error from the call to AssumeRole.
func TakeRole(c context.Context, api STSAssumeRoleAPI, input *sts.AssumeRoleInput) (*sts.AssumeRoleOutput, error) {
	return api.AssumeRole(c, input)
}

func (p *provider) Configure(ctx context.Context, req tfsdk.ConfigureProviderRequest, resp *tfsdk.ConfigureProviderResponse) {
	// Retrieve provider data from configuration
	var config providerData
	err := req.Config.Get(ctx, &config)
	if err != nil {
		resp.Diagnostics = append(resp.Diagnostics, &tfprotov6.Diagnostic{
			Severity: tfprotov6.DiagnosticSeverityError,
			Summary:  "Error parsing configuration",
			Detail:   "Error parsing the configuration, this is an error in the provider. Please report the following to the provider developer:\n\n" + err.Error(),
		})
		return
	}

	var roleArn string
	if config.RoleArn.Unknown {
		resp.Diagnostics = append(resp.Diagnostics, &tfprotov6.Diagnostic{
			Severity: tfprotov6.DiagnosticSeverityWarning,
			Summary:  "Unable to create client",
			Detail:   "Cannot use unknown value as role_arn",
		})
		return
	}

	if config.RoleArn.Null {
		// Unused
		roleArn = os.Getenv("HASHICUPS_USERNAME")
	} else {
		roleArn = config.RoleArn.Value
	}

	if roleArn == "" {
		resp.Diagnostics = append(resp.Diagnostics, &tfprotov6.Diagnostic{
			// Error vs warning - empty value must stop execution
			Severity: tfprotov6.DiagnosticSeverityError,
			Summary:  "Unable to find roleArn",
			Detail:   "role_arn cannot be an empty string",
		})
	}

	// User must provide a password to the provider
	var sessionName string
	if config.SessionName.Unknown {
		// Cannot connect to client with an unknown value
		resp.Diagnostics = append(resp.Diagnostics, &tfprotov6.Diagnostic{
			Severity: tfprotov6.DiagnosticSeverityWarning,
			Summary:  "Unable to create client",
			Detail:   "Cannot use unknown value as sessionName",
		})
		return
	}

	if config.SessionName.Null {
		sessionName = os.Getenv("HASHICUPS_PASSWORD")
	} else {
		sessionName = config.SessionName.Value
	}

	if sessionName == "" {
		resp.Diagnostics = append(resp.Diagnostics, &tfprotov6.Diagnostic{
			// Error vs warning - empty value must stop execution
			Severity: tfprotov6.DiagnosticSeverityError,
			Summary:  "Unable to find sessionName",
			Detail:   "session_name cannot be an empty string",
		})
	}

	cfg, err := awsConfig.LoadDefaultConfig(context.TODO())
	if err != nil {
		resp.Diagnostics = append(resp.Diagnostics, &tfprotov6.Diagnostic{
			Severity: tfprotov6.DiagnosticSeverityError,
			Summary:  "Unable to create client",
			Detail:   "Unable to create aws config:\n\n" + err.Error(),
		})
		return
	}

	client := sts.NewFromConfig(cfg)

	input := &sts.AssumeRoleInput{
		RoleArn:         &roleArn,
		RoleSessionName: &sessionName,
	}

	result, err := TakeRole(context.TODO(), client, input)
	if err != nil {
		resp.Diagnostics = append(resp.Diagnostics, &tfprotov6.Diagnostic{
			Severity: tfprotov6.DiagnosticSeverityError,
			Summary:  "Unable to create client",
			Detail:   "Unable to assume role:\n\n" + err.Error(),
		})
		return
	}

	cfg2, err := awsConfig.LoadDefaultConfig(context.TODO(),
		awsConfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(*result.Credentials.AccessKeyId, *result.Credentials.SecretAccessKey, *result.Credentials.SessionToken)))
	if err != nil {
		resp.Diagnostics = append(resp.Diagnostics, &tfprotov6.Diagnostic{
			Severity: tfprotov6.DiagnosticSeverityError,
			Summary:  "Unable to create client",
			Detail:   "Unable to load config for assumed role:\n\n" + err.Error(),
		})
		return
	}
	p.client = &cfg2
	p.configured = true
}

func (p *provider) GetResources(_ context.Context) (map[string]tfsdk.ResourceType, []*tfprotov6.Diagnostic) {
	return map[string]tfsdk.ResourceType{
		"blackdarkaws_oidc_provider_client_id": resourceAwsOidcProviderClientIdEntryType{},
	}, nil
}

func (p *provider) GetDataSources(_ context.Context) (map[string]tfsdk.DataSourceType, []*tfprotov6.Diagnostic) {
	return map[string]tfsdk.DataSourceType{}, nil
}
