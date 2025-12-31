package provider

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/ssoriche/terraform-provider-kanidm/internal/client"
)

// Ensure the implementation satisfies the provider.Provider interface
var _ provider.Provider = (*kanidmProvider)(nil)

// kanidmProvider is the provider implementation
type kanidmProvider struct {
	version string
}

// kanidmProviderModel describes the provider data model
type kanidmProviderModel struct {
	URL   types.String `tfsdk:"url"`
	Token types.String `tfsdk:"token"`
}

// New creates a new provider instance
func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &kanidmProvider{
			version: version,
		}
	}
}

// Metadata returns the provider type name
func (p *kanidmProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "kanidm"
	resp.Version = p.version
}

// Schema defines the provider-level schema for configuration data
func (p *kanidmProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Terraform provider for managing Kanidm identity resources.",
		Attributes: map[string]schema.Attribute{
			"url": schema.StringAttribute{
				Description: "Kanidm server URL. May also be provided via KANIDM_URL environment variable.",
				Optional:    true,
			},
			"token": schema.StringAttribute{
				Description: "Kanidm API token for authentication. May also be provided via KANIDM_TOKEN environment variable.",
				Optional:    true,
				Sensitive:   true,
			},
		},
	}
}

// Configure prepares the Kanidm API client for data sources and resources
func (p *kanidmProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	tflog.Info(ctx, "Configuring Kanidm client")

	var config kanidmProviderModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Resolve URL from configuration or environment variable
	url := os.Getenv("KANIDM_URL")
	if !config.URL.IsNull() {
		url = config.URL.ValueString()
	}

	if url == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("url"),
			"Missing Kanidm URL",
			"The provider cannot create the Kanidm API client as there is a missing or empty value for the Kanidm URL. "+
				"Set the url value in the configuration or use the KANIDM_URL environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	// Resolve token from configuration or environment variable
	token := os.Getenv("KANIDM_TOKEN")
	if !config.Token.IsNull() {
		token = config.Token.ValueString()
	}

	if token == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("token"),
			"Missing Kanidm Token",
			"The provider cannot create the Kanidm API client as there is a missing or empty value for the Kanidm token. "+
				"Set the token value in the configuration or use the KANIDM_TOKEN environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Create Kanidm client
	tflog.Debug(ctx, "Creating Kanidm client", map[string]any{
		"url": url,
	})

	apiClient := client.NewClient(url, token)

	// Make the client available to data sources and resources
	resp.DataSourceData = apiClient
	resp.ResourceData = apiClient

	tflog.Info(ctx, "Configured Kanidm client", map[string]any{
		"success": true,
	})
}

// DataSources defines the data sources implemented in the provider
func (p *kanidmProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		// Data sources will be implemented later
	}
}

// Resources defines the resources implemented in the provider
func (p *kanidmProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewPersonResource,
		NewServiceAccountResource,
		NewGroupResource,
		NewOAuth2BasicResource,
	}
}
