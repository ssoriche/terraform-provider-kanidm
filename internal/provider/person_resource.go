package provider

import (
	"context"
	"errors"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/ssoriche/terraform-provider-kanidm/internal/client"
)

// Ensure the implementation satisfies the required interfaces
var (
	_ resource.Resource                = (*personResource)(nil)
	_ resource.ResourceWithImportState = (*personResource)(nil)
)

// NewPersonResource creates a new person resource
func NewPersonResource() resource.Resource {
	return &personResource{}
}

// personResource is the resource implementation
type personResource struct {
	client *client.Client
}

// personResourceModel describes the resource data model
type personResourceModel struct {
	ID                           types.String `tfsdk:"id"`
	DisplayName                  types.String `tfsdk:"displayname"`
	Mail                         types.List   `tfsdk:"mail"`
	Password                     types.String `tfsdk:"password"`
	GenerateCredentialResetToken types.Bool   `tfsdk:"generate_credential_reset_token"`
	CredentialResetToken         types.String `tfsdk:"credential_reset_token"`
	CredentialResetTokenTTL      types.Int64  `tfsdk:"credential_reset_token_ttl"`
}

// Metadata returns the resource type name
func (r *personResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_person"
}

// Schema defines the schema for the resource
func (r *personResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `Manages a Kanidm person account.

## Authentication Setup

Kanidm supports two credential setup workflows:

### Password-Based Authentication
Set the ` + "`password`" + ` attribute to create a password-based account:

` + "```hcl" + `
resource "kanidm_person" "example" {
  id          = "jdoe"
  displayname = "John Doe"
  password    = var.initial_password
}
` + "```" + `

### Passkey/Modern Authentication (Recommended)
Set ` + "`generate_credential_reset_token = true`" + ` to generate a one-time token for credential setup via the Kanidm web UI:

` + "```hcl" + `
resource "kanidm_person" "example" {
  id                            = "jdoe"
  displayname                   = "John Doe"
  generate_credential_reset_token = true
}

output "credential_reset_token" {
  value     = kanidm_person.example.credential_reset_token
  sensitive = true
}
` + "```" + `

The user can then visit the Kanidm web UI with the token to set up passkeys or passwords.`,

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Unique identifier for the person account (username). Cannot be changed after creation.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"displayname": schema.StringAttribute{
				MarkdownDescription: "Display name of the person.",
				Required:            true,
			},
			"mail": schema.ListAttribute{
				MarkdownDescription: "Email addresses for the person.",
				Optional:            true,
				ElementType:         types.StringType,
			},
			"password": schema.StringAttribute{
				MarkdownDescription: "Password for the person account. **Note:** This is write-only and will not be stored in state. " +
					"Mutually exclusive with `generate_credential_reset_token`. " +
					"Consider using `lifecycle { ignore_changes = [password] }` if the password is managed externally.",
				Optional:  true,
				Sensitive: true,
			},
			"generate_credential_reset_token": schema.BoolAttribute{
				MarkdownDescription: "Whether to generate a credential reset token for passkey/password setup via the web UI. " +
					"Mutually exclusive with `password`. Defaults to `false`.",
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(false),
			},
			"credential_reset_token": schema.StringAttribute{
				MarkdownDescription: "The credential reset token (generated when `generate_credential_reset_token` is `true`). " +
					"This token can be used once to set up credentials via the Kanidm web UI. **Computed value only.**",
				Computed:  true,
				Sensitive: true,
			},
			"credential_reset_token_ttl": schema.Int64Attribute{
				MarkdownDescription: "Time-to-live for the credential reset token in seconds. Defaults to 3600 (1 hour).",
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(3600),
			},
		},
	}
}

