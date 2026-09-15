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

var (
	_ resource.Resource                = &UserResource{}
	_ resource.ResourceWithImportState = &UserResource{}
)

type UserResource struct{ client *client.Client }

type userResourceModel struct {
	ID           types.String `tfsdk:"id"`
	Login        types.String `tfsdk:"login"`
	Name         types.String `tfsdk:"name"`
	Profile      types.String `tfsdk:"profile"`
	Organisation types.String `tfsdk:"organisation"`
	Email        types.String `tfsdk:"email"`
	Type         types.String `tfsdk:"type"`
	Locked       types.Bool   `tfsdk:"locked"`
}

func NewUserResource() resource.Resource { return &UserResource{} }

func (r *UserResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user"
}

func (r *UserResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a TheHive user account.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"login": schema.StringAttribute{
				Description: "Login identifier (typically an email address).",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Display name of the user.",
				Required:    true,
			},
			"profile": schema.StringAttribute{
				Description: "Profile to assign (e.g. analyst, admin, org-admin).",
				Required:    true,
			},
			"organisation": schema.StringAttribute{
				Description: "Organisation the user belongs to.",
				Required:    true,
			},
			"email": schema.StringAttribute{
				Description: "Email address.",
				Optional:    true,
				Computed:    true,
			},
			"type": schema.StringAttribute{
				Description: "Account type: normal or service.",
				Optional:    true,
				Computed:    true,
			},
			"locked": schema.BoolAttribute{
				Description: "Whether the account is locked.",
				Computed:    true,
			},
		},
	}
}

func (r *UserResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *UserResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan userResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	u, err := r.client.CreateUser(ctx, client.UserInput{
		Login:        plan.Login.ValueString(),
		Name:         plan.Name.ValueString(),
		Profile:      plan.Profile.ValueString(),
		Organisation: plan.Organisation.ValueString(),
		Email:        plan.Email.ValueString(),
		Type:         plan.Type.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Error creating user", err.Error())
		return
	}

	tflog.Trace(ctx, "created user", map[string]any{"id": u.ID})
	plan.ID = types.StringValue(u.ID)
	plan.Login = types.StringValue(u.Login)
	plan.Name = types.StringValue(u.Name)
	plan.Profile = types.StringValue(u.Profile)
	plan.Organisation = types.StringValue(u.Organisation)
	plan.Email = types.StringValue(u.Email)
	plan.Type = types.StringValue(u.Type)
	plan.Locked = types.BoolValue(u.Locked)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *UserResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state userResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	u, err := r.client.GetUser(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading user", err.Error())
		return
	}
	state.Login = types.StringValue(u.Login)
	state.Name = types.StringValue(u.Name)
	state.Profile = types.StringValue(u.Profile)
	state.Organisation = types.StringValue(u.Organisation)
	state.Email = types.StringValue(u.Email)
	state.Type = types.StringValue(u.Type)
	state.Locked = types.BoolValue(u.Locked)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *UserResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan userResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	u, err := r.client.UpdateUser(ctx, plan.ID.ValueString(), client.UserUpdateInput{
		Name:    plan.Name.ValueString(),
		Profile: plan.Profile.ValueString(),
		Email:   plan.Email.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Error updating user", err.Error())
		return
	}
	plan.Name = types.StringValue(u.Name)
	plan.Profile = types.StringValue(u.Profile)
	plan.Email = types.StringValue(u.Email)
	plan.Locked = types.BoolValue(u.Locked)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *UserResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state userResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteUser(ctx, state.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Error deleting user", err.Error())
	}
}

func (r *UserResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
