package provider

import (
	"context"
	"errors"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/ssoriche/terraform-provider-kanidm/internal/client"
)

var (
	_ resource.Resource                = (*serviceAccountResource)(nil)
	_ resource.ResourceWithImportState = (*serviceAccountResource)(nil)
)

func NewServiceAccountResource() resource.Resource {
	return &serviceAccountResource{}
}

type serviceAccountResource struct {
	client *client.Client
}

type serviceAccountResourceModel struct {
	ID       types.String `tfsdk:"id"`
	APIToken types.String `tfsdk:"api_token"`
}

func (r *serviceAccountResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_service_account"
}

func (r *serviceAccountResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `Manages a Kanidm service account.

Service accounts are used for automated systems and applications to authenticate with Kanidm.
An API token is automatically generated on creation and can be used for authentication.

## Example Usage

` + "```hcl" + `
resource "kanidm_service_account" "terraform" {
  id          = "terraform-automation"
  displayname = "Terraform Automation Account"
}

# Store the API token in 1Password or another secret manager
output "terraform_token" {
  value     = kanidm_service_account.terraform.api_token
  sensitive = true
}
` + "```" + `

**Important:** The API token is only available during creation and cannot be recovered later.
Store it securely immediately after creation.`,

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Unique identifier for the service account. Cannot be changed after creation.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"api_token": schema.StringAttribute{
				MarkdownDescription: "API token for the service account. **Only available during creation.** " +
					"Store this token securely as it cannot be retrieved later.",
				Computed:  true,
				Sensitive: true,
			},
		},
	}
}

func (r *serviceAccountResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			"Expected *client.Client. Please report this issue to the provider developers.",
		)
		return
	}

	r.client = c
}

func (r *serviceAccountResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan serviceAccountResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Creating service account", map[string]any{
		"id": plan.ID.ValueString(),
	})

	// Create the service account (this also generates an initial API token)
	sa, err := r.client.CreateServiceAccount(ctx, plan.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Service Account",
			"Could not create service account: "+err.Error(),
		)
		return
	}

	// Map response to state
	plan.ID = types.StringValue(sa.ID)
	plan.APIToken = types.StringValue(sa.APIToken)

	tflog.Debug(ctx, "Service account created successfully", map[string]any{
		"id": plan.ID.ValueString(),
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *serviceAccountResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state serviceAccountResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Reading service account", map[string]any{
		"id": state.ID.ValueString(),
	})

	// Get current service account from API
	sa, err := r.client.GetServiceAccount(ctx, state.ID.ValueString())
	if err != nil {
		if errors.Is(err, client.ErrNotFound) {
			tflog.Warn(ctx, "Service account not found, removing from state", map[string]any{
				"id": state.ID.ValueString(),
			})
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError(
			"Error Reading Service Account",
			"Could not read service account: "+err.Error(),
		)
		return
	}

	// Update state with current values
	state.ID = types.StringValue(sa.ID)
	// API token is write-only and cannot be read back, preserve existing state value

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *serviceAccountResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state serviceAccountResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Updating service account", map[string]any{
		"id": plan.ID.ValueString(),
	})

	// Service accounts have no updatable attributes (ID requires replacement)
	// Just preserve state values
	plan.APIToken = state.APIToken

	tflog.Debug(ctx, "Service account updated successfully", map[string]any{
		"id": plan.ID.ValueString(),
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *serviceAccountResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state serviceAccountResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Deleting service account", map[string]any{
		"id": state.ID.ValueString(),
	})

	// Delete the service account
	if err := r.client.DeleteServiceAccount(ctx, state.ID.ValueString()); err != nil {
		if errors.Is(err, client.ErrNotFound) {
			tflog.Warn(ctx, "Service account not found during delete, removing from state", map[string]any{
				"id": state.ID.ValueString(),
			})
			return
		}

		resp.Diagnostics.AddError(
			"Error Deleting Service Account",
			"Could not delete service account: "+err.Error(),
		)
		return
	}

	tflog.Debug(ctx, "Service account deleted successfully", map[string]any{
		"id": state.ID.ValueString(),
	})
}

func (r *serviceAccountResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Use the ID directly as the import identifier
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)

	tflog.Debug(ctx, "Imported service account", map[string]any{
		"id": req.ID,
	})

	// Add a warning about the API token
	resp.Diagnostics.AddWarning(
		"API Token Not Available",
		"The API token for this service account is not available after import. "+
			"If you need the token, you must regenerate it manually using the Kanidm CLI or web interface.",
	)
}
