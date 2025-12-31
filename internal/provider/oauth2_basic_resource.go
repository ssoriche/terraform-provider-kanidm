package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/ssoriche/terraform-provider-kanidm/internal/client"
)

var _ resource.Resource = (*oauth2BasicResource)(nil)

func NewOAuth2BasicResource() resource.Resource {
	return &oauth2BasicResource{}
}

type oauth2BasicResource struct {
	client *client.Client
}

//nolint:unused // Will be used when CRUD operations are implemented
type oauth2BasicResourceModel struct {
	Name         types.String `tfsdk:"name"`
	DisplayName  types.String `tfsdk:"displayname"`
	Origin       types.String `tfsdk:"origin"`
	ClientID     types.String `tfsdk:"client_id"`
	ClientSecret types.String `tfsdk:"client_secret"`
}

func (r *oauth2BasicResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_oauth2_basic"
}

func (r *oauth2BasicResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Kanidm OAuth2 basic (confidential) client.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Description: "Resource server name (OAuth2 client identifier).",
				Required:    true,
			},
			"displayname": schema.StringAttribute{
				Description: "Display name for the OAuth2 client.",
				Required:    true,
			},
			"origin": schema.StringAttribute{
				Description: "Origin URL for the OAuth2 client.",
				Required:    true,
			},
			"client_id": schema.StringAttribute{
				Description: "OAuth2 client ID (computed).",
				Computed:    true,
			},
			"client_secret": schema.StringAttribute{
				Description: "OAuth2 client secret (generated on creation).",
				Computed:    true,
				Sensitive:   true,
			},
		},
	}
}

func (r *oauth2BasicResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Resource Configure Type", "Expected *client.Client")
		return
	}
	r.client = client
}

func (r *oauth2BasicResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	resp.Diagnostics.AddError("Not Implemented", "OAuth2Basic resource Create is not yet implemented")
}

func (r *oauth2BasicResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	resp.Diagnostics.AddError("Not Implemented", "OAuth2Basic resource Read is not yet implemented")
}

func (r *oauth2BasicResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError("Not Implemented", "OAuth2Basic resource Update is not yet implemented")
}

func (r *oauth2BasicResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	resp.Diagnostics.AddError("Not Implemented", "OAuth2Basic resource Delete is not yet implemented")
}
