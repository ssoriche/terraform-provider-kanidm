package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/ssoriche/terraform-provider-kanidm/internal/client"
)

// Ensure the implementation satisfies the resource.Resource interface
var _ resource.Resource = (*personResource)(nil)

// NewPersonResource creates a new person resource
func NewPersonResource() resource.Resource {
	return &personResource{}
}

// personResource is the resource implementation
type personResource struct {
	client *client.Client
}

// personResourceModel describes the resource data model
//
//nolint:unused // Will be used when CRUD operations are implemented
type personResourceModel struct {
	ID          types.String `tfsdk:"id"`
	DisplayName types.String `tfsdk:"displayname"`
	Mail        types.List   `tfsdk:"mail"`
}

// Metadata returns the resource type name
func (r *personResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_person"
}

// Schema defines the schema for the resource
func (r *personResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Kanidm person account.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Unique identifier for the person account (username).",
				Required:    true,
			},
			"displayname": schema.StringAttribute{
				Description: "Display name of the person.",
				Required:    true,
			},
			"mail": schema.ListAttribute{
				Description: "Email addresses for the person.",
				Optional:    true,
				ElementType: types.StringType,
			},
		},
	}
}

// Configure adds the provider configured client to the resource
func (r *personResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			"Expected *client.Client, got something else. Please report this issue to the provider developers.",
		)
		return
	}

	r.client = client
}

// Create creates the resource and sets the initial Terraform state
func (r *personResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// TODO: Implement Create logic
	resp.Diagnostics.AddError("Not Implemented", "Person resource Create is not yet implemented")
}

// Read refreshes the Terraform state with the latest data
func (r *personResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// TODO: Implement Read logic
	resp.Diagnostics.AddError("Not Implemented", "Person resource Read is not yet implemented")
}

// Update updates the resource and sets the updated Terraform state
func (r *personResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// TODO: Implement Update logic
	resp.Diagnostics.AddError("Not Implemented", "Person resource Update is not yet implemented")
}

// Delete deletes the resource and removes the Terraform state
func (r *personResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// TODO: Implement Delete logic
	resp.Diagnostics.AddError("Not Implemented", "Person resource Delete is not yet implemented")
}
