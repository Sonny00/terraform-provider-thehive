package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/Sonny00/terraform-provider-thehive/internal/client"
)

// Compile-time interface checks.
var (
	_ resource.Resource                = &OrganisationResource{}
	_ resource.ResourceWithImportState = &OrganisationResource{}
)

// OrganisationResource manages a TheHive organisation.
type OrganisationResource struct {
	client *client.Client
}

// organisationResourceModel maps the Terraform schema to Go types.
type organisationResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
}

func NewOrganisationResource() resource.Resource {
	return &OrganisationResource{}
}

func (r *OrganisationResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_organisation"
}

func (r *OrganisationResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a TheHive organisation.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Internal TheHive identifier (read-only, set by the API).",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Name of the organisation. Must be unique on the instance.",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "Free-text description of the organisation.",
				Optional:    true,
				Computed:    true, // API may return "" when unset.
			},
		},
	}
}

// Configure receives the provider-level client.
func (r *OrganisationResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected provider data type",
			fmt.Sprintf("Expected *client.Client, got: %T", req.ProviderData),
		)
		return
	}

	r.client = c
}

// ---------- CRUD ----------

func (r *OrganisationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan organisationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	org, err := r.client.CreateOrganisation(ctx, client.OrganisationInput{
		Name:        plan.Name.ValueString(),
		Description: plan.Description.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Error creating organisation", err.Error())
		return
	}

	tflog.Trace(ctx, "created organisation", map[string]any{"id": org.ID})

	plan.ID = types.StringValue(org.ID)
	plan.Name = types.StringValue(org.Name)
	plan.Description = types.StringValue(org.Description)

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *OrganisationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state organisationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	org, err := r.client.GetOrganisation(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading organisation", err.Error())
		return
	}

	state.Name = types.StringValue(org.Name)
	state.Description = types.StringValue(org.Description)

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *OrganisationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan organisationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	org, err := r.client.UpdateOrganisation(ctx, plan.ID.ValueString(), client.OrganisationInput{
		Name:        plan.Name.ValueString(),
		Description: plan.Description.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Error updating organisation", err.Error())
		return
	}

	plan.Name = types.StringValue(org.Name)
	plan.Description = types.StringValue(org.Description)

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *OrganisationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state organisationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteOrganisation(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error deleting organisation", err.Error())
		return
	}

	tflog.Trace(ctx, "deleted organisation", map[string]any{"id": state.ID.ValueString()})
}

// ImportState lets users run `terraform import thehive_organisation.foo <id>`.
func (r *OrganisationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
