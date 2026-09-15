package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/Sonny00/terraform-provider-thehive/internal/client"
)

var (
	_ resource.Resource                = &ProfileResource{}
	_ resource.ResourceWithImportState = &ProfileResource{}
)

type ProfileResource struct{ client *client.Client }

type profileResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Permissions types.Set    `tfsdk:"permissions"`
	IsAdmin     types.Bool   `tfsdk:"is_admin"`
}

func NewProfileResource() resource.Resource { return &ProfileResource{} }

func (r *ProfileResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_profile"
}

func (r *ProfileResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a TheHive permission profile.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"name": schema.StringAttribute{
				Description: "Profile name.",
				Required:    true,
			},
			"permissions": schema.SetAttribute{
				Description: "Permission strings (e.g. manageCase/create, manageAlert/create).",
				Required:    true,
				ElementType: types.StringType,
			},
			"is_admin": schema.BoolAttribute{
				Description: "Whether this is an admin-level profile.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
		},
	}
}

func (r *ProfileResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected provider data type", fmt.Sprintf("Expected *client.Client, got: %T", req.ProviderData))
		return
	}
	r.client = c
}

func (r *ProfileResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan profileResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var perms []string
	resp.Diagnostics.Append(plan.Permissions.ElementsAs(ctx, &perms, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	p, err := r.client.CreateProfile(ctx, client.ProfileInput{
		Name:        plan.Name.ValueString(),
		Permissions: perms,
		IsAdmin:     plan.IsAdmin.ValueBool(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Error creating profile", err.Error())
		return
	}

	tflog.Trace(ctx, "created profile", map[string]any{"id": p.ID})
	plan.ID = types.StringValue(p.ID)
	plan.Name = types.StringValue(p.Name)
	// Don't remap permissions from API response — it reorders them alphabetically.
	// plan.Permissions stays as the user wrote it in HCL.
	plan.IsAdmin = types.BoolValue(p.IsAdmin)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *ProfileResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state profileResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	p, err := r.client.GetProfile(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading profile", err.Error())
		return
	}
	state.Name = types.StringValue(p.Name)
	permsList, diags := types.SetValueFrom(ctx, types.StringType, p.Permissions)
	resp.Diagnostics.Append(diags...)
	state.Permissions = permsList
	state.IsAdmin = types.BoolValue(p.IsAdmin)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *ProfileResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan profileResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var perms []string
	resp.Diagnostics.Append(plan.Permissions.ElementsAs(ctx, &perms, false)...)
	if resp.Diagnostics.HasError() {
		return
	}
	_, err := r.client.UpdateProfile(ctx, plan.ID.ValueString(), client.ProfileInput{
		Name:        plan.Name.ValueString(),
		Permissions: perms,
		IsAdmin:     plan.IsAdmin.ValueBool(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Error updating profile", err.Error())
		return
	}
	// Don't remap permissions — same reason as Create.
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *ProfileResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state profileResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteProfile(ctx, state.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Error deleting profile", err.Error())
	}
}

func (r *ProfileResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
