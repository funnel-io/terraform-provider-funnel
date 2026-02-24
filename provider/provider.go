package provider

import (
	"context"
	"fmt"

	"terraform-provider-funnel/provider/auth"
	"terraform-provider-funnel/provider/common"
	"terraform-provider-funnel/provider/datasources"
	"terraform-provider-funnel/provider/resources"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ provider.Provider = &funnelProvider{}
)

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &funnelProvider{
			version: version,
		}
	}
}

type funnelProvider struct {
	version     string
	environment string
}

func (p *funnelProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "funnel"
	resp.Version = p.version
}

func (p *funnelProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manage your Funnel setup.",
		Attributes: map[string]schema.Attribute{
			"environment": schema.StringAttribute{
				MarkdownDescription: "Funnel environment to manage. Default `us`. One of `us`, `eu`, `stage`, or `dev`.",
				Optional:            true,
				Validators: []validator.String{
					stringvalidator.OneOf("us", "eu", "stage", "dev"),
				},
			},
			"subscription_id": schema.StringAttribute{
				MarkdownDescription: "Funnel subscription ID. E.g. `fsXXXXXXXXXXX`.",
				Required:            true,
			},
			"client_id": schema.StringAttribute{
				MarkdownDescription: "Auth0 Client ID for Funnel.",
				Required:            true,
			},
			"client_secret": schema.StringAttribute{
				MarkdownDescription: "Auth0 Client Secret for Funnel.",
				Required:            true,
				Sensitive:           true,
			},
		},
	}
}

func (p *funnelProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	tflog.Info(ctx, "Configuring Terraform Provider Funnel client")

	var config common.FunnelProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Store configuration values
	if config.Environment.IsNull() || config.Environment.IsUnknown() {
		config.Environment = types.StringValue("us")
	}
	p.environment = config.Environment.ValueString()

	// Get Auth0 token
	token, err := auth.GetAccessToken(config.ClientId.ValueString(), config.ClientSecret.ValueString(), config.Environment.ValueString(), ctx)
	if err != nil {
		resp.Diagnostics.AddError("Failed to get Auth0 token", fmt.Sprintf("Could not get Auth0 token: %v", err))
		return
	}
	config.Token = token

	// Make the configuration available to resources and data sources
	resp.DataSourceData = &config
	resp.ResourceData = &config
}

func (p *funnelProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		datasources.NewExportFieldDataSource,
		datasources.NewWorkspacesDataSource,
	}
}

func (p *funnelProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		resources.NewWorkspaceResource,
		resources.NewGCSResource,
		resources.NewBigqueryResource,
		resources.NewSnowflakeResource,
	}
}