// Configure adds the provider configured client to the resource
func (r *personResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// Create creates the resource and sets the initial Terraform state
func (r *personResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan personResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Validate mutually exclusive options
	hasPassword := !plan.Password.IsNull() && !plan.Password.IsUnknown()
	generateToken := plan.GenerateCredentialResetToken.ValueBool()

	if hasPassword && generateToken {
		resp.Diagnostics.AddError(
			"Conflicting Configuration",
			"Cannot specify both 'password' and 'generate_credential_reset_token'. Choose one authentication setup method.",
		)
		return
	}

	tflog.Debug(ctx, "Creating person", map[string]any{
		"id": plan.ID.ValueString(),
	})

	// Create the person account
	person, err := r.client.CreatePerson(ctx, plan.ID.ValueString(), plan.DisplayName.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Person",
			"Could not create person: "+err.Error(),
		)
		return
	}

	// Set password if provided
	if hasPassword {
		tflog.Debug(ctx, "Setting initial password for person")
		if err := r.client.SetPersonPassword(ctx, person.ID, plan.Password.ValueString()); err != nil {
			resp.Diagnostics.AddError(
				"Error Setting Password",
				"Person was created but password could not be set: "+err.Error(),
			)
			return
		}
	}

	// Generate credential reset token if requested
	if generateToken {
		tflog.Debug(ctx, "Generating credential reset token for person")
		ttl := int(plan.CredentialResetTokenTTL.ValueInt64())
		token, err := r.client.CreatePersonCredentialResetToken(ctx, person.ID, &ttl)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Generating Credential Reset Token",
				"Person was created but credential reset token could not be generated: "+err.Error(),
			)
			return
		}
		plan.CredentialResetToken = types.StringValue(token)
	}

	// Update mail if provided
	if !plan.Mail.IsNull() && !plan.Mail.IsUnknown() {
		var mailAddrs []string
		resp.Diagnostics.Append(plan.Mail.ElementsAs(ctx, &mailAddrs, false)...)
		if resp.Diagnostics.HasError() {
			return
		}

		if len(mailAddrs) > 0 {
			tflog.Debug(ctx, "Updating mail addresses for person")
			if err := r.client.UpdatePerson(ctx, person.ID, "", mailAddrs); err != nil {
				resp.Diagnostics.AddError(
					"Error Updating Mail",
					"Person was created but mail addresses could not be set: "+err.Error(),
				)
				return
			}
		}
	}

	// Read back the person to get the current state
	createdPerson, err := r.client.GetPerson(ctx, person.ID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Person",
			"Person was created but could not be read back: "+err.Error(),
		)
		return
	}

	// Map response to state
	plan.ID = types.StringValue(createdPerson.ID)
	plan.DisplayName = types.StringValue(createdPerson.DisplayName)

	if len(createdPerson.Mail) > 0 {
		mailList, diags := types.ListValueFrom(ctx, types.StringType, createdPerson.Mail)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		plan.Mail = mailList
	}

	// Password is write-only, keep the planned value but don't try to read it back

	tflog.Debug(ctx, "Person created successfully", map[string]any{
		"id": plan.ID.ValueString(),
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Read refreshes the Terraform state with the latest data
func (r *personResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state personResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Reading person", map[string]any{
		"id": state.ID.ValueString(),
	})

	// Get current person from API
	person, err := r.client.GetPerson(ctx, state.ID.ValueString())
	if err != nil {
		if errors.Is(err, client.ErrNotFound) {
			tflog.Warn(ctx, "Person not found, removing from state", map[string]any{
				"id": state.ID.ValueString(),
			})
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError(
			"Error Reading Person",
			"Could not read person: "+err.Error(),
		)
		return
	}

	// Update state with current values
	state.ID = types.StringValue(person.ID)
	state.DisplayName = types.StringValue(person.DisplayName)

	if len(person.Mail) > 0 {
		mailList, diags := types.ListValueFrom(ctx, types.StringType, person.Mail)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		state.Mail = mailList
	} else {
		state.Mail = types.ListNull(types.StringType)
	}

	// Password and credential_reset_token are write-only, preserve existing state values

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update updates the resource and sets the updated Terraform state
func (r *personResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state personResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Updating person", map[string]any{
		"id": plan.ID.ValueString(),
	})

	// Prepare mail addresses
	var mailAddrs []string
	if !plan.Mail.IsNull() && !plan.Mail.IsUnknown() {
		resp.Diagnostics.Append(plan.Mail.ElementsAs(ctx, &mailAddrs, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	// Update person attributes (displayname and mail)
	if err := r.client.UpdatePerson(ctx, plan.ID.ValueString(), plan.DisplayName.ValueString(), mailAddrs); err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Person",
			"Could not update person: "+err.Error(),
		)
		return
	}

	// Update password if changed
	if !plan.Password.Equal(state.Password) && !plan.Password.IsNull() {
		tflog.Debug(ctx, "Updating password for person")
		if err := r.client.SetPersonPassword(ctx, plan.ID.ValueString(), plan.Password.ValueString()); err != nil {
			resp.Diagnostics.AddError(
				"Error Updating Password",
				"Person was updated but password could not be changed: "+err.Error(),
			)
			return
		}
	}

	// Generate new credential reset token if requested and changed
	if plan.GenerateCredentialResetToken.ValueBool() && !plan.GenerateCredentialResetToken.Equal(state.GenerateCredentialResetToken) {
		tflog.Debug(ctx, "Generating new credential reset token for person")
		ttl := int(plan.CredentialResetTokenTTL.ValueInt64())
		token, err := r.client.CreatePersonCredentialResetToken(ctx, plan.ID.ValueString(), &ttl)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Generating Credential Reset Token",
				"Person was updated but credential reset token could not be generated: "+err.Error(),
			)
			return
		}
		plan.CredentialResetToken = types.StringValue(token)
	}

	// Read back the updated person
	updatedPerson, err := r.client.GetPerson(ctx, plan.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Person",
			"Person was updated but could not be read back: "+err.Error(),
		)
		return
	}

	// Update state
	plan.ID = types.StringValue(updatedPerson.ID)
	plan.DisplayName = types.StringValue(updatedPerson.DisplayName)

	if len(updatedPerson.Mail) > 0 {
		mailList, diags := types.ListValueFrom(ctx, types.StringType, updatedPerson.Mail)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		plan.Mail = mailList
	} else {
		plan.Mail = types.ListNull(types.StringType)
	}

	tflog.Debug(ctx, "Person updated successfully", map[string]any{
		"id": plan.ID.ValueString(),
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete deletes the resource and removes the Terraform state
func (r *personResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state personResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Deleting person", map[string]any{
		"id": state.ID.ValueString(),
	})

	// Delete the person
	if err := r.client.DeletePerson(ctx, state.ID.ValueString()); err != nil {
		if errors.Is(err, client.ErrNotFound) {
			// Person already deleted, just remove from state
			tflog.Warn(ctx, "Person not found during delete, removing from state", map[string]any{
				"id": state.ID.ValueString(),
			})
			return
		}

		resp.Diagnostics.AddError(
			"Error Deleting Person",
			"Could not delete person: "+err.Error(),
		)
		return
	}

	tflog.Debug(ctx, "Person deleted successfully", map[string]any{
		"id": state.ID.ValueString(),
	})
}

// ImportState imports an existing person into Terraform state
func (r *personResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Use the ID (username) directly as the import identifier
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)

	tflog.Debug(ctx, "Imported person", map[string]any{
		"id": req.ID,
	})
}
